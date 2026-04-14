package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"os"

	billgrpc "github.com/aoyo/qp/internal/bill/grpc"
	"github.com/aoyo/qp/internal/bill/handler"
	"github.com/aoyo/qp/internal/bill/service"
	"github.com/aoyo/qp/pkg/db"
	"github.com/aoyo/qp/pkg/envmode"
	"github.com/aoyo/qp/pkg/etcd"
	"github.com/aoyo/qp/pkg/logger"
	"github.com/aoyo/qp/pkg/proto/bill"
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
		Bill struct {
			Port     int `yaml:"port"`
			GrpcPort int `yaml:"grpc_port"`
		} `yaml:"bill"`
		Logger struct {
			UdpPort int `yaml:"udp_port"`
		} `yaml:"logger"`
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
	grpcPort := config.Server.Bill.GrpcPort
	if grpcPort == 0 {
		grpcPort = port + 1000
	}
	var dbURI string
	if config.Database.User != "" && config.Database.Password != "" {
		dbURI = fmt.Sprintf(
			"mongodb://%s:%s@%s:%d/%s?authSource=admin",
			config.Database.User,
			config.Database.Password,
			config.Database.Host,
			config.Database.Port,
			config.Database.Dbname,
		)
	} else {
		dbURI = fmt.Sprintf(
			"mongodb://%s:%d/%s",
			config.Database.Host,
			config.Database.Port,
			config.Database.Dbname,
		)
	}
	dbName := config.Database.Dbname

	// 初始化数据库连接
	dbInstance, err := db.InitDB(dbURI)
	if err != nil {
		log.Printf("Warning: Failed to connect to database: %v", err)
		log.Println("Continuing without database connection...")
	} else {
		defer dbInstance.Close()
	}

	// 打印欢迎日志
	printWelcomeLog("Bill", port, grpcPort, config.Database.Host, config.Database.Port, config.Database.Dbname)

	// 向gateway注册协议编号段
	go registerToGateway("bill", fmt.Sprintf("localhost:%d", port), 301, 400)

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
	tokenService := service.NewTokenService(dbInstance, dbName)
	paymentService := service.NewPaymentService(dbInstance, dbName, tokenService)

	// 初始化日志客户端
	logClient, err := logger.NewClient(fmt.Sprintf("localhost:%d", config.Server.Logger.UdpPort), fmt.Sprintf("bill://localhost:%d", port))
	if err != nil {
		log.Printf("Warning: Failed to initialize log client: %v", err)
	} else {
		defer logClient.Close()
		// 发送测试日志
		logClient.Warn("Bill server starting")
	}

	// 注册服务到 etcd
	if etcdClient != nil {
		serviceAddress := fmt.Sprintf("localhost:%d", port)
		if err := etcdClient.RegisterService("bill", serviceAddress); err != nil {
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
