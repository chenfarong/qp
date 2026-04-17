package equip

import (
	"context"
	"zagame/inside/gamelogic/common"
	"zagame/inside/gamelogic/grpc"
	pb "zagame/pb/golang/gamelogic"
	"zagame/proto"
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

// NewHandler 创建装备处理器
func NewHandler() *Handler {
	return &Handler{
		Service: NewService(),
	}
}

// RegisterHandlers 注册消息处理器
func (h *Handler) RegisterHandlers(router *grpc.Router, handler common.Handler) {
	router.RegisterHandler(proto.MSG_GetEquipRequest, h, "GetEquip",
		func() interface{} { return &pb.GetEquipRequest{} },
		func() interface{} { return &pb.GetEquipResponse{} })
	router.RegisterHandler(proto.MSG_UpgradeEquipRequest, h, "UpgradeEquip",
		func() interface{} { return &pb.UpgradeEquipRequest{} },
		func() interface{} { return &pb.UpgradeEquipResponse{} })
}
