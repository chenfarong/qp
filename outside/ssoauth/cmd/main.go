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
