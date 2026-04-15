package actor

import (
	"fmt"
	"sync"
)

// Model 角色模型
type Model struct {
	// 存储角色信息
	actors map[string]bool
	// 自增ID
	idCounter int
	// 互斥锁
	mu sync.Mutex
}

func NewModel() *Model {
	return &Model{
		actors:    make(map[string]bool),
		idCounter: 0,
	}
}

// CreateActor 创建角色
func (m *Model) CreateActor() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.idCounter++
	aid := fmt.Sprintf("actor_%d", m.idCounter)
	m.actors[aid] = true
	return aid
}

// UseActor 使用角色
func (m *Model) UseActor(aid string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.actors[aid]; !exists {
		return fmt.Errorf("actor not found: %s", aid)
	}
	
	// 实现使用角色的逻辑
	return nil
}
