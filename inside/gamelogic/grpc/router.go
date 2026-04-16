package grpc

import (
	"context"
	"fmt"
	"log"
	"sync"

	"zagame/inside/gamelogic/actor"
	"zagame/inside/gamelogic/bag"
	"zagame/inside/gamelogic/base"
	"zagame/inside/gamelogic/equip"
	"zagame/inside/gamelogic/hero"
	pb "zagame/pb/golang/gamelogic"

	"google.golang.org/protobuf/proto"
)

// MessageHandler 消息处理器类型
type MessageHandler func(ctx context.Context, session string, messageContent []byte) ([]byte, error)

// Router 消息路由器
type Router struct {
	handlers map[int32]MessageHandler
	mu       sync.RWMutex
}

// NewRouter 创建消息路由器
func NewRouter() *Router {
	return &Router{
		handlers: make(map[int32]MessageHandler),
	}
}

// RegisterHandler 注册消息处理器
func (r *Router) RegisterHandler(messageID int32, handler MessageHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[messageID] = handler
	log.Printf("注册消息处理器: messageID=%d\n", messageID)
}

// HandleMessage 处理消息
func (r *Router) HandleMessage(ctx context.Context, messageID int32, session string, messageContent []byte) ([]byte, error) {
	r.mu.RLock()
	handler, ok := r.handlers[messageID]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("未找到消息处理器: messageID=%d", messageID)
	}

	return handler(ctx, session, messageContent)
}

// InitHandlers 初始化消息处理器
func InitHandlers(router *Router, baseHandler *base.Handler, heroHandler *hero.Handler, bagHandler *bag.Handler, actorHandler *actor.Handler, equipHandler *equip.Handler) {
	// 基础消息
	router.RegisterHandler(MessageIDLoginRequest, handleLoginRequest(baseHandler))
	router.RegisterHandler(MessageIDGetRoleInfoRequest, handleGetRoleInfoRequest(baseHandler))

	// 角色消息
	router.RegisterHandler(MessageIDActorCreateRequest, handleActorCreateRequest(actorHandler))
	router.RegisterHandler(MessageIDActorUseRequest, handleActorUseRequest(actorHandler))
	router.RegisterHandler(MessageIDActorUseWithNameRequest, handleActorUseWithNameRequest(actorHandler))

	// 背包消息
	router.RegisterHandler(MessageIDGetBagRequest, handleGetBagRequest(bagHandler))
	router.RegisterHandler(MessageIDBagItemUseRequest, handleBagItemUseRequest(bagHandler))

	// 装备消息
	router.RegisterHandler(MessageIDGetEquipRequest, handleGetEquipRequest(equipHandler))
	router.RegisterHandler(MessageIDUpgradeEquipRequest, handleUpgradeEquipRequest(equipHandler))

	// 英雄消息
	router.RegisterHandler(MessageIDGetHeroesRequest, handleGetHeroesRequest(heroHandler))
	router.RegisterHandler(MessageIDRecruitHeroRequest, handleRecruitHeroRequest(heroHandler))
	router.RegisterHandler(MessageIDUpStarHeroRequest, handleUpStarHeroRequest(heroHandler))
	router.RegisterHandler(MessageIDOpenSkillHeroRequest, handleOpenSkillHeroRequest(heroHandler))
	router.RegisterHandler(MessageIDUpSkillHeroRequest, handleUpSkillHeroRequest(heroHandler))

	// 货币消息
	router.RegisterHandler(MessageIDGetGameMoneyRequest, handleGetGameMoneyRequest(baseHandler))
}

// 处理登录请求
func handleLoginRequest(handler *base.Handler) MessageHandler {
	return func(ctx context.Context, session string, messageContent []byte) ([]byte, error) {
		req := &pb.LoginRequest{}
		if err := proto.Unmarshal(messageContent, req); err != nil {
			return nil, err
		}

		resp, err := handler.Login(ctx, req)
		if err != nil {
			return nil, err
		}

		return proto.Marshal(resp)
	}
}

// 处理获取角色信息请求
func handleGetRoleInfoRequest(handler *base.Handler) MessageHandler {
	return func(ctx context.Context, session string, messageContent []byte) ([]byte, error) {
		req := &pb.GetRoleInfoRequest{}
		if err := proto.Unmarshal(messageContent, req); err != nil {
			return nil, err
		}

		resp, err := handler.GetRoleInfo(ctx, req)
		if err != nil {
			return nil, err
		}

		return proto.Marshal(resp)
	}
}

// 处理创建角色请求
func handleActorCreateRequest(handler *actor.Handler) MessageHandler {
	return func(ctx context.Context, session string, messageContent []byte) ([]byte, error) {
		req := &pb.ActorCreateRequest{}
		if err := proto.Unmarshal(messageContent, req); err != nil {
			return nil, err
		}

		resp, err := handler.ActorCreate(ctx, req)
		if err != nil {
			return nil, err
		}

		return proto.Marshal(resp)
	}
}

// 处理使用角色请求
func handleActorUseRequest(handler *actor.Handler) MessageHandler {
	return func(ctx context.Context, session string, messageContent []byte) ([]byte, error) {
		req := &pb.ActorUseRequest{}
		if err := proto.Unmarshal(messageContent, req); err != nil {
			return nil, err
		}

		resp, err := handler.ActorUse(ctx, req)
		if err != nil {
			return nil, err
		}

		return proto.Marshal(resp)
	}
}

// 处理使用角色请求（通过名称）
func handleActorUseWithNameRequest(handler *actor.Handler) MessageHandler {
	return func(ctx context.Context, session string, messageContent []byte) ([]byte, error) {
		req := &pb.ActorUseWithNameRequest{}
		if err := proto.Unmarshal(messageContent, req); err != nil {
			return nil, err
		}

		resp, err := handler.ActorUseWithName(ctx, req)
		if err != nil {
			return nil, err
		}

		return proto.Marshal(resp)
	}
}

// 处理获取背包请求
func handleGetBagRequest(handler *bag.Handler) MessageHandler {
	return func(ctx context.Context, session string, messageContent []byte) ([]byte, error) {
		req := &pb.GetBagRequest{}
		if err := proto.Unmarshal(messageContent, req); err != nil {
			return nil, err
		}

		resp, err := handler.GetBag(ctx, req)
		if err != nil {
			return nil, err
		}

		return proto.Marshal(resp)
	}
}

// 处理背包物品使用请求
func handleBagItemUseRequest(handler *bag.Handler) MessageHandler {
	return func(ctx context.Context, session string, messageContent []byte) ([]byte, error) {
		req := &pb.BagItemUseRequest{}
		if err := proto.Unmarshal(messageContent, req); err != nil {
			return nil, err
		}

		resp, err := handler.BagItemUse(ctx, req)
		if err != nil {
			return nil, err
		}

		return proto.Marshal(resp)
	}
}

// 处理获取装备请求
func handleGetEquipRequest(handler *equip.Handler) MessageHandler {
	return func(ctx context.Context, session string, messageContent []byte) ([]byte, error) {
		req := &pb.GetEquipRequest{}
		if err := proto.Unmarshal(messageContent, req); err != nil {
			return nil, err
		}

		resp, err := handler.GetEquip(ctx, req)
		if err != nil {
			return nil, err
		}

		return proto.Marshal(resp)
	}
}

// 处理升级装备请求
func handleUpgradeEquipRequest(handler *equip.Handler) MessageHandler {
	return func(ctx context.Context, session string, messageContent []byte) ([]byte, error) {
		req := &pb.UpgradeEquipRequest{}
		if err := proto.Unmarshal(messageContent, req); err != nil {
			return nil, err
		}

		resp, err := handler.UpgradeEquip(ctx, req)
		if err != nil {
			return nil, err
		}

		return proto.Marshal(resp)
	}
}

// 处理获取英雄请求
func handleGetHeroesRequest(handler *hero.Handler) MessageHandler {
	return func(ctx context.Context, session string, messageContent []byte) ([]byte, error) {
		req := &pb.GetHeroesRequest{}
		if err := proto.Unmarshal(messageContent, req); err != nil {
			return nil, err
		}

		resp, err := handler.GetHeroes(ctx, req)
		if err != nil {
			return nil, err
		}

		return proto.Marshal(resp)
	}
}

// 处理招募英雄请求
func handleRecruitHeroRequest(handler *hero.Handler) MessageHandler {
	return func(ctx context.Context, session string, messageContent []byte) ([]byte, error) {
		req := &pb.RecruitHeroRequest{}
		if err := proto.Unmarshal(messageContent, req); err != nil {
			return nil, err
		}

		resp, err := handler.RecruitHero(ctx, req)
		if err != nil {
			return nil, err
		}

		return proto.Marshal(resp)
	}
}

// 处理英雄升星请求
func handleUpStarHeroRequest(handler *hero.Handler) MessageHandler {
	return func(ctx context.Context, session string, messageContent []byte) ([]byte, error) {
		req := &pb.UpStarHeroRequest{}
		if err := proto.Unmarshal(messageContent, req); err != nil {
			return nil, err
		}

		resp, err := handler.UpStarHero(ctx, req)
		if err != nil {
			return nil, err
		}

		return proto.Marshal(resp)
	}
}

// 处理开启技能请求
func handleOpenSkillHeroRequest(handler *hero.Handler) MessageHandler {
	return func(ctx context.Context, session string, messageContent []byte) ([]byte, error) {
		req := &pb.OpenSkillHeroRequest{}
		if err := proto.Unmarshal(messageContent, req); err != nil {
			return nil, err
		}

		resp, err := handler.OpenSkillHero(ctx, req)
		if err != nil {
			return nil, err
		}

		return proto.Marshal(resp)
	}
}

// 处理升级技能请求
func handleUpSkillHeroRequest(handler *hero.Handler) MessageHandler {
	return func(ctx context.Context, session string, messageContent []byte) ([]byte, error) {
		req := &pb.UpSkillHeroRequest{}
		if err := proto.Unmarshal(messageContent, req); err != nil {
			return nil, err
		}

		resp, err := handler.UpSkillHero(ctx, req)
		if err != nil {
			return nil, err
		}

		return proto.Marshal(resp)
	}
}

// 处理获取游戏货币请求
func handleGetGameMoneyRequest(handler *base.Handler) MessageHandler {
	return func(ctx context.Context, session string, messageContent []byte) ([]byte, error) {
		req := &pb.GetGameMoneyRequest{}
		if err := proto.Unmarshal(messageContent, req); err != nil {
			return nil, err
		}

		resp, err := handler.GetGameMoney(ctx, req)
		if err != nil {
			return nil, err
		}

		return proto.Marshal(resp)
	}
}
