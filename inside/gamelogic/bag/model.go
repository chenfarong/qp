package bag

import (
	"fmt"
	pb "zagame/pb/golang/gamelogic"
)

// Model 背包模型
type Model struct {
	// 存储背包物品
	bag map[string]int32
}

func NewModel() *Model {
	return &Model{
		bag: make(map[string]int32),
	}
}

// GetBag 获取背包
func (m *Model) GetBag() map[string]int32 {
	return m.bag
}

// GetItems 获取物品列表
func (m *Model) GetItems() []*pb.ItemData {
	var items []*pb.ItemData
	for itemId, count := range m.bag {
		// 尝试将itemId转换为int64作为ItemId
		itemIdInt := int64(0)
		_, err := fmt.Sscanf(itemId, "%d", &itemIdInt)
		if err != nil {
			itemIdInt = 0
		}

		item := &pb.ItemData{
			ItemId:    itemIdInt, // 使用itemId作为物品ID
			ItemCfgId: 1,         // 这里应该使用配置ID，暂时设为1
			Num:       int64(count),
		}
		items = append(items, item)
	}
	return items
}
