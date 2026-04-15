package hero

import (
	pb "zagame/pb/golang/gamelogic"
)

// Model 英雄模型
type Model struct {
	// 存储英雄列表
	heroes []*pb.HeroData
}

func NewModel() *Model {
	return &Model{
		heroes: []*pb.HeroData{},
	}
}

// GetHeroes 获取英雄列表
func (m *Model) GetHeroes() []*pb.HeroData {
	return m.heroes
}

// RecruitHero 招募英雄
func (m *Model) RecruitHero(cfgId int32) *pb.HeroData {
	hero := &pb.HeroData{
		Uid:   int64(len(m.heroes) + 1),
		CfgId: cfgId,
		Level: 1,
		Star:  1,
	}
	m.heroes = append(m.heroes, hero)
	return hero
}

// UpStarHero 英雄升星
func (m *Model) UpStarHero(uid int64) *pb.HeroData {
	for _, hero := range m.heroes {
		if hero.Uid == uid {
			hero.Star++
			return hero
		}
	}
	return nil
}

// OpenSkillHero 英雄技能开启
func (m *Model) OpenSkillHero(uid int64, slotId int32) *pb.HeroData {
	for _, hero := range m.heroes {
		if hero.Uid == uid {
			// 实现技能开启逻辑
			return hero
		}
	}
	return nil
}

// UpSkillHero 英雄技能升级
func (m *Model) UpSkillHero(uid int64, slotId int32) *pb.HeroData {
	for _, hero := range m.heroes {
		if hero.Uid == uid {
			// 实现技能升级逻辑
			return hero
		}
	}
	return nil
}
