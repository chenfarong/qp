package service

import (
	"errors"
	"time"

	"github.com/aoyo/qp/internal/gamelogic/model"
	"github.com/aoyo/qp/pkg/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GameService 游戏服务
type GameService struct {
	db     *db.DB
	dbName string
}

// NewGameService 创建游戏服务实例
func NewGameService(db *db.DB, dbName string) *GameService {
	return &GameService{
		db:     db,
		dbName: dbName,
	}
}

// CreateCharacterRequest 创建角色请求
type CreateCharacterRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Name   string `json:"name" binding:"required,min=2,max=50"`
}

// CharacterResponse 角色响应
type CharacterResponse struct {
	Character model.Character `json:"character"`
}

// CreateCharacter 创建角色
func (s *GameService) CreateCharacter(req CreateCharacterRequest) (*CharacterResponse, error) {
	collection := s.db.GetCollection(s.dbName, model.Character{}.CollectionName())

	// 解析用户ID
	userID, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		return nil, err
	}

	// 检查用户是否已存在角色
	cursor, err := collection.Find(s.db.Ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(s.db.Ctx)

	var characters []model.Character
	if err := cursor.All(s.db.Ctx, &characters); err != nil {
		return nil, err
	}

	if len(characters) >= 3 {
		return nil, errors.New("user can only create up to 3 characters")
	}

	// 检查角色名是否已存在
	var existingCharacter model.Character
	if err := collection.FindOne(s.db.Ctx, bson.M{"name": req.Name}).Decode(&existingCharacter); err == nil {
		return nil, errors.New("character name already exists")
	} else if err != mongo.ErrNoDocuments {
		return nil, err
	}

	// 创建角色
	now := time.Now()
	character := model.Character{
		ID:           primitive.NewObjectID(),
		CreatedAt:    now,
		UpdatedAt:    now,
		UserID:       userID,
		Name:         req.Name,
		Level:        1,
		Exp:          0,
		HP:           100,
		MP:           50,
		Strength:     10,
		Agility:      10,
		Intelligence: 10,
		Gold:         0,
		Status:       1,
	}

	if _, err := collection.InsertOne(s.db.Ctx, character); err != nil {
		return nil, err
	}

	return &CharacterResponse{
		Character: character,
	}, nil
}

// GetCharactersByUserID 获取用户的所有角色
func (s *GameService) GetCharactersByUserID(userID string) ([]model.Character, error) {
	collection := s.db.GetCollection(s.dbName, model.Character{}.CollectionName())

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	cursor, err := collection.Find(s.db.Ctx, bson.M{"user_id": objectID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(s.db.Ctx)

	var characters []model.Character
	if err := cursor.All(s.db.Ctx, &characters); err != nil {
		return nil, err
	}

	return characters, nil
}

// GetCharacterByID 根据ID获取角色信息
func (s *GameService) GetCharacterByID(characterID string) (*model.Character, error) {
	collection := s.db.GetCollection(s.dbName, model.Character{}.CollectionName())

	objectID, err := primitive.ObjectIDFromHex(characterID)
	if err != nil {
		return nil, err
	}

	var character model.Character
	if err := collection.FindOne(s.db.Ctx, bson.M{"_id": objectID}).Decode(&character); err != nil {
		return nil, err
	}

	return &character, nil
}

// UpdateCharacterStatus 更新角色状态
func (s *GameService) UpdateCharacterStatus(characterID string, status int) error {
	collection := s.db.GetCollection(s.dbName, model.Character{}.CollectionName())

	objectID, err := primitive.ObjectIDFromHex(characterID)
	if err != nil {
		return err
	}

	_, err = collection.UpdateOne(s.db.Ctx, bson.M{"_id": objectID}, bson.M{"$set": bson.M{"status": status, "updated_at": time.Now()}})
	return err
}

// AddExp 添加经验值
func (s *GameService) AddExp(characterID string, exp int) error {
	collection := s.db.GetCollection(s.dbName, model.Character{}.CollectionName())

	objectID, err := primitive.ObjectIDFromHex(characterID)
	if err != nil {
		return err
	}

	var character model.Character
	if err := collection.FindOne(s.db.Ctx, bson.M{"_id": objectID}).Decode(&character); err != nil {
		return err
	}

	character.Exp += exp

	// 检查是否升级
	for character.Exp >= s.getRequiredExp(character.Level) {
		character.Exp -= s.getRequiredExp(character.Level)
		character.Level++
		character.HP += 10
		character.MP += 5
		character.Strength += 2
		character.Agility += 1
		character.Intelligence += 1
	}

	character.UpdatedAt = time.Now()
	_, err = collection.UpdateOne(s.db.Ctx, bson.M{"_id": objectID}, bson.M{"$set": character})
	return err
}

// getRequiredExp 获取升级所需经验值
func (s *GameService) getRequiredExp(level int) int {
	return level * 100
}

// BattleRequest 战斗请求
type BattleRequest struct {
	CharacterID string `json:"character_id" binding:"required"`
	EnemyLevel  int    `json:"enemy_level" binding:"required,min=1"`
}

// BattleResponse 战斗响应
type BattleResponse struct {
	Victory    bool   `json:"victory"`
	ExpGained  int    `json:"exp_gained"`
	GoldGained int    `json:"gold_gained"`
	Message    string `json:"message"`
}

// Battle 战斗逻辑
func (s *GameService) Battle(req BattleRequest) (*BattleResponse, error) {
	collection := s.db.GetCollection(s.dbName, model.Character{}.CollectionName())

	objectID, err := primitive.ObjectIDFromHex(req.CharacterID)
	if err != nil {
		return nil, err
	}

	var character model.Character
	if err := collection.FindOne(s.db.Ctx, bson.M{"_id": objectID}).Decode(&character); err != nil {
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
		character.Exp += expGained
		character.Gold += goldGained

		// 检查是否升级
		for character.Exp >= s.getRequiredExp(character.Level) {
			character.Exp -= s.getRequiredExp(character.Level)
			character.Level++
			character.HP += 10
			character.MP += 5
			character.Strength += 2
			character.Agility += 1
			character.Intelligence += 1
			message += " You leveled up!"
		}

		character.UpdatedAt = time.Now()
		_, err = collection.UpdateOne(s.db.Ctx, bson.M{"_id": objectID}, bson.M{"$set": character})
		if err != nil {
			return nil, err
		}
	} else {
		victory = false
		expGained = req.EnemyLevel * 5
		goldGained = 0
		message = "You were defeated by the enemy!"

		// 失败也获得少量经验
		character.Exp += expGained
		character.UpdatedAt = time.Now()
		_, err = collection.UpdateOne(s.db.Ctx, bson.M{"_id": objectID}, bson.M{"$set": character})
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
