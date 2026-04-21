package gm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"zagame/common/logger"
)

// GMCommand GM命令结构
type GMCommand struct {
	Command string          `json:"command"`
	Params  json.RawMessage `json:"params"`
}

// GMResponse GM响应结构
type GMResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Server GM服务器
type Server struct {
	port    int
	service *Service
}

// NewServer 创建GM服务器
func NewServer(port int, service *Service) *Server {
	return &Server{
		port:    port,
		service: service,
	}
}

// Start 启动GM服务器
func (s *Server) Start() error {
	// 注册HTTP路由
	http.HandleFunc("/gm", s.handleGMCommand)

	// 启动HTTP服务器
	serverAddr := fmt.Sprintf(":%d", s.port)
	logger.Info("============================================================")
	logger.Info("                          GM Server                          ")
	logger.Info("============================================================")
	logger.Info("服务名称: GM Server")
	logger.Info("服务类型: HTTP Server")
	logger.Info("监听地址: 0.0.0.0")
	logger.Info("监听端口: %d", s.port)
	logger.Info("============================================================")
	logger.Info("GM服务器已成功启动，等待命令...")
	logger.Info("============================================================")

	// 启动HTTP服务器
	server := &http.Server{
		Addr:         serverAddr,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return server.ListenAndServe()
}

// handleGMCommand 处理GM命令
func (s *Server) handleGMCommand(w http.ResponseWriter, r *http.Request) {
	// 解析请求
	var cmd GMCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		s.sendResponse(w, false, "无效的命令格式", nil)
		return
	}

	// 处理命令
	result, err := s.service.HandleCommand(cmd.Command, cmd.Params)
	if err != nil {
		s.sendResponse(w, false, err.Error(), nil)
		return
	}

	// 返回响应
	s.sendResponse(w, true, "命令执行成功", result)
}

// sendResponse 发送响应
func (s *Server) sendResponse(w http.ResponseWriter, success bool, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	response := GMResponse{
		Success: success,
		Message: message,
		Data:    data,
	}
	json.NewEncoder(w).Encode(response)
}
