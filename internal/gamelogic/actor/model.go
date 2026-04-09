package actor

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// InventoryItem 背包物品，与角色同一文档（characters 集合一行）存储
type InventoryItem struct {
	ID         string    `bson:"id" json:"id"`
	ItemType   string    `bson:"item_type" json:"item_type"`
	ItemID     string    `bson:"item_id" json:"item_id"`
	Quantity   int64     `bson:"quantity" json:"quantity"`
	IsEquipped bool      `bson:"is_equipped" json:"is_equipped"`
	CreatedAt  time.Time `bson:"created_at" json:"created_at"`
}

// Character 游戏角色模型
type Character struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
	DeletedAt    *time.Time         `bson:"deleted_at,omitempty" json:"-"`
	UserID       primitive.ObjectID `bson:"user_id" json:"user_id"`
	Name         string             `bson:"name" json:"name"`
	Level        int                `bson:"level" json:"level"`
	Exp          int                `bson:"exp" json:"exp"`
	HP           int                `bson:"hp" json:"hp"`
	MP           int                `bson:"mp" json:"mp"`
	Strength     int                `bson:"strength" json:"strength"`
	Agility      int                `bson:"agility" json:"agility"`
	Intelligence int                `bson:"intelligence" json:"intelligence"`
	Gold         int                `bson:"gold" json:"gold"`
	Status       int                `bson:"status" json:"status"` // 1: 正常, 0: 离线
	Items        []InventoryItem    `bson:"items,omitempty" json:"items,omitempty"`
}

// CollectionName 指定集合名
func (Character) CollectionName() string {
	return "characters"
}
