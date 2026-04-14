package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"os"

	authgrpc "github.com/aoyo/qp/internal/ssoauth/grpc"
	"github.com/aoyo/qp/internal/ssoauth/handler"
	"github.com/aoyo/qp/internal/ssoauth/service"
	"github.com/aoyo/qp/pkg/db"
	"github.com/aoyo/qp/pkg/envmode"
	"github.com/aoyo/qp/pkg/etcd"
	"github.com/aoyo/qp/pkg/proto/auth"
	"github.com/aoyo/qp/pkg/proto/gateway"
	"github.com/aoyo/qp/pkg/utils"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	grpc_lib "google.golang.org/grpc"
	"gopkg.in/yaml.v3"
)

// Config 配置结构
type Config struct {
	Sandbox bool `yaml:"sandbox"`
	Server  struct {
		Ssoauth struct {
			Port     int `yaml:"port"`
			GrpcPort int `yaml:"grpc_port"`
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
	var uri string
	if config.Database.User != "" && config.Database.Password != "" {
		uri = fmt.Sprintf(
			"mongodb://%s:%s@%s:%d/%s?authSource=admin",
			config.Database.User,
			config.Database.Password,
			config.Database.Host,
			config.Database.Port,
			config.Database.Dbname,
		)
	} else {
		uri = fmt.Sprintf(
			"mongodb://%s:%d/%s",
			config.Database.Host,
			config.Database.Port,
			config.Database.Dbname,
		)
	}

	dbInstance, err := db.InitDB(uri)
	if err != nil {
		log.Printf("Warning: Failed to connect to database: %v", err)
		log.Println("Continuing without database connection...")
	} else {
		defer dbInstance.Close()
	}

	// 仅 sandbox 跳过 etcd；缺省或错误配置视为生产并连接 etcd
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

	grpcPort := config.Server.Ssoauth.GrpcPort
	if grpcPort == 0 {
		grpcPort = config.Server.Ssoauth.Port + 1000
	}
	go startGRPCServer(authService, grpcPort)

	// 初始化路由
	router := gin.Default()
	authHandler.RegisterRoutes(router)

	// 打印欢迎日志
	printWelcomeLog("SSO Auth", config.Server.Ssoauth.Port, grpcPort, config.Database.Host, config.Database.Port, config.Database.Dbname)

	// 向gateway注册协议编号段
	go registerToGateway("ssoauth", fmt.Sprintf("localhost:%d", config.Server.Ssoauth.Port), 1, 100)

	// 启动HTTP服务
	port := config.Server.Ssoauth.Port
	log.Printf("SSO Auth service starting on port %d...", port)
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

// registerToGateway 向gateway注册协议编号段
func registerToGateway(serviceName string, serviceAddress string, startProtocol, endProtocol int32) {
	// 连接gateway的gRPC服务
	conn, err := grpc_lib.Dial("localhost:50051", grpc_lib.WithInsecure(), grpc_lib.WithTimeout(5*time.Second))
	if err != nil {
		log.Printf("Warning: Failed to connect to gateway: %v", err)
		return
	}
	defer conn.Close()

	// 创建gateway客户端
	client := gateway.NewGatewayServiceClient(conn)

	// 创建注册请求
	req := &gateway.RegisterProtocolRangeRequest{
		ServiceName:    serviceName,
		ServiceAddress: serviceAddress,
		StartProtocol:  startProtocol,
		EndProtocol:    endProtocol,
	}

	// 发送注册请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.RegisterProtocolRange(ctx, req)
	if err != nil {
		log.Printf("Warning: Failed to register protocol range to gateway: %v", err)
		return
	}

	if resp.Success {
		log.Printf("Successfully registered protocol range %d-%d to gateway", startProtocol, endProtocol)
	} else {
		log.Printf("Failed to register protocol range to gateway: %s", resp.Error)
	}
}
