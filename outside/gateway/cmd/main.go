package main

import (
	"fmt"
	"log"

	"zgame/config"
	"zgame/database"
	"zgame/internet/gateway/internal/handler"
)

func main() {
	// 加载配置
	config.LoadConfig()

	// 初始化数据库连接
	if err := database.InitDatabase(); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer database.CloseDatabase()

	// 启动gRPC服务器
	go func() {
		grpcPort := 8082 // gRPC服务器端口
		if err := handler.StartGRPCServer(grpcPort); err != nil {
			log.Fatalf("gRPC服务器启动失败: %v", err)
		}
	}()

	// 启动WebSocket服务器
	addr := fmt.Sprintf("%s:%d", config.AppConfig.Game.Host, config.AppConfig.Game.Port)
	fmt.Printf("Gateway WebSocket Server started on %s\n", addr)
	if err := handler.StartWebSocketServer(); err != nil {
		log.Fatalf("WebSocket server error: %v", err)
	}
}
