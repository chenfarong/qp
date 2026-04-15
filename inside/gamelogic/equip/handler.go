package equip

import (
	"context"
	pb "zagame/pb/golang/gamelogic"
)

// Handler 装备处理器
type Handler struct {
	Service *Service
}

// GetEquip 获取装备请求处理
func (h *Handler) GetEquip(ctx context.Context, req *pb.GetEquipRequest) (*pb.GetEquipResponse, error) {
	return h.Service.GetEquip(ctx, req)
}

// UpgradeEquip 装备升级请求处理
func (h *Handler) UpgradeEquip(ctx context.Context, req *pb.UpgradeEquipRequest) (*pb.UpgradeEquipResponse, error) {
	return h.Service.UpgradeEquip(ctx, req)
}