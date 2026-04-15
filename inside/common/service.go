package common

// ServiceMessageResult 服务消息结果
type ServiceMessageResult struct {
	ServiceName    string `json:"serviceName"`
	MessageContent []byte `json:"messageContent"`
}

// Service 服务接口
type Service interface {
	OnStartup()
	OnShutdown()
	// 内部消息处理接口
	HandleMessage(messageID string, actorID string, messageContent []byte, results []ServiceMessageResult)
}

// Handler 消息处理器接口
type Handler interface {
	RegisterHandlers()
	OnActorUse(actorID string)
	OnActorLogout(actorID string)
	OnActorOnline(actorID string)
	OnActorOffline(actorID string)
}
