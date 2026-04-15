package bag

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

// AddItem 添加物品
func (m *Model) AddItem(itemId string, count int32) {
	m.bag[itemId] += count
}

// RemoveItem 移除物品
func (m *Model) RemoveItem(itemId string, count int32) {
	if m.bag[itemId] > count {
		m.bag[itemId] -= count
	} else {
		delete(m.bag, itemId)
	}
}

// GetBag 获取背包
func (m *Model) GetBag() map[string]int32 {
	return m.bag
}
