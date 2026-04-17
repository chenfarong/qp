package main

import (
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

	// 打印欢迎信息
	logger.Info("========================================================")
	logger.Info("                      Gateway Server                    ")
	logger.Info("========================================================")
	logger.Info("服务名称: Gateway Server")
	logger.Info("服务类型: WebSocket + gRPC Server")
	logger.Info("监听地址: %s", config.AppConfig.Gateway.Host)
	logger.Info("WebSocket端口: %d", config.AppConfig.Gateway.WsPort)
	logger.Info("gRPC端口: %d", config.AppConfig.Gateway.GrpcPort)
	logger.Info("最大连接数: %d", config.AppConfig.Gateway.MaxConnections)
	logger.Info("会话超时时间: %d秒", config.AppConfig.Gateway.SessionTimeout)
	logger.Info("========================================================")
	logger.Info("服务器已成功启动，等待客户端连接...")
	logger.Info("========================================================")

	// 启动gRPC服务器
	go func() {
		grpcPort := config.AppConfig.Gateway.GrpcPort // gRPC服务器端口
		if err := handler.StartGRPCServer(grpcPort); err != nil {
			logger.Fatalf("gRPC服务器启动失败: %v", err)
		}
	}()

	// 启动WebSocket服务器
	if err := handler.StartWebSocketServer(); err != nil {
		logger.Fatalf("WebSocket server error: %v", err)
	}
}
