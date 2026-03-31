package service

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/aoyo/qp/internal/gamelogic/model"
	"github.com/aoyo/qp/pkg/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// CharacterChannel 角色通道结构
type CharacterChannel struct {
	Character *model.Character
	Ch        chan interface{}
	LastUsed  time.Time
}

// GameService 游戏服务
type GameService struct {
	db              *db.DB
	dbName          string
	characterCache  map[string]*CharacterChannel // 角色缓存，key为角色ID
	cacheMutex      sync.RWMutex                 // 缓存互斥锁
	cleanupInterval time.Duration                // 缓存清理间隔
}

// NewGameService 创建游戏服务实例
func NewGameService(db *db.DB, dbName string) *GameService {
	service := &GameService{
		db:              db,
		dbName:          dbName,
		characterCache:  make(map[string]*CharacterChannel),
		cleanupInterval: time.Hour, // 每小时清理一次
	}

	// 启动缓存清理 goroutine
	go service.cleanupCache()

	return service
}

// cleanupCache 清理过期的角色缓存
func (s *GameService) cleanupCache() {
	ticker := time.NewTicker(s.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		s.cacheMutex.Lock()
		for characterID, cc := range s.characterCache {
			// 清理超过24小时未使用的角色缓存
			if time.Since(cc.LastUsed) > 24*time.Hour {
				close(cc.Ch)
				delete(s.characterCache, characterID)
				log.Printf("Cleaned up character cache for character ID: %s", characterID)
			}
		}
		s.cacheMutex.Unlock()
	}
}

// cacheCharacter 将角色缓存到内存并分配channel
func (s *GameService) cacheCharacter(character *model.Character) {
	characterID := character.ID.Hex()

	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	// 检查角色是否已在缓存中
	if _, exists := s.characterCache[characterID]; exists {
		// 更新缓存中的角色信息
		s.characterCache[characterID].Character = character
		s.characterCache[characterID].LastUsed = time.Now()
		return
	}

	// 创建角色通道
	ch := make(chan interface{}, 100)

	// 创建角色通道结构
	cc := &CharacterChannel{
		Character: character,
		Ch:        ch,
		LastUsed:  time.Now(),
	}

	// 存储到缓存
	s.characterCache[characterID] = cc

	// 启动角色消息处理 goroutine
	go s.handleCharacterMessages(characterID, ch)

	log.Printf("Character cached: %s", characterID)
}

// handleCharacterMessages 处理角色消息
func (s *GameService) handleCharacterMessages(characterID string, ch chan interface{}) {
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				// 通道已关闭
				log.Printf("Character channel closed: %s", characterID)
				return
			}

			// 处理消息
			log.Printf("Processing message for character %s: %v", characterID, msg)

			// 这里可以根据消息类型执行不同的处理逻辑
			// 例如：更新角色状态、处理战斗请求、使用物品等

			// 模拟消息处理延迟
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// getCharacterFromCache 从缓存中获取角色
func (s *GameService) getCharacterFromCache(characterID string) (*CharacterChannel, bool) {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	cc, exists := s.characterCache[characterID]
	if exists {
		// 更新最后使用时间
		cc.LastUsed = time.Now()
	}

	return cc, exists
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

	// 将角色缓存到内存并分配channel
	s.cacheCharacter(&character)

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
	// 先从缓存中获取角色
	cc, exists := s.getCharacterFromCache(characterID)
	if exists {
		return cc.Character, nil
	}

	// 缓存中不存在，从数据库获取
	collection := s.db.GetCollection(s.dbName, model.Character{}.CollectionName())

	objectID, err := primitive.ObjectIDFromHex(characterID)
	if err != nil {
		return nil, err
	}

	var character model.Character
	if err := collection.FindOne(s.db.Ctx, bson.M{"_id": objectID}).Decode(&character); err != nil {
		return nil, err
	}

	// 将角色缓存到内存并分配channel
	s.cacheCharacter(&character)

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

// GetInventoryRequest 获取背包请求
type GetInventoryRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// InventoryResponse 背包响应
type InventoryResponse struct {
	Inventory model.Inventory `json:"inventory"`
}

// GetInventory 获取用户背包
func (s *GameService) GetInventory(userID string) (*InventoryResponse, error) {
	collection := s.db.GetCollection(s.dbName, "inventories")

	// 查找用户背包
	var inventory model.Inventory
	err := collection.FindOne(s.db.Ctx, bson.M{"user_id": userID}).Decode(&inventory)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// 创建新背包
			inventory = model.Inventory{
				ID:        primitive.NewObjectID(),
				UserID:    userID,
				Items:     []model.Item{},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			_, err = collection.InsertOne(s.db.Ctx, inventory)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return &InventoryResponse{
		Inventory: inventory,
	}, nil
}

// AddItemRequest 添加物品请求
type AddItemRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	ItemType string `json:"item_type" binding:"required"`
	ItemID   string `json:"item_id" binding:"required"`
	Quantity int64  `json:"quantity" binding:"required,min=1"`
}

// AddItem 添加物品到背包
func (s *GameService) AddItem(req AddItemRequest) error {
	collection := s.db.GetCollection(s.dbName, "inventories")

	// 查找用户背包
	var inventory model.Inventory
	err := collection.FindOne(s.db.Ctx, bson.M{"user_id": req.UserID}).Decode(&inventory)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// 创建新背包
			inventory = model.Inventory{
				ID:        primitive.NewObjectID(),
				UserID:    req.UserID,
				Items:     []model.Item{},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		} else {
			return err
		}
	}

	// 检查物品是否已存在
	itemExists := false
	for i, item := range inventory.Items {
		if item.ItemType == req.ItemType && item.ItemID == req.ItemID {
			// 物品已存在，增加数量
			inventory.Items[i].Quantity += req.Quantity
			itemExists = true
			break
		}
	}

	if !itemExists {
		// 添加新物品
		newItem := model.Item{
			ID:         primitive.NewObjectID().Hex(),
			ItemType:   req.ItemType,
			ItemID:     req.ItemID,
			Quantity:   req.Quantity,
			IsEquipped: false,
			CreatedAt:  time.Now(),
		}
		inventory.Items = append(inventory.Items, newItem)
	}

	// 更新背包
	inventory.UpdatedAt = time.Now()
	_, err = collection.UpdateOne(s.db.Ctx, bson.M{"user_id": req.UserID}, bson.M{"$set": inventory})
	if err != nil {
		// 如果更新失败，尝试插入
		_, err = collection.InsertOne(s.db.Ctx, inventory)
		if err != nil {
			return err
		}
	}

	return nil
}

// UseItemRequest 使用物品请求
type UseItemRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	ItemID   string `json:"item_id" binding:"required"`
	Quantity int64  `json:"quantity" binding:"required,min=1"`
}

// UseItem 使用物品
func (s *GameService) UseItem(req UseItemRequest) error {
	collection := s.db.GetCollection(s.dbName, "inventories")

	// 查找用户背包
	var inventory model.Inventory
	err := collection.FindOne(s.db.Ctx, bson.M{"user_id": req.UserID}).Decode(&inventory)
	if err != nil {
		return err
	}

	// 查找物品
	itemFound := false
	for i, item := range inventory.Items {
		if item.ID == req.ItemID {
			if item.Quantity < req.Quantity {
				return errors.New("insufficient item quantity")
			}

			// 减少物品数量
			inventory.Items[i].Quantity -= req.Quantity
			if inventory.Items[i].Quantity <= 0 {
				// 移除物品
				inventory.Items = append(inventory.Items[:i], inventory.Items[i+1:]...)
			}
			itemFound = true
			break
		}
	}

	if !itemFound {
		return errors.New("item not found")
	}

	// 更新背包
	inventory.UpdatedAt = time.Now()
	_, err = collection.UpdateOne(s.db.Ctx, bson.M{"user_id": req.UserID}, bson.M{"$set": inventory})
	return err
}

// RemoveItemRequest 删除物品请求
type RemoveItemRequest struct {
	UserID string `json:"user_id" binding:"required"`
	ItemID string `json:"item_id" binding:"required"`
}

// RemoveItem 从背包中删除物品
func (s *GameService) RemoveItem(req RemoveItemRequest) error {
	collection := s.db.GetCollection(s.dbName, "inventories")

	// 查找用户背包
	var inventory model.Inventory
	err := collection.FindOne(s.db.Ctx, bson.M{"user_id": req.UserID}).Decode(&inventory)
	if err != nil {
		return err
	}

	// 查找并删除物品
	itemFound := false
	for i, item := range inventory.Items {
		if item.ID == req.ItemID {
			// 移除物品
			inventory.Items = append(inventory.Items[:i], inventory.Items[i+1:]...)
			itemFound = true
			break
		}
	}

	if !itemFound {
		return errors.New("item not found")
	}

	// 更新背包
	inventory.UpdatedAt = time.Now()
	_, err = collection.UpdateOne(s.db.Ctx, bson.M{"user_id": req.UserID}, bson.M{"$set": inventory})
	return err
}

// EquipItemRequest 装备物品请求
type EquipItemRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	ItemID   string `json:"item_id" binding:"required"`
	Equipped bool   `json:"equipped" binding:"required"`
}

// EquipItem 装备或卸下物品
func (s *GameService) EquipItem(req EquipItemRequest) error {
	collection := s.db.GetCollection(s.dbName, "inventories")

	// 查找用户背包
	var inventory model.Inventory
	err := collection.FindOne(s.db.Ctx, bson.M{"user_id": req.UserID}).Decode(&inventory)
	if err != nil {
		return err
	}

	// 查找物品并更新装备状态
	itemFound := false
	for i, item := range inventory.Items {
		if item.ID == req.ItemID {
			inventory.Items[i].IsEquipped = req.Equipped
			itemFound = true
			break
		}
	}

	if !itemFound {
		return errors.New("item not found")
	}

	// 更新背包
	inventory.UpdatedAt = time.Now()
	_, err = collection.UpdateOne(s.db.Ctx, bson.M{"user_id": req.UserID}, bson.M{"$set": inventory})
	return err
}

// UseCharacterRequest 使用角色请求
type UseCharacterRequest struct {
	CharacterID string `json:"character_id" binding:"required"`
	UserID      string `json:"user_id" binding:"required"`
}

// UseCharacterResponse 使用角色响应
type UseCharacterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// UseCharacter 使用角色，触发其他服务的角色使用事件，并向客户端推送消息
func (s *GameService) UseCharacter(req UseCharacterRequest) (*UseCharacterResponse, error) {
	// 1. 验证角色是否存在
	character, err := s.GetCharacterByID(req.CharacterID)
	if err != nil {
		return nil, err
	}

	// 2. 验证角色是否属于该用户
	userID, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		return nil, err
	}

	if character.UserID != userID {
		return nil, errors.New("character does not belong to the user")
	}

	// 3. 更新角色状态为使用中
	err = s.UpdateCharacterStatus(req.CharacterID, 2) // 2 表示使用中
	if err != nil {
		return nil, err
	}

	// 4. 将角色缓存到内存并分配channel
	s.cacheCharacter(character)

	// 5. 向角色通道发送使用事件消息
	cc, exists := s.getCharacterFromCache(req.CharacterID)
	if exists {
		cc.Ch <- map[string]interface{}{
			"type":         "character_used",
			"character_id": req.CharacterID,
			"user_id":      req.UserID,
			"timestamp":    time.Now(),
		}
	}

	// 6. 触发其他服务的角色使用事件（这里可以通过消息队列或RPC调用其他服务）
	// 例如：向bill服务发送角色使用事件，向ssoauth服务发送角色使用事件等
	// 这里为了简化，只记录日志

	// 7. 构造响应
	response := &UseCharacterResponse{
		Success: true,
		Message: "Character used successfully",
	}

	// 8. 向客户端推送消息（这里可以通过WebSocket或其他推送机制）
	// 例如：通过gateway服务的WebSocket连接向客户端推送角色使用事件

	return response, nil
}
