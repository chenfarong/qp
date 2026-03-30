package main

import (
	"fmt"
	"log"
	"net"

	"github.com/aoyo/qp/internal/bill/grpc"
	"github.com/aoyo/qp/internal/bill/handler"
	"github.com/aoyo/qp/internal/bill/service"
	"github.com/aoyo/qp/pkg/db"
	"github.com/aoyo/qp/pkg/proto/bill"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func main() {
	// 加载配置
	// TODO: 从配置文件加载配置
	host := "localhost"
	port := 9100
	grpcPort := port + 1000
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
	
	// 启动gRPC服务器
	go startGRPCServer(paymentService, grpcPort)
	
	// 初始化处理器
	billHandler := handler.NewBillHandler(tokenService, paymentService)
	
	// 设置路由
	router := gin.Default()
	billHandler.RegisterRoutes(router)
	
	// 启动HTTP服务器
	serverAddr := fmt.Sprintf("%s:%d", host, port)
	log.Printf("Bill server starting on %s", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// startGRPCServer 启动gRPC服务器
func startGRPCServer(paymentService *service.PaymentService, port int) {
	// 创建gRPC服务器
	grpcServer := grpc.NewServer()

	// 注册账单服务
	bill.RegisterBillServiceServer(grpcServer, grpc.NewBillServer(paymentService))

	// 监听端口
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("Bill gRPC service starting on port %d...", port)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}
