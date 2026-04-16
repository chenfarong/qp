package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"zgame/config"
	"zgame/database"

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
	// 注册HTTP API路由
	http.HandleFunc("/actor/list", handleGetActorList)
	http.HandleFunc("/actor/create", handleCreateActor)

	// 注册WebSocket路由
	http.HandleFunc("/ws", handleWebSocket)

	addr := fmt.Sprintf("%s:%d", config.AppConfig.Game.Host, config.AppConfig.Game.Port)
	return http.ListenAndServe(addr, nil)
}

// handleGetActorList 处理获取角色列表请求
func handleGetActorList(w http.ResponseWriter, r *http.Request) {
	// 从Authorization header中获取token
	token := r.Header.Get("Authorization")
	if token == "" {
		log.Println("获取角色列表失败: 缺少Authorization header")
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("{\"success\": false, \"message\": \"Authorization header is required\"}")
		return
	}

	// 移除Bearer前缀
	token = strings.TrimPrefix(token, "Bearer ")

	// 这里应该验证token，实际项目中应该使用JWT验证
	// 为了测试，我们假设token是有效的

	log.Println("收到获取角色列表请求")

	// 检查数据库连接是否初始化
	if database.DB == nil {
		log.Println("Database connection is not initialized")
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("{\"success\": false, \"message\": \"Database connection not initialized\"}")
		return
	}

	// 从数据库中获取角色列表
	rows, err := database.DB.Query(`
		SELECT actor_id, name, level, realm, created_at, updated_at, online_at, offline_at 
		FROM actors
	`)
	if err != nil {
		log.Printf("Database query error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("{\"success\": false, \"message\": \"Database error: %v\"}", err)
		return
	}
	defer rows.Close()

	// 解析角色列表
	var actors []map[string]interface{}
	for rows.Next() {
		var actorID, name, realm string
		var level int
		var createdAt, updatedAt, onlineAt time.Time
		var offlineAt *time.Time // 使用指针类型来存储可能为nil的值

		err := rows.Scan(&actorID, &name, &level, &realm, &createdAt, &updatedAt, &onlineAt, &offlineAt)
		if err != nil {
			log.Printf("Database scan error: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("{\"success\": false, \"message\": \"Database error: %v\"}", err)
			return
		}

		// 处理offline_at字段
		offlineAtUnix := int64(0)
		if offlineAt != nil {
			offlineAtUnix = offlineAt.Unix()
		}

		actor := map[string]interface{}{
			"actor_id":   actorID,
			"name":       name,
			"level":      level,
			"realm":      realm,
			"created_at": createdAt.Unix(),
			"updated_at": updatedAt.Unix(),
			"online_at":  onlineAt.Unix(),
			"offline_at": offlineAtUnix,
		}
		actors = append(actors, actor)
	}

	log.Printf("获取角色列表成功，共 %d 个角色\n", len(actors))

	// 生成响应
	response := map[string]interface{}{
		"success": true,
		"message": "获取角色列表成功",
		"data":    actors,
	}

	// 编码为JSON
	responseJSON, err := json.Marshal(response)
	if err != nil {
		log.Printf("JSON encoding error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("{\"success\": false, \"message\": \"JSON encoding error\"}")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}

// handleCreateActor 处理创建角色请求
func handleCreateActor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Println("创建角色失败: 方法不允许")
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Printf("{\"success\": false, \"message\": \"Method not allowed\"}")
		return
	}

	// 从Authorization header中获取token
	token := r.Header.Get("Authorization")
	if token == "" {
		log.Println("创建角色失败: 缺少Authorization header")
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("{\"success\": false, \"message\": \"Authorization header is required\"}")
		return
	}

	// 移除Bearer前缀
	token = strings.TrimPrefix(token, "Bearer ")

	// 解析请求体
	var req struct {
		Name string `json:"name"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("创建角色失败: 请求体解析错误: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("{\"success\": false, \"message\": \"Invalid request body\"}")
		return
	}

	log.Printf("收到创建角色请求: name=%s\n", req.Name)

	// 这里应该验证token，实际项目中应该使用JWT验证
	// 为了测试，我们假设token是有效的

	// 生成角色ID
	actorID := fmt.Sprintf("actor_%d", time.Now().UnixNano())

	// 保存角色到数据库
	_, err = database.DB.Exec(`
		INSERT INTO actors (actor_id, user_id, name, level, realm, created_at, updated_at, online_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, actorID, 1, req.Name, 1, "realm_1", time.Now(), time.Now(), time.Now())
	if err != nil {
		log.Printf("创建角色失败: 数据库错误: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("{\"success\": false, \"message\": \"Database error\"}")
		return
	}

	log.Printf("创建角色成功: name=%s, actor_id=%s\n", req.Name, actorID)

	// 生成响应
	response := map[string]interface{}{
		"success": true,
		"message": "创建角色成功",
		"data": map[string]interface{}{
			"actor_id":   actorID,
			"name":       req.Name,
			"level":      1,
			"realm":      "realm_1",
			"created_at": time.Now().Unix(),
			"updated_at": time.Now().Unix(),
			"online_at":  time.Now().Unix(),
			"offline_at": 0,
		},
	}

	// 编码为JSON
	responseJSON, err := json.Marshal(response)
	if err != nil {
		log.Printf("JSON encoding error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("{\"success\": false, \"message\": \"JSON encoding error\"}")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
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

	// 读取session
	var msg struct {
		Session string `json:"session"`
	}

	if err := conn.ReadJSON(&msg); err != nil {
		log.Printf("Read session error: %v\n", err)
		return
	}

	session := msg.Session
	if session == "" {
		log.Println("Empty session")
		return
	}

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

		// 处理收到的消息（这里可以根据消息类型进行处理）
		log.Printf("Received message from session %s: %s\n", session, message)

		// 回复客户端
		if err := conn.WriteMessage(messageType, message); err != nil {
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
