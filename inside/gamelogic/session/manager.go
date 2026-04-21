package session

import (
	"context"
	"encoding/json"
	"sync"

	"zagame/common/logger"
	"zagame/proto"
)

// ActorInfo 角色信息
type ActorInfo struct {
	ActorID   string
	ActorName string
}

// SessionManager 会话管理器
type SessionManager struct {
	sessions map[string]*Session
	mutex    sync.RWMutex
}

// Session 会话结构
type Session struct {
	sessionID string
	msgChan   chan *Message
	closeChan chan struct{}
	actorInfo *ActorInfo
}

// Message 消息结构
type Message struct {
	ctx            context.Context
	messageId      int32
	session        string
	messageContent []byte
	responseChan   chan []byte
	errorChan      chan error
}

var (
	sessionManager *SessionManager
	once           sync.Once
)

// GetSessionManager 获取会话管理器实例
func GetSessionManager() *SessionManager {
	once.Do(func() {
		sessionManager = &SessionManager{
			sessions: make(map[string]*Session),
		}
	})
	return sessionManager
}

// GetActorInfo 获取角色信息
func GetActorInfo(sessionID string) (ActorInfo, bool) {
	sm := GetSessionManager()
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists || session.actorInfo == nil {
		return ActorInfo{}, false
	}

	return *session.actorInfo, true
}

// SetActorInfo 设置角色信息
func SetActorInfo(sessionID string, actorInfo ActorInfo) {
	sm := GetSessionManager()
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[sessionID]
	if exists {
		session.actorInfo = &actorInfo
	}
}

// GetOrCreateSession 获取或创建会话
func (sm *SessionManager) GetOrCreateSession(sessionID string) *Session {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		session = &Session{
			sessionID: sessionID,
			msgChan:   make(chan *Message, 100),
			closeChan: make(chan struct{}),
		}
		sm.sessions[sessionID] = session

		// 为每个会话启动一个goroutine
		go session.processMessages()

		logger.Infof("创建新会话: %s", sessionID)
	}

	return session
}

// ProcessMessage 处理消息
func (sm *SessionManager) ProcessMessage(ctx context.Context, messageId int32, session string, messageContent []byte) ([]byte, error) {
	s := sm.GetOrCreateSession(session)

	responseChan := make(chan []byte, 1)
	errorChan := make(chan error, 1)

	s.msgChan <- &Message{
		ctx:            ctx,
		messageId:      messageId,
		session:        session,
		messageContent: messageContent,
		responseChan:   responseChan,
		errorChan:      errorChan,
	}

	select {
	case response := <-responseChan:
		return response, nil
	case err := <-errorChan:
		return nil, err
	}
}

// processMessages 处理会话消息
func (s *Session) processMessages() {
	// 添加panic恢复机制，防止单个会话的panic影响其他会话和整个程序
	defer func() {
		if r := recover(); r != nil {
			logger.Panicf("会话 %s 处理线程发生panic，已恢复: %v", s.sessionID, r)
		}
	}()

	logger.Infof("会话处理线程启动: %s", s.sessionID)
	defer logger.Infof("会话处理线程结束: %s", s.sessionID)

	for {
		select {
		case msg := <-s.msgChan:
			s.processMessage(msg)
		case <-s.closeChan:
			return
		}
	}
}

// processMessage 处理单个消息
func (s *Session) processMessage(msg *Message) {
	// 从router包中导入HandleMessage函数
	// 这里需要注意循环导入的问题，稍后会解决
	// 暂时使用一个全局变量来存储router实例
	if router == nil {
		msg.errorChan <- nil
		return
	}

	// 将actor信息添加到上下文中
	if s.actorInfo != nil {
		msg.ctx = context.WithValue(msg.ctx, "actor_id", s.actorInfo.ActorID)
		msg.ctx = context.WithValue(msg.ctx, "actor_name", s.actorInfo.ActorName)
	}

	responseContent, err := router.HandleMessage(msg.ctx, msg.messageId, msg.session, msg.messageContent)
	if err != nil {
		msg.errorChan <- err
		return
	}

	// 检查是否需要更新actor信息
	if msg.messageId == proto.MSG_ActorCreateRequest || msg.messageId == proto.MSG_ActorUseRequest || msg.messageId == proto.MSG_ActorUseWithNameRequest {
		if responseContent != nil && len(*responseContent) > 0 {
			var respContent map[string]interface{}
			err := json.Unmarshal(*responseContent, &respContent)
			if err == nil {
				if errInfo, ok := respContent["err"].(map[string]interface{}); ok {
					if errCode, ok := errInfo["errCode"].(float64); ok && errCode == 0 {
						if data, ok := respContent["data"].(map[string]interface{}); ok {
							if actorId, ok := data["actorId"].(string); ok {
								if actorName, ok := data["name"].(string); ok {
									s.actorInfo = &ActorInfo{
										ActorID:   actorId,
										ActorName: actorName,
									}
									logger.Infof("会话 %s 关联角色: %s(%s)", s.sessionID, actorName, actorId)
								}
							}
						}
					}
				}
			}
		}
	}

	msg.responseChan <- *responseContent
}

// Close 关闭会话
func (s *Session) Close() {
	close(s.closeChan)
}

// 全局变量，用于存储router实例
var router Router

// SetRouter 设置router实例
func SetRouter(r Router) {
	router = r
}

// Router 路由接口
type Router interface {
	HandleMessage(ctx context.Context, messageId int32, session string, messageContent []byte) (*[]byte, error)
}
