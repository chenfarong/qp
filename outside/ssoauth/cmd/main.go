package main

import (
	"fmt"
	"net/http"

	"zagame/common/logger"
	"zagame/config"
	"zgame/internet/ssoauth/internal/handler"
	"zgame/internet/ssoauth/internal/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	config.LoadConfig()

	// 初始化日志系统
	logger.Init(logger.Config{
		ServerName: "ssoauth",
		Level:      logger.DEBUG,
		Outputs: []logger.OutputConfig{
			{Type: logger.Console},
			{Type: logger.File},
		},
		UDPServer: "",
		UDPPort:   0,
	})
	defer logger.Close()

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

	logger.InfoKV("启动SSO Auth服务器", "address", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("listen: %s\n", err)
	}
}
