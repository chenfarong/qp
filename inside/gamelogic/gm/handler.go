package gm

// Handler GM处理器
type Handler struct {
	Service *Service
}

// NewHandler 创建GM处理器
func NewHandler() *Handler {
	return &Handler{
		Service: NewService(),
	}
}

// RegisterHandlers 注册处理器
func (h *Handler) RegisterHandlers(router interface{}, instance interface{}) {
	// 这里可以注册GM相关的消息处理器
	// 但GM服务器是独立的HTTP服务器，不需要注册到消息路由器
}
