package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"os"

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
				if messageType == websocket.TextMessage {
					// 广播消息给所有客户端
					wsManager.broadcast <- message
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
