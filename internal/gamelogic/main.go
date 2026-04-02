package gamelogic

import (
	"fmt"
	"log"

	"github.com/aoyo/qp/pkg/db"
	"github.com/gin-gonic/gin"
)

// StartGameLogic 启动游戏逻辑服务
func StartGameLogic(db *db.DB, dbName string) error {
	// 创建游戏逻辑应用
	app := NewApp(db, dbName)

	// 启动应用
	if err := app.Start(); err != nil {
		return fmt.Errorf("failed to start game logic app: %w", err)
	}

	log.Println("Game logic service started successfully")
	return nil
}

// NewRouter 创建并配置路由器
func NewRouter(app *App) *gin.Engine {
	router := gin.Default()

	// 注册路由
	app.RegisterRoutes(router)

	return router
}

// GetApp 获取游戏逻辑应用实例
func GetApp(db *db.DB, dbName string) *App {
	return NewApp(db, dbName)
}
