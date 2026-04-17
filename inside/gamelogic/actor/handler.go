package actor

import (
	"context"

	"zagame/common/logger"
	"zagame/inside/gamelogic/common"
	"zagame/inside/gamelogic/grpc"
	pb "zagame/pb/golang/gamelogic"
	"zagame/proto"
)

// Handler 角色处理器
type Handler struct {
	Service *Service
}

// ActorCreate 创建角色服务
func (h *Handler) ActorCreate(ctx context.Context, req *pb.ActorCreateRequest) (*pb.ActorUseResponse, error) {
	logger.Debug("创建角色请求", "user", ctx.Value("username"), "actorName", req.GetActorName())

	return h.Service.ActorCreate(ctx, req)
}

// ActorUse 使用角色服务
func (h *Handler) ActorUse(ctx context.Context, req *pb.ActorUseRequest) (*pb.ActorUseResponse, error) {

	logger.Debug("使用角色请求", "user", ctx.Value("username"), "actorID", req.GetAid())

	return h.Service.ActorUse(ctx, req)
}

// ActorUseWithName 使用角色服务（通过名称）
func (h *Handler) ActorUseWithName(ctx context.Context, req *pb.ActorUseWithNameRequest) (*pb.ActorUseResponse, error) {
	return h.Service.ActorUseWithName(ctx, req)
}

// NewHandler 创建角色处理器
func NewHandler() *Handler {
	return &Handler{
		Service: NewService(),
	}
}

// RegisterHandlers 注册消息处理器
func (h *Handler) RegisterHandlers(router *grpc.Router, handler common.Handler) {
	// 角色消息
	router.RegisterHandler(proto.MSG_ActorCreateRequest, h, "ActorCreate",
		func() interface{} { return &pb.ActorCreateRequest{} },
		func() interface{} { return &pb.ActorUseResponse{} })
	router.RegisterHandler(proto.MSG_ActorUseRequest, h, "ActorUse",
		func() interface{} { return &pb.ActorUseRequest{} },
		func() interface{} { return &pb.ActorUseResponse{} })
	router.RegisterHandler(proto.MSG_ActorUseWithNameRequest, h, "ActorUseWithName",
		func() interface{} { return &pb.ActorUseWithNameRequest{} },
		func() interface{} { return &pb.ActorUseResponse{} })
}
