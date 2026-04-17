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
var sessionActor sync.Map

// GetActorInfo 获取会话对应的角色信息
func GetActorInfo(session string) (ActorInfo, bool) {
	value, exists := sessionActor.Load(session)
	if !exists {
		return ActorInfo{}, false
	}
	actorInfo, ok := value.(ActorInfo)
	return actorInfo, ok
}

// SetActorInfo 设置会话对应的角色信息
func SetActorInfo(session string, actorInfo ActorInfo) {
	sessionActor.Store(session, actorInfo)
}

// RemoveActorInfo 移除会话对应的角色信息
func RemoveActorInfo(session string) {
	sessionActor.Delete(session)
}
