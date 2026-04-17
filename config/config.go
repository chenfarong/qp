package config

import (
	"flag"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

// Config 服务器配置结构
type Config struct {
	Auth       AuthConfig       `yaml:"auth"`
	Game       GameConfig       `yaml:"game"`
	Gateway    GatewayConfig    `yaml:"gateway"`
	Database   DatabaseConfig   `yaml:"database"`
	Log        LogConfig        `yaml:"log"`
	GameConfig GameConfigConfig `yaml:"game_config"`
}

// GatewayConfig 网关服务器配置
type GatewayConfig struct {
	Host           string `yaml:"host"`
	WsPort         int    `yaml:"ws_port"`
	GrpcPort       int    `yaml:"grpc_port"`
	MaxConnections int    `yaml:"max_connections"`
	SessionTimeout int    `yaml:"session_timeout"`
}

// AuthConfig 验证服务器配置
type AuthConfig struct {
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	SecretKey   string `yaml:"secret_key"`
	TokenExpiry int    `yaml:"token_expiry"`
}

// GameConfig 游戏服务器配置
type GameConfig struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	MaxConnections int    `yaml:"max_connections"`
	SessionTimeout int    `yaml:"session_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type     string `yaml:"type"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Dbname   string `yaml:"dbname"`
	Sslmode  string `yaml:"sslmode"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `yaml:"level"`
	File       string `yaml:"file"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
}

// GameConfigConfig 游戏配置
type GameConfigConfig struct {
	StartingGold  int `yaml:"starting_gold"`
	StartingLevel int `yaml:"starting_level"`
	MaxBagSize    int `yaml:"max_bag_size"`
	MaxHeroes     int `yaml:"max_heroes"`
}

// 全局配置变量
var AppConfig Config

// LoadConfig 加载配置文件
func LoadConfig() {
	// 解析命令行参数
	configFile := flag.String("config", "../../config.yml", "配置文件路径")
	flag.Parse()

	// 读取配置文件
	data, err := os.ReadFile(*configFile)
	if err != nil {
		// 如果指定的路径不存在，尝试在当前目录查找
		if _, err := os.Stat("config.yml"); err == nil {
			configFile = flag.String("config", "config.yml", "配置文件路径")
			data, err = os.ReadFile(*configFile)
		}
		if err != nil {
			log.Fatalf("无法读取配置文件: %v", err)
		}
	}

	// 解析YAML配置
	err = yaml.Unmarshal(data, &AppConfig)
	if err != nil {
		log.Fatalf("无法解析配置文件: %v", err)
	}

	fmt.Printf("配置文件加载成功: %s\n", *configFile)
}
