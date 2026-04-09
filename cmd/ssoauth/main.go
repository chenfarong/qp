package main

import (
	"fmt"
	"log"
	"net"
	"strconv"

	"os"

	authgrpc "github.com/aoyo/qp/internal/ssoauth/grpc"
	"github.com/aoyo/qp/internal/ssoauth/handler"
	"github.com/aoyo/qp/internal/ssoauth/service"
	"github.com/aoyo/qp/pkg/db"
	"github.com/aoyo/qp/pkg/etcd"
	"github.com/aoyo/qp/pkg/proto/auth"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
)

// Config 配置结构
type Config struct {
	Server struct {
		Ssoauth struct {
			Port int `yaml:"port"`
		} `yaml:"ssoauth"`
	} `yaml:"server"`
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Dbname   string `yaml:"dbname"`
	}
	Jwt struct {
		Secret      string `yaml:"secret"`
		ExpireHours int    `yaml:"expire_hours"`
	} `yaml:"jwt"`
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

	dbInstance, err := db.InitDB(uri)
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
	authService := service.NewAuthService(dbInstance, config.Jwt.Secret, config.Jwt.ExpireHours, config.Database.Dbname)
	authHandler := handler.NewAuthHandler(authService)

	// 注册服务到 etcd
	if etcdClient != nil {
		serviceAddress := fmt.Sprintf("localhost:%d", config.Server.Ssoauth.Port)
		if err := etcdClient.RegisterService("ssoauth", serviceAddress); err != nil {
			log.Printf("Warning: Failed to register service to etcd: %v", err)
		} else {
			log.Println("Service registered to etcd successfully")
		}
	}

	// 启动gRPC服务器
	go startGRPCServer(authService, config.Server.Ssoauth.Port+1000)

	// 初始化路由
	router := gin.Default()
	authHandler.RegisterRoutes(router)

	// 启动HTTP服务
	port := config.Server.Ssoauth.Port
	log.Printf("SSO Auth service starting on port %d...", port)
	if err := router.Run(":" + strconv.Itoa(port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// startGRPCServer 启动gRPC服务器
func startGRPCServer(authService *service.AuthService, port int) {
	// 创建gRPC服务器
	grpcServer := grpc.NewServer()

	// 注册认证服务
	auth.RegisterAuthServiceServer(grpcServer, authgrpc.NewAuthServer(authService))

	// 监听端口
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("SSO Auth gRPC service starting on port %d...", port)
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
