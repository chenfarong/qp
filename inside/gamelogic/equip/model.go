package equip

import (
	pb "zagame/pb/golang/gamelogic"
)

// Model 装备模型
type Model struct {
	// 存储装备列表
	equips []*pb.EquipData
}

func NewModel() *Model {
	return &Model{
		equips: []*pb.EquipData{},
	}
}

// GetEquips 获取装备列表
func (m *Model) GetEquips() []*pb.EquipData {
	return m.equips
}

// UpgradeEquip 装备升级
func (m *Model) UpgradeEquip(equipId int64) *pb.EquipData {
	for _, equip := range m.equips {
		if equip.EquipId == equipId {
			// 实现装备升级逻辑
			return equip
		}
	}
	// 如果找不到装备，创建一个新的
	equip := &pb.EquipData{
		EquipId:    equipId,
		EquipCfgId: 1,
	}
	m.equips = append(m.equips, equip)
	return equip
}