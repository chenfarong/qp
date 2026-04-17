package bag

import (
	"context"
	"zagame/inside/gamelogic/grpc"
	pb "zagame/pb/golang/gamelogic"
	"zagame/proto"

	"github.com/bytedance/gopkg/util/logger"
)

// Handler 背包处理器
type Handler struct {
	Service *Service
}

// GetBag 获取背包请求处理
func (h *Handler) GetBag(ctx context.Context, req *pb.GetBagRequest) (*pb.GetBagResponse, error) {

	logger.Debug("获取背包请求", "user", ctx.Value("username"))

	return h.Service.GetBag(ctx, req)
}

// BagItemUse 使用道具请求处理
func (h *Handler) BagItemUse(ctx context.Context, req *pb.BagItemUseRequest) (*pb.BagItemUseResponse, error) {
	return h.Service.BagItemUse(ctx, req)
}

// NewHandler 创建背包处理器
func NewHandler() *Handler {
	return &Handler{
		Service: NewService(),
	}
}

// RegisterHandlers 注册消息处理器
func (h *Handler) RegisterHandlers(router *grpc.Router) {
	// 背包消息
	router.RegisterHandler(proto.MSG_GetBagRequest, h.handleGetBagRequest)
	router.RegisterHandler(proto.MSG_BagItemUseRequest, h.handleBagItemUseRequest)
}

// handleGetBagRequest 处理获取背包请求
func (h *Handler) handleGetBagRequest(ctx context.Context, session string, messageContent []byte) ([]byte, error) {
	req := &pb.GetBagRequest{}
	if err := grpc.Unmarshal(messageContent, req); err != nil {
		return nil, err
	}

	resp, err := h.GetBag(ctx, req)
	if err != nil {
		return nil, err
	}

	return grpc.Marshal(resp)
}

// handleBagItemUseRequest 处理背包物品使用请求
func (h *Handler) handleBagItemUseRequest(ctx context.Context, session string, messageContent []byte) ([]byte, error) {
	req := &pb.BagItemUseRequest{}
	if err := grpc.Unmarshal(messageContent, req); err != nil {
		return nil, err
	}

	resp, err := h.BagItemUse(ctx, req)
	if err != nil {
		return nil, err
	}

	return grpc.Marshal(resp)
}
