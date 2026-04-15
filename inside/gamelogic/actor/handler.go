package actor

import (
	"context"
	pb "zagame/pb/golang/gamelogic"
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
