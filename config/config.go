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
	// 尝试从多个位置加载配置文件
	configPaths := []string{
		"config.yml",           // 当前目录
		"test_config.yml",      // 测试配置文件
		"../../config.yml",     // 相对于 cmd 目录
		"../../../config.yml",  // 相对于 outside/cmd 目录
		"../test_config.yml",   // 相对于 test/gamelogic 目录
	}

	var data []byte
	var err error
	var configFile string

	// 尝试读取配置文件
	for _, path := range configPaths {
		data, err = os.ReadFile(path)
		if err == nil {
			configFile = path
			break
		}
	}

	// 如果所有路径都失败，尝试解析命令行参数
	if err != nil {
		configFileFlag := flag.String("config", "config.yml", "配置文件路径")
		flag.Parse()
		data, err = os.ReadFile(*configFileFlag)
		if err != nil {
			log.Fatalf("无法读取配置文件: %v", err)
		}
		configFile = *configFileFlag
	}

	// 解析YAML配置
	err = yaml.Unmarshal(data, &AppConfig)
	if err != nil {
		log.Fatalf("无法解析配置文件: %v", err)
	}

	fmt.Printf("配置文件加载成功: %s\n", configFile)
}
