package main

import (
	"fmt"
	"log"

	"os"
	"strconv"

	"github.com/aoyo/qp/internal/ssoauth/handler"
	"github.com/aoyo/qp/internal/ssoauth/service"
	"github.com/aoyo/qp/pkg/db"
	"github.com/gin-gonic/gin"
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
	authService := service.NewAuthService(dbInstance, config.Jwt.Secret, config.Jwt.ExpireHours, config.Database.Dbname)
	authHandler := handler.NewAuthHandler(authService)

	// 初始化路由
	router := gin.Default()
	authHandler.RegisterRoutes(router)

	// 启动服务
	port := config.Server.Ssoauth.Port
	log.Printf("SSO Auth service starting on port %d...", port)
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
