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

// 确保 BattleService 实现了 Service 接口
var _ interface{} = (*BattleService)(nil)

// NewBattleService 创建战斗服务实例（characterService 须与 App 共用，以保证在线角色内存一致）
func NewBattleService(db *db.DB, dbName string, characterService *actor.CharacterService) *BattleService {
	return &BattleService{
		db:               db,
		dbName:           dbName,
		characterService: characterService,
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

// CharacterLogin 角色登录
func (s *BattleService) CharacterLogin(characterID string) error {
	return s.characterService.CharacterLogin(characterID)
}

// CharacterLogout 角色登出
func (s *BattleService) CharacterLogout(characterID string) error {
	return s.characterService.CharacterLogout(characterID)
}

// CharacterOnline 角色上线
func (s *BattleService) CharacterOnline(characterID string) error {
	return s.characterService.CharacterOnline(characterID)
}

// CharacterOffline 角色下线
func (s *BattleService) CharacterOffline(characterID string) error {
	return s.characterService.CharacterOffline(characterID)
}

// HandleInternalMessage 处理内部消息；未识别类型时返回 handled=false
func (s *BattleService) HandleInternalMessage(messageType string, messageData []byte) (bool, error) {
	return s.characterService.HandleInternalMessage(messageType, messageData)
}
