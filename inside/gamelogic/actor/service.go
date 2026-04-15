package actor

import (
	"context"
	pb "zgame/pb/golang/gamelogic"
)

type Service struct {
	model *Model
}

func NewService() *Service {
	return &Service{
		model: NewModel(),
	}
}

// ActorCreate 创建角色服务
func (s *Service) ActorCreate(ctx context.Context, req *pb.ActorCreateRequest) (*pb.ActorCreateResponse, error) {
	// 实现创建角色逻辑
	aid := s.model.CreateActor()
	return &pb.ActorCreateResponse{
		Aid: aid,
	}, nil
}

// ActorUse 使用角色服务
func (s *Service) ActorUse(ctx context.Context, req *pb.ActorUseRequest) error {
	// 实现使用角色逻辑
	return s.model.UseActor(req.Aid)
}
