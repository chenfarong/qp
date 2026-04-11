package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aoyo/qp/internal/gamelogic/actor"
	"github.com/aoyo/qp/internal/gamelogic/bag"
)

// MockInventoryService 模拟背包服务，用于测试
type MockInventoryService struct{}

// GetInventory 模拟获取背包的方法
func (s *MockInventoryService) GetInventory(characterID string) (*bag.InventoryResponse, error) {
	// 创建模拟数据
	items := []actor.InventoryItem{
		{
			ID:         "1",
			ItemType:   "weapon",
			ItemID:     "sword_001",
			Quantity:   1,
			IsEquipped: true,
			CreatedAt:  time.Now(),
		},
		{
			ID:         "2",
			ItemType:   "potion",
			ItemID:     "hp_potion_001",
			Quantity:   10,
			IsEquipped: false,
			CreatedAt:  time.Now(),
		},
		{
			ID:         "3",
			ItemType:   "armor",
			ItemID:     "shield_001",
			Quantity:   1,
			IsEquipped: true,
			CreatedAt:  time.Now(),
		},
	}

	return &bag.InventoryResponse{
		Inventory: bag.InventoryPayload{
			CharacterID: characterID,
			Items:       items,
		},
	}, nil
}

func main() {
	// 创建模拟背包服务
	inventoryService := &MockInventoryService{}

	// 测试获取背包
	characterID := "6618c5c07a6c8a2a2a2a2a2a"

	response, err := inventoryService.GetInventory(characterID)
	if err != nil {
		fmt.Printf("Failed to get inventory: %v\n", err)
		return
	}

	// 将结果转换为JSON并打印
	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshal JSON: %v\n", err)
		return
	}

	fmt.Println("Inventory Response:")
	fmt.Println(string(jsonData))
}
