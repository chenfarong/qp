package handler

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"zgame/internet/gateway/proto"

	"google.golang.org/grpc"
)

// 服务器信息存储
type ServerInfo struct {
	ServerID   string
	ServerName string
	StartMsgID int32
	EndMsgID   int32
	Address    string
	Port       int32
	Client     proto.GatewayServiceClient
	Conn       *grpc.ClientConn
}

// 服务器管理器
type ServerManager struct {
	servers  map[string]*ServerInfo
	msgIDMap map[int32]string // 消息ID到服务器ID的映射
	mu       sync.RWMutex
}

// 全局服务器管理器
var serverManager = &ServerManager{
	servers:  make(map[string]*ServerInfo),
	msgIDMap: make(map[int32]string),
}

// GatewayServer gRPC服务器实现
type GatewayServer struct {
	proto.UnimplementedGatewayServiceServer
}

// RegisterServer 注册服务器
func (s *GatewayServer) RegisterServer(ctx context.Context, req *proto.RegisterServerRequest) (*proto.RegisterServerResponse, error) {
	serverInfo := req.ServerInfo

	// 检查服务器是否已存在
	serverManager.mu.Lock()
	defer serverManager.mu.Unlock()

	// 检查消息ID范围是否冲突
	for msgID := serverInfo.StartMsgId; msgID <= serverInfo.EndMsgId; msgID++ {
		if _, exists := serverManager.msgIDMap[msgID]; exists {
			return &proto.RegisterServerResponse{
				Success: false,
				Message: fmt.Sprintf("消息ID %d 已被注册", msgID),
			}, nil
		}
	}

	// 连接到注册的服务器
	address := fmt.Sprintf("%s:%d", serverInfo.Address, serverInfo.Port)
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return &proto.RegisterServerResponse{
			Success: false,
			Message: fmt.Sprintf("无法连接到服务器: %v", err),
		}, nil
	}

	// 创建客户端
	client := proto.NewGatewayServiceClient(conn)

	// 存储服务器信息
	server := &ServerInfo{
		ServerID:   serverInfo.ServerId,
		ServerName: serverInfo.ServerName,
		StartMsgID: serverInfo.StartMsgId,
		EndMsgID:   serverInfo.EndMsgId,
		Address:    serverInfo.Address,
		Port:       serverInfo.Port,
		Client:     client,
		Conn:       conn,
	}

	serverManager.servers[serverInfo.ServerId] = server

	// 更新消息ID映射
	for msgID := serverInfo.StartMsgId; msgID <= serverInfo.EndMsgId; msgID++ {
		serverManager.msgIDMap[msgID] = serverInfo.ServerId
	}

	log.Printf("服务器 %s (%s) 注册成功，消息ID范围: %d-%d",
		serverInfo.ServerName, serverInfo.ServerId, serverInfo.StartMsgId, serverInfo.EndMsgId)

	return &proto.RegisterServerResponse{
		Success: true,
		Message: "服务器注册成功",
	}, nil
}

// ForwardMessage 转发消息
func (s *GatewayServer) ForwardMessage(ctx context.Context, req *proto.ForwardMessageRequest) (*proto.ForwardMessageResponse, error) {
	messageID := req.MessageId
	session := req.Session
	messageContent := req.MessageContent

	// 查找处理该消息的服务器
	serverManager.mu.RLock()
	serverID, exists := serverManager.msgIDMap[messageID]
	if !exists {
		serverManager.mu.RUnlock()
		return &proto.ForwardMessageResponse{
			Success: false,
			Message: fmt.Sprintf("消息ID %d 未找到对应的服务器", messageID),
		}, nil
	}

	server, exists := serverManager.servers[serverID]
	if !exists {
		serverManager.mu.RUnlock()
		return &proto.ForwardMessageResponse{
			Success: false,
			Message: fmt.Sprintf("服务器 %s 不存在", serverID),
		}, nil
	}
	serverManager.mu.RUnlock()

	// 转发消息到目标服务器
	response, err := server.Client.ForwardMessage(ctx, &proto.ForwardMessageRequest{
		MessageId:      messageID,
		Session:        session,
		MessageContent: messageContent,
	})

	if err != nil {
		return &proto.ForwardMessageResponse{
			Success: false,
			Message: fmt.Sprintf("转发消息失败: %v", err),
		}, nil
	}

	return response, nil
}

// StartGRPCServer 启动gRPC服务器
func StartGRPCServer(port int) error {
	// 创建gRPC服务器
	server := grpc.NewServer()

	// 注册GatewayService
	proto.RegisterGatewayServiceServer(server, &GatewayServer{})

	// 监听端口
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("无法监听端口 %d: %v", port, err)
	}

	log.Printf("gRPC服务器启动，监听端口 %d", port)

	// 启动服务器
	return server.Serve(listener)
}
