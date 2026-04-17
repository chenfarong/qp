package common

import (
	"zagame/inside/gamelogic/grpc"
)

// Handler 统一的处理器接口
type Handler interface {
	// RegisterHandlers 注册消息处理器
	RegisterHandlers(router *grpc.Router, handler Handler)
}
