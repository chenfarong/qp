package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"zagame/config"
	"zagame/outside/gateway/proto"

	"github.com/gorilla/websocket"
)

// 客户端连接映射
var clients = make(map[string]*websocket.Conn)
var clientsMutex sync.RWMutex

// 升级HTTP连接为WebSocket连接
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// StartWebSocketServer 启动WebSocket服务器
func StartWebSocketServer() error {
	// 注册WebSocket路由
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// 从URL参数中获取token
		// token := r.URL.Query().Get("token")
		// if token == "" {
		// 	log.Println("Empty token")
		// 	w.WriteHeader(http.StatusBadRequest)
		// 	return
		// }

		// 升级HTTP连接为WebSocket连接
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v\n", err)
			return
		}

		// 使用token作为session，如果不存在则生成一个随机ID
		token := r.URL.Query().Get("token")
		session := token
		if session == "" {
			session, err = newSessionID()
			if err != nil {
				log.Printf("生成session ID失败: %v\n", err)
				conn.Close()
				return
			}
		}
		clientIP := r.RemoteAddr

		// 存储客户端连接
		clientsMutex.Lock()
		clients[session] = conn
		clientsMutex.Unlock()

		log.Printf("Client connected with session: %s, IP: %s\n", session, clientIP)

		// 处理WebSocket消息
		go handleWebSocket(conn, session, clientIP)
	})

	addr := fmt.Sprintf(":%d", config.AppConfig.Gateway.WsPort)
	log.Printf("Gateway WebSocket Server starting on %s", addr)
	return http.ListenAndServe(addr, nil)
}

// handleWebSocket 处理WebSocket连接
func handleWebSocket(conn *websocket.Conn, session string, clientIP string) {
	defer func() {
		// 关闭连接
		conn.Close()

		// 移除客户端连接
		clientsMutex.Lock()
		delete(clients, session)
		clientsMutex.Unlock()

		log.Printf("Client disconnected with session: %s, IP: %s\n", session, clientIP)
	}()

	// 循环读取消息
	for {
		// 读取消息
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v\n", err)
			break
		}

		// 处理收到的消息
		log.Printf("Received message from session %s (IP: %s): %s\n", session, clientIP, message)

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
			conn.WriteMessage(websocket.TextMessage, errorJSON)
			continue
		}

		// 转发消息到gamelogic
		response, err := forwardToGameLogic(wsMsg.MsgID, session, clientIP, message)
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
			conn.WriteMessage(websocket.TextMessage, errorJSON)
			continue
		}

		// 发送响应给客户端
		if err := conn.WriteMessage(websocket.TextMessage, response); err != nil {
			log.Printf("WebSocket write error: %v\n", err)
			break
		}
	}
}

func newSessionID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// forwardToGameLogic 转发消息到游戏逻辑服务
func forwardToGameLogic(msgID int32, session string, clientIP string, message []byte) ([]byte, error) {
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
		ClientIp:       clientIP,
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
