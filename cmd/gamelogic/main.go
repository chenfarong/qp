package main

import (
	"fmt"
	"log"
	"strconv"

	"os"

	"github.com/aoyo/qp/internal/gamelogic"
	"github.com/aoyo/qp/pkg/db"
	"github.com/aoyo/qp/pkg/etcd"
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

	// 启动gRPC服务器（暂时注释掉，因为缺少proto文件）
	// go startGRPCServer(app.CharacterService, config.Server.Gamelogic.Port+1000)

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

// startGRPCServer 启动gRPC服务器（暂时注释掉，因为缺少proto文件）
/*
func startGRPCServer(characterService *actor.CharacterService, port int) {
	// 创建gRPC服务器
	grpcServer := grpc.NewServer()

	// 注册游戏服务
	// game.RegisterGameServiceServer(grpcServer, grpc.NewGameServer(characterService))

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
*/

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
