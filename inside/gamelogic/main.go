package gamelogic

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"zagame/common/logger"
	"zagame/inside/gamelogic/actor"
	"zagame/inside/gamelogic/bag"
	"zagame/inside/gamelogic/base"
	"zagame/inside/gamelogic/common"
	"zagame/inside/gamelogic/equip"
	"zagame/inside/gamelogic/gm"
	"zagame/inside/gamelogic/grpc"
	"zagame/inside/gamelogic/grpc/client"
	"zagame/inside/gamelogic/hero"
	"zagame/inside/gamelogic/session"
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

// StartClient 启动gRPC客户端，连接到gateway
func StartClient(gatewayAddress string) error {
	// 创建各个处理器
	baseHandler := base.NewHandler()
	heroHandler := hero.NewHandler()
	bagHandler := bag.NewHandler()
	actorHandler := actor.NewHandler()
	equipHandler := equip.NewHandler()

	// 创建消息路由器
	router := grpc.NewRouter()

	// 统一注册所有处理器
	handlers := []common.Handler{
		baseHandler,
		heroHandler,
		bagHandler,
		actorHandler,
		equipHandler,
	}

	for _, h := range handlers {
		h.RegisterHandlers(router, h)
	}

	// 设置router实例到session包
	session.SetRouter(router)

	// 打印欢迎信息
	logger.Info("============================================================")
	logger.Info("                      GameLogic Server                      ")
	logger.Info("============================================================")
	logger.Info("服务名称: GameLogic Server")
	logger.Info("服务类型: gRPC Client")
	logger.Info("连接地址: %s", gatewayAddress)
	logger.Info("============================================================")
	logger.Info("正在连接到Gateway服务器...")
	logger.Info("============================================================")

	// 启动GM服务器
	go func() {
		gmService := gm.NewService()
		gmServer := gm.NewServer(8084, gmService)
		if err := gmServer.Start(); err != nil {
			logger.Errorf("启动GM服务器失败: %v", err)
		}
	}()

	// 连接到gateway服务
	cli, err := client.NewClient(gatewayAddress)
	if err != nil {
		logger.Errorf("连接到gateway服务失败: %v", err)
		return fmt.Errorf("连接到gateway服务失败: %v", err)
	}
	defer cli.Close()

	// 注册服务器
	err = cli.RegisterServer("gamelogic", "GameLogic Server", "", 0)
	if err != nil {
		logger.Errorf("注册服务器失败: %v", err)
		return fmt.Errorf("注册服务器失败: %v", err)
	}

	logger.Info("成功连接到Gateway服务器")
	logger.Info("服务器已成功启动，等待处理消息...")
	logger.Info("============================================================")

	// 等待信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("收到退出信号，正在关闭服务器...")
	return nil
}
