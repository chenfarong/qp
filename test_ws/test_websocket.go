package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	// WebSocket服务器地址
	wsAddr := "ws://localhost:8061/ws"

	// 添加token参数
	params := url.Values{}
	params.Add("token", "test_session_123")
	wsAddr = wsAddr + "?" + params.Encode()

	// 连接到WebSocket服务器
	fmt.Printf("连接到WebSocket服务器: %s\n", wsAddr)
	conn, _, err := websocket.DefaultDialer.Dial(wsAddr, nil)
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer conn.Close()

	// 启动一个goroutine接收消息
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("读取消息失败: %v", err)
				return
			}
			log.Printf("收到消息: %s", message)
		}
	}()

	// 发送测试消息
	testMsg := map[string]interface{}{
		"msg_id": 1001, // LoginRequest
		"data": map[string]interface{}{
			"session": "test_session_123",
			"name":    "TestPlayer",
		},
	}

	if err := conn.WriteJSON(testMsg); err != nil {
		log.Printf("发送消息失败: %v", err)
		return
	}
	fmt.Println("发送测试消息成功")

	// 等待信号
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	select {
	case <-interrupt:
		fmt.Println("收到中断信号，关闭连接")
	case <-time.After(10 * time.Second):
		fmt.Println("10秒后超时，关闭连接")
	}
}
