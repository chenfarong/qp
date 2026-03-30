package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Message 聊天消息模型
type Message struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SenderID  string             `bson:"sender_id" json:"sender_id"`
	ReceiverID string             `bson:"receiver_id" json:"receiver_id"`
	Content   string             `bson:"content" json:"content"`
	Type      string             `bson:"type" json:"type"` // text, image, voice, etc.
	Status    string             `bson:"status" json:"status"` // sent, delivered, read
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// CollectionName 返回消息集合名称
func (Message) CollectionName() string {
	return "messages"
}

// Conversation 会话模型
type Conversation struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserIDs   []string           `bson:"user_ids" json:"user_ids"`
	LastMessage string             `bson:"last_message" json:"last_message"`
	LastMessageTime time.Time          `bson:"last_message_time" json:"last_message_time"`
	UnreadCount int               `bson:"unread_count" json:"unread_count"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

// CollectionName 返回会话集合名称
func (Conversation) CollectionName() string {
	return "conversations"
}
