package actor

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

// ActorCreate 创建角色服务
func (s *Service) ActorCreate(ctx context.Context, req *pb.ActorCreateRequest) (*pb.ActorUseResponse, error) {
	// 实现创建角色逻辑
	aid := s.model.CreateActor()
	actorData := &pb.ActorData{
		ActorId: aid,
		Name:    req.GetActorName(),
		Level:   1,
	}
	return &pb.ActorUseResponse{
		Err:  &pb.ResultErr{ErrCode: 0, ErrText: ""},
		Data: actorData,
	}, nil
}

// ActorUse 使用角色服务
func (s *Service) ActorUse(ctx context.Context, req *pb.ActorUseRequest) (*pb.ActorUseResponse, error) {
	// 实现使用角色逻辑
	err := s.model.UseActor(req.GetAid())
	if err != nil {
		return &pb.ActorUseResponse{
			Err: &pb.ResultErr{ErrCode: 1, ErrText: err.Error()},
		}, err
	}
	actorData := &pb.ActorData{
		ActorId: req.GetAid(),
		Name:    "TestActor",
		Level:   1,
	}
	return &pb.ActorUseResponse{
		Err:  &pb.ResultErr{ErrCode: 0, ErrText: ""},
		Data: actorData,
	}, nil
}

// ActorUseWithName 使用角色服务（通过名称）
func (s *Service) ActorUseWithName(ctx context.Context, req *pb.ActorUseWithNameRequest) (*pb.ActorUseResponse, error) {
	// 实现使用角色逻辑（通过名称）
	aid := s.model.GetActorIdByName(req.Name)
	if aid == "" {
		return &pb.ActorUseResponse{
			Err: &pb.ResultErr{ErrCode: 1, ErrText: "Actor not found"},
		}, nil
	}
	actorData := &pb.ActorData{
		ActorId: aid,
		Name:    req.Name,
		Realm:   req.Realm,
		Level:   1,
	}
	return &pb.ActorUseResponse{
		Err:  &pb.ResultErr{ErrCode: 0, ErrText: ""},
		Data: actorData,
	}, nil
}
