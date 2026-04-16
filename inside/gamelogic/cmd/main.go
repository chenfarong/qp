package main

import (
	"log"

	"zagame/inside/gamelogic"
)

func main() {
	// 启动gRPC服务器
	port := int32(8083) // 游戏逻辑服务的gRPC端口
	gatewayAddress := "localhost:8082" // gateway服务的gRPC地址

	log.Printf("启动游戏逻辑服务，端口: %d\n", port)

	// 启动gRPC服务器
	if err := gamelogic.StartGRPCServer(port, gatewayAddress); err != nil {
		log.Fatalf("启动游戏逻辑服务失败: %v\n", err)
	}
}
