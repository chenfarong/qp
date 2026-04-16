package main

import (
	"fmt"
	"log"
	"net/http"

	"zgame/config"
	"zgame/database"
	"zgame/internet/ssoauth/internal/handler"
	"zgame/internet/ssoauth/internal/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	config.LoadConfig()

	// 初始化数据库连接
	if err := database.InitDatabase(); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer database.CloseDatabase()

	// 检查并创建默认的za_admin账号
	handler.CheckAndCreateDefaultAdmin()

	r := gin.Default()

	// 中间件
	r.Use(middleware.CORS())

	// 路由
	r.POST("/login", handler.Login)
	r.POST("/register", handler.Register)

	// 需要JWT验证的路由
	auth := r.Group("/auth")
	auth.Use(middleware.JWT())
	{
		auth.GET("/profile", handler.Profile)
	}

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", config.AppConfig.Auth.Host, config.AppConfig.Auth.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	fmt.Printf("SSO Auth Server started on %s\n", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
}
