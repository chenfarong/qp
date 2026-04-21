package gm

import (
	"encoding/json"
	"fmt"

	"zagame/common/logger"
)

// Service GM服务
type Service struct {
	// 可以在这里添加其他依赖，比如数据库连接等
}

// NewService 创建GM服务
func NewService() *Service {
	return &Service{}
}

// HandleCommand 处理GM命令
func (s *Service) HandleCommand(command string, params json.RawMessage) (interface{}, error) {
	logger.Infof("收到GM命令: %s, 参数: %s", command, params)

	// 根据命令类型处理
	switch command {
	case "ping":
		return s.handlePing()
	case "get_server_status":
		return s.handleGetServerStatus()
	case "broadcast":
		return s.handleBroadcast(params)
	case "kick_player":
		return s.handleKickPlayer(params)
	case "give_item":
		return s.handleGiveItem(params)
	default:
		return nil, fmt.Errorf("未知命令: %s", command)
	}
}

// handlePing 处理ping命令
func (s *Service) handlePing() (interface{}, error) {
	return map[string]string{"message": "pong"}, nil
}

// handleGetServerStatus 处理获取服务器状态命令
func (s *Service) handleGetServerStatus() (interface{}, error) {
	// 这里可以返回服务器的各种状态信息
	return map[string]interface{}{
		"status":  "running",
		"version": "1.0.0",
		"time":    fmt.Sprintf("%v", json.RawMessage(`"2026-04-21 12:00:00"`)),
	}, nil
}

// BroadcastParams 广播命令参数
type BroadcastParams struct {
	Message string `json:"message"`
}

// handleBroadcast 处理广播命令
func (s *Service) handleBroadcast(params json.RawMessage) (interface{}, error) {
	var p BroadcastParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("参数解析失败: %v", err)
	}

	// 这里实现广播逻辑
	logger.Infof("广播消息: %s", p.Message)

	return map[string]string{"message": "广播成功"}, nil
}

// KickPlayerParams 踢人命令参数
type KickPlayerParams struct {
	PlayerID string `json:"player_id"`
	Reason   string `json:"reason"`
}

// handleKickPlayer 处理踢人命令
func (s *Service) handleKickPlayer(params json.RawMessage) (interface{}, error) {
	var p KickPlayerParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("参数解析失败: %v", err)
	}

	// 这里实现踢人逻辑
	logger.Infof("踢人: %s, 原因: %s", p.PlayerID, p.Reason)

	return map[string]string{"message": "踢人成功"}, nil
}

// GiveItemParams 给物品命令参数
type GiveItemParams struct {
	PlayerID string `json:"player_id"`
	ItemID   string `json:"item_id"`
	Count    int    `json:"count"`
}

// handleGiveItem 处理给物品命令
func (s *Service) handleGiveItem(params json.RawMessage) (interface{}, error) {
	var p GiveItemParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("参数解析失败: %v", err)
	}

	// 这里实现给物品逻辑
	logger.Infof("给物品: 玩家=%s, 物品ID=%s, 数量=%d", p.PlayerID, p.ItemID, p.Count)

	return map[string]string{"message": "给物品成功"}, nil
}
