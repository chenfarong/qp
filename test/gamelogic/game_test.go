package gamelogic

import (
	"testing"

	"github.com/aoyo/qp/internal/gamelogic/service"
	"github.com/aoyo/qp/pkg/db"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestGameService(t *testing.T) {
	// 连接测试数据库
	uri := "mongodb://admin:password@localhost:27017/qp_game?authSource=admin"
	dbInstance, err := db.InitDB(uri)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbInstance.Close()

	// 初始化游戏服务
	gameService := service.NewGameService(dbInstance, "qp_game")

	// 生成测试用户ID
	userID := primitive.NewObjectID().Hex()

	// 测试创建角色
	createCharReq := service.CreateCharacterRequest{
		UserID: userID,
		Name:   "Test Character",
	}

	createCharResp, err := gameService.CreateCharacter(createCharReq)
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	if createCharResp.Character.ID.IsZero() {
		t.Fatalf("Character should have an ID")
	}

	// 测试获取角色列表
	characters, err := gameService.GetCharactersByUserID(userID)
	if err != nil {
		t.Fatalf("Failed to get characters: %v", err)
	}

	if len(characters) == 0 {
		t.Fatalf("Should have at least one character")
	}

	// 测试战斗
	battleReq := service.BattleRequest{
		CharacterID: createCharResp.Character.ID.Hex(),
		EnemyLevel:  1,
	}

	battleResp, err := gameService.Battle(battleReq)
	if err != nil {
		t.Fatalf("Failed to battle: %v", err)
	}

	if battleResp.ExpGained == 0 {
		t.Fatalf("Battle should reward experience")
	}

	t.Log("Game service tests passed!")
}
