package grpc

import (
	"context"
	"fmt"
	"log"
	"sync"

	"google.golang.org/protobuf/proto"
)

// Marshal 序列化消息
func Marshal(m proto.Message) ([]byte, error) {
	return proto.Marshal(m)
}

// Unmarshal 反序列化消息
func Unmarshal(data []byte, m proto.Message) error {
	return proto.Unmarshal(data, m)
}

// MessageHandler 消息处理器类型
type MessageHandler func(ctx context.Context, session string, messageContent []byte) ([]byte, error)

// Router 消息路由器
type Router struct {
	handlers map[int32]MessageHandler
	mu       sync.RWMutex
}

// NewRouter 创建消息路由器
func NewRouter() *Router {
	return &Router{
		handlers: make(map[int32]MessageHandler),
	}
}

// RegisterHandler 注册消息处理器
func (r *Router) RegisterHandler(messageID int32, handler MessageHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[messageID] = handler
	log.Printf("注册消息处理器: messageID=%d\n", messageID)
}

// HandleMessage 处理消息
func (r *Router) HandleMessage(ctx context.Context, messageID int32, session string, messageContent []byte) ([]byte, error) {
	r.mu.RLock()
	handler, ok := r.handlers[messageID]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("未找到消息处理器: messageID=%d", messageID)
	}

	return handler(ctx, session, messageContent)
}
