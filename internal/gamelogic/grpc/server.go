package grpc

import (
	"context"

	"github.com/aoyo/qp/internal/gamelogic"
	"github.com/aoyo/qp/pkg/proto/game"
)

// GameServer 游戏服务gRPC服务器
type GameServer struct {
	game.UnimplementedGameServiceServer
	app *gamelogic.App
}

// NewGameServer 创建游戏服务gRPC服务器实例
func NewGameServer(app *gamelogic.App) *GameServer {
	return &GameServer{
		app: app,
	}
}

// CreateCharacter 创建角色
func (s *GameServer) CreateCharacter(ctx context.Context, req *game.CreateCharacterRequest) (*game.CreateCharacterResponse, error) {
	// 转换请求参数
	createReq := struct {
		UserID string `json:"user_id" binding:"required"`
		Name   string `json:"name" binding:"required,min=2,max=50"`
	}{
		UserID: string(req.UserId),
		Name:   req.Name,
	}

	// 调用服务
	resp, err := s.app.CharacterService.CreateCharacter(createReq)
	if err != nil {
		return &game.CreateCharacterResponse{
			Error: err.Error(),
		}, nil
	}

	// 转换响应
	character := &game.Character{
		Id:           1, // 临时值，实际应该从数据库获取
		UserId:       req.UserId,
		Name:         resp.Character.Name,
		Level:        int32(resp.Character.Level),
		Exp:          int32(resp.Character.Exp),
		Hp:           int32(resp.Character.HP),
		Mp:           int32(resp.Character.MP),
		Strength:     int32(resp.Character.Strength),
		Agility:      int32(resp.Character.Agility),
		Intelligence: int32(resp.Character.Intelligence),
		Gold:         int32(resp.Character.Gold),
		Status:       int32(resp.Character.Status),
	}

	return &game.CreateCharacterResponse{
		Character: character,
	}, nil
}

// GetCharacters 获取用户的所有角色
func (s *GameServer) GetCharacters(ctx context.Context, req *game.GetCharactersRequest) (*game.GetCharactersResponse, error) {
	// 调用服务
	characters, err := s.app.CharacterService.GetCharactersByUserID(string(req.UserId))
	if err != nil {
		return &game.GetCharactersResponse{
			Error: err.Error(),
		}, nil
	}

	// 转换响应
	var gameCharacters []*game.Character
	for i, character := range characters {
		gameCharacter := &game.Character{
			Id:           uint32(i + 1), // 临时值，实际应该从数据库获取
			UserId:       req.UserId,
			Name:         character.Name,
			Level:        int32(character.Level),
			Exp:          int32(character.Exp),
			Hp:           int32(character.HP),
			Mp:           int32(character.MP),
			Strength:     int32(character.Strength),
			Agility:      int32(character.Agility),
			Intelligence: int32(character.Intelligence),
			Gold:         int32(character.Gold),
			Status:       int32(character.Status),
		}
		gameCharacters = append(gameCharacters, gameCharacter)
	}

	return &game.GetCharactersResponse{
		Characters: gameCharacters,
	}, nil
}

// GetCharacter 获取角色详情
func (s *GameServer) GetCharacter(ctx context.Context, req *game.GetCharacterRequest) (*game.GetCharacterResponse, error) {
	// 调用服务
	character, err := s.app.CharacterService.GetCharacterByID(string(req.CharacterId))
	if err != nil {
		return &game.GetCharacterResponse{
			Error: err.Error(),
		}, nil
	}

	// 转换响应
	gameCharacter := &game.Character{
		Id:           req.CharacterId,
		UserId:       1, // 临时值，实际应该从数据库获取
		Name:         character.Name,
		Level:        int32(character.Level),
		Exp:          int32(character.Exp),
		Hp:           int32(character.HP),
		Mp:           int32(character.MP),
		Strength:     int32(character.Strength),
		Agility:      int32(character.Agility),
		Intelligence: int32(character.Intelligence),
		Gold:         int32(character.Gold),
		Status:       int32(character.Status),
	}

	return &game.GetCharacterResponse{
		Character: gameCharacter,
	}, nil
}

// UpdateCharacterStatus 更新角色状态
func (s *GameServer) UpdateCharacterStatus(ctx context.Context, req *game.UpdateCharacterStatusRequest) (*game.UpdateCharacterStatusResponse, error) {
	// 调用服务
	err := s.app.CharacterService.UpdateCharacterStatus(string(req.CharacterId), int(req.Status))
	if err != nil {
		return &game.UpdateCharacterStatusResponse{
			Error: err.Error(),
		}, nil
	}

	return &game.UpdateCharacterStatusResponse{
		Message: "Status updated successfully",
	}, nil
}

// Battle 战斗
func (s *GameServer) Battle(ctx context.Context, req *game.BattleRequest) (*game.BattleResponse, error) {
	// 转换请求参数
	battleReq := struct {
		CharacterID string `json:"character_id" binding:"required"`
		EnemyLevel  int    `json:"enemy_level" binding:"required,min=1"`
	}{
		CharacterID: string(req.CharacterId),
		EnemyLevel:  int(req.EnemyLevel),
	}

	// 调用服务
	resp, err := s.app.BattleService.Battle(battleReq)
	if err != nil {
		return &game.BattleResponse{
			Error: err.Error(),
		}, nil
	}

	return &game.BattleResponse{
		Victory:    resp.Victory,
		ExpGained:  int32(resp.ExpGained),
		GoldGained: int32(resp.GoldGained),
		Message:    resp.Message,
	}, nil
}
