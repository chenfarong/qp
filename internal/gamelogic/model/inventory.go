package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Inventory 背包模型
type Inventory struct {
	ID        primitive.ObjectID `bson:"_id" json:"id"`
	UserID    string             `bson:"user_id" json:"user_id"`
	Items     []Item             `bson:"items" json:"items"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// Item 物品模型
type Item struct {
	ID         string    `bson:"id" json:"id"`
	ItemType   string    `bson:"item_type" json:"item_type"`
	ItemID     string    `bson:"item_id" json:"item_id"`
	Quantity   int64     `bson:"quantity" json:"quantity"`
	IsEquipped bool      `bson:"is_equipped" json:"is_equipped"`
	CreatedAt  time.Time `bson:"created_at" json:"created_at"`
}
