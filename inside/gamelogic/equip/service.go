package equip

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

// GetEquip 获取装备服务
func (s *Service) GetEquip(ctx context.Context, req *pb.GetEquipRequest) (*pb.GetEquipResponse, error) {
	// 实现获取装备逻辑
	equips := s.model.GetEquips()
	return &pb.GetEquipResponse{
		Err: &pb.ResultErr{ErrCode: 0, ErrText: ""},
		Data: equips,
	}, nil
}

// UpgradeEquip 装备升级服务
func (s *Service) UpgradeEquip(ctx context.Context, req *pb.UpgradeEquipRequest) (*pb.UpgradeEquipResponse, error) {
	// 实现装备升级逻辑
	equip := s.model.UpgradeEquip(req.EquipId)
	return &pb.UpgradeEquipResponse{
		Err: &pb.ResultErr{ErrCode: 0, ErrText: ""},
		Data: equip,
	}, nil
}