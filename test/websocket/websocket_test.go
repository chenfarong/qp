package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	proto "github.com/aoyo/qp/proto"
	protoutil "github.com/aoyo/qp/pkg/proto"
	"github.com/gorilla/websocket"
)

func TestWebSocketProtobuf(t *testing.T) {
	// 启动Gateway服务（在实际测试中，应该先启动服务）
	// TODO: 启动Gateway服务

	// 连接WebSocket
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}
	log.Printf("Connecting to %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// 测试注册请求
	testRegister(t, conn)

	// 测试登录请求
	testLogin(t, conn)
}

func testRegister(t *testing.T, conn *websocket.Conn) {
	// 创建注册请求
	req := &proto.Message{
		Type: proto.MessageType_MSG_TYPE_AUTH_REGISTER,
		Data: &proto.Message_AuthRegister{
			AuthRegister: &proto.AuthRegisterRequest{
				Username: "testuser",
				Password: "password123",
				Email:    "test@example.com",
				Nickname: "Test User",
			},
		},
	}

	// 序列化请求
	data, err := protoutil.Serialize(req)
	if err != nil {
		t.Fatalf("Failed to serialize request: %v", err)
	}

	// 发送请求
	err = conn.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		t.Fatalf("Failed to write message: %v", err)
	}

	// 接收响应
	messageType, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read message: %v", err)
	}

	// 验证消息类型
	if messageType != websocket.BinaryMessage {
		t.Fatalf("Expected binary message, got %d", messageType)
	}

	// 反序列化响应
	resp, err := protoutil.Deserialize(message)
	if err != nil {
		t.Fatalf("Failed to deserialize response: %v", err)
	}

	// 验证响应类型
	if resp.Type != proto.MessageType_MSG_TYPE_RESPONSE {
		t.Fatalf("Expected response message, got %v", resp.Type)
	}

	// 验证响应数据
	response := resp.GetResponse()
	if response == nil {
		t.Fatalf("Expected response data, got nil")
	}

	if response.Code != 200 {
		t.Fatalf("Expected code 200, got %d", response.Code)
	}

	if response.Message != "success" {
		t.Fatalf("Expected message 'success', got '%s'", response.Message)
	}

	// 验证认证响应
	authResp := response.GetAuthResponse()
	if authResp == nil {
		t.Fatalf("Expected auth response, got nil")
	}

	if authResp.Token == "" {
		t.Fatalf("Expected token, got empty")
	}

	if authResp.UserID == "" {
		t.Fatalf("Expected user ID, got empty")
	}

	if authResp.Username != "testuser" {
		t.Fatalf("Expected username 'testuser', got '%s'", authResp.Username)
	}

	if authResp.Nickname != "Test User" {
		t.Fatalf("Expected nickname 'Test User', got '%s'", authResp.Nickname)
	}

	log.Println("Register test passed!")
}

func testLogin(t *testing.T, conn *websocket.Conn) {
	// 创建登录请求
	req := &proto.Message{
		Type: proto.MessageType_MSG_TYPE_AUTH_LOGIN,
		Data: &proto.Message_AuthLogin{
			AuthLogin: &proto.AuthLoginRequest{
				Username: "testuser",
				Password: "password123",
			},
		},
	}

	// 序列化请求
	data, err := protoutil.Serialize(req)
	if err != nil {
		t.Fatalf("Failed to serialize request: %v", err)
	}

	// 发送请求
	err = conn.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		t.Fatalf("Failed to write message: %v", err)
	}

	// 接收响应
	messageType, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read message: %v", err)
	}

	// 验证消息类型
	if messageType != websocket.BinaryMessage {
		t.Fatalf("Expected binary message, got %d", messageType)
	}

	// 反序列化响应
	resp, err := protoutil.Deserialize(message)
	if err != nil {
		t.Fatalf("Failed to deserialize response: %v", err)
	}

	// 验证响应类型
	if resp.Type != proto.MessageType_MSG_TYPE_RESPONSE {
		t.Fatalf("Expected response message, got %v", resp.Type)
	}

	// 验证响应数据
	response := resp.GetResponse()
	if response == nil {
		t.Fatalf("Expected response data, got nil")
	}

	if response.Code != 200 {
		t.Fatalf("Expected code 200, got %d", response.Code)
	}

	if response.Message != "success" {
		t.Fatalf("Expected message 'success', got '%s'", response.Message)
	}

	// 验证认证响应
	authResp := response.GetAuthResponse()
	if authResp == nil {
		t.Fatalf("Expected auth response, got nil")
	}

	if authResp.Token == "" {
		t.Fatalf("Expected token, got empty")
	}

	if authResp.UserID == "" {
		t.Fatalf("Expected user ID, got empty")
	}

	if authResp.Username != "testuser" {
		t.Fatalf("Expected username 'testuser', got '%s'", authResp.Username)
	}

	log.Println("Login test passed!")
}
