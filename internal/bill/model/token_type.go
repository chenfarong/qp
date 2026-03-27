package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TokenType 代币类型模型
type TokenType struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
	DeletedAt   *time.Time         `bson:"deleted_at,omitempty" json:"-"`
	Name        string             `bson:"name" json:"name"`             // 代币名称
	Symbol      string             `bson:"symbol" json:"symbol"`         // 代币符号
	Type        string             `bson:"type" json:"type"`             // 代币类型（如：金币、钻石、经验值等）
	Description string             `bson:"description" json:"description"` // 代币描述
	IsActive    bool               `bson:"is_active" json:"is_active"`   // 是否激活
}
