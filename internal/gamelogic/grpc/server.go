package grpc

import (
	"context"
	"log"

	"github.com/aoyo/qp/internal/gamelogic/service"
	"github.com/aoyo/qp/pkg/proto/game"
)

// GameServer 游戏服务gRPC服务器
type GameServer struct {
	game.UnimplementedGameServiceServer
	gameService *service.GameService
}

// NewGameServer 创建游戏服务gRPC服务器实例
func NewGameServer(gameService *service.GameService) *GameServer {
	return &GameServer{
		gameService: gameService,
	}
}

// CreateCharacter 创建角色
func (s *GameServer) CreateCharacter(ctx context.Context, req *game.CreateCharacterRequest) (*game.CreateCharacterResponse, error) {
	// 转换请求参数
	createReq := service.CreateCharacterRequest{
		UserID: req.UserId,
		Name:   req.Name,
	}

	// 调用服务
	resp, err := s.gameService.CharacterService.CreateCharacter(createReq)
	if err != nil {
		return &game.CreateCharacterResponse{
			Error: err.Error(),
		}, nil
	}

	// 转换响应
	character := &game.Character{
		Id:           resp.Character.ID,
		UserId:       resp.Character.UserID,
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
	characters, err := s.gameService.CharacterService.GetCharactersByUserID(req.UserId)
	if err != nil {
		return &game.GetCharactersResponse{
			Error: err.Error(),
		}, nil
	}

	// 转换响应
	var gameCharacters []*game.Character
	for _, character := range characters {
		gameCharacter := &game.Character{
			Id:           character.ID,
			UserId:       character.UserID,
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
	character, err := s.gameService.CharacterService.GetCharacterByID(req.CharacterId)
	if err != nil {
		return &game.GetCharacterResponse{
			Error: err.Error(),
		}, nil
	}

	// 转换响应
	gameCharacter := &game.Character{
		Id:           character.ID,
		UserId:       character.UserID,
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
	err := s.gameService.CharacterService.UpdateCharacterStatus(req.CharacterId, int(req.Status))
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
	battleReq := service.BattleRequest{
		CharacterID: req.CharacterId,
		EnemyLevel:  int(req.EnemyLevel),
	}

	// 调用服务
	resp, err := s.gameService.BattleService.Battle(battleReq)
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

