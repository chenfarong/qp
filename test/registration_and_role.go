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

	pb "zagame/pb/golang/gamelogic"
	"zagame/proto"

	"github.com/gorilla/websocket"
)

// 全局变量存储命令行参数
var (
	authServerURL = flag.String("auth", "http://localhost:8080", "验证服务器URL")
	gameServerURL = flag.String("game", "http://localhost:8081", "游戏服务器URL")
	username      = flag.String("username", "testuser", "用户名")
	password      = flag.String("password", "testpassword", "密码")
	actorName     = flag.String("actor", "test_actor", "角色名称")
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

// ActorCreateRequest 创建角色请求结构
type ActorCreateRequest struct {
	Name string `json:"name"`
}

// ActorCreateResponse 创建角色响应结构
type ActorCreateResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	Data    ActorInfo `json:"data"`
}

// ActorInfo 角色信息结构
type ActorInfo struct {
	ActorId string `json:"actor_id"`
	Name    string `json:"name"`
	Level   int32  `json:"level"`
	Realm   string `json:"realm"`
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

	fmt.Println("=== 注册账号和创建角色测试 ===")
	fmt.Printf("验证服务器: %s\n", *authServerURL)
	fmt.Printf("游戏服务器: %s\n", *gameServerURL)
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

	// 步骤3: 创建角色
	fmt.Println("\n=== 步骤3: 创建角色 ===")
	actor, err := createActor(token)
	if err != nil {
		fmt.Printf("创建角色失败: %v\n", err)
		return
	}
	fmt.Println("创建角色成功!")

	// 显示结果
	fmt.Println("\n=== 测试完成 ===")
	fmt.Printf("用户名: %s\n", *username)
	fmt.Printf("角色ID: %s\n", actor.ActorId)
	fmt.Printf("角色名称: %s\n", actor.Name)
	fmt.Printf("角色等级: %d\n", actor.Level)
	fmt.Printf("所在服区: %s\n", actor.Realm)
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
	resp, err := http.Post(*authServerURL+"/register", "application/json", bytes.NewBuffer(jsonData))
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
	resp, err := http.Post(*authServerURL+"/login", "application/json", bytes.NewBuffer(jsonData))
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

// createActor 创建角色
func createActor(token string) (*ActorInfo, error) {
	// 连接到WebSocket服务器
	dialer := websocket.DefaultDialer
	url := fmt.Sprintf("ws://localhost:8081/ws?token=%s", token)
	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("连接WebSocket失败: %v", err)
	}
	defer conn.Close()
	wsConn = conn

	// 构造创建角色请求
	req := pb.ActorCreateRequest{
		Name: *actorName,
	}

	// 发送WebSocket消息
	err = sendWebSocketMessage(proto.MessageIDActorCreateRequest, req)
	if err != nil {
		return nil, fmt.Errorf("发送创建角色请求失败: %v", err)
	}

	// 等待响应
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	var response WebSocketMessage
	err = conn.ReadJSON(&response)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析响应
	var createResp pb.ActorCreateResponse
	err = json.Unmarshal(response.Data, &createResp)
	if err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if !createResp.Success {
		return nil, fmt.Errorf("创建角色失败: %s", createResp.Message)
	}

	// 转换响应格式
	actor := &ActorInfo{
		ActorId: createResp.Data.ActorId,
		Name:    createResp.Data.Name,
		Level:   createResp.Data.Level,
		Realm:   createResp.Data.Realm,
	}

	return actor, nil
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
	fmt.Println("注册账号和创建角色测试客户端")
	fmt.Println("Usage:")
	fmt.Println("  registration_and_role.exe [help] [-auth=auth_url] [-game=game_url] [-username=username] [-password=password] [-actor=actor_name]")
	fmt.Println("\nOptions:")
	fmt.Println("  help                Show this help message")
	fmt.Println("  -auth=auth_url      Auth server URL (default: http://localhost:8080)")
	fmt.Println("  -game=game_url      Game server URL (default: http://localhost:8081)")
	fmt.Println("  -username=username  Username (default: testuser)")
	fmt.Println("  -password=password  Password (default: testpassword)")
	fmt.Println("  -actor=actor_name   Actor name (default: test_actor)")
	fmt.Println("\nExamples:")
	fmt.Println("  registration_and_role.exe help")
	fmt.Println("  registration_and_role.exe")
	fmt.Println("  registration_and_role.exe -username=newuser -password=newpass -actor=myhero")
}
