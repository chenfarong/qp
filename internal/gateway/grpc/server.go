package grpc

import (
	"context"

	"github.com/aoyo/qp/pkg/proto/gateway"
)

// GatewayServer 网关服务gRPC服务器
type GatewayServer struct {
	gateway.UnimplementedGatewayServiceServer
	messagePusher MessagePusher
}

// MessagePusher 消息推送接口
type MessagePusher interface {
	PushMessage(userID uint32, messageType string, messageData []byte, title, content string) (bool, error)
	BroadcastMessage(messageType string, messageData []byte, title, content string) (int, error)
	GetConnectedUsers() []uint32
}

// NewGatewayServer 创建网关服务gRPC服务器实例
func NewGatewayServer(messagePusher MessagePusher) *GatewayServer {
	return &GatewayServer{
		messagePusher: messagePusher,
	}
}

// PushMessage 推送消息给指定用户
func (s *GatewayServer) PushMessage(ctx context.Context, req *gateway.PushMessageRequest) (*gateway.PushMessageResponse, error) {
	// 调用消息推送接口
	success, err := s.messagePusher.PushMessage(
		req.UserId,
		req.MessageType,
		req.MessageData,
		req.Title,
		req.Content,
	)

	if err != nil {
		return &gateway.PushMessageResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 获取当前连接的客户端数量
	connectedUsers := s.messagePusher.GetConnectedUsers()

	return &gateway.PushMessageResponse{
		Success:         success,
		ConnectedClients: int32(len(connectedUsers)),
	}, nil
}

// BroadcastMessage 广播消息给所有用户
func (s *GatewayServer) BroadcastMessage(ctx context.Context, req *gateway.BroadcastMessageRequest) (*gateway.BroadcastMessageResponse, error) {
	// 调用消息推送接口
	broadcastCount, err := s.messagePusher.BroadcastMessage(
		req.MessageType,
		req.MessageData,
		req.Title,
		req.Content,
	)

	if err != nil {
		return &gateway.BroadcastMessageResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &gateway.BroadcastMessageResponse{
		Success:        true,
		BroadcastCount: int32(broadcastCount),
	}, nil
}

// GetConnectedUsers 获取当前连接的用户列表
func (s *GatewayServer) GetConnectedUsers(ctx context.Context, req *gateway.GetConnectedUsersRequest) (*gateway.GetConnectedUsersResponse, error) {
	// 调用消息推送接口
	connectedUsers := s.messagePusher.GetConnectedUsers()

	return &gateway.GetConnectedUsersResponse{
		UserIds:    connectedUsers,
		TotalUsers: int32(len(connectedUsers)),
	}, nil
}
