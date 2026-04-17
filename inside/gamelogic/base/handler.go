package base

import (
	"context"
	"zagame/inside/gamelogic/common"
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
func (h *Handler) RegisterHandlers(router *grpc.Router, handler common.Handler) {
	// 基础消息
	router.RegisterHandler(proto.MSG_LoginRequest, h, "Login",
		func() interface{} { return &pb.LoginRequest{} },
		func() interface{} { return &pb.LoginResponse{} })
	router.RegisterHandler(proto.MSG_GetRoleInfoRequest, h, "GetRoleInfo",
		func() interface{} { return &pb.GetRoleInfoRequest{} },
		func() interface{} { return &pb.GetRoleInfoResponse{} })
	router.RegisterHandler(proto.MSG_GetGameMoneyRequest, h, "GetGameMoney",
		func() interface{} { return &pb.GetGameMoneyRequest{} },
		func() interface{} { return &pb.GetGameMoneyResponse{} })
}
