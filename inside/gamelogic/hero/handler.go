package hero

import (
	"context"
	pb "zagame/pb/golang/gamelogic"
)

// Handler 英雄处理器
type Handler struct {
	Service *Service
}

// GetHeroes 获取英雄列表请求处理
func (h *Handler) GetHeroes(ctx context.Context, req *pb.GetHeroesRequest) (*pb.GetHeroesResponse, error) {
	return h.Service.GetHeroes(ctx, req)
}

// RecruitHero 招募英雄请求处理
func (h *Handler) RecruitHero(ctx context.Context, req *pb.RecruitHeroRequest) (*pb.RecruiHeroesResponse, error) {
	return h.Service.RecruitHero(ctx, req)
}

// UpStarHero 英雄升星请求处理
func (h *Handler) UpStarHero(ctx context.Context, req *pb.UpStarHeroRequest) (*pb.UpStarHeroesResponse, error) {
	return h.Service.UpStarHero(ctx, req)
}

// OpenSkillHero 英雄技能开启请求处理
func (h *Handler) OpenSkillHero(ctx context.Context, req *pb.OpenSkillHeroRequest) (*pb.OpenSkillHeroesResponse, error) {
	return h.Service.OpenSkillHero(ctx, req)
}

// UpSkillHero 英雄技能升级请求处理
func (h *Handler) UpSkillHero(ctx context.Context, req *pb.UpSkillHeroRequest) (*pb.OpenSkillHeroesResponse, error) {
	return h.Service.UpSkillHero(ctx, req)
}
