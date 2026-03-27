package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User 用户模型
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time         `bson:"deleted_at,omitempty" json:"-"`
	Username  string             `bson:"username" json:"username"`
	Password  string             `bson:"password" json:"-"`
	Email     string             `bson:"email" json:"email"`
	Nickname  string             `bson:"nickname" json:"nickname"`
	Avatar    string             `bson:"avatar" json:"avatar"`
	Status    int                `bson:"status" json:"status"` // 1: 正常, 0: 禁用
}

// CollectionName 指定集合名
func (User) CollectionName() string {
	return "users"
}
