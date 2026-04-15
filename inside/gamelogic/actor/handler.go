package actor

import (
	"context"
	pb "zgame/pb/golang/gamelogic"
)

// Handler 角色处理器
type Handler struct {
	Service *Service
}

// ActorCreate 创建角色请求处理
func (h *Handler) ActorCreate(ctx context.Context, req *pb.ActorCreateRequest) (*pb.ActorCreateResponse, error) {
	return h.Service.ActorCreate(ctx, req)
}

// ActorUse 使用角色请求处理
func (h *Handler) ActorUse(ctx context.Context, req *pb.ActorUseRequest) error {
	return h.Service.ActorUse(ctx, req)
}
