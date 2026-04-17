package session

import (
	"sync"
)

// ActorInfo 角色信息
type ActorInfo struct {
	ActorID   string
	ActorName string
}

// 全局 session -> actor 映射
var (
	sessionActor map[string]ActorInfo
	sessionActorMu sync.RWMutex
)

func init() {
	sessionActor = make(map[string]ActorInfo)
}

// GetActorInfo 获取会话对应的角色信息
func GetActorInfo(session string) (ActorInfo, bool) {
	sessionActorMu.RLock()
	defer sessionActorMu.RUnlock()
	actorInfo, exists := sessionActor[session]
	return actorInfo, exists
}

// SetActorInfo 设置会话对应的角色信息
func SetActorInfo(session string, actorInfo ActorInfo) {
	sessionActorMu.Lock()
	defer sessionActorMu.Unlock()
	sessionActor[session] = actorInfo
}

// RemoveActorInfo 移除会话对应的角色信息
func RemoveActorInfo(session string) {
	sessionActorMu.Lock()
	defer sessionActorMu.Unlock()
	delete(sessionActor, session)
}