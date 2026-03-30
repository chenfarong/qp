package main

import (
	"fmt"
	"log"
	"strconv"

	"os"

	"github.com/aoyo/qp/internal/chat/handler"
	"github.com/aoyo/qp/internal/chat/service"
	"github.com/aoyo/qp/pkg/db"
	"github.com/gin-gonic/gin"
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

	// 初始化服务
	chatService := service.NewChatService(dbClient, config.Database.Dbname)
	chatHandler := handler.NewChatHandler(chatService)

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

	// 启动服务
	port := config.Server.Chat.Port
	log.Printf("Chat service starting on port %d...", port)
	if err := router.Run(":" + strconv.Itoa(port)); err != nil {
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
