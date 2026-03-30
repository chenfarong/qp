package service

import (
	"time"

	"github.com/aoyo/qp/internal/chat/model"
	"github.com/aoyo/qp/pkg/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ChatService 聊天服务
type ChatService struct {
	db     *db.DB
	dbName string
}

// NewChatService 创建聊天服务实例
func NewChatService(db *db.DB, dbName string) *ChatService {
	return &ChatService{
		db:     db,
		dbName: dbName,
	}
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	SenderID   string `json:"sender_id" binding:"required"`
	ReceiverID string `json:"receiver_id" binding:"required"`
	Content    string `json:"content" binding:"required"`
	Type       string `json:"type" binding:"required"`
}

// MessageResponse 消息响应
type MessageResponse struct {
	Message model.Message `json:"message"`
}

// SendMessage 发送消息
func (s *ChatService) SendMessage(req SendMessageRequest) (*MessageResponse, error) {
	messageCollection := s.db.GetCollection(s.dbName, model.Message{}.CollectionName())
	conversationCollection := s.db.GetCollection(s.dbName, model.Conversation{}.CollectionName())

	// 创建消息
	now := time.Now()
	message := model.Message{
		ID:         primitive.NewObjectID(),
		SenderID:   req.SenderID,
		ReceiverID: req.ReceiverID,
		Content:    req.Content,
		Type:       req.Type,
		Status:     "sent",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// 保存消息
	_, err := messageCollection.InsertOne(s.db.Ctx, message)
	if err != nil {
		return nil, err
	}

	// 查找或创建会话
	userIDs := []string{req.SenderID, req.ReceiverID}
	var conversation model.Conversation
	err = conversationCollection.FindOne(s.db.Ctx, bson.M{"user_ids": bson.M{"$all": userIDs}}).Decode(&conversation)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			// 创建新会话
			conversation = model.Conversation{
				ID:              primitive.NewObjectID(),
				UserIDs:         userIDs,
				LastMessage:     req.Content,
				LastMessageTime: now,
				UnreadCount:     1,
				CreatedAt:       now,
				UpdatedAt:       now,
			}
			_, err = conversationCollection.InsertOne(s.db.Ctx, conversation)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		// 更新会话
		conversation.LastMessage = req.Content
		conversation.LastMessageTime = now
		conversation.UnreadCount++
		conversation.UpdatedAt = now
		_, err = conversationCollection.UpdateOne(s.db.Ctx, bson.M{"_id": conversation.ID}, bson.M{"$set": conversation})
		if err != nil {
			return nil, err
		}
	}

	return &MessageResponse{
		Message: message,
	}, nil
}

// GetMessagesRequest 获取消息请求
type GetMessagesRequest struct {
	UserID      string `json:"user_id" binding:"required"`
	OtherUserID string `json:"other_user_id" binding:"required"`
	Limit       int    `json:"limit" binding:"required,min=1,max=100"`
	Offset      int    `json:"offset" binding:"min=0"`
}

// GetMessages 获取与特定用户的消息历史
func (s *ChatService) GetMessages(req GetMessagesRequest) ([]model.Message, error) {
	collection := s.db.GetCollection(s.dbName, model.Message{}.CollectionName())

	// 构建查询条件：只获取两个用户之间的消息
	filter := bson.M{
		"$or": []bson.M{
			{"sender_id": req.UserID, "receiver_id": req.OtherUserID},
			{"sender_id": req.OtherUserID, "receiver_id": req.UserID},
		},
	}

	// 按时间倒序排序
	cursor, err := collection.Find(s.db.Ctx, filter, nil)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(s.db.Ctx)

	var messages []model.Message
	if err := cursor.All(s.db.Ctx, &messages); err != nil {
		return nil, err
	}

	// 反转消息顺序，使最早的消息在前
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	// 更新消息状态为已读
	for _, msg := range messages {
		if msg.ReceiverID == req.UserID && msg.Status != "read" {
			collection.UpdateOne(s.db.Ctx, bson.M{"_id": msg.ID}, bson.M{"$set": bson.M{"status": "read", "updated_at": time.Now()}})
		}
	}

	return messages, nil
}

// GetConversationsRequest 获取会话请求
type GetConversationsRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// GetConversations 获取用户的所有会话
func (s *ChatService) GetConversations(req GetConversationsRequest) ([]model.Conversation, error) {
	collection := s.db.GetCollection(s.dbName, model.Conversation{}.CollectionName())

	// 查找用户参与的所有会话
	cursor, err := collection.Find(s.db.Ctx, bson.M{"user_ids": req.UserID}, nil)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(s.db.Ctx)

	var conversations []model.Conversation
	if err := cursor.All(s.db.Ctx, &conversations); err != nil {
		return nil, err
	}

	return conversations, nil
}

// UpdateMessageStatusRequest 更新消息状态请求
type UpdateMessageStatusRequest struct {
	MessageID string `json:"message_id" binding:"required"`
	Status    string `json:"status" binding:"required"`
}

// UpdateMessageStatus 更新消息状态
func (s *ChatService) UpdateMessageStatus(req UpdateMessageStatusRequest) error {
	collection := s.db.GetCollection(s.dbName, model.Message{}.CollectionName())

	messageID, err := primitive.ObjectIDFromHex(req.MessageID)
	if err != nil {
		return err
	}

	_, err = collection.UpdateOne(s.db.Ctx, bson.M{"_id": messageID}, bson.M{"$set": bson.M{"status": req.Status, "updated_at": time.Now()}})
	return err
}
