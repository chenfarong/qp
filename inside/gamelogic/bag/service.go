package bag

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

// AddItem 添加物品服务
func (s *Service) AddItem(ctx context.Context, req *pb.AddItemRequest) (*pb.AddItemResponse, error) {
	// 实现添加物品逻辑
	s.model.AddItem(req.ItemId, req.Count)
	return &pb.AddItemResponse{
		Success: true,
		Message: "添加物品成功",
	}, nil
}

// RemoveItem 移除物品服务
func (s *Service) RemoveItem(ctx context.Context, req *pb.RemoveItemRequest) (*pb.RemoveItemResponse, error) {
	// 实现移除物品逻辑
	s.model.RemoveItem(req.ItemId, req.Count)
	return &pb.RemoveItemResponse{
		Success: true,
		Message: "移除物品成功",
	}, nil
}

// GetBag 获取背包服务
func (s *Service) GetBag(ctx context.Context, req *pb.GetBagRequest) (*pb.GetBagResponse, error) {
	// 实现获取背包逻辑
	bag := s.model.GetBag()
	return &pb.GetBagResponse{
		Success: true,
		Message: "获取背包成功",
		Bag:     bag,
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
