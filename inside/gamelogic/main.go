package gamelogic

import (
	"fmt"
	"net"

	"zagame/common/logger"
	"zagame/inside/gamelogic/actor"
	"zagame/inside/gamelogic/bag"
	"zagame/inside/gamelogic/base"
	"zagame/inside/gamelogic/common"
	"zagame/inside/gamelogic/equip"
	"zagame/inside/gamelogic/grpc"
	"zagame/inside/gamelogic/grpc/client"
	"zagame/inside/gamelogic/hero"

	gateway "zagame/inside/gamelogic/grpc/gateway"

	grpcserver "google.golang.org/grpc"
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

// StartGRPCServer 启动gRPC服务器
func StartGRPCServer(port int32, gatewayAddress string) error {
	// 添加panic恢复机制
	defer func() {
		if r := recover(); r != nil {
			logger.Panic("发生panic，已恢复: %v", r)
		}
	}()

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
		h.RegisterHandlers(router)
	}

	// 监听端口
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	// 创建gRPC服务器
	server := grpcserver.NewServer()

	// 注册Gateway服务
	gwServer := grpc.NewGatewayServer(router)
	gateway.RegisterGatewayServiceServer(server, gwServer)

	// 启动服务器
	logger.Infof("GameLogic gRPC Server started on %s", addr)

	// 连接到gateway服务并注册消息处理号段
	client, err := client.NewClient(gatewayAddress)
	if err != nil {
		logger.Errorf("连接到gateway服务失败: %v", err)
		// 继续启动服务器，因为客户端会自动重连
	} else {
		defer client.Close()

		// 注册服务器
		err = client.RegisterServer("gamelogic", "GameLogic Server", "localhost", port)
		if err != nil {
			logger.Errorf("注册服务器失败: %v", err)
			// 继续启动服务器，因为客户端会自动重连
		}
	}

	// 启动服务器
	return server.Serve(listener)
}
