package base

import (
	"context"
	pb "zgame/pb/golang/gamelogic"
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
