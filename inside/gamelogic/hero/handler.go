package hero

import (
	"context"
	"zagame/inside/gamelogic/common"
	"zagame/inside/gamelogic/grpc"
	pb "zagame/pb/golang/gamelogic"
	"zagame/proto"
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

// NewHandler 创建英雄处理器
func NewHandler() *Handler {
	return &Handler{
		Service: NewService(),
	}
}

// RegisterHandlers 注册消息处理器
func (h *Handler) RegisterHandlers(router *grpc.Router, handler common.Handler) {
	router.RegisterHandler(proto.MSG_GetHeroesRequest, h, "GetHeroes",
		func() interface{} { return &pb.GetHeroesRequest{} },
		func() interface{} { return &pb.GetHeroesResponse{} })
	router.RegisterHandler(proto.MSG_RecruitHeroRequest, h, "RecruitHero",
		func() interface{} { return &pb.RecruitHeroRequest{} },
		func() interface{} { return &pb.RecruiHeroesResponse{} })
	router.RegisterHandler(proto.MSG_UpStarHeroRequest, h, "UpStarHero",
		func() interface{} { return &pb.UpStarHeroRequest{} },
		func() interface{} { return &pb.UpStarHeroesResponse{} })
	router.RegisterHandler(proto.MSG_OpenSkillHeroRequest, h, "OpenSkillHero",
		func() interface{} { return &pb.OpenSkillHeroRequest{} },
		func() interface{} { return &pb.OpenSkillHeroesResponse{} })
	router.RegisterHandler(proto.MSG_UpSkillHeroRequest, h, "UpSkillHero",
		func() interface{} { return &pb.UpSkillHeroRequest{} },
		func() interface{} { return &pb.OpenSkillHeroesResponse{} })
}
