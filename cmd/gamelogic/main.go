package main

import (
	"fmt"
	"log"
	"net"
	"strconv"

	"os"

	"github.com/aoyo/qp/internal/gamelogic/grpc"
	"github.com/aoyo/qp/internal/gamelogic/handler"
	"github.com/aoyo/qp/internal/gamelogic/service"
	"github.com/aoyo/qp/pkg/db"
	"github.com/aoyo/qp/pkg/proto/game"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
)

// Config 配置结构
type Config struct {
	Server struct {
		Gamelogic struct {
			Port int `yaml:"port"`
		} `yaml:"gamelogic"`
	} `yaml:"server"`
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Dbname   string `yaml:"dbname"`
	} `yaml:"database"`
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

	dbInstance, err := db.InitDB(uri)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 初始化服务
	gameService := service.NewGameService(dbInstance, config.Database.Dbname)
	gameHandler := handler.NewGameHandler(gameService)

	// 启动gRPC服务器
	go startGRPCServer(gameService, config.Server.Gamelogic.Port+1000)

	// 初始化路由
	router := gin.Default()
	gameHandler.RegisterRoutes(router)

	// 启动HTTP服务
	port := config.Server.Gamelogic.Port
	log.Printf("Game Logic service starting on port %d...", port)
	if err := router.Run(":" + strconv.Itoa(port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// startGRPCServer 启动gRPC服务器
func startGRPCServer(gameService *service.GameService, port int) {
	// 创建gRPC服务器
	grpcServer := grpc.NewServer()

	// 注册游戏服务
	game.RegisterGameServiceServer(grpcServer, grpc.NewGameServer(gameService))

	// 监听端口
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("Game Logic gRPC service starting on port %d...", port)
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
