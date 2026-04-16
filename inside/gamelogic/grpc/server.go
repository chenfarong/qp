package grpc

import (
	"context"
	"log"

	gateway "zagame/inside/gamelogic/grpc/gateway"
)

// GatewayServer Gateway服务器实现
type GatewayServer struct {
	gateway.UnimplementedGatewayServiceServer
	router *Router
}

// NewGatewayServer 创建Gateway服务器
func NewGatewayServer(router *Router) *GatewayServer {
	return &GatewayServer{
		router: router,
	}
}

// RegisterServer 注册服务器
func (s *GatewayServer) RegisterServer(ctx context.Context, req *gateway.RegisterServerRequest) (*gateway.RegisterServerResponse, error) {
	log.Printf("收到注册服务器请求: serverID=%s, serverName=%s\n", req.ServerInfo.ServerId, req.ServerInfo.ServerName)

	resp := &gateway.RegisterServerResponse{
		Success: true,
		Message: "注册成功",
	}

	return resp, nil
}

// ForwardMessage 转发消息
func (s *GatewayServer) ForwardMessage(ctx context.Context, req *gateway.ForwardMessageRequest) (*gateway.ForwardMessageResponse, error) {
	log.Printf("收到转发消息: messageID=%d, session=%s\n", req.MessageId, req.Session)

	responseContent, err := s.router.HandleMessage(ctx, req.MessageId, req.Session, req.MessageContent)
	if err != nil {
		log.Printf("处理消息失败: messageID=%d, error=%v\n", req.MessageId, err)
		return &gateway.ForwardMessageResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	log.Printf("处理消息成功: messageID=%d\n", req.MessageId)

	return &gateway.ForwardMessageResponse{
		Success:         true,
		Message:         "处理成功",
		ResponseContent: responseContent,
	}, nil
}
