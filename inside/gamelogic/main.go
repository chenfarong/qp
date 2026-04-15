package gamelogic

import (
	"zagame/inside/gamelogic/actor"
	"zagame/inside/gamelogic/bag"
	"zagame/inside/gamelogic/base"
	"zagame/inside/gamelogic/equip"
	"zagame/inside/gamelogic/hero"
)

// Handler 游戏逻辑处理器
type Handler struct {
	Base  *base.Handler
	Hero  *hero.Handler
	Bag   *bag.Handler
	Actor *actor.Handler
	Equip *equip.Handler
}

// Service 游戏逻辑服务
type Service struct {
	Base  *base.Service
	Hero  *hero.Service
	Bag   *bag.Service
	Actor *actor.Service
	Equip *equip.Service
}

// NewHandler 创建游戏逻辑处理器
func NewHandler() *Handler {
	return &Handler{
		Base:  &base.Handler{Service: base.NewService()},
		Hero:  &hero.Handler{Service: hero.NewService()},
		Bag:   &bag.Handler{Service: bag.NewService()},
		Actor: &actor.Handler{Service: actor.NewService()},
		Equip: &equip.Handler{Service: equip.NewService()},
	}
}

// NewService 创建游戏逻辑服务
func NewService() *Service {
	return &Service{
		Base:  base.NewService(),
		Hero:  hero.NewService(),
		Bag:   bag.NewService(),
		Actor: actor.NewService(),
		Equip: equip.NewService(),
	}
}
