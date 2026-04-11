package main

import (
	"fmt"
	"log"
	"net"
	"strconv"

	"os"

	"github.com/aoyo/qp/internal/gamelogic"
	gamelogicgrpc "github.com/aoyo/qp/internal/gamelogic/grpc"
	"github.com/aoyo/qp/pkg/db"
	"github.com/aoyo/qp/pkg/envmode"
	"github.com/aoyo/qp/pkg/etcd"
	"github.com/aoyo/qp/pkg/proto/game"
	"google.golang.org/grpc"
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
	uri := fmt.Sprintf(
		"mongodb://%s:%s@%s:%d/%s?authSource=admin",
		config.Database.User,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.Dbname,
	)
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

	// 注册服务到 etcd
	if etcdClient != nil {
		serviceAddress := fmt.Sprintf("localhost:%d", config.Server.Gamelogic.Port)
		if err := etcdClient.RegisterService("gamelogic", serviceAddress); err != nil {
			log.Printf("Warning: Failed to register service to etcd: %v", err)
		} else {
			log.Println("Service registered to etcd successfully")
		}
	}

	grpcPort := config.Server.Gamelogic.GrpcPort
	if grpcPort == 0 {
		grpcPort = config.Server.Gamelogic.Port + 1000
	}
	go startGRPCServer(app, grpcPort)

	// 初始化路由
	log.Println("Initializing router...")
	router := gamelogic.NewRouter(app)
	log.Println("Router initialized successfully")

	// 启动HTTP服务
	port := config.Server.Gamelogic.Port
	log.Printf("Game Logic service starting on port %d...", port)
	if err := router.Run(":" + strconv.Itoa(port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
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
