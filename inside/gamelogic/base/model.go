package base

import (
	pb "zagame/pb/golang/gamelogic"
)

type Model struct {
	// 存储角色信息
	roles map[string]*pb.Role
}

func NewModel() *Model {
	return &Model{
		roles: make(map[string]*pb.Role),
	}
}

// GetRole 获取角色信息
func (m *Model) GetRole(aid string) *pb.Role {
	return m.roles[aid]
}

// SaveRole 保存角色信息
func (m *Model) SaveRole(role *pb.Role) {
	m.roles[role.Aid] = role
}
