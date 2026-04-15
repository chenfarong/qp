package actor

import (
	"fmt"
	"sync"
)

// Model 角色模型
type Model struct {
	// 存储角色信息
	actors map[string]bool
	// 存储名称到ID的映射
	nameToId map[string]string
	// 自增ID
	idCounter int
	// 互斥锁
	mu sync.Mutex
}

func NewModel() *Model {
	return &Model{
		actors:    make(map[string]bool),
		nameToId:  make(map[string]string),
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

// GetActorIdByName 根据名称获取角色ID
func (m *Model) GetActorIdByName(name string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if aid, exists := m.nameToId[name]; exists {
		return aid
	}
	
	// 如果没有找到，创建一个新的角色
	aid := m.CreateActor()
	m.nameToId[name] = aid
	return aid
}
