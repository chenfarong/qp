package handler

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"zagame/outside/gateway/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
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
	Cancel     context.CancelFunc
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

	// 如果服务器已存在，先删除旧的连接
	if oldServer, exists := serverManager.servers[serverInfo.ServerId]; exists {
		log.Printf("服务器 %s (%s) 已存在，删除旧连接", oldServer.ServerName, oldServer.ServerID)
		removeServer(oldServer)
	}

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

	// 创建监控上下文
	monitorCtx, cancel := context.WithCancel(context.Background())

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
		Cancel:     cancel,
	}

	serverManager.servers[serverInfo.ServerId] = server

	// 更新消息ID映射
	for msgID := serverInfo.StartMsgId; msgID <= serverInfo.EndMsgId; msgID++ {
		serverManager.msgIDMap[msgID] = serverInfo.ServerId
	}

	// 启动连接状态监控
	go monitorConnection(monitorCtx, server)

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
		ClientIp:       req.ClientIp,
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

// monitorConnection 监控服务器连接状态
func monitorConnection(ctx context.Context, server *ServerInfo) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			state := server.Conn.GetState()
			if state == connectivity.Shutdown || state == connectivity.TransientFailure {
				log.Printf("服务器 %s (%s) 连接断开，状态: %s", server.ServerName, server.ServerID, state)
				removeServer(server)
				return
			}
		}
	}
}

// removeServer 删除服务器注册
func removeServer(server *ServerInfo) {
	serverManager.mu.Lock()
	defer serverManager.mu.Unlock()

	// 停止监控
	if server.Cancel != nil {
		server.Cancel()
	}

	// 关闭连接
	if server.Conn != nil {
		server.Conn.Close()
	}

	// 删除消息ID映射
	for msgID := server.StartMsgID; msgID <= server.EndMsgID; msgID++ {
		delete(serverManager.msgIDMap, msgID)
	}

	// 删除服务器信息
	delete(serverManager.servers, server.ServerID)

	log.Printf("服务器 %s (%s) 已删除注册", server.ServerName, server.ServerID)
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
