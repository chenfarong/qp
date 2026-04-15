package handler

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocket连接升级器
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源的请求
	},
}

// 客户端连接映射
var clients = make(map[string]*websocket.Conn)
var clientsMutex sync.RWMutex

// StartWebSocketServer 启动WebSocket服务器
func StartWebSocketServer() error {
	http.HandleFunc("/ws", handleWebSocket)
	return http.ListenAndServe(":8081", nil)
}

// handleWebSocket 处理WebSocket连接
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 升级HTTP连接为WebSocket连接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("WebSocket upgrade error: %v\n", err)
		return
	}
	defer conn.Close()

	// 读取session
	var msg struct {
		Session string `json:"session"`
	}

	if err := conn.ReadJSON(&msg); err != nil {
		fmt.Printf("Read session error: %v\n", err)
		return
	}

	session := msg.Session
	if session == "" {
		fmt.Println("Empty session")
		return
	}

	// 存储客户端连接
	clientsMutex.Lock()
	clients[session] = conn
	clientsMutex.Unlock()

	fmt.Printf("Client connected with session: %s\n", session)

	// 处理消息
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("Read message error: %v\n", err)
			break
		}

		// 处理收到的消息（这里可以根据消息类型进行处理）
		fmt.Printf("Received message from session %s: %s\n", session, message)

		// 回复客户端
		if err := conn.WriteMessage(messageType, message); err != nil {
			fmt.Printf("Write message error: %v\n", err)
			break
		}
	}

	// 移除客户端连接
	clientsMutex.Lock()
	delete(clients, session)
	clientsMutex.Unlock()

	fmt.Printf("Client disconnected with session: %s\n", session)
}

// SendToClient 发送消息到客户端
func SendToClient(session string, message []byte) error {
	clientsMutex.RLock()
	conn, ok := clients[session]
	clientsMutex.RUnlock()

	if !ok {
		return fmt.Errorf("client not found")
	}

	return conn.WriteMessage(websocket.TextMessage, message)
}