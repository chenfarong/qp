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
	"github.com/aoyo/qp/pkg/etcd"
	"github.com/aoyo/qp/pkg/proto/chat"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
)

// Config 配置结构
type Config struct {
	Server struct {
		Chat struct {
			Port int `yaml:"port"`
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

	// 初始化 etcd 客户端
	log.Println("Connecting to etcd...")
	etcdClient, err := etcd.NewClient(config.Etcd.Endpoints)
	if err != nil {
		log.Printf("Warning: Failed to connect to etcd: %v", err)
		log.Println("Continuing without etcd connection...")
	} else {
		log.Println("Etcd connected successfully")
		defer etcdClient.Close()
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

	// 启动gRPC服务器
	go startGRPCServer(chatService, config.Server.Chat.Port+1000)

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

	// 启动HTTP服务
	port := config.Server.Chat.Port
	log.Printf("Chat service starting on port %d...", port)
	if err := router.Run(":" + strconv.Itoa(port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
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
