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

// CharacterChannel 在线角色会话：每个角色独立 channel，由唯一 goroutine 消费。
type CharacterChannel struct {
	Character *Character
	Ch        chan interface{}
}

// CharacterService 角色服务
type CharacterService struct {
	db             *db.DB
	dbName         string
	characterCache map[string]*CharacterChannel // 仅在线角色；登出时删除
	cacheMutex     sync.RWMutex
}

// 确保 CharacterService 实现了 Service 接口
var _ interface{} = (*CharacterService)(nil)

// NewCharacterService 创建角色服务实例
func NewCharacterService(db *db.DB, dbName string) *CharacterService {
	return &CharacterService{
		db:             db,
		dbName:         dbName,
		characterCache: make(map[string]*CharacterChannel),
	}
}

// loadCharacterFromDB 从数据库加载角色（不写入在线缓存）
func (s *CharacterService) loadCharacterFromDB(characterID string) (*Character, error) {
	collection := s.db.GetCollection(s.dbName, Character{}.CollectionName())

	objectID, err := primitive.ObjectIDFromHex(characterID)
	if err != nil {
		return nil, err
	}

	var character Character
	if err := collection.FindOne(s.db.Ctx, bson.M{"_id": objectID}).Decode(&character); err != nil {
		return nil, err
	}
	return &character, nil
}

// removeCharacterFromCache 登出时关闭 channel、移除内存（对应 goroutine 随之退出）
func (s *CharacterService) removeCharacterFromCache(characterID string) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	cc, ok := s.characterCache[characterID]
	if !ok {
		return
	}
	close(cc.Ch)
	delete(s.characterCache, characterID)
}

// SyncCachedCharacter 若角色在线，用数据库最新数据刷新内存中的 Character
func (s *CharacterService) SyncCachedCharacter(characterID string) {
	char, err := s.loadCharacterFromDB(characterID)
	if err != nil {
		return
	}
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	if cc, ok := s.characterCache[characterID]; ok {
		cc.Character = char
	}
}

// cacheCharacter 将角色缓存到内存并分配channel
func (s *CharacterService) cacheCharacter(character *Character) {
	characterID := character.ID.Hex()

	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	// 已在线则只刷新内存中的角色数据，不新建 goroutine/channel
	if cc, exists := s.characterCache[characterID]; exists {
		cc.Character = character
		return
	}

	ch := make(chan interface{}, 100)
	cc := &CharacterChannel{
		Character: character,
		Ch:        ch,
	}

	s.characterCache[characterID] = cc

	// 每个在线角色仅一个 goroutine，消费该角色专属 channel
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
		Items:        []InventoryItem{},
	}

	if _, err := collection.InsertOne(s.db.Ctx, character); err != nil {
		return nil, err
	}

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

// GetCharacterByID 根据ID获取角色信息：在线则读内存，否则只读库（不会把非在线角色写入内存）
func (s *CharacterService) GetCharacterByID(characterID string) (*Character, error) {
	if cc, exists := s.getCharacterFromCache(characterID); exists {
		return cc.Character, nil
	}
	return s.loadCharacterFromDB(characterID)
}

// UpdateCharacterStatus 更新角色状态
func (s *CharacterService) UpdateCharacterStatus(characterID string, status int) error {
	collection := s.db.GetCollection(s.dbName, Character{}.CollectionName())

	objectID, err := primitive.ObjectIDFromHex(characterID)
	if err != nil {
		return err
	}

	_, err = collection.UpdateOne(s.db.Ctx, bson.M{"_id": objectID}, bson.M{"$set": bson.M{"status": status, "updated_at": time.Now()}})
	if err != nil {
		return err
	}
	s.SyncCachedCharacter(characterID)
	return nil
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
	if err != nil {
		return err
	}
	s.SyncCachedCharacter(characterID)
	return nil
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
	character, err := s.GetCharacterByID(req.CharacterID)
	if err != nil {
		return nil, err
	}

	userID, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		return nil, err
	}

	if character.UserID != userID {
		return nil, errors.New("character does not belong to the user")
	}

	err = s.UpdateCharacterStatus(req.CharacterID, 2) // 2 表示使用中
	if err != nil {
		return nil, err
	}

	character, err = s.loadCharacterFromDB(req.CharacterID)
	if err != nil {
		return nil, err
	}
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

// CharacterLogin 角色登录
func (s *CharacterService) CharacterLogin(characterID string) error {
	if _, err := s.loadCharacterFromDB(characterID); err != nil {
		return err
	}

	if err := s.UpdateCharacterStatus(characterID, 1); err != nil { // 1 表示登录
		return err
	}

	character, err := s.loadCharacterFromDB(characterID)
	if err != nil {
		return err
	}
	s.cacheCharacter(character)

	// 4. 向角色通道发送登录事件消息
	cc, exists := s.getCharacterFromCache(characterID)
	if exists {
		cc.Ch <- map[string]interface{}{
			"type":         "character_login",
			"character_id": characterID,
			"timestamp":    time.Now(),
		}
	}

	log.Printf("Character login: %s", characterID)
	return nil
}

// CharacterLogout 角色登出：更新库后从内存删除并关闭 channel（唯一清理在线会话的路径）
func (s *CharacterService) CharacterLogout(characterID string) error {
	if _, err := s.loadCharacterFromDB(characterID); err != nil {
		return err
	}

	if err := s.UpdateCharacterStatus(characterID, 0); err != nil { // 0 表示离线
		return err
	}

	s.removeCharacterFromCache(characterID)
	log.Printf("Character logout: %s", characterID)
	return nil
}

// CharacterOnline 角色上线
func (s *CharacterService) CharacterOnline(characterID string) error {
	if _, err := s.loadCharacterFromDB(characterID); err != nil {
		return err
	}

	if err := s.UpdateCharacterStatus(characterID, 2); err != nil { // 2 表示上线
		return err
	}

	character, err := s.loadCharacterFromDB(characterID)
	if err != nil {
		return err
	}
	s.cacheCharacter(character)

	// 4. 向角色通道发送上线事件消息
	cc, exists := s.getCharacterFromCache(characterID)
	if exists {
		cc.Ch <- map[string]interface{}{
			"type":         "character_online",
			"character_id": characterID,
			"timestamp":    time.Now(),
		}
	}

	log.Printf("Character online: %s", characterID)
	return nil
}

// CharacterOffline 角色下线：只更新状态，保留内存中的在线数据与 channel/goroutine；登出才清内存
func (s *CharacterService) CharacterOffline(characterID string) error {
	if _, err := s.loadCharacterFromDB(characterID); err != nil {
		return err
	}

	if err := s.UpdateCharacterStatus(characterID, 0); err != nil { // 0 表示离线
		return err
	}

	if cc, ok := s.getCharacterFromCache(characterID); ok {
		select {
		case cc.Ch <- map[string]interface{}{
			"type":         "character_offline",
			"character_id": characterID,
			"timestamp":    time.Now(),
		}:
		default:
		}
	}

	log.Printf("Character offline: %s", characterID)
	return nil
}

// HandleInternalMessage 处理内部消息；未识别类型时返回 handled=false
func (s *CharacterService) HandleInternalMessage(messageType string, messageData []byte) (bool, error) {
	_ = messageType
	_ = messageData
	return false, nil
}
