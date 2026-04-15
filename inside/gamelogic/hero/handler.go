package hero

import (
	"context"
	pb "zgame/pb/golang/gamelogic"
)

// Handler 英雄处理器
type Handler struct {
	Service *Service
}

// AddHero 添加英雄请求处理
func (h *Handler) AddHero(ctx context.Context, req *pb.AddHeroRequest) (*pb.AddHeroResponse, error) {
	return h.Service.AddHero(ctx, req)
}

// RemoveHero 移除英雄请求处理
func (h *Handler) RemoveHero(ctx context.Context, req *pb.RemoveHeroRequest) (*pb.RemoveHeroResponse, error) {
	return h.Service.RemoveHero(ctx, req)
}

// GetHeroes 获取英雄列表请求处理
func (h *Handler) GetHeroes(ctx context.Context, req *pb.GetHeroesRequest) (*pb.GetHeroesResponse, error) {
	return h.Service.GetHeroes(ctx, req)
}
