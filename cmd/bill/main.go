package main

import (
	"fmt"
	"log"

	"github.com/aoyo/qp/internal/bill/handler"
	"github.com/aoyo/qp/internal/bill/service"
	"github.com/aoyo/qp/pkg/db"
	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	// TODO: 从配置文件加载配置
	host := "localhost"
	port := 9100
	dbURI := "mongodb://admin:password@localhost:27017/qp_game?authSource=admin"
	dbName := "qp_game"
	
	// 初始化数据库连接
	dbInstance, err := db.InitDB(dbURI)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbInstance.Close()
	
	// 初始化服务
	tokenService := service.NewTokenService(dbInstance, dbName)
	paymentService := service.NewPaymentService(dbInstance, dbName, tokenService)
	
	// 初始化处理器
	billHandler := handler.NewBillHandler(tokenService, paymentService)
	
	// 设置路由
	router := gin.Default()
	billHandler.RegisterRoutes(router)
	
	// 启动服务器
	serverAddr := fmt.Sprintf("%s:%d", host, port)
	log.Printf("Bill server starting on %s", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
