package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aoyo/qp/internal/gamelogic/actor"
	"github.com/aoyo/qp/internal/gamelogic/bag"
	"github.com/aoyo/qp/pkg/db"
)

func main() {
	// 初始化数据库连接
	dbClient, err := db.InitDB("mongodb://admin:password@localhost:27017/admin")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbClient.Close()

	// 初始化角色服务
	characterService := actor.NewCharacterService(dbClient, "qp_game")

	// 初始化背包服务
	inventoryService := bag.NewInventoryService(dbClient, "qp_game", characterService)

	// 测试获取背包
	// 注意：这里需要使用一个已存在的角色ID
	characterID := "6618c5c07a6c8a2a2a2a2a2a" // 替换为实际的角色ID

	response, err := inventoryService.GetInventory(characterID)
	if err != nil {
		log.Fatalf("Failed to get inventory: %v", err)
	}

	// 将结果转换为JSON并打印
	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	fmt.Println("Inventory Response:")
	fmt.Println(string(jsonData))
}