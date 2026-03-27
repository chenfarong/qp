package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserToken 用户代币余额模型
type UserToken struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CreatedAt time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time         `bson:"updated_at" json:"updated_at"`
	UserID    string            `bson:"user_id" json:"user_id"`           // 用户ID
	TokenType string            `bson:"token_type" json:"token_type"`     // 代币类型
	Balance   int64             `bson:"balance" json:"balance"`           // 代币余额
	Locked    int64             `bson:"locked" json:"locked"`             // 锁定的代币数量
}
