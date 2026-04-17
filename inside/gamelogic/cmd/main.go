package main

import (
	"os"

	"zagame/common/logger"
	"zagame/inside/gamelogic"
)

func main() {
	// 初始化日志系统
	logger.Init(logger.Config{
		ServerName: "gamelogic",
		Level:      logger.DEBUG,
		Outputs: []logger.OutputConfig{
			{Type: logger.Console},
		},
		UDPServer: "",
		UDPPort:   0,
	})
	defer logger.Close()

	// 启动gRPC服务器
	port := int32(8083)                // 游戏逻辑服务的gRPC端口
	gatewayAddress := "localhost:8082" // gateway服务的gRPC地址

	// 测试 key-value 格式的日志
	logger.InfoKV("启动游戏逻辑服务", "port", port, "gateway", gatewayAddress, "version", "1.0.0")

	// 启动gRPC服务器
	if err := gamelogic.StartGRPCServer(port, gatewayAddress); err != nil {
		logger.Fatalf("启动游戏逻辑服务失败: %v", err)
		os.Exit(1)
	}
}
