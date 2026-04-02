package battle

import (
	"github.com/aoyo/qp/internal/gamelogic/actor"
	"github.com/aoyo/qp/pkg/db"
)

// BattleService 战斗服务
type BattleService struct {
	db               *db.DB
	dbName           string
	characterService *actor.CharacterService
}

// NewBattleService 创建战斗服务实例
func NewBattleService(db *db.DB, dbName string) *BattleService {
	return &BattleService{
		db:               db,
		dbName:           dbName,
		characterService: actor.NewCharacterService(db, dbName),
	}
}

// Battle 战斗逻辑
func (s *BattleService) Battle(req BattleRequest) (*BattleResponse, error) {
	// 获取角色信息
	character, err := s.characterService.GetCharacterByID(req.CharacterID)
	if err != nil {
		return nil, err
	}

	// 简单的战斗逻辑：基于等级和属性计算胜负
	playerPower := character.Level*10 + character.Strength + character.Agility/2 + character.Intelligence/3
	enemyPower := req.EnemyLevel*10 + 50 // 敌人基础属性

	var victory bool
	var expGained, goldGained int
	var message string

	if playerPower > enemyPower {
		victory = true
		expGained = req.EnemyLevel * 20
		goldGained = req.EnemyLevel * 5
		message = "You defeated the enemy!"

		// 添加经验和金币
		err = s.characterService.AddExp(req.CharacterID, expGained)
		if err != nil {
			return nil, err
		}
	} else {
		victory = false
		expGained = req.EnemyLevel * 5
		goldGained = 0
		message = "You were defeated by the enemy!"

		// 失败也获得少量经验
		err = s.characterService.AddExp(req.CharacterID, expGained)
		if err != nil {
			return nil, err
		}
	}

	return &BattleResponse{
		Victory:    victory,
		ExpGained:  expGained,
		GoldGained: goldGained,
		Message:    message,
	}, nil
}

// CreateCharacter 创建角色
func (s *BattleService) CreateCharacter(req actor.CreateCharacterRequest) (*actor.CharacterResponse, error) {
	return s.characterService.CreateCharacter(req)
}

// UseCharacter 使用角色
func (s *BattleService) UseCharacter(req actor.UseCharacterRequest) (*actor.UseCharacterResponse, error) {
	return s.characterService.UseCharacter(req)
}

// CharacterOffline 角色下线
func (s *BattleService) CharacterOffline(characterID string) error {
	return s.characterService.CharacterOffline(characterID)
}
