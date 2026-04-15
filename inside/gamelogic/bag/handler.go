package bag

import (
	"context"
	pb "zgame/pb/golang/gamelogic"
)

// Handler 背包处理器
type Handler struct {
	Service *Service
}

// AddItem 添加物品请求处理
func (h *Handler) AddItem(ctx context.Context, req *pb.AddItemRequest) (*pb.AddItemResponse, error) {
	return h.Service.AddItem(ctx, req)
}

// RemoveItem 移除物品请求处理
func (h *Handler) RemoveItem(ctx context.Context, req *pb.RemoveItemRequest) (*pb.RemoveItemResponse, error) {
	return h.Service.RemoveItem(ctx, req)
}

// GetBag 获取背包请求处理
func (h *Handler) GetBag(ctx context.Context, req *pb.GetBagRequest) (*pb.GetBagResponse, error) {
	return h.Service.GetBag(ctx, req)
}

// BagItemUse 使用道具请求处理
func (h *Handler) BagItemUse(ctx context.Context, req *pb.BagItemUseRequest) (*pb.BagItemUseResponse, error) {
	return h.Service.BagItemUse(ctx, req)
}
