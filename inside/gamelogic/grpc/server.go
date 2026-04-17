package grpc

import (
	"context"
	"encoding/json"

	"zagame/common/logger"
	"zagame/inside/gamelogic/session"
	"zagame/proto"

	gateway "zagame/inside/gamelogic/gateway"
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
	logger.Infof("收到注册服务器请求: serverID=%s, serverName=%s", req.ServerInfo.ServerId, req.ServerInfo.ServerName)

	resp := &gateway.RegisterServerResponse{
		Success: true,
		Message: "注册成功",
	}

	return resp, nil
}

// ForwardMessage 转发消息
func (s *GatewayServer) ForwardMessage(ctx context.Context, req *gateway.ForwardMessageRequest) (*gateway.ForwardMessageResponse, error) {
	// 打印收到的消息
	logger.Infof("收到转发消息: messageID=%d, messageName=%s, session=%s, clientIP=%s", req.MessageId, proto.GetMessageName(req.MessageId), req.Session, req.ClientIp)

	// 尝试将消息内容解析为JSON并打印
	if len(req.MessageContent) > 0 {
		var msgContent interface{}
		err := json.Unmarshal(req.MessageContent, &msgContent)
		if err != nil {
			logger.Errorf("消息内容解析失败: %v, 原始内容: %s", err, string(req.MessageContent))
		} else {
			jsonContent, err := json.Marshal(msgContent)
			if err != nil {
				logger.Errorf("消息内容序列化失败: %v", err)
			} else {
				//logger.Debugf("收到消息内容: %s", string(jsonContent))
				_ = jsonContent
			}
		}
	}

	// 从sessionActor映射中获取actor信息
	actorInfo, exists := session.GetActorInfo(req.Session)

	// 将actor信息和客户端IP添加到上下文中
	if exists {
		ctx = context.WithValue(ctx, "actor_id", actorInfo.ActorID)
		ctx = context.WithValue(ctx, "actor_name", actorInfo.ActorName)
		//		logger.Debugf("会话 %s 关联的角色: %s(%s), 客户端IP: %s", req.Session, actorInfo.ActorName, actorInfo.ActorID, req.ClientIp)
	} else {
		// 即使没有角色信息，也添加客户端IP
		ctx = context.WithValue(ctx, "client_ip", req.ClientIp)
		//		logger.Debugf("会话 %s 客户端IP: %s", req.Session, req.ClientIp)
	}

	responseContent, err := s.router.HandleMessage(ctx, req.MessageId, req.Session, req.MessageContent)
	if err != nil {
		logger.Errorf("处理消息失败: messageID=%d, messageName=%s, error=%v", req.MessageId, proto.GetMessageName(req.MessageId), err)
		return &gateway.ForwardMessageResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 尝试将响应内容解析为JSON并打印
	if responseContent != nil && len(*responseContent) > 0 {
		var respContent interface{}
		err := json.Unmarshal(*responseContent, &respContent)
		if err != nil {
			logger.Errorf("响应内容解析失败: %v, 原始内容: %s", err, string(*responseContent))
		} else {
			jsonContent, err := json.Marshal(respContent)
			if err != nil {
				logger.Errorf("响应内容序列化失败: %v", err)
			} else {
				logger.Debugf("响应内容: %s", string(jsonContent))
			}
		}

		// 检查是否需要更新sessionActor映射
		if req.MessageId == proto.MSG_ActorCreateRequest || req.MessageId == proto.MSG_ActorUseRequest || req.MessageId == proto.MSG_ActorUseWithNameRequest {
			if respMap, ok := respContent.(map[string]interface{}); ok {
				if errInfo, ok := respMap["err"].(map[string]interface{}); ok {
					if errCode, ok := errInfo["errCode"].(float64); ok && errCode == 0 {
						if data, ok := respMap["data"].(map[string]interface{}); ok {
							if actorId, ok := data["actorId"].(string); ok {
								if actorName, ok := data["name"].(string); ok {
									session.SetActorInfo(req.Session, session.ActorInfo{
										ActorID:   actorId,
										ActorName: actorName,
									})
									logger.Infof("会话 %s 关联角色: %s(%s)", req.Session, actorName, actorId)
								}
							}
						}
					}
				}
			}
		}
	}

	//	logger.Infof("处理消息成功: messageID=%d, messageName=%s, session=%s, clientIP=%s", req.MessageId, proto.GetMessageName(req.MessageId), req.Session, req.ClientIp)

	return &gateway.ForwardMessageResponse{
		Success:         true,
		Message:         "处理成功",
		ResponseContent: *responseContent,
	}, nil
}
