package hero

// Model 英雄模型
type Model struct {
	// 存储英雄列表
	heroes []string
}

func NewModel() *Model {
	return &Model{
		heroes: []string{},
	}
}

// AddHero 添加英雄
func (m *Model) AddHero(heroId string) {
	m.heroes = append(m.heroes, heroId)
}

// RemoveHero 移除英雄
func (m *Model) RemoveHero(heroId string) {
	for i, hero := range m.heroes {
		if hero == heroId {
			m.heroes = append(m.heroes[:i], m.heroes[i+1:]...)
			break
		}
	}
}

// GetHeroes 获取英雄列表
func (m *Model) GetHeroes() []string {
	return m.heroes
}
