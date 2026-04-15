package router

import (
	"github.com/aoyo/qp/pkg/proto/proto"
)

// Handler 处理函数类型
type Handler func(*proto.Message) *proto.Message

// Router 路由器
type Router struct {
	handlers map[proto.MessageType]Handler
}

// NewRouter 创建新的路由器
func NewRouter() *Router {
	return &Router{
		handlers: make(map[proto.MessageType]Handler),
	}
}

// Register 注册处理函数
func (r *Router) Register(messageType proto.MessageType, handler Handler) {
	r.handlers[messageType] = handler
}

// Handle 处理消息
func (r *Router) Handle(message *proto.Message) *proto.Message {
	if handler, ok := r.handlers[message.Type]; ok {
		return handler(message)
	}
	return &proto.Message{
		Type: proto.MessageType_MSG_TYPE_RESPONSE,
		Data: &proto.Message_Response{
			Response: &proto.Response{
				Code:    404,
				Message: "Handler not found",
			},
		},
	}
}

// GetHandlers 获取所有处理函数
func (r *Router) GetHandlers() map[proto.MessageType]Handler {
	return r.handlers
}

// GetMessageTypes 获取所有注册的消息类型
func (r *Router) GetMessageTypes() []proto.MessageType {
	types := make([]proto.MessageType, 0, len(r.handlers))
	for msgType := range r.handlers {
		types = append(types, msgType)
	}
	return types
}
