package gamelogic

import (
	"github.com/aoyo/qp/internal/gamelogic/actor"
	"github.com/aoyo/qp/internal/gamelogic/bag"
	"github.com/aoyo/qp/internal/gamelogic/battle"
	"github.com/aoyo/qp/pkg/db"
	"github.com/gin-gonic/gin"
)

// Service 统一服务接口
type Service interface {
	// 创建角色
	CreateCharacter(req actor.CreateCharacterRequest) (*actor.CharacterResponse, error)
	// 使用角色
	UseCharacter(req actor.UseCharacterRequest) (*actor.UseCharacterResponse, error)
	// 角色下线
	CharacterOffline(characterID string) error
}

// App 游戏逻辑应用
type App struct {
	db              *db.DB
	dbName          string
	services        map[string]Service
	CharacterService *actor.CharacterService
	BattleService    *battle.BattleService
	InventoryService *bag.InventoryService
}

// NewApp 创建游戏逻辑应用实例
func NewApp(db *db.DB, dbName string) *App {
	// 创建角色服务
	characterService := actor.NewCharacterService(db, dbName)
	
	// 创建战斗服务
	battleService := battle.NewBattleService(db, dbName)
	
	// 创建背包服务
	inventoryService := bag.NewInventoryService(db, dbName)

	// 初始化服务映射
	services := make(map[string]Service)
	services["character"] = characterService
	services["battle"] = battleService
	services["inventory"] = inventoryService

	return &App{
		db:              db,
		dbName:          dbName,
		services:        services,
		CharacterService: characterService,
		BattleService:    battleService,
		InventoryService: inventoryService,
	}
}

// GetService 根据名称获取服务
func (a *App) GetService(name string) Service {
	return a.services[name]
}

// RegisterRoutes 注册路由
func (a *App) RegisterRoutes(router *gin.Engine) {
	gameGroup := router.Group("/api/game")
	{
		// 注册角色相关路由
		characterHandler := actor.NewCharacterHandler(a.CharacterService)
		characterHandler.RegisterRoutes(gameGroup)
		
		// 注册战斗相关路由
		battleHandler := battle.NewBattleHandler(a.BattleService)
		battleHandler.RegisterRoutes(gameGroup)
		
		// 注册背包相关路由
		inventoryHandler := bag.NewInventoryHandler(a.InventoryService)
		inventoryHandler.RegisterRoutes(gameGroup)
	}
}

// Start 启动应用
func (a *App) Start() error {
	// 这里可以添加启动逻辑，例如初始化服务、加载配置等
	return nil
}

// Stop 停止应用
func (a *App) Stop() error {
	// 这里可以添加停止逻辑，例如关闭连接、清理资源等
	return nil
}
