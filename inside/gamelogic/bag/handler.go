package bag

import (
	"context"
	"zagame/inside/gamelogic/common"
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
func (h *Handler) RegisterHandlers(router *grpc.Router, handler common.Handler) {
	// 背包消息
	router.RegisterHandler(proto.MSG_GetBagRequest, h, "GetBag",
		func() interface{} { return &pb.GetBagRequest{} },
		func() interface{} { return &pb.GetBagResponse{} })
	router.RegisterHandler(proto.MSG_BagItemUseRequest, h, "BagItemUse",
		func() interface{} { return &pb.BagItemUseRequest{} },
		func() interface{} { return &pb.BagItemUseResponse{} })
}
