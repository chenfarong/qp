package main

import (
	"fmt"
	"log"

	"zgame/internet/gateway/internal/handler"
)

func main() {
	// 启动WebSocket服务器
	fmt.Println("Gateway WebSocket Server started on port 8081")
	if err := handler.StartWebSocketServer(); err != nil {
		log.Fatalf("WebSocket server error: %v", err)
	}
}
