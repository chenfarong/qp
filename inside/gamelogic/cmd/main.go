package main

import (
	"fmt"
	"os"

	"zagame/common/logger"
	"zagame/config"
	"zagame/inside/gamelogic"
)

func main() {
	// 加载配置
	config.LoadConfig()

	// 初始化日志系统
	logger.Init(logger.Config{
		ServerName: "gamelogic",
		Level:      logger.DEBUG,
		Outputs: []logger.OutputConfig{
			{Type: logger.Console},
			{Type: logger.File},
		},
		UDPServer: "",
		UDPPort:   0,
	})
	defer logger.Close()

	// gateway服务地址
	gatewayAddress := fmt.Sprintf("%s:%d", config.AppConfig.Gateway.Host, config.AppConfig.Gateway.GrpcPort) // gateway服务的gRPC地址

	// 测试 key-value 格式的日志
	logger.InfoKV("启动游戏逻辑服务", "gateway", gatewayAddress, "version", "1.0.0")

	// 启动gRPC客户端
	if err := gamelogic.StartClient(gatewayAddress); err != nil {
		logger.Fatalf("启动游戏逻辑服务失败: %v", err)
		os.Exit(1)
	}
}
