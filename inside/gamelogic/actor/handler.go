package actor

import (
	"context"
	"encoding/json"

	"zagame/common/logger"
	"zagame/inside/gamelogic/grpc"
	"zagame/inside/gamelogic/session"
	pb "zagame/pb/golang/gamelogic"
	"zagame/proto"
)

// Handler 角色处理器
type Handler struct {
	Service *Service
}

// ActorCreate 创建角色请求处理
func (h *Handler) ActorCreate(ctx context.Context, req *pb.ActorCreateRequest) (*pb.ActorUseResponse, error) {
	return h.Service.ActorCreate(ctx, req)
}

// ActorUse 使用角色请求处理
func (h *Handler) ActorUse(ctx context.Context, req *pb.ActorUseRequest) (*pb.ActorUseResponse, error) {
	return h.Service.ActorUse(ctx, req)
}

// ActorUseWithName 使用角色请求处理（通过名称）
func (h *Handler) ActorUseWithName(ctx context.Context, req *pb.ActorUseWithNameRequest) (*pb.ActorUseResponse, error) {
	return h.Service.ActorUseWithName(ctx, req)
}

// NewHandler 创建角色处理器
func NewHandler() *Handler {
	return &Handler{
		Service: NewService(),
	}
}

// RegisterHandlers 注册消息处理器
func (h *Handler) RegisterHandlers(router *grpc.Router) {
	// 角色消息
	router.RegisterHandler(proto.MSG_ActorCreateRequest, h.handleActorCreateRequest)
	router.RegisterHandler(proto.MSG_ActorUseRequest, h.handleActorUseRequest)
	router.RegisterHandler(proto.MSG_ActorUseWithNameRequest, h.handleActorUseWithNameRequest)
}

// handleActorCreateRequest 处理创建角色请求
func (h *Handler) handleActorCreateRequest(ctx context.Context, sess string, messageContent []byte) ([]byte, error) {
	req := &pb.ActorCreateRequest{}
	if err := grpc.Unmarshal(messageContent, req); err != nil {
		return nil, err
	}

	// 打印请求日志
	reqJSON, err := json.Marshal(req)
	if err == nil {
		logger.Debugf("角色创建请求: %s", string(reqJSON))
	}

	resp, err := h.ActorCreate(ctx, req)
	if err != nil {
		return nil, err
	}

	// 打印响应日志
	respJSON, err := json.Marshal(resp)
	if err == nil {
		logger.Debugf("角色创建响应: %s", string(respJSON))
	}

	// 如果创建成功，更新sessionActor映射
	if resp.Err != nil && resp.Err.ErrCode == 0 && resp.Data != nil {
		session.SetActorInfo(sess, session.ActorInfo{
			ActorID:   resp.Data.ActorId,
			ActorName: resp.Data.Name,
		})
		logger.Infof("会话 %s 关联角色: %s(%s)", sess, resp.Data.Name, resp.Data.ActorId)
	}

	return grpc.Marshal(resp)
}

// handleActorUseRequest 处理使用角色请求
func (h *Handler) handleActorUseRequest(ctx context.Context, sess string, messageContent []byte) ([]byte, error) {
	req := &pb.ActorUseRequest{}
	if err := grpc.Unmarshal(messageContent, req); err != nil {
		return nil, err
	}

	// 打印请求日志
	reqJSON, err := json.Marshal(req)
	if err == nil {
		logger.Debugf("角色使用请求: %s", string(reqJSON))
	}

	resp, err := h.ActorUse(ctx, req)
	if err != nil {
		return nil, err
	}

	// 打印响应日志
	respJSON, err := json.Marshal(resp)
	if err == nil {
		logger.Debugf("角色使用响应: %s", string(respJSON))
	}

	// 如果使用成功，更新sessionActor映射
	if resp.Err != nil && resp.Err.ErrCode == 0 && resp.Data != nil {
		session.SetActorInfo(sess, session.ActorInfo{
			ActorID:   resp.Data.ActorId,
			ActorName: resp.Data.Name,
		})
		logger.Infof("会话 %s 关联角色: %s(%s)", sess, resp.Data.Name, resp.Data.ActorId)
	}

	return grpc.Marshal(resp)
}

// handleActorUseWithNameRequest 处理使用角色请求（通过名称）
func (h *Handler) handleActorUseWithNameRequest(ctx context.Context, sess string, messageContent []byte) ([]byte, error) {
	req := &pb.ActorUseWithNameRequest{}
	if err := grpc.Unmarshal(messageContent, req); err != nil {
		return nil, err
	}

	resp, err := h.ActorUseWithName(ctx, req)
	if err != nil {
		return nil, err
	}

	// 如果使用成功，更新sessionActor映射
	if resp.Err != nil && resp.Err.ErrCode == 0 && resp.Data != nil {
		session.SetActorInfo(sess, session.ActorInfo{
			ActorID:   resp.Data.ActorId,
			ActorName: resp.Data.Name,
		})
		logger.Infof("会话 %s 关联角色: %s(%s)", sess, resp.Data.Name, resp.Data.ActorId)
	}

	return grpc.Marshal(resp)
}
