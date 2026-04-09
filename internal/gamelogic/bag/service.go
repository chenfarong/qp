package bag

import (
	"errors"
	"time"

	"github.com/aoyo/qp/internal/gamelogic/actor"
	"github.com/aoyo/qp/pkg/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// InventoryService 背包服务（读写 characters 集合中与角色同行的 items）
type InventoryService struct {
	db         *db.DB
	dbName     string
	characters *actor.CharacterService
}

var _ interface{} = (*InventoryService)(nil)

// NewInventoryService 创建背包服务实例；characters 用于背包变更后同步在线角色内存
func NewInventoryService(db *db.DB, dbName string, characters *actor.CharacterService) *InventoryService {
	return &InventoryService{
		db:         db,
		dbName:     dbName,
		characters: characters,
	}
}

// InventoryResponse 背包响应
type InventoryResponse struct {
	Inventory InventoryPayload `json:"inventory"`
}

func (s *InventoryService) characterCollection() *mongo.Collection {
	return s.db.GetCollection(s.dbName, actor.Character{}.CollectionName())
}

func (s *InventoryService) loadCharacterForInventory(characterID string) (*actor.Character, error) {
	objectID, err := primitive.ObjectIDFromHex(characterID)
	if err != nil {
		return nil, err
	}
	var char actor.Character
	if err := s.characterCollection().FindOne(s.db.Ctx, bson.M{"_id": objectID}).Decode(&char); err != nil {
		return nil, err
	}
	if char.Items == nil {
		char.Items = []actor.InventoryItem{}
	}
	return &char, nil
}

func (s *InventoryService) saveInventoryItems(characterID string, items []actor.InventoryItem) error {
	objectID, err := primitive.ObjectIDFromHex(characterID)
	if err != nil {
		return err
	}
	now := time.Now()
	_, err = s.characterCollection().UpdateOne(s.db.Ctx, bson.M{"_id": objectID}, bson.M{"$set": bson.M{
		"items":      items,
		"updated_at": now,
	}})
	if err != nil {
		return err
	}
	if s.characters != nil {
		s.characters.SyncCachedCharacter(characterID)
	}
	return nil
}

// GetInventory 按角色 ID 获取背包（角色须已存在）
func (s *InventoryService) GetInventory(characterID string) (*InventoryResponse, error) {
	char, err := s.loadCharacterForInventory(characterID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("character not found")
		}
		return nil, err
	}
	return &InventoryResponse{
		Inventory: InventoryPayload{
			CharacterID: characterID,
			Items:       char.Items,
		},
	}, nil
}

// AddItemRequest 添加物品请求
type AddItemRequest struct {
	CharacterID string `json:"character_id" binding:"required"`
	ItemType    string `json:"item_type" binding:"required"`
	ItemID      string `json:"item_id" binding:"required"`
	Quantity    int64  `json:"quantity" binding:"required,min=1"`
}

// AddItem 添加物品到背包
func (s *InventoryService) AddItem(req AddItemRequest) error {
	char, err := s.loadCharacterForInventory(req.CharacterID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("character not found")
		}
		return err
	}
	items := char.Items

	itemExists := false
	for i, item := range items {
		if item.ItemType == req.ItemType && item.ItemID == req.ItemID {
			items[i].Quantity += req.Quantity
			itemExists = true
			break
		}
	}
	if !itemExists {
		items = append(items, actor.InventoryItem{
			ID:         primitive.NewObjectID().Hex(),
			ItemType:   req.ItemType,
			ItemID:     req.ItemID,
			Quantity:   req.Quantity,
			IsEquipped: false,
			CreatedAt:  time.Now(),
		})
	}
	return s.saveInventoryItems(req.CharacterID, items)
}

// UseItemRequest 使用物品请求
type UseItemRequest struct {
	CharacterID string `json:"character_id" binding:"required"`
	ItemID      string `json:"item_id" binding:"required"`
	Quantity    int64  `json:"quantity" binding:"required,min=1"`
}

// UseItem 使用物品
func (s *InventoryService) UseItem(req UseItemRequest) error {
	char, err := s.loadCharacterForInventory(req.CharacterID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("character not found")
		}
		return err
	}
	items := char.Items

	itemFound := false
	for i, item := range items {
		if item.ID == req.ItemID {
			if item.Quantity < req.Quantity {
				return errors.New("insufficient item quantity")
			}
			items[i].Quantity -= req.Quantity
			if items[i].Quantity <= 0 {
				items = append(items[:i], items[i+1:]...)
			}
			itemFound = true
			break
		}
	}
	if !itemFound {
		return errors.New("item not found")
	}
	return s.saveInventoryItems(req.CharacterID, items)
}

// RemoveItemRequest 删除物品请求
type RemoveItemRequest struct {
	CharacterID string `json:"character_id" binding:"required"`
	ItemID      string `json:"item_id" binding:"required"`
}

// RemoveItem 从背包中删除物品
func (s *InventoryService) RemoveItem(req RemoveItemRequest) error {
	char, err := s.loadCharacterForInventory(req.CharacterID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("character not found")
		}
		return err
	}
	items := char.Items

	itemFound := false
	for i, item := range items {
		if item.ID == req.ItemID {
			items = append(items[:i], items[i+1:]...)
			itemFound = true
			break
		}
	}
	if !itemFound {
		return errors.New("item not found")
	}
	return s.saveInventoryItems(req.CharacterID, items)
}

// EquipItemRequest 装备物品请求
type EquipItemRequest struct {
	CharacterID string `json:"character_id" binding:"required"`
	ItemID      string `json:"item_id" binding:"required"`
	Equipped    bool   `json:"equipped" binding:"required"`
}

// EquipItem 装备或卸下物品
func (s *InventoryService) EquipItem(req EquipItemRequest) error {
	char, err := s.loadCharacterForInventory(req.CharacterID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("character not found")
		}
		return err
	}
	items := char.Items

	itemFound := false
	for i, item := range items {
		if item.ID == req.ItemID {
			items[i].IsEquipped = req.Equipped
			itemFound = true
			break
		}
	}
	if !itemFound {
		return errors.New("item not found")
	}
	return s.saveInventoryItems(req.CharacterID, items)
}

// CreateCharacter 创建角色
func (s *InventoryService) CreateCharacter(req actor.CreateCharacterRequest) (*actor.CharacterResponse, error) {
	return nil, errors.New("inventory service does not handle character creation")
}

// UseCharacter 使用角色
func (s *InventoryService) UseCharacter(req actor.UseCharacterRequest) (*actor.UseCharacterResponse, error) {
	return nil, errors.New("inventory service does not handle character usage")
}

// CharacterLogin 角色登录
func (s *InventoryService) CharacterLogin(characterID string) error {
	return errors.New("inventory service does not handle character login")
}

// CharacterLogout 角色登出
func (s *InventoryService) CharacterLogout(characterID string) error {
	return errors.New("inventory service does not handle character logout")
}

// CharacterOnline 角色上线
func (s *InventoryService) CharacterOnline(characterID string) error {
	return errors.New("inventory service does not handle character online")
}

// CharacterOffline 角色下线
func (s *InventoryService) CharacterOffline(characterID string) error {
	return errors.New("inventory service does not handle character offline")
}

// HandleInternalMessage 处理内部消息；未识别类型时返回 handled=false
func (s *InventoryService) HandleInternalMessage(messageType string, messageData []byte) (bool, error) {
	_ = messageType
	_ = messageData
	return false, nil
}
