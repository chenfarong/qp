package main

import (
	"fmt"
	"log"
	"net/http"

	"zgame/config"
	"zgame/internet/ssoauth/internal/handler"
	"zgame/internet/ssoauth/internal/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	config.LoadConfig()

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

	// 打印欢迎信息
	fmt.Println("========================================================")
	fmt.Println("                      SSO Auth Server                    ")
	fmt.Println("========================================================")
	fmt.Printf("服务名称: SSO Auth Server\n")
	fmt.Printf("服务类型: HTTP Server\n")
	fmt.Printf("监听地址: %s\n", config.AppConfig.Auth.Host)
	fmt.Printf("监听端口: %d\n", config.AppConfig.Auth.Port)
	fmt.Printf("Token过期时间: %d秒\n", config.AppConfig.Auth.TokenExpiry)
	fmt.Println("========================================================")
	fmt.Println("服务器已成功启动，等待客户端连接...")
	fmt.Println("========================================================")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
}
