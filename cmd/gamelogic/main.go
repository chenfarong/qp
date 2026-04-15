package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"runtime/debug"
	"strconv"
	"time"

	"os"

	"github.com/aoyo/qp/internal/gamelogic"
	gamelogicgrpc "github.com/aoyo/qp/internal/gamelogic/grpc"
	"github.com/aoyo/qp/pkg/db"
	"github.com/aoyo/qp/pkg/envmode"
	"github.com/aoyo/qp/pkg/etcd"
	"github.com/aoyo/qp/pkg/logger"
	"github.com/aoyo/qp/pkg/proto/game"
	"github.com/aoyo/qp/pkg/proto/gateway"
	"github.com/aoyo/qp/pkg/utils"
	"google.golang.org/grpc"
	grpc_lib "google.golang.org/grpc"
	"gopkg.in/yaml.v3"
)

// Config 配置结构
type Config struct {
	Sandbox bool `yaml:"sandbox"`
	Server  struct {
		Gamelogic struct {
			Port     int `yaml:"port"`
			GrpcPort int `yaml:"grpc_port"`
		} `yaml:"gamelogic"`
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
	log.Println("Loading configuration...")
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Println("Configuration loaded successfully")

	// 初始化数据库
	log.Println("Connecting to database...")
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
	log.Printf("Database URI: %s", uri)

	// 尝试连接数据库
	dbInstance, err := db.InitDB(uri)
	if err != nil {
		log.Printf("Warning: Failed to connect to database: %v", err)
		log.Println("Continuing without database connection...")
	} else {
		log.Println("Database connected successfully")
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

	// 初始化游戏逻辑应用
	log.Println("Initializing game logic application...")
	app := gamelogic.GetApp(dbInstance, config.Database.Dbname)
	log.Println("Game logic application initialized successfully")

	// 启动游戏逻辑服务
	log.Println("Starting game logic service...")
	if err := gamelogic.StartGameLogic(dbInstance, config.Database.Dbname); err != nil {
		log.Printf("Warning: Failed to start game logic service: %v", err)
		log.Println("Continuing without game logic service...")
	} else {
		log.Println("Game logic service started successfully")
	}

	// 初始化日志客户端
	logClient, err := logger.NewClient(fmt.Sprintf("localhost:%d", config.Server.Logger.UdpPort), fmt.Sprintf("gamelogic://localhost:%d", config.Server.Gamelogic.Port))
	if err != nil {
		log.Printf("Warning: Failed to initialize log client: %v", err)
	} else {
		defer logClient.Close()
		// 发送测试日志
		logClient.Warn("Game Logic server starting")
	}

	// 注册服务到 etcd
	if etcdClient != nil {
		serviceAddress := fmt.Sprintf("localhost:%d", config.Server.Gamelogic.Port)
		if err := etcdClient.RegisterService("gamelogic", serviceAddress); err != nil {
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

	grpcPort := config.Server.Gamelogic.GrpcPort
	if grpcPort == 0 {
		grpcPort = config.Server.Gamelogic.Port + 1000
	}
	go func() {
		defer recoverPanic(logClient, "Game Logic gRPC")
		startGRPCServer(app, grpcPort)
	}()

	// 初始化路由
	log.Println("Initializing router...")
	router := gamelogic.NewRouter(app)
	log.Println("Router initialized successfully")

	// 打印欢迎日志
	printWelcomeLog("Game Logic", config.Server.Gamelogic.Port, grpcPort, config.Database.Host, config.Database.Port, config.Database.Dbname)

	// 向gateway注册协议编号段
	go func() {
		defer recoverPanic(logClient, "Game Logic Gateway Registration")
		registerToGateway("gamelogic", fmt.Sprintf("localhost:%d", config.Server.Gamelogic.Port), 101, 200)
	}()

	// 启动HTTP服务
	port := config.Server.Gamelogic.Port
	log.Printf("Game Logic service starting on port %d...", port)
	
	// 主服务器启动，添加panic recovery
	defer recoverPanic(logClient, "Game Logic HTTP")
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

func startGRPCServer(app *gamelogic.App, port int) {
	grpcServer := grpc.NewServer()
	game.RegisterGameServiceServer(grpcServer, gamelogicgrpc.NewGameServer(app))
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen gRPC: %v", err)
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
