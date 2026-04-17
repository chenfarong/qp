package base

import (
	"context"
	"zagame/inside/gamelogic/grpc"
	pb "zagame/pb/golang/gamelogic"
	"zagame/proto"
)

// Handler 基础处理器
type Handler struct {
	Service *Service
}

// Login 登录请求处理
func (h *Handler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	return h.Service.Login(ctx, req)
}

// GetRoleInfo 获取角色信息请求处理
func (h *Handler) GetRoleInfo(ctx context.Context, req *pb.GetRoleInfoRequest) (*pb.GetRoleInfoResponse, error) {
	return h.Service.GetRoleInfo(ctx, req)
}

// GetGameMoney 获取游戏货币请求处理
func (h *Handler) GetGameMoney(ctx context.Context, req *pb.GetGameMoneyRequest) (*pb.GetGameMoneyResponse, error) {
	return h.Service.GetGameMoney(ctx, req)
}

// NewHandler 创建基础处理器
func NewHandler() *Handler {
	return &Handler{
		Service: NewService(),
	}
}

// RegisterHandlers 注册消息处理器
func (h *Handler) RegisterHandlers(router *grpc.Router) {
	// 基础消息
	router.RegisterHandler(proto.MSG_LoginRequest, h.handleLoginRequest)
	router.RegisterHandler(proto.MSG_GetRoleInfoRequest, h.handleGetRoleInfoRequest)
	router.RegisterHandler(proto.MSG_GetGameMoneyRequest, h.handleGetGameMoneyRequest)
}

// handleLoginRequest 处理登录请求
func (h *Handler) handleLoginRequest(ctx context.Context, session string, messageContent []byte) ([]byte, error) {
	req := &pb.LoginRequest{}
	if err := grpc.Unmarshal(messageContent, req); err != nil {
		return nil, err
	}

	resp, err := h.Login(ctx, req)
	if err != nil {
		return nil, err
	}

	return grpc.Marshal(resp)
}

// handleGetRoleInfoRequest 处理获取角色信息请求
func (h *Handler) handleGetRoleInfoRequest(ctx context.Context, session string, messageContent []byte) ([]byte, error) {
	req := &pb.GetRoleInfoRequest{}
	if err := grpc.Unmarshal(messageContent, req); err != nil {
		return nil, err
	}

	resp, err := h.GetRoleInfo(ctx, req)
	if err != nil {
		return nil, err
	}

	return grpc.Marshal(resp)
}

// handleGetGameMoneyRequest 处理获取游戏货币请求
func (h *Handler) handleGetGameMoneyRequest(ctx context.Context, session string, messageContent []byte) ([]byte, error) {
	req := &pb.GetGameMoneyRequest{}
	if err := grpc.Unmarshal(messageContent, req); err != nil {
		return nil, err
	}

	resp, err := h.GetGameMoney(ctx, req)
	if err != nil {
		return nil, err
	}

	return grpc.Marshal(resp)
}
