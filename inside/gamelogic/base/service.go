package base

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

// Login 登录服务
func (s *Service) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// 实现登录逻辑
	return &pb.LoginResponse{
		Success: true,
		Message: "登录成功",
		Role: &pb.Role{
			Aid:      "1",
			Name:     req.Name,
			Level:    1,
			Gold:     1000,
			Bag:      make(map[string]int32),
			Heroes:   []string{},
			Session:  "session_1",
		},
	}, nil
}

// GetRoleInfo 获取角色信息服务
func (s *Service) GetRoleInfo(ctx context.Context, req *pb.GetRoleInfoRequest) (*pb.GetRoleInfoResponse, error) {
	// 实现获取角色信息逻辑
	return &pb.GetRoleInfoResponse{
		Success: true,
		Message: "获取角色信息成功",
		Role: &pb.Role{
			Aid:      "1",
			Name:     "TestRole",
			Level:    1,
			Gold:     1000,
			Bag:      make(map[string]int32),
			Heroes:   []string{},
			Session:  "session_1",
		},
	}, nil
}
