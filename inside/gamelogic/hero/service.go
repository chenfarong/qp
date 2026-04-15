package hero

import (
	"context"
	pb "zagame/pb/golang/gamelogic"
)

type Service struct {
	model *Model
}

func NewService() *Service {
	return &Service{
		model: NewModel(),
	}
}

// GetHeroes 获取英雄列表服务
func (s *Service) GetHeroes(ctx context.Context, req *pb.GetHeroesRequest) (*pb.GetHeroesResponse, error) {
	// 实现获取英雄列表逻辑
	heroes := s.model.GetHeroes()
	return &pb.GetHeroesResponse{
		Err:  &pb.ResultErr{ErrCode: 0, ErrText: ""},
		Data: heroes,
	}, nil
}

// RecruitHero 招募英雄服务
func (s *Service) RecruitHero(ctx context.Context, req *pb.RecruitHeroRequest) (*pb.RecruiHeroesResponse, error) {
	// 实现招募英雄逻辑
	hero := s.model.RecruitHero(req.CfgId)
	return &pb.RecruiHeroesResponse{
		Err:  &pb.ResultErr{ErrCode: 0, ErrText: ""},
		Data: hero,
	}, nil
}

// UpStarHero 英雄升星服务
func (s *Service) UpStarHero(ctx context.Context, req *pb.UpStarHeroRequest) (*pb.UpStarHeroesResponse, error) {
	// 实现英雄升星逻辑
	hero := s.model.UpStarHero(req.Uid)
	return &pb.UpStarHeroesResponse{
		Err:  &pb.ResultErr{ErrCode: 0, ErrText: ""},
		Data: hero,
	}, nil
}

// OpenSkillHero 英雄技能开启服务
func (s *Service) OpenSkillHero(ctx context.Context, req *pb.OpenSkillHeroRequest) (*pb.OpenSkillHeroesResponse, error) {
	// 实现英雄技能开启逻辑
	hero := s.model.OpenSkillHero(req.Uid, req.SlotId)
	return &pb.OpenSkillHeroesResponse{
		Err:  &pb.ResultErr{ErrCode: 0, ErrText: ""},
		Data: hero,
	}, nil
}

// UpSkillHero 英雄技能升级服务
func (s *Service) UpSkillHero(ctx context.Context, req *pb.UpSkillHeroRequest) (*pb.OpenSkillHeroesResponse, error) {
	// 实现英雄技能升级逻辑
	hero := s.model.UpSkillHero(req.Uid, req.SlotId)
	return &pb.OpenSkillHeroesResponse{
		Err:  &pb.ResultErr{ErrCode: 0, ErrText: ""},
		Data: hero,
	}, nil
}
