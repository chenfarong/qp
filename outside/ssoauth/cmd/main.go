package main

import (
	"fmt"
	"log"
	"net/http"

	"zgame/internet/ssoauth/internal/handler"
	"zgame/internet/ssoauth/internal/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
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
	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	fmt.Println("SSO Auth Server started on port 8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
}
