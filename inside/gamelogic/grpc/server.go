package grpc

import (
	"context"
	"encoding/json"
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
	// 打印收到的消息
	log.Printf("收到转发消息: messageID=%d, session=%s\n", req.MessageId, req.Session)
	
	// 尝试将消息内容解析为JSON并打印
	if len(req.MessageContent) > 0 {
		var msgContent interface{}
		err := json.Unmarshal(req.MessageContent, &msgContent)
		if err != nil {
			log.Printf("消息内容解析失败: %v, 原始内容: %s\n", err, string(req.MessageContent))
		} else {
			jsonContent, err := json.MarshalIndent(msgContent, "  ", "  ")
			if err != nil {
				log.Printf("消息内容序列化失败: %v\n", err)
			} else {
				log.Printf("消息内容: %s\n", string(jsonContent))
			}
		}
	}

	responseContent, err := s.router.HandleMessage(ctx, req.MessageId, req.Session, req.MessageContent)
	if err != nil {
		log.Printf("处理消息失败: messageID=%d, error=%v\n", req.MessageId, err)
		return &gateway.ForwardMessageResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 尝试将响应内容解析为JSON并打印
	if len(responseContent) > 0 {
		var respContent interface{}
		err := json.Unmarshal(responseContent, &respContent)
		if err != nil {
			log.Printf("响应内容解析失败: %v, 原始内容: %s\n", err, string(responseContent))
		} else {
			jsonContent, err := json.MarshalIndent(respContent, "  ", "  ")
			if err != nil {
				log.Printf("响应内容序列化失败: %v\n", err)
			} else {
				log.Printf("响应内容: %s\n", string(jsonContent))
			}
		}
	}

	log.Printf("处理消息成功: messageID=%d\n", req.MessageId)

	return &gateway.ForwardMessageResponse{
		Success:         true,
		Message:         "处理成功",
		ResponseContent: responseContent,
	}, nil
}
