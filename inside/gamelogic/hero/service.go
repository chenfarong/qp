package hero

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

// AddHero 添加英雄服务
func (s *Service) AddHero(ctx context.Context, req *pb.AddHeroRequest) (*pb.AddHeroResponse, error) {
	// 实现添加英雄逻辑
	s.model.AddHero(req.HeroId)
	return &pb.AddHeroResponse{
		Success: true,
		Message: "添加英雄成功",
	}, nil
}

// RemoveHero 移除英雄服务
func (s *Service) RemoveHero(ctx context.Context, req *pb.RemoveHeroRequest) (*pb.RemoveHeroResponse, error) {
	// 实现移除英雄逻辑
	s.model.RemoveHero(req.HeroId)
	return &pb.RemoveHeroResponse{
		Success: true,
		Message: "移除英雄成功",
	}, nil
}

// GetHeroes 获取英雄列表服务
func (s *Service) GetHeroes(ctx context.Context, req *pb.GetHeroesRequest) (*pb.GetHeroesResponse, error) {
	// 实现获取英雄列表逻辑
	heroes := s.model.GetHeroes()
	return &pb.GetHeroesResponse{
		Success: true,
		Message: "获取英雄列表成功",
		Heroes:  heroes,
	}, nil
}
