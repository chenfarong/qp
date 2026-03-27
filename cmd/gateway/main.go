package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"os"

	protoutil "github.com/aoyo/qp/pkg/proto"
	proto "github.com/aoyo/qp/proto"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gopkg.in/yaml.v3"
)

// Config 配置结构
type Config struct {
	Server struct {
		Gateway struct {
			Port int `yaml:"port"`
		} `yaml:"gateway"`
		Ssoauth struct {
			Port int `yaml:"port"`
		} `yaml:"ssoauth"`
		Gamelogic struct {
			Port int `yaml:"port"`
		} `yaml:"gamelogic"`
	} `yaml:"server"`
}

// WebSocket连接管理器
type WebSocketManager struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mutex      sync.Mutex
}

// NewWebSocketManager 创建WebSocket管理器
func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

// Run 运行WebSocket管理器
func (manager *WebSocketManager) Run() {
	for {
		select {
		case client := <-manager.register:
			manager.mutex.Lock()
			manager.clients[client] = true
			manager.mutex.Unlock()
		case client := <-manager.unregister:
			manager.mutex.Lock()
			if _, ok := manager.clients[client]; ok {
				delete(manager.clients, client)
				client.Close()
			}
			manager.mutex.Unlock()
		case message := <-manager.broadcast:
			manager.mutex.Lock()
			for client := range manager.clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					client.Close()
					delete(manager.clients, client)
				}
			}
			manager.mutex.Unlock()
		}
	}
}

// 升级HTTP连接为WebSocket连接
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源的WebSocket连接
	},
}

func main() {
	// 加载配置
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化WebSocket管理器
	wsManager := NewWebSocketManager()
	go wsManager.Run()

	// 初始化路由
	router := gin.Default()

	// 注册路由
	registerRoutes(router, config, wsManager)

	// 启动服务
	port := config.Server.Gateway.Port
	log.Printf("Gateway service starting on port %d...", port)
	if err := router.Run(":" + strconv.Itoa(port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// loadConfig 加载配置文件
func loadConfig() (*Config, error) {
	file, err := os.Open("configs/config.yaml")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// registerRoutes 注册路由
func registerRoutes(router *gin.Engine, config *Config, wsManager *WebSocketManager) {
	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// WebSocket路由
	router.GET("/ws", func(c *gin.Context) {
		// 升级HTTP连接为WebSocket连接
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println("Failed to upgrade to WebSocket:", err)
			return
		}

		// 注册新的WebSocket连接
		wsManager.register <- conn

		// 处理WebSocket消息
		go func() {
			defer func() {
				wsManager.unregister <- conn
			}()

			for {
				messageType, message, err := conn.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						log.Printf("WebSocket error: %v", err)
					}
					break
				}

				// 处理接收到的消息
				if messageType == websocket.BinaryMessage {
					// 反序列化protobuf消息
					msg, err := protoutil.Deserialize(message)
					if err != nil {
						log.Printf("Failed to deserialize message: %v", err)
						continue
					}

					// 处理消息
					response := handleMessage(msg, config)

					// 序列化响应消息
					responseData, err := protoutil.Serialize(response)
					if err != nil {
						log.Printf("Failed to serialize response: %v", err)
						continue
					}

					// 发送响应消息
					err = conn.WriteMessage(websocket.BinaryMessage, responseData)
					if err != nil {
						log.Printf("Failed to write message: %v", err)
						break
					}
				}
			}
		}()
	})

	// API路由组
	apiGroup := router.Group("/api")

	// 认证相关路由（转发到SsoAuth服务）
	authGroup := apiGroup.Group("/auth")
	{
		authGroup.POST("/register", proxyToService(fmt.Sprintf("http://localhost:%d", config.Server.Ssoauth.Port)))
		authGroup.POST("/login", proxyToService(fmt.Sprintf("http://localhost:%d", config.Server.Ssoauth.Port)))
		authGroup.GET("/validate", proxyToService(fmt.Sprintf("http://localhost:%d", config.Server.Ssoauth.Port)))
	}

	// 游戏相关路由（转发到GameLogic服务）
	gameGroup := apiGroup.Group("/game")
	{
		gameGroup.POST("/characters", proxyToService(fmt.Sprintf("http://localhost:%d", config.Server.Gamelogic.Port)))
		gameGroup.GET("/characters", proxyToService(fmt.Sprintf("http://localhost:%d", config.Server.Gamelogic.Port)))
		gameGroup.GET("/characters/:id", proxyToService(fmt.Sprintf("http://localhost:%d", config.Server.Gamelogic.Port)))
		gameGroup.PUT("/characters/:id/status", proxyToService(fmt.Sprintf("http://localhost:%d", config.Server.Gamelogic.Port)))
		gameGroup.POST("/battle", proxyToService(fmt.Sprintf("http://localhost:%d", config.Server.Gamelogic.Port)))
	}
}

// handleMessage 处理接收到的protobuf消息
func handleMessage(msg *proto.Message, config *Config) *proto.Message {
	// 根据消息类型处理
	switch msg.Type {
	case proto.MessageType_MSG_TYPE_AUTH_REGISTER:
		// 处理注册请求
		if req := msg.GetAuthRegister(); req != nil {
			// 构建请求数据
			data, err := json.Marshal(req)
			if err != nil {
				return createErrorResponse(500, "Failed to marshal request")
			}

			// 发送请求到SsoAuth服务
			resp, err := http.Post(fmt.Sprintf("http://localhost:%d/api/auth/register", config.Server.Ssoauth.Port), "application/json", bytes.NewBuffer(data))
			if err != nil {
				return createErrorResponse(500, "Failed to send request")
			}
			defer resp.Body.Close()

			// 解析响应
			var authResp struct {
				Token    string `json:"token"`
				UserID   string `json:"user_id"`
				Username string `json:"username"`
				Nickname string `json:"nickname"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
				return createErrorResponse(500, "Failed to decode response")
			}

			// 构建响应消息
			return &proto.Message{
				Type: proto.MessageType_MSG_TYPE_RESPONSE,
				Data: &proto.Message_Response{
					Response: &proto.Response{
						Code:    200,
						Message: "success",
						Data: &proto.Response_AuthResponse{
							AuthResponse: &proto.AuthResponse{
								Token:    authResp.Token,
								UserID:   authResp.UserID,
								Username: authResp.Username,
								Nickname: authResp.Nickname,
							},
						},
					},
				},
			}
		}
	case proto.MessageType_MSG_TYPE_AUTH_LOGIN:
		// 处理登录请求
		if req := msg.GetAuthLogin(); req != nil {
			// 构建请求数据
			data, err := json.Marshal(req)
			if err != nil {
				return createErrorResponse(500, "Failed to marshal request")
			}

			// 发送请求到SsoAuth服务
			resp, err := http.Post(fmt.Sprintf("http://localhost:%d/api/auth/login", config.Server.Ssoauth.Port), "application/json", bytes.NewBuffer(data))
			if err != nil {
				return createErrorResponse(500, "Failed to send request")
			}
			defer resp.Body.Close()

			// 解析响应
			var authResp struct {
				Token    string `json:"token"`
				UserID   string `json:"user_id"`
				Username string `json:"username"`
				Nickname string `json:"nickname"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
				return createErrorResponse(500, "Failed to decode response")
			}

			// 构建响应消息
			return &proto.Message{
				Type: proto.MessageType_MSG_TYPE_RESPONSE,
				Data: &proto.Message_Response{
					Response: &proto.Response{
						Code:    200,
						Message: "success",
						Data: &proto.Response_AuthResponse{
							AuthResponse: &proto.AuthResponse{
								Token:    authResp.Token,
								UserID:   authResp.UserID,
								Username: authResp.Username,
								Nickname: authResp.Nickname,
							},
						},
					},
				},
			}
		}
	// 其他消息类型的处理...
	default:
		return createErrorResponse(400, "Unknown message type")
	}

	// 默认返回错误响应
	return createErrorResponse(400, "Invalid message")
}

// createErrorResponse 创建错误响应
func createErrorResponse(code int32, message string) *proto.Message {
	return &proto.Message{
		Type: proto.MessageType_MSG_TYPE_RESPONSE,
		Data: &proto.Message_Response{
			Response: &proto.Response{
				Code:    code,
				Message: message,
			},
		},
	}
}

// proxyToService 代理到指定服务
func proxyToService(targetURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 构建目标URL
		target := fmt.Sprintf("%s%s", targetURL, c.Request.URL.Path)
		if c.Request.URL.RawQuery != "" {
			target += "?" + c.Request.URL.RawQuery
		}

		// 创建新的请求
		req, err := http.NewRequest(c.Request.Method, target, c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return
		}

		// 复制请求头
		for key, values := range c.Request.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		// 发送请求
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to proxy request"})
			return
		}
		defer resp.Body.Close()

		// 复制响应头
		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}

		// 设置响应状态码
		c.Status(resp.StatusCode)

		// 复制响应体
		c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
	}
}
