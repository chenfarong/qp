package bag

import (
	"context"
	pb "zagame/pb/golang/gamelogic"
)

// Handler 背包处理器
type Handler struct {
	Service *Service
}

// GetBag 获取背包请求处理
func (h *Handler) GetBag(ctx context.Context, req *pb.GetBagRequest) (*pb.GetBagResponse, error) {
	return h.Service.GetBag(ctx, req)
}

// BagItemUse 使用道具请求处理
func (h *Handler) BagItemUse(ctx context.Context, req *pb.BagItemUseRequest) (*pb.BagItemUseResponse, error) {
	return h.Service.BagItemUse(ctx, req)
}
