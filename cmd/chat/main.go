package main

import (
	"fmt"
	"log"
	"net"
	"strconv"

	"os"

	chatgrpc "github.com/aoyo/qp/internal/chat/grpc"
	"github.com/aoyo/qp/internal/chat/handler"
	"github.com/aoyo/qp/internal/chat/service"
	"github.com/aoyo/qp/pkg/db"
	"github.com/aoyo/qp/pkg/envmode"
	"github.com/aoyo/qp/pkg/etcd"
	"github.com/aoyo/qp/pkg/proto/chat"
	"github.com/aoyo/qp/pkg/utils"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
)

// Config 配置结构
type Config struct {
	Sandbox bool `yaml:"sandbox"`
	Server  struct {
		Chat struct {
			Port     int `yaml:"port"`
			GrpcPort int `yaml:"grpc_port"`
		} `yaml:"chat"`
	} `yaml:"server"`
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Dbname   string `yaml:"dbname"`
	} `yaml:"database"`
	Etcd struct {
		Endpoints []string `yaml:"endpoints"`
	} `yaml:"etcd"`
}

func main() {
	// 加载配置
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库
	uri := fmt.Sprintf(
		"mongodb://%s:%s@%s:%d/%s?authSource=admin",
		config.Database.User,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.Dbname,
	)

	dbClient, err := db.InitDB(uri)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
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

	// 初始化服务
	chatService := service.NewChatService(dbClient, config.Database.Dbname)
	chatHandler := handler.NewChatHandler(chatService)

	// 注册服务到 etcd
	if etcdClient != nil {
		serviceAddress := fmt.Sprintf("localhost:%d", config.Server.Chat.Port)
		if err := etcdClient.RegisterService("chat", serviceAddress); err != nil {
			log.Printf("Warning: Failed to register service to etcd: %v", err)
		} else {
			log.Println("Service registered to etcd successfully")
		}
	}

	grpcPort := config.Server.Chat.GrpcPort
	if grpcPort == 0 {
		grpcPort = config.Server.Chat.Port + 1000
	}
	go startGRPCServer(chatService, grpcPort)

	// 初始化路由
	router := gin.Default()

	// 注册路由
	chatHandler.RegisterRoutes(router)

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// 打印欢迎日志
	printWelcomeLog("Chat", config.Server.Chat.Port, grpcPort, config.Database.Host, config.Database.Port, config.Database.Dbname)

	// 启动HTTP服务
	port := config.Server.Chat.Port
	log.Printf("Chat service starting on port %d...", port)
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
	log.Printf("🗄️  Database: %s:%d/%s", dbHost, dbPort, dbName)
	if gitInfo != nil {
		log.Printf("📝 Git Branch: %s", gitInfo.Branch)
		log.Printf("🔖 Git Commit: %s", gitInfo.CommitHash)
		log.Printf("💬 Git Message: %s", gitInfo.CommitMsg)
	}
	log.Println("===============================================================")
	log.Println("")
}

// startGRPCServer 启动gRPC服务器
func startGRPCServer(chatService *service.ChatService, port int) {
	// 创建gRPC服务器
	grpcServer := grpc.NewServer()

	// 注册聊天服务
	chat.RegisterChatServiceServer(grpcServer, chatgrpc.NewChatServer(chatService))

	// 监听端口
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("Chat gRPC service starting on port %d...", port)
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
