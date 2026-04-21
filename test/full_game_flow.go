package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

// 全局变量存储命令行参数
var (
	authServerURL = flag.String("auth", "http://localhost:8080", "验证服务器URL")
	gatewayURL    = flag.String("gateway", "ws://localhost:8061", "网关服务器WebSocket URL")
	username      = flag.String("username", "fulltestuser", "用户名")
	password      = flag.String("password", "fulltestpass", "密码")
	actorName     = flag.String("actor", "fulltestactor", "角色名称")
)

// LoginResponse 登录响应结构
type LoginResponse struct {
	Success bool   `json:"success"`
	Session string `json:"session"`
	Token   string `json:"token"`
	Message string `json:"message"`
}

// RegisterResponse 注册响应结构
type RegisterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ActorInfo 角色信息结构
type ActorInfo struct {
	ActorId string `json:"actor_id"`
	Name    string `json:"name"`
	Level   int32  `json:"level"`
	Realm   string `json:"realm"`
}

// ActorListResponse 角色列表响应结构
type ActorListResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    []ActorInfo `json:"data"`
}

// WebSocketMessage WebSocket消息结构
type WebSocketMessage struct {
	MsgID int32           `json:"msg_id"`
	Data  json.RawMessage `json:"data"`
}

var wsConn *websocket.Conn

func main() {
	// 检查是否有help参数
	for _, arg := range os.Args[1:] {
		if arg == "help" {
			printHelp()
			return
		}
	}

	// 解析命令行参数
	flag.Parse()

	fmt.Println("=== 完整游戏流程测试 ===")
	fmt.Printf("验证服务器: %s\n", *authServerURL)
	fmt.Printf("网关服务器: %s\n", *gatewayURL)
	fmt.Printf("用户名: %s\n", *username)
	fmt.Printf("密码: %s\n", *password)
	fmt.Printf("角色名称: %s\n", *actorName)

	// 步骤1: 注册账号
	fmt.Println("\n=== 步骤1: 注册账号 ===")
	if err := register(); err != nil {
		fmt.Printf("注册失败: %v\n", err)
		return
	}
	fmt.Println("注册成功!")

	// 步骤2: 登录获取token
	fmt.Println("\n=== 步骤2: 登录验证 ===")
	token, err := login()
	if err != nil {
		fmt.Printf("登录失败: %v\n", err)
		return
	}
	fmt.Println("登录成功!")

	// 步骤3: WebSocket连接到网关
	fmt.Println("\n=== 步骤3: 连接网关 ===")
	if err := connectWebSocket(token); err != nil {
		fmt.Printf("连接网关失败: %v\n", err)
		return
	}
	fmt.Println("网关连接成功!")

	// 步骤4: 创建角色
	fmt.Println("\n=== 步骤4: 创建角色 ===")
	actor, err := createActor()
	if err != nil {
		fmt.Printf("创建角色失败: %v\n", err)
		return
	}
	fmt.Printf("创建角色成功: %s (ID: %s)\n", actor.Name, actor.ActorId)

	// 步骤5: 获取角色列表并选择角色
	fmt.Println("\n=== 步骤5: 获取角色列表 ===")
	actors, err := getActorList()
	if err != nil {
		fmt.Printf("获取角色列表失败: %v\n", err)
		return
	}
	fmt.Printf("找到 %d 个角色\n", len(actors))

	// 自动选择刚创建的角色
	var selectedActor *ActorInfo
	for i := range actors {
		if actors[i].Name == *actorName {
			selectedActor = &actors[i]
			break
		}
	}
	if selectedActor == nil && len(actors) > 0 {
		selectedActor = &actors[0]
	}

	if selectedActor != nil {
		fmt.Printf("选择角色: %s (ID: %s, 等级: %d)\n", selectedActor.Name, selectedActor.ActorId, selectedActor.Level)
	} else {
		fmt.Println("未找到可用的角色")
		return
	}

	// 步骤6: 获取背包数据
	fmt.Println("\n=== 步骤6: 获取背包数据 ===")
	bagData, err := getBagData()
	if err != nil {
		fmt.Printf("获取背包数据失败: %v\n", err)
		return
	}
	fmt.Println("获取背包数据成功!")
	fmt.Printf("背包物品数量: %d\n", len(bagData))

	// 显示背包内容
	for itemID, count := range bagData {
		fmt.Printf("  物品ID: %s, 数量: %d\n", itemID, count)
	}

	// 关闭连接
	wsConn.Close()

	fmt.Println("\n=== 测试完成 ===")
	fmt.Printf("用户名: %s\n", *username)
	fmt.Printf("角色ID: %s\n", selectedActor.ActorId)
	fmt.Printf("角色名称: %s\n", selectedActor.Name)
	fmt.Printf("角色等级: %d\n", selectedActor.Level)
	fmt.Printf("背包物品数: %d\n", len(bagData))
}

// register 注册账号
func register() error {
	// 构造注册请求
	registerReq := map[string]string{
		"username": *username,
		"password": *password,
	}

	jsonData, err := json.Marshal(registerReq)
	if err != nil {
		return fmt.Errorf("构造注册请求失败: %v", err)
	}

	// 发送注册请求
	resp, err := http.Post(*authServerURL+"/register", "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("发送注册请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取注册响应失败: %v", err)
	}

	var registerResp RegisterResponse
	if err := json.Unmarshal(body, &registerResp); err != nil {
		return fmt.Errorf("解析注册响应失败: %v", err)
	}

	if !registerResp.Success {
		return fmt.Errorf("注册失败: %s", registerResp.Message)
	}

	return nil
}

// login 登录验证
func login() (string, error) {
	// 构造登录请求
	loginReq := map[string]string{
		"username": *username,
		"password": *password,
	}

	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return "", fmt.Errorf("构造登录请求失败: %v", err)
	}

	// 发送登录请求
	resp, err := http.Post(*authServerURL+"/login", "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("发送登录请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取登录响应失败: %v", err)
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return "", fmt.Errorf("解析登录响应失败: %v", err)
	}

	if !loginResp.Success {
		return "", fmt.Errorf("登录失败: %s", loginResp.Message)
	}

	return loginResp.Token, nil
}

// connectWebSocket 连接到WebSocket网关
func connectWebSocket(token string) error {
	dialer := websocket.DefaultDialer
	url := fmt.Sprintf("%s/ws?token=%s", *gatewayURL, token)
	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("连接WebSocket失败: %v", err)
	}
	wsConn = conn
	return nil
}

// createActor 创建角色
func createActor() (*ActorInfo, error) {
	// 构造创建角色请求 (包含session信息)
	createReq := map[string]interface{}{
		"session":    "", // session通过WebSocket token验证，不需要在body中
		"actor_name": *actorName,
	}

	// 发送WebSocket消息 (消息ID: 2001 - ActorCreateRequest)
	err := sendWebSocketMessage(2001, createReq)
	if err != nil {
		return nil, fmt.Errorf("发送创建角色请求失败: %v", err)
	}

	// 等待响应 (消息ID: 2004 - ActorUseResponse，因为创建后自动使用角色)
	wsConn.SetReadDeadline(time.Now().Add(10 * time.Second))
	var response WebSocketMessage
	err = wsConn.ReadJSON(&response)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 调试信息
	fmt.Printf("收到消息 - ID: %d, 数据: %s\n", response.MsgID, string(response.Data))

	// 检查是否是使用角色的响应
	if response.MsgID != 2004 {
		return nil, fmt.Errorf("收到意外的消息ID: %d, 期望: 2004", response.MsgID)
	}

	// 解析响应
	var createResp map[string]interface{}
	err = json.Unmarshal(response.Data, &createResp)
	if err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if success, ok := createResp["success"].(bool); !ok || !success {
		message := "未知错误"
		if msg, ok := createResp["message"].(string); ok {
			message = msg
		}
		return nil, fmt.Errorf("创建角色失败: %s", message)
	}

	// 解析角色数据
	data, ok := createResp["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("响应数据格式错误")
	}

	actor := &ActorInfo{
		ActorId: getStringValue(data, "actor_id"),
		Name:    getStringValue(data, "name"),
		Level:   int32(getFloatValue(data, "level")),
		Realm:   getStringValue(data, "realm"),
	}

	return actor, nil
}

// getActorList 获取角色列表
func getActorList() ([]ActorInfo, error) {
	// 发送获取角色列表请求 (消息ID: 2003 - ActorListRequest)
	req := map[string]interface{}{}
	err := sendWebSocketMessage(2003, req)
	if err != nil {
		return nil, fmt.Errorf("发送获取角色列表请求失败: %v", err)
	}

	// 等待响应 (消息ID: 2004 - ActorListResponse)
	wsConn.SetReadDeadline(time.Now().Add(10 * time.Second))
	var response WebSocketMessage
	err = wsConn.ReadJSON(&response)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 检查是否是角色列表的响应
	if response.MsgID != 2004 {
		return nil, fmt.Errorf("收到意外的消息ID: %d", response.MsgID)
	}

	// 解析响应
	var listResp map[string]interface{}
	err = json.Unmarshal(response.Data, &listResp)
	if err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if success, ok := listResp["success"].(bool); !ok || !success {
		message := "未知错误"
		if msg, ok := listResp["message"].(string); ok {
			message = msg
		}
		return nil, fmt.Errorf("获取角色列表失败: %s", message)
	}

	// 解析角色数据
	data, ok := listResp["data"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("响应数据格式错误")
	}

	var actors []ActorInfo
	for _, item := range data {
		if actorData, ok := item.(map[string]interface{}); ok {
			actor := ActorInfo{
				ActorId: getStringValue(actorData, "actor_id"),
				Name:    getStringValue(actorData, "name"),
				Level:   int32(getFloatValue(actorData, "level")),
				Realm:   getStringValue(actorData, "realm"),
			}
			actors = append(actors, actor)
		}
	}

	return actors, nil
}

// getBagData 获取背包数据
func getBagData() (map[string]int32, error) {
	// 发送获取背包数据请求 (消息ID: 3001 - GetBagRequest)
	req := map[string]interface{}{}
	err := sendWebSocketMessage(3001, req)
	if err != nil {
		return nil, fmt.Errorf("发送获取背包数据请求失败: %v", err)
	}

	// 等待响应 (消息ID: 3002 - GetBagResponse)
	wsConn.SetReadDeadline(time.Now().Add(10 * time.Second))
	var response WebSocketMessage
	err = wsConn.ReadJSON(&response)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 检查是否是背包数据的响应
	if response.MsgID != 3002 {
		return nil, fmt.Errorf("收到意外的消息ID: %d", response.MsgID)
	}

	// 解析响应
	var bagResp map[string]interface{}
	err = json.Unmarshal(response.Data, &bagResp)
	if err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if success, ok := bagResp["success"].(bool); !ok || !success {
		message := "未知错误"
		if msg, ok := bagResp["message"].(string); ok {
			message = msg
		}
		return nil, fmt.Errorf("获取背包数据失败: %s", message)
	}

	// 解析背包数据
	data, ok := bagResp["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("响应数据格式错误")
	}

	bagData := make(map[string]int32)
	for itemID, count := range data {
		if countFloat, ok := count.(float64); ok {
			bagData[itemID] = int32(countFloat)
		}
	}

	return bagData, nil
}

// 辅助函数
func getStringValue(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func getFloatValue(data map[string]interface{}, key string) float64 {
	if val, ok := data[key].(float64); ok {
		return val
	}
	return 0
}

// sendWebSocketMessage 发送WebSocket消息
func sendWebSocketMessage(msgID int32, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	message := WebSocketMessage{
		MsgID: msgID,
		Data:  jsonData,
	}

	return wsConn.WriteJSON(message)
}

// printHelp 打印帮助信息
func printHelp() {
	fmt.Println("完整游戏流程测试客户端")
	fmt.Println("Usage:")
	fmt.Println("  full_game_test.exe [help] [-auth=auth_url] [-gateway=gateway_url] [-username=username] [-password=password] [-actor=actor_name]")
	fmt.Println("\nOptions:")
	fmt.Println("  help                Show this help message")
	fmt.Println("  -auth=auth_url      Auth server URL (default: http://localhost:8080)")
	fmt.Println("  -gateway=gateway_url Gateway WebSocket URL (default: ws://localhost:8081)")
	fmt.Println("  -username=username  Username (default: fulltestuser)")
	fmt.Println("  -password=password  Password (default: fulltestpass)")
	fmt.Println("  -actor=actor_name   Actor name (default: fulltestactor)")
	fmt.Println("\nExamples:")
	fmt.Println("  full_game_test.exe help")
	fmt.Println("  full_game_test.exe")
	fmt.Println("  full_game_test.exe -username=newuser -password=newpass -actor=myhero")
}
