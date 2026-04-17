package equip

import (
	"context"
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
func (h *Handler) RegisterHandlers(router *grpc.Router) {
	// 装备消息
	router.RegisterHandler(proto.MSG_GetEquipRequest, h.handleGetEquipRequest)
	router.RegisterHandler(proto.MSG_UpgradeEquipRequest, h.handleUpgradeEquipRequest)
}

// handleGetEquipRequest 处理获取装备请求
func (h *Handler) handleGetEquipRequest(ctx context.Context, session string, messageContent []byte) ([]byte, error) {
	req := &pb.GetEquipRequest{}
	if err := grpc.Unmarshal(messageContent, req); err != nil {
		return nil, err
	}

	resp, err := h.GetEquip(ctx, req)
	if err != nil {
		return nil, err
	}

	return grpc.Marshal(resp)
}

// handleUpgradeEquipRequest 处理升级装备请求
func (h *Handler) handleUpgradeEquipRequest(ctx context.Context, session string, messageContent []byte) ([]byte, error) {
	req := &pb.UpgradeEquipRequest{}
	if err := grpc.Unmarshal(messageContent, req); err != nil {
		return nil, err
	}

	resp, err := h.UpgradeEquip(ctx, req)
	if err != nil {
		return nil, err
	}

	return grpc.Marshal(resp)
}
