package actor

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/aoyo/qp/pkg/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// CharacterChannel 角色通道结构
type CharacterChannel struct {
	Character *Character
	Ch        chan interface{}
	LastUsed  time.Time
}

// CharacterService 角色服务
type CharacterService struct {
	db              *db.DB
	dbName          string
	characterCache  map[string]*CharacterChannel // 角色缓存，key为角色ID
	cacheMutex      sync.RWMutex                 // 缓存互斥锁
	cleanupInterval time.Duration                // 缓存清理间隔
}

// NewCharacterService 创建角色服务实例
func NewCharacterService(db *db.DB, dbName string) *CharacterService {
	service := &CharacterService{
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
func (s *CharacterService) cleanupCache() {
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
func (s *CharacterService) cacheCharacter(character *Character) {
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
func (s *CharacterService) handleCharacterMessages(characterID string, ch chan interface{}) {
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
func (s *CharacterService) getCharacterFromCache(characterID string) (*CharacterChannel, bool) {
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
	Character Character `json:"character"`
}

// CreateCharacter 创建角色
func (s *CharacterService) CreateCharacter(req CreateCharacterRequest) (*CharacterResponse, error) {
	collection := s.db.GetCollection(s.dbName, Character{}.CollectionName())

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

	var characters []Character
	if err := cursor.All(s.db.Ctx, &characters); err != nil {
		return nil, err
	}

	if len(characters) >= 3 {
		return nil, errors.New("user can only create up to 3 characters")
	}

	// 检查角色名是否已存在
	var existingCharacter Character
	if err := collection.FindOne(s.db.Ctx, bson.M{"name": req.Name}).Decode(&existingCharacter); err == nil {
		return nil, errors.New("character name already exists")
	} else if err != mongo.ErrNoDocuments {
		return nil, err
	}

	// 创建角色
	now := time.Now()
	character := Character{
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
func (s *CharacterService) GetCharactersByUserID(userID string) ([]Character, error) {
	collection := s.db.GetCollection(s.dbName, Character{}.CollectionName())

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	cursor, err := collection.Find(s.db.Ctx, bson.M{"user_id": objectID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(s.db.Ctx)

	var characters []Character
	if err := cursor.All(s.db.Ctx, &characters); err != nil {
		return nil, err
	}

	return characters, nil
}

// GetCharacterByID 根据ID获取角色信息
func (s *CharacterService) GetCharacterByID(characterID string) (*Character, error) {
	// 先从缓存中获取角色
	cc, exists := s.getCharacterFromCache(characterID)
	if exists {
		return cc.Character, nil
	}

	// 缓存中不存在，从数据库获取
	collection := s.db.GetCollection(s.dbName, Character{}.CollectionName())

	objectID, err := primitive.ObjectIDFromHex(characterID)
	if err != nil {
		return nil, err
	}

	var character Character
	if err := collection.FindOne(s.db.Ctx, bson.M{"_id": objectID}).Decode(&character); err != nil {
		return nil, err
	}

	// 将角色缓存到内存并分配channel
	s.cacheCharacter(&character)

	return &character, nil
}

// UpdateCharacterStatus 更新角色状态
func (s *CharacterService) UpdateCharacterStatus(characterID string, status int) error {
	collection := s.db.GetCollection(s.dbName, Character{}.CollectionName())

	objectID, err := primitive.ObjectIDFromHex(characterID)
	if err != nil {
		return err
	}

	_, err = collection.UpdateOne(s.db.Ctx, bson.M{"_id": objectID}, bson.M{"$set": bson.M{"status": status, "updated_at": time.Now()}})
	return err
}

// AddExp 添加经验值
func (s *CharacterService) AddExp(characterID string, exp int) error {
	collection := s.db.GetCollection(s.dbName, Character{}.CollectionName())

	objectID, err := primitive.ObjectIDFromHex(characterID)
	if err != nil {
		return err
	}

	var character Character
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
func (s *CharacterService) getRequiredExp(level int) int {
	return level * 100
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
func (s *CharacterService) UseCharacter(req UseCharacterRequest) (*UseCharacterResponse, error) {
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

// CharacterOffline 角色下线
func (s *CharacterService) CharacterOffline(characterID string) error {
	// 1. 验证角色是否存在
	_, err := s.GetCharacterByID(characterID)
	if err != nil {
		return err
	}

	// 2. 更新角色状态为离线
	err = s.UpdateCharacterStatus(characterID, 0) // 0 表示离线
	if err != nil {
		return err
	}

	// 3. 清理角色缓存
	s.cacheMutex.Lock()
	cc, exists := s.characterCache[characterID]
	if exists {
		close(cc.Ch)
		delete(s.characterCache, characterID)
		log.Printf("Character offline: %s", characterID)
	}
	s.cacheMutex.Unlock()

	// 4. 触发其他服务的角色下线事件（这里可以通过消息队列或RPC调用其他服务）
	// 例如：向bill服务发送角色下线事件，向ssoauth服务发送角色下线事件等
	// 这里为了简化，只记录日志

	return nil
}
