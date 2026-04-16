package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"zgame/config"
	"zgame/internet/gateway/proto"

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
	// 注册WebSocket路由
	http.HandleFunc("/ws", handleWebSocket)

	addr := fmt.Sprintf("%s:%d", config.AppConfig.Game.Host, config.AppConfig.Game.Port)
	return http.ListenAndServe(addr, nil)
}

// forwardToGameLogic 转发消息到游戏逻辑服务
func forwardToGameLogic(msgID int32, session string, message []byte) ([]byte, error) {
	// 查找处理该消息的服务器
	serverManager.mu.RLock()
	serverID, exists := serverManager.msgIDMap[msgID]
	if !exists {
		serverManager.mu.RUnlock()
		return nil, fmt.Errorf("消息ID %d 未找到对应的服务器", msgID)
	}

	server, exists := serverManager.servers[serverID]
	if !exists {
		serverManager.mu.RUnlock()
		return nil, fmt.Errorf("服务器 %s 不存在", serverID)
	}
	serverManager.mu.RUnlock()

	// 转发消息到目标服务器
	response, err := server.Client.ForwardMessage(context.Background(), &proto.ForwardMessageRequest{
		MessageId:      msgID,
		Session:        session,
		MessageContent: message,
	})

	if err != nil {
		return nil, fmt.Errorf("转发消息失败: %v", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("服务器处理消息失败: %s", response.Message)
	}

	return response.ResponseContent, nil
}

// handleWebSocket 处理WebSocket连接
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 升级HTTP连接为WebSocket连接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v\n", err)
		return
	}
	defer conn.Close()

	// 从URL参数中获取token
	token := r.URL.Query().Get("token")
	if token == "" {
		log.Println("Empty token")
		return
	}

	// 使用token作为session
	session := token

	// 存储客户端连接
	clientsMutex.Lock()
	clients[session] = conn
	clientsMutex.Unlock()

	log.Printf("Client connected with session: %s\n", session)

	// 处理消息
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Read message error: %v\n", err)
			break
		}

		// 处理收到的消息
		log.Printf("Received message from session %s: %s\n", session, message)

		// 解析消息
		var wsMsg struct {
			MsgID int32           `json:"msg_id"`
			Data  json.RawMessage `json:"data"`
		}

		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("解析消息失败: %v\n", err)
			// 发送错误响应
			errorResp := map[string]interface{}{
				"msg_id": wsMsg.MsgID,
				"data": map[string]interface{}{
					"success": false,
					"message": "Invalid message format",
				},
			}
			errorJSON, _ := json.Marshal(errorResp)
			conn.WriteMessage(messageType, errorJSON)
			continue
		}

		// 转发消息到gamelogic
		response, err := forwardToGameLogic(wsMsg.MsgID, session, message)
		if err != nil {
			log.Printf("转发消息失败: %v\n", err)
			// 发送错误响应
			errorResp := map[string]interface{}{
				"msg_id": wsMsg.MsgID,
				"data": map[string]interface{}{
					"success": false,
					"message": "Failed to forward message",
				},
			}
			errorJSON, _ := json.Marshal(errorResp)
			conn.WriteMessage(messageType, errorJSON)
			continue
		}

		// 发送响应给客户端
		if err := conn.WriteMessage(messageType, response); err != nil {
			log.Printf("Write message error: %v\n", err)
			break
		}
	}

	// 移除客户端连接
	clientsMutex.Lock()
	delete(clients, session)
	clientsMutex.Unlock()

	log.Printf("Client disconnected with session: %s\n", session)
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
