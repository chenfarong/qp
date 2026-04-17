package hero

import (
	"context"
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
func (h *Handler) RegisterHandlers(router *grpc.Router) {
	// 英雄消息
	router.RegisterHandler(proto.MSG_GetHeroesRequest, h.handleGetHeroesRequest)
	router.RegisterHandler(proto.MSG_RecruitHeroRequest, h.handleRecruitHeroRequest)
	router.RegisterHandler(proto.MSG_UpStarHeroRequest, h.handleUpStarHeroRequest)
	router.RegisterHandler(proto.MSG_OpenSkillHeroRequest, h.handleOpenSkillHeroRequest)
	router.RegisterHandler(proto.MSG_UpSkillHeroRequest, h.handleUpSkillHeroRequest)
}

// handleGetHeroesRequest 处理获取英雄请求
func (h *Handler) handleGetHeroesRequest(ctx context.Context, session string, messageContent []byte) ([]byte, error) {
	req := &pb.GetHeroesRequest{}
	if err := grpc.Unmarshal(messageContent, req); err != nil {
		return nil, err
	}

	resp, err := h.GetHeroes(ctx, req)
	if err != nil {
		return nil, err
	}

	return grpc.Marshal(resp)
}

// handleRecruitHeroRequest 处理招募英雄请求
func (h *Handler) handleRecruitHeroRequest(ctx context.Context, session string, messageContent []byte) ([]byte, error) {
	req := &pb.RecruitHeroRequest{}
	if err := grpc.Unmarshal(messageContent, req); err != nil {
		return nil, err
	}

	resp, err := h.RecruitHero(ctx, req)
	if err != nil {
		return nil, err
	}

	return grpc.Marshal(resp)
}

// handleUpStarHeroRequest 处理英雄升星请求
func (h *Handler) handleUpStarHeroRequest(ctx context.Context, session string, messageContent []byte) ([]byte, error) {
	req := &pb.UpStarHeroRequest{}
	if err := grpc.Unmarshal(messageContent, req); err != nil {
		return nil, err
	}

	resp, err := h.UpStarHero(ctx, req)
	if err != nil {
		return nil, err
	}

	return grpc.Marshal(resp)
}

// handleOpenSkillHeroRequest 处理开启技能请求
func (h *Handler) handleOpenSkillHeroRequest(ctx context.Context, session string, messageContent []byte) ([]byte, error) {
	req := &pb.OpenSkillHeroRequest{}
	if err := grpc.Unmarshal(messageContent, req); err != nil {
		return nil, err
	}

	resp, err := h.OpenSkillHero(ctx, req)
	if err != nil {
		return nil, err
	}

	return grpc.Marshal(resp)
}

// handleUpSkillHeroRequest 处理升级技能请求
func (h *Handler) handleUpSkillHeroRequest(ctx context.Context, session string, messageContent []byte) ([]byte, error) {
	req := &pb.UpSkillHeroRequest{}
	if err := grpc.Unmarshal(messageContent, req); err != nil {
		return nil, err
	}

	resp, err := h.UpSkillHero(ctx, req)
	if err != nil {
		return nil, err
	}

	return grpc.Marshal(resp)
}
