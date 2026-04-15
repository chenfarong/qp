package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"runtime/debug"
	"strconv"
	"sync"

	"os"

	"github.com/aoyo/qp/internal/gateway/grpc"
	"github.com/aoyo/qp/pkg/envmode"
	"github.com/aoyo/qp/pkg/etcd"
	"github.com/aoyo/qp/pkg/logger"
	"github.com/aoyo/qp/pkg/proto/game"
	"github.com/aoyo/qp/pkg/proto/gateway"
	"github.com/aoyo/qp/pkg/proto/proto"
	"github.com/aoyo/qp/pkg/protocol"
	"github.com/aoyo/qp/pkg/router"
	"github.com/aoyo/qp/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	grpc_lib "google.golang.org/grpc"
	pb "google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"
)

// protoutil 提供protobuf序列化和反序列化功能
var protoutil = struct {
	Serialize   func(*proto.Message) ([]byte, error)
	Deserialize func([]byte) (*proto.Message, error)
}{
	Serialize: func(msg *proto.Message) ([]byte, error) {
		return pb.Marshal(msg)
	},
	Deserialize: func(data []byte) (*proto.Message, error) {
		msg := &proto.Message{}
		err := pb.Unmarshal(data, msg)
		return msg, err
	},
}

// Config 配置结构
type Config struct {
	Sandbox bool `yaml:"sandbox"`
	Server  struct {
		Gateway struct {
			Port     int `yaml:"port"`
			GrpcPort int `yaml:"grpc_port"`
		} `yaml:"gateway"`
		Ssoauth struct {
			Port int `yaml:"port"`
		} `yaml:"ssoauth"`
		Gamelogic struct {
			Port     int `yaml:"port"`
			GrpcPort int `yaml:"grpc_port"`
		} `yaml:"gamelogic"`
		Chat struct {
			Port int `yaml:"port"`
		} `yaml:"chat"`
		Logger struct {
			UdpPort int `yaml:"udp_port"`
		} `yaml:"logger"`
	} `yaml:"server"`
	Etcd struct {
		Endpoints []string `yaml:"endpoints"`
	} `yaml:"etcd"`
}

// WebSocket连接管理器
type WebSocketManager struct {
	clients    map[*websocket.Conn]bool
	userConns  map[uint32][]*websocket.Conn // 用户ID到连接的映射
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mutex      sync.RWMutex
}

// NewWebSocketManager 创建WebSocket管理器
func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		clients:    make(map[*websocket.Conn]bool),
		userConns:  make(map[uint32][]*websocket.Conn),
		register:   make(chan *websocket.Conn, 100),
		unregister: make(chan *websocket.Conn, 100),
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
			log.Printf("Client connected. Total clients: %d", len(manager.clients))
		case client := <-manager.unregister:
			manager.mutex.Lock()
			if _, ok := manager.clients[client]; ok {
				delete(manager.clients, client)
				client.Close()
			}
			manager.mutex.Unlock()
			log.Printf("Client disconnected. Total clients: %d", len(manager.clients))
		}
	}
}

// GetClientCount 获取当前连接数
func (manager *WebSocketManager) GetClientCount() int {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	return len(manager.clients)
}

// RegisterUserConn 注册用户连接
func (manager *WebSocketManager) RegisterUserConn(userID uint32, conn *websocket.Conn) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	manager.userConns[userID] = append(manager.userConns[userID], conn)
}

// UnregisterUserConn 注销用户连接
func (manager *WebSocketManager) UnregisterUserConn(conn *websocket.Conn) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	// 从所有用户的连接列表中移除该连接
	for userID, conns := range manager.userConns {
		newConns := []*websocket.Conn{}
		for _, c := range conns {
			if c != conn {
				newConns = append(newConns, c)
			}
		}
		if len(newConns) > 0 {
			manager.userConns[userID] = newConns
		} else {
			delete(manager.userConns, userID)
		}
	}
}

// PushMessage 推送消息给指定用户
func (manager *WebSocketManager) PushMessage(userID uint32, messageType string, messageData []byte, title, content string) (bool, error) {
	manager.mutex.RLock()
	conns, ok := manager.userConns[userID]
	manager.mutex.RUnlock()

	if !ok || len(conns) == 0 {
		return false, nil
	}

	// 编码协议数据包
	encodedData, err := protocol.Encode(
		protocol.MessageTypeNotify,
		protocol.CompressFlagNone,
		0, // 通知消息使用0作为消息ID
		messageData,
	)
	if err != nil {
		return false, err
	}

	success := false
	for _, conn := range conns {
		err := conn.WriteMessage(websocket.BinaryMessage, encodedData)
		if err == nil {
			success = true
		}
	}

	return success, nil
}

// BroadcastMessage 广播消息给所有用户
func (manager *WebSocketManager) BroadcastMessage(messageType string, messageData []byte, title, content string) (int, error) {
	manager.mutex.RLock()
	clients := make([]*websocket.Conn, 0, len(manager.clients))
	for client := range manager.clients {
		clients = append(clients, client)
	}
	manager.mutex.RUnlock()

	// 编码协议数据包
	encodedData, err := protocol.Encode(
		protocol.MessageTypeNotify,
		protocol.CompressFlagNone,
		0, // 通知消息使用0作为消息ID
		messageData,
	)
	if err != nil {
		return 0, err
	}

	broadcastCount := 0
	for _, client := range clients {
		err := client.WriteMessage(websocket.BinaryMessage, encodedData)
		if err == nil {
			broadcastCount++
		}
	}

	return broadcastCount, nil
}

// GetConnectedUsers 获取当前连接的用户列表
func (manager *WebSocketManager) GetConnectedUsers() []uint32 {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	userIDs := make([]uint32, 0, len(manager.userConns))
	for userID := range manager.userConns {
		userIDs = append(userIDs, userID)
	}

	return userIDs
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

	var etcdClient *etcd.Client
	if envmode.UseEtcd(config.Sandbox) {
		log.Printf("%s：连接 etcd", envmode.SandboxLabel(config.Sandbox))
		var errEtcd error
		etcdClient, errEtcd = etcd.NewClient(config.Etcd.Endpoints)
		if errEtcd != nil {
			log.Printf("Warning: Failed to connect to etcd: %v", errEtcd)
			log.Println("Continuing without etcd connection...")
		} else {
			defer etcdClient.Close()
		}
	} else {
		log.Printf("%s：跳过 etcd", envmode.SandboxLabel(config.Sandbox))
	}

	// 初始化日志客户端
	logClient, err := logger.NewClient(fmt.Sprintf("localhost:%d", config.Server.Logger.UdpPort), fmt.Sprintf("gateway://localhost:%d", config.Server.Gateway.Port))
	if err != nil {
		log.Printf("Warning: Failed to initialize log client: %v", err)
	} else {
		defer logClient.Close()
		// 发送测试日志
		logClient.Warn("Gateway server starting")
	}

	// 注册服务到 etcd
	if etcdClient != nil {
		serviceAddress := fmt.Sprintf("localhost:%d", config.Server.Gateway.Port)
		if err := etcdClient.RegisterService("gateway", serviceAddress); err != nil {
			log.Printf("Warning: Failed to register service to etcd: %v", err)
			if logClient != nil {
				logClient.Warn(fmt.Sprintf("Failed to register service to etcd: %v", err))
			}
		} else {
			log.Println("Service registered to etcd successfully")
			if logClient != nil {
				logClient.Warn("Service registered to etcd successfully")
			}
		}
	}

	// 初始化WebSocket管理器
	wsManager := NewWebSocketManager()
	go wsManager.Run()

	// 初始化路由器
	r := setupRouter()

	// 启动gRPC服务器
	grpcPort := config.Server.Gateway.GrpcPort
	if grpcPort == 0 {
		grpcPort = config.Server.Gateway.Port + 1000
	}
	go func() {
		defer recoverPanic(logClient, "Gateway gRPC")
		startGRPCServer(wsManager, grpcPort)
	}()

	// 初始化gamelogic gRPC客户端
	gamelogicGrpcPort := config.Server.Gamelogic.GrpcPort
	if gamelogicGrpcPort == 0 {
		gamelogicGrpcPort = config.Server.Gamelogic.Port + 1000
	}
	gamelogicConn, err := grpc_lib.Dial(fmt.Sprintf("localhost:%d", gamelogicGrpcPort), grpc_lib.WithInsecure())
	if err != nil {
		log.Printf("Warning: Failed to connect to gamelogic gRPC: %v", err)
	} else {
		gamelogicClient = game.NewGameServiceClient(gamelogicConn)
		log.Println("Connected to gamelogic gRPC server successfully")
	}

	// 初始化路由
	router := gin.Default()

	// 注册路由
	registerRoutes(router, config, wsManager, r)

	// 打印欢迎日志
	printWelcomeLog("Gateway", config.Server.Gateway.Port, grpcPort, "", 0, "")

	// 启动HTTP/WebSocket服务
	port := config.Server.Gateway.Port
	log.Printf("Gateway service starting on port %d...", port)

	// 主服务器启动，添加panic recovery
	defer recoverPanic(logClient, "Gateway HTTP")
	if err := router.Run(":" + strconv.Itoa(port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// printWelcomeLog 打印欢迎日志
func printWelcomeLog(serverType string, httpPort, grpcPort int, dbHost string, dbPort int, dbName string) {
	// 获取git信息
	gitInfo, err := utils.GetGitInfo()
	if err != nil {
		log.Printf("Warning: Failed to get git info: %v", err)
	}

	// 打印欢迎日志
	log.Println("")
	log.Println("===============================================================")
	log.Printf("🎉 %s Server Welcome! 🎉", serverType)
	log.Println("===============================================================")
	log.Printf("🌐 Server Type: %s", serverType)
	log.Printf("🚪 HTTP Port: %d", httpPort)
	log.Printf("🔗 gRPC Port: %d", grpcPort)
	if dbHost != "" {
		log.Printf("🗄️  Database: %s:%d/%s", dbHost, dbPort, dbName)
	}
	if gitInfo != nil {
		log.Printf("📝 Git Branch: %s", gitInfo.Branch)
		log.Printf("🔖 Git Commit: %s", gitInfo.CommitHash)
		log.Printf("💬 Git Message: %s", gitInfo.CommitMsg)
	}
	log.Println("===============================================================")
	log.Println("")
}

// recoverPanic 恢复panic并打印错误信息和调用堆栈
func recoverPanic(logClient *logger.Client, serverName string) {
	if r := recover(); r != nil {
		// 捕获panic信息
		panicMsg := fmt.Sprintf("Panic recovered in %s server: %v\n%s", serverName, r, string(debug.Stack()))

		// 打印到控制台
		log.Printf("ERROR: %s", panicMsg)

		// 发送到日志服务器
		if logClient != nil {
			logClient.Error(panicMsg)
		}
	}
}

// startGRPCServer 启动gRPC服务器
func startGRPCServer(wsManager *WebSocketManager, port int) {
	// 创建gRPC服务器
	grpcServer := grpc_lib.NewServer()

	// 注册网关服务
	gateway.RegisterGatewayServiceServer(grpcServer, grpc.NewGatewayServer(wsManager))

	// 监听端口
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("Gateway gRPC service starting on port %d...", port)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
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
func registerRoutes(router *gin.Engine, config *Config, wsManager *WebSocketManager, r *router.Router) {
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

		// 为每个客户端连接分配一个独立的goroutine处理消息
		// 这个goroutine会持续运行，直到连接关闭
		go handleClientConnection(conn, wsManager, config, r)
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
		gameGroup.POST("/characters/use", proxyToService(fmt.Sprintf("http://localhost:%d", config.Server.Gamelogic.Port)))
		gameGroup.POST("/battle", proxyToService(fmt.Sprintf("http://localhost:%d", config.Server.Gamelogic.Port)))
	}

	// 聊天相关路由（转发到Chat服务）
	chatGroup := apiGroup.Group("/chat")
	{
		chatGroup.POST("/messages", proxyToService(fmt.Sprintf("http://localhost:%d", config.Server.Chat.Port)))
		chatGroup.POST("/messages/history", proxyToService(fmt.Sprintf("http://localhost:%d", config.Server.Chat.Port)))
		chatGroup.POST("/conversations", proxyToService(fmt.Sprintf("http://localhost:%d", config.Server.Chat.Port)))
		chatGroup.POST("/messages/status", proxyToService(fmt.Sprintf("http://localhost:%d", config.Server.Chat.Port)))
	}
}

// messageToJSON 将Message转换为JSON格式
func messageToJSON(msg *proto.Message) string {
	// 创建一个map来存储消息内容
	msgMap := make(map[string]interface{})
	msgMap["type"] = msg.Type.String()

	// 根据消息类型处理Data字段
	switch msg.Type {
	case proto.MessageType_MSG_TYPE_AUTH_REGISTER:
		if req := msg.GetAuthRegister(); req != nil {
			reqMap := make(map[string]interface{})
			reqMap["username"] = req.Username
			reqMap["password"] = "******" // 密码脱敏
			reqMap["email"] = req.Email
			reqMap["nickname"] = req.Nickname
			msgMap["data"] = reqMap
		}
	case proto.MessageType_MSG_TYPE_AUTH_LOGIN:
		if req := msg.GetAuthLogin(); req != nil {
			reqMap := make(map[string]interface{})
			reqMap["username"] = req.Username
			reqMap["password"] = "******" // 密码脱敏
			msgMap["data"] = reqMap
		}
	case proto.MessageType_MSG_TYPE_AUTH_VALIDATE:
		if req := msg.GetAuthValidate(); req != nil {
			reqMap := make(map[string]interface{})
			reqMap["token"] = req.Token
			msgMap["data"] = reqMap
		}
	case proto.MessageType_MSG_TYPE_GAME_CREATE_CHARACTER:
		if req := msg.GetGameCreateCharacter(); req != nil {
			reqMap := make(map[string]interface{})
			reqMap["user_id"] = req.UserId
			reqMap["name"] = req.Name
			msgMap["data"] = reqMap
		}
	case proto.MessageType_MSG_TYPE_GAME_GET_CHARACTERS:
		if req := msg.GetGameGetCharacters(); req != nil {
			reqMap := make(map[string]interface{})
			reqMap["user_id"] = req.UserId
			msgMap["data"] = reqMap
		}
	case proto.MessageType_MSG_TYPE_GAME_GET_CHARACTER:
		if req := msg.GetGameGetCharacter(); req != nil {
			reqMap := make(map[string]interface{})
			reqMap["character_id"] = req.CharacterId
			msgMap["data"] = reqMap
		}
	case proto.MessageType_MSG_TYPE_GAME_UPDATE_CHARACTER_STATUS:
		if req := msg.GetGameUpdateCharacterStatus(); req != nil {
			reqMap := make(map[string]interface{})
			reqMap["character_id"] = req.CharacterId
			reqMap["status"] = req.Status
			msgMap["data"] = reqMap
		}
	case proto.MessageType_MSG_TYPE_GAME_BATTLE:
		if req := msg.GetGameBattle(); req != nil {
			reqMap := make(map[string]interface{})
			reqMap["character_id"] = req.CharacterId
			reqMap["enemy_level"] = req.EnemyLevel
			msgMap["data"] = reqMap
		}
	case proto.MessageType_MSG_TYPE_BILL_CREATE_PAYMENT:
		if req := msg.GetBillCreatePayment(); req != nil {
			reqMap := make(map[string]interface{})
			reqMap["user_id"] = req.UserId
			reqMap["product_id"] = req.ProductId
			reqMap["amount"] = req.Amount
			reqMap["payment_method"] = req.PaymentMethod
			msgMap["data"] = reqMap
		}
	case proto.MessageType_MSG_TYPE_BILL_GET_PAYMENT:
		if req := msg.GetBillGetPayment(); req != nil {
			reqMap := make(map[string]interface{})
			reqMap["payment_id"] = req.PaymentId
			msgMap["data"] = reqMap
		}
	case proto.MessageType_MSG_TYPE_RESPONSE:
		if resp := msg.GetResponse(); resp != nil {
			respMap := make(map[string]interface{})
			respMap["code"] = resp.Code
			respMap["message"] = resp.Message

			// 处理响应数据
			switch {
			case resp.GetAuthResponse() != nil:
				authResp := resp.GetAuthResponse()
				authRespMap := make(map[string]interface{})
				authRespMap["token"] = authResp.Token
				authRespMap["user_id"] = authResp.UserId
				authRespMap["username"] = authResp.Username
				authRespMap["nickname"] = authResp.Nickname
				respMap["data"] = authRespMap
			case resp.GetGameResponse() != nil:
				gameResp := resp.GetGameResponse()
				gameRespMap := make(map[string]interface{})
				switch {
				case gameResp.GetCharacter() != nil:
					char := gameResp.GetCharacter()
					charMap := make(map[string]interface{})
					charMap["id"] = char.Id
					charMap["user_id"] = char.UserId
					charMap["name"] = char.Name
					charMap["level"] = char.Level
					charMap["exp"] = char.Exp
					charMap["hp"] = char.Hp
					charMap["mp"] = char.Mp
					charMap["strength"] = char.Strength
					charMap["agility"] = char.Agility
					charMap["intelligence"] = char.Intelligence
					charMap["gold"] = char.Gold
					charMap["status"] = char.Status
					gameRespMap["character"] = charMap
				case gameResp.GetCharacters() != nil:
					chars := gameResp.GetCharacters()
					characters := make([]interface{}, 0, len(chars.Items))
					for _, char := range chars.Items {
						charMap := make(map[string]interface{})
						charMap["id"] = char.Id
						charMap["user_id"] = char.UserId
						charMap["name"] = char.Name
						charMap["level"] = char.Level
						charMap["exp"] = char.Exp
						charMap["hp"] = char.Hp
						charMap["mp"] = char.Mp
						charMap["strength"] = char.Strength
						charMap["agility"] = char.Agility
						charMap["intelligence"] = char.Intelligence
						charMap["gold"] = char.Gold
						charMap["status"] = char.Status
						characters = append(characters, charMap)
					}
					gameRespMap["characters"] = characters
				case gameResp.GetBattleResult() != nil:
					battleResult := gameResp.GetBattleResult()
					battleMap := make(map[string]interface{})
					battleMap["victory"] = battleResult.Victory
					battleMap["exp_gained"] = battleResult.ExpGained
					battleMap["gold_gained"] = battleResult.GoldGained
					battleMap["message"] = battleResult.Message
					gameRespMap["battle_result"] = battleMap
				case gameResp.GetBagItems() != nil:
					bagItems := gameResp.GetBagItems()
					items := make([]interface{}, 0, len(bagItems.Items))
					for _, item := range bagItems.Items {
						itemMap := make(map[string]interface{})
						itemMap["item_id"] = item.ItemId
						itemMap["item_cfg_id"] = item.ItemCfgId
						itemMap["num"] = item.Num
						items = append(items, itemMap)
					}
					gameRespMap["bag_items"] = items
				}
				respMap["data"] = gameRespMap
			case resp.GetBillResponse() != nil:
				billResp := resp.GetBillResponse()
				billRespMap := make(map[string]interface{})
				switch {
				case billResp.GetPayment() != nil:
					payment := billResp.GetPayment()
					paymentMap := make(map[string]interface{})
					paymentMap["id"] = payment.Id
					paymentMap["user_id"] = payment.UserId
					paymentMap["product_id"] = payment.ProductId
					paymentMap["amount"] = payment.Amount
					paymentMap["payment_method"] = payment.PaymentMethod
					paymentMap["status"] = payment.Status
					paymentMap["transaction_id"] = payment.TransactionId
					paymentMap["created_at"] = payment.CreatedAt
					paymentMap["updated_at"] = payment.UpdatedAt
					billRespMap["payment"] = paymentMap
				case billResp.GetPayments() != nil:
					payments := billResp.GetPayments()
					paymentList := make([]interface{}, 0, len(payments.Items))
					for _, payment := range payments.Items {
						paymentMap := make(map[string]interface{})
						paymentMap["id"] = payment.Id
						paymentMap["user_id"] = payment.UserId
						paymentMap["product_id"] = payment.ProductId
						paymentMap["amount"] = payment.Amount
						paymentMap["payment_method"] = payment.PaymentMethod
						paymentMap["status"] = payment.Status
						paymentMap["transaction_id"] = payment.TransactionId
						paymentMap["created_at"] = payment.CreatedAt
						paymentMap["updated_at"] = payment.UpdatedAt
						paymentList = append(paymentList, paymentMap)
					}
					billRespMap["payments"] = paymentList
				case billResp.GetTotal() != 0:
					billRespMap["total"] = billResp.GetTotal()
				}
				respMap["data"] = billRespMap
			case resp.GetChatResponse() != nil:
				chatResp := resp.GetChatResponse()
				chatRespMap := make(map[string]interface{})
				switch {
				case chatResp.GetChatMessage() != nil:
					// 这里应该是获取ChatMessage，但是我们的protobuf定义中可能没有这个字段
					// 暂时跳过
				case chatResp.GetChatMessages() != nil:
					// 这里应该是获取ChatMessages，但是我们的protobuf定义中可能没有这个字段
					// 暂时跳过
				case chatResp.GetConversations() != nil:
					conversations := chatResp.GetConversations()
					conversationList := make([]interface{}, 0, len(conversations.Items))
					for _, conversation := range conversations.Items {
						conversationMap := make(map[string]interface{})
						conversationMap["id"] = conversation.Id
						conversationMap["user_ids"] = conversation.UserIds
						conversationMap["last_message"] = conversation.LastMessage
						conversationMap["last_message_time"] = conversation.LastMessageTime
						conversationMap["unread_count"] = conversation.UnreadCount
						conversationMap["created_at"] = conversation.CreatedAt
						conversationMap["updated_at"] = conversation.UpdatedAt
						conversationList = append(conversationList, conversationMap)
					}
					chatRespMap["conversations"] = conversationList
				}
				respMap["data"] = chatRespMap
			}

			msgMap["data"] = respMap
		}
	default:
		msgMap["data"] = "Unknown message type"
	}

	// 转换为JSON
	jsonData, err := json.MarshalIndent(msgMap, "", "  ")
	if err != nil {
		return fmt.Sprintf("{\"error\": \"Failed to marshal message: %v\"}", err)
	}

	return string(jsonData)
}

// handleMessage 处理接收到的protobuf消息
func handleMessage(msg *proto.Message, config *Config, r *router.Router) *proto.Message {
	// 使用路由器处理消息
	return r.Handle(msg)
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

// handleClientConnection 处理客户端连接
// 每个客户端连接分配一个独立的goroutine，持续运行直到连接关闭
func handleClientConnection(conn *websocket.Conn, wsManager *WebSocketManager, config *Config, r *router.Router) {
	// 添加panic recovery
	defer func() {
		if r := recover(); r != nil {
			// 捕获panic信息
			panicMsg := fmt.Sprintf("Panic recovered in WebSocket connection: %v\n%s", r, string(debug.Stack()))

			// 打印到控制台
			log.Printf("ERROR: %s", panicMsg)
		}
	}()

	var userID uint32

	// 连接建立时发送欢迎消息
	welcomeMsg := map[string]string{
		"type":    "welcome",
		"message": "WebSocket connection established successfully!",
	}
	welcomeData, _ := json.Marshal(welcomeMsg)
	if err := conn.WriteMessage(websocket.TextMessage, welcomeData); err != nil {
		log.Printf("Failed to send welcome message: %v", err)
	}

	// 连接关闭时注销
	defer func() {
		wsManager.unregister <- conn
		wsManager.UnregisterUserConn(conn)
	}()

	// 持续读取和处理消息
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// 处理接收到的消息
		switch messageType {
		case websocket.BinaryMessage:
			// 解码协议数据包
			packet, err := protocol.Decode(bytes.NewReader(message))
			if err != nil {
				log.Printf("Failed to decode protocol packet: %v", err)
				continue
			}

			// 反序列化protobuf消息
			msg, err := protoutil.Deserialize(packet.Data)
			if err != nil {
				log.Printf("Failed to deserialize message: %v", err)
				continue
			}

			// 转换为JSON并打印日志
			jsonMsg := messageToJSON(msg)
			log.Printf("Received protobuf message (Type: %c, Compress: %c, ID: %d): %s",
				packet.MessageType, packet.CompressFlag, packet.MessageID, jsonMsg)

			// 处理消息
			response := handleMessage(msg, config, r)

			// 序列化响应消息
			responseData, err := protoutil.Serialize(response)
			if err != nil {
				log.Printf("Failed to serialize response: %v", err)
				continue
			}

			// 编码协议数据包
			encodedData, err := protocol.Encode(
				protocol.MessageTypeResponse,
				protocol.CompressFlagNone,
				packet.MessageID, // 使用相同的消息ID作为响应
				responseData,
			)
			if err != nil {
				log.Printf("Failed to encode protocol packet: %v", err)
				continue
			}

			// 发送响应消息
			err = conn.WriteMessage(websocket.BinaryMessage, encodedData)
			if err != nil {
				log.Printf("Failed to write message: %v", err)
				break
			}

			// 从响应中提取用户ID并注册连接
			if response.Type == proto.MessageType_MSG_TYPE_RESPONSE {
				if resp := response.GetResponse(); resp != nil {
					if authResp := resp.GetAuthResponse(); authResp != nil {
						// 解析用户ID
						userIDStr := authResp.UserId
						if userIDStr != "" {
							userIDInt, err := strconv.ParseUint(userIDStr, 10, 32)
							if err == nil {
								userID = uint32(userIDInt)
								// 注册用户连接
								wsManager.RegisterUserConn(userID, conn)
								log.Printf("User %d connected via WebSocket", userID)
							}
						}
					}
				}
			}
		case websocket.TextMessage:
			// 处理文本消息
			log.Printf("Received text message: %s", string(message))

			// 解析文本消息
			var textMsg map[string]interface{}
			if err := json.Unmarshal(message, &textMsg); err != nil {
				log.Printf("Failed to parse text message: %v", err)
				continue
			}

			// 发送确认消息
			ackMsg := map[string]interface{}{
				"type":    "ack",
				"message": "Message received",
				"data":    textMsg,
			}
			ackData, _ := json.Marshal(ackMsg)
			if err := conn.WriteMessage(websocket.TextMessage, ackData); err != nil {
				log.Printf("Failed to send ack message: %v", err)
			}
		}
	}
}
