package bag

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

// GetBag 获取背包服务
func (s *Service) GetBag(ctx context.Context, req *pb.GetBagRequest) (*pb.GetBagResponse, error) {
	// 实现获取背包逻辑
	items := s.model.GetItems()
	errCode := int32(0)
	errText := ""
	return &pb.GetBagResponse{
		Err:  &pb.ResultErr{ErrCode: &errCode, ErrText: &errText},
		Data: items,
	}, nil
}

// BagItemUse 使用道具服务
func (s *Service) BagItemUse(ctx context.Context, req *pb.BagItemUseRequest) (*pb.BagItemUseResponse, error) {
	// 实现使用道具逻辑
	itemData := &pb.ItemData{
		ItemId:    req.ItemId,
		ItemCfgId: 1,
		Num:       req.Num,
	}
	return &pb.BagItemUseResponse{
		Success:  true,
		Message:  "使用道具成功",
		ItemData: itemData,
	}, nil
}
