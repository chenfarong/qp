package main

import (
	"fmt"
	"log"
	"net"

	"os"

	billgrpc "github.com/aoyo/qp/internal/bill/grpc"
	"github.com/aoyo/qp/internal/bill/handler"
	"github.com/aoyo/qp/internal/bill/service"
	"github.com/aoyo/qp/pkg/db"
	"github.com/aoyo/qp/pkg/etcd"
	"github.com/aoyo/qp/pkg/proto/bill"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
)

// Config 配置结构
type Config struct {
	Server struct {
		Bill struct {
			Port int `yaml:"port"`
		} `yaml:"bill"`
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

	host := "localhost"
	port := config.Server.Bill.Port
	grpcPort := port + 1000
	dbURI := fmt.Sprintf(
		"mongodb://%s:%s@%s:%d/%s?authSource=admin",
		config.Database.User,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.Dbname,
	)
	dbName := config.Database.Dbname

	// 初始化数据库连接
	dbInstance, err := db.InitDB(dbURI)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbInstance.Close()

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
	tokenService := service.NewTokenService(dbInstance, dbName)
	paymentService := service.NewPaymentService(dbInstance, dbName, tokenService)

	// 注册服务到 etcd
	if etcdClient != nil {
		serviceAddress := fmt.Sprintf("localhost:%d", port)
		if err := etcdClient.RegisterService("bill", serviceAddress); err != nil {
			log.Printf("Warning: Failed to register service to etcd: %v", err)
		} else {
			log.Println("Service registered to etcd successfully")
		}
	}

	// 启动gRPC服务器
	go startGRPCServer(paymentService, grpcPort)

	// 初始化处理器
	billHandler := handler.NewBillHandler(tokenService, paymentService)

	// 设置路由
	router := gin.Default()
	billHandler.RegisterRoutes(router)

	// 启动HTTP服务器
	serverAddr := fmt.Sprintf("%s:%d", host, port)
	log.Printf("Bill server starting on %s", serverAddr)
	if err := router.Run(serverAddr); err != nil {
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

// startGRPCServer 启动gRPC服务器
func startGRPCServer(paymentService *service.PaymentService, port int) {
	// 创建gRPC服务器
	grpcServer := grpc.NewServer()

	// 注册账单服务
	bill.RegisterBillServiceServer(grpcServer, billgrpc.NewBillServer(paymentService))

	// 监听端口
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("Bill gRPC service starting on port %d...", port)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}
