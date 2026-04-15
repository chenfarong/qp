package gamelogic

import (
	"zgame/inside/gamelogic/actor"
	"zgame/inside/gamelogic/bag"
	"zgame/inside/gamelogic/base"
	"zgame/inside/gamelogic/hero"
)

// Handler 游戏逻辑处理器
type Handler struct {
	Base  *base.Handler
	Hero  *hero.Handler
	Bag   *bag.Handler
	Actor *actor.Handler
}

// Service 游戏逻辑服务
type Service struct {
	Base  *base.Service
	Hero  *hero.Service
	Bag   *bag.Service
	Actor *actor.Service
}

// NewHandler 创建游戏逻辑处理器
func NewHandler() *Handler {
	return &Handler{
		Base:  &base.Handler{Service: base.NewService()},
		Hero:  &hero.Handler{Service: hero.NewService()},
		Bag:   &bag.Handler{Service: bag.NewService()},
		Actor: &actor.Handler{Service: actor.NewService()},
	}
}

// NewService 创建游戏逻辑服务
func NewService() *Service {
	return &Service{
		Base:  base.NewService(),
		Hero:  hero.NewService(),
		Bag:   bag.NewService(),
		Actor: actor.NewService(),
	}
}
