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

// InventoryService 背包服务
type InventoryService struct {
	db     *db.DB
	dbName string
}

// NewInventoryService 创建背包服务实例
func NewInventoryService(db *db.DB, dbName string) *InventoryService {
	return &InventoryService{
		db:     db,
		dbName: dbName,
	}
}

// GetInventoryRequest 获取背包请求
type GetInventoryRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// InventoryResponse 背包响应
type InventoryResponse struct {
	Inventory Inventory `json:"inventory"`
}

// GetInventory 获取用户背包
func (s *InventoryService) GetInventory(userID string) (*InventoryResponse, error) {
	collection := s.db.GetCollection(s.dbName, "inventories")

	// 查找用户背包
	var inventory Inventory
	err := collection.FindOne(s.db.Ctx, bson.M{"user_id": userID}).Decode(&inventory)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// 创建新背包
			inventory = Inventory{
				ID:        primitive.NewObjectID(),
				UserID:    userID,
				Items:     []Item{},
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
func (s *InventoryService) AddItem(req AddItemRequest) error {
	collection := s.db.GetCollection(s.dbName, "inventories")

	// 查找用户背包
	var inventory Inventory
	err := collection.FindOne(s.db.Ctx, bson.M{"user_id": req.UserID}).Decode(&inventory)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// 创建新背包
			inventory = Inventory{
				ID:        primitive.NewObjectID(),
				UserID:    req.UserID,
				Items:     []Item{},
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
		newItem := Item{
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
func (s *InventoryService) UseItem(req UseItemRequest) error {
	collection := s.db.GetCollection(s.dbName, "inventories")

	// 查找用户背包
	var inventory Inventory
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
func (s *InventoryService) RemoveItem(req RemoveItemRequest) error {
	collection := s.db.GetCollection(s.dbName, "inventories")

	// 查找用户背包
	var inventory Inventory
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
func (s *InventoryService) EquipItem(req EquipItemRequest) error {
	collection := s.db.GetCollection(s.dbName, "inventories")

	// 查找用户背包
	var inventory Inventory
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

// CreateCharacter 创建角色
func (s *InventoryService) CreateCharacter(req actor.CreateCharacterRequest) (*actor.CharacterResponse, error) {
	// 由于InventoryService不直接处理角色创建，这里返回错误
	return nil, errors.New("inventory service does not handle character creation")
}

// UseCharacter 使用角色
func (s *InventoryService) UseCharacter(req actor.UseCharacterRequest) (*actor.UseCharacterResponse, error) {
	// 由于InventoryService不直接处理角色使用，这里返回错误
	return nil, errors.New("inventory service does not handle character usage")
}

// CharacterOffline 角色下线
func (s *InventoryService) CharacterOffline(characterID string) error {
	// 由于InventoryService不直接处理角色下线，这里返回错误
	return errors.New("inventory service does not handle character offline")
}
