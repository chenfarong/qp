package main

import (
	"fmt"

	"zagame/common/logger"
	"zagame/config"
	"zagame/outside/gateway/internal/handler"
)

func main() {
	// 加载配置
	config.LoadConfig()

	// 初始化日志系统
	logger.Init(logger.Config{
		ServerName: "gateway",
		Level:      logger.DEBUG,
		Outputs: []logger.OutputConfig{
			{Type: logger.Console},
			{Type: logger.File},
		},
		UDPServer: "",
		UDPPort:   0,
	})
	defer logger.Close()

	// 初始化数据库连接（注释掉，使用内存存储）
	// if err := database.InitDatabase(); err != nil {
	//  logger.Fatalf("初始化数据库失败: %v", err)
	// }
	// defer database.CloseDatabase()

	// 启动gRPC服务器
	go func() {
		grpcPort := config.AppConfig.Gateway.GrpcPort // gRPC服务器端口
		logger.InfoKV("启动gRPC服务器", "port", grpcPort)
		if err := handler.StartGRPCServer(grpcPort); err != nil {
			logger.Fatalf("gRPC服务器启动失败: %v", err)
		}
	}()

	// 启动WebSocket服务器
	addr := fmt.Sprintf("%s:%d", config.AppConfig.Gateway.Host, config.AppConfig.Gateway.WsPort)
	logger.InfoKV("启动WebSocket服务器", "address", addr)
	if err := handler.StartWebSocketServer(); err != nil {
		logger.Fatalf("WebSocket server error: %v", err)
	}
}
