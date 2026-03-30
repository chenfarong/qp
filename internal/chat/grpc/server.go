package grpc

import (
	"context"

	"github.com/aoyo/qp/internal/chat/service"
	"github.com/aoyo/qp/pkg/proto/chat"
)

// ChatServer 聊天服务gRPC服务器
type ChatServer struct {
	chat.UnimplementedChatServiceServer
	chatService *service.ChatService
}

// NewChatServer 创建聊天服务gRPC服务器实例
func NewChatServer(chatService *service.ChatService) *ChatServer {
	return &ChatServer{
		chatService: chatService,
	}
}

// SendMessage 发送消息
func (s *ChatServer) SendMessage(ctx context.Context, req *chat.SendMessageRequest) (*chat.SendMessageResponse, error) {
	// 转换请求参数
	sendReq := service.SendMessageRequest{
		SenderID:   req.SenderId,
		ReceiverID: req.ReceiverId,
		Content:    req.Content,
		Type:       req.Type,
	}

	// 调用服务
	message, err := s.chatService.SendMessage(sendReq)
	if err != nil {
		return &chat.SendMessageResponse{
			Error: err.Error(),
		}, nil
	}

	// 转换响应
	chatMessage := &chat.Message{
		Id:        message.ID,
		SenderId:  message.SenderID,
		ReceiverId: message.ReceiverID,
		Content:   message.Content,
		Type:      message.Type,
		Status:    message.Status,
		CreatedAt: message.CreatedAt.Unix(),
		UpdatedAt: message.UpdatedAt.Unix(),
	}

	return &chat.SendMessageResponse{
		Message: chatMessage,
	}, nil
}

// GetMessages 获取消息历史
func (s *ChatServer) GetMessages(ctx context.Context, req *chat.GetMessagesRequest) (*chat.GetMessagesResponse, error) {
	// 调用服务
	messages, err := s.chatService.GetMessages(req.UserId, req.OtherUserId, int(req.Limit), int(req.Offset))
	if err != nil {
		return &chat.GetMessagesResponse{
			Error: err.Error(),
		}, nil
	}

	// 转换响应
	var chatMessages []*chat.Message
	for _, message := range messages {
		chatMessage := &chat.Message{
			Id:        message.ID,
			SenderId:  message.SenderID,
			ReceiverId: message.ReceiverID,
			Content:   message.Content,
			Type:      message.Type,
			Status:    message.Status,
			CreatedAt: message.CreatedAt.Unix(),
			UpdatedAt: message.UpdatedAt.Unix(),
		}
		chatMessages = append(chatMessages, chatMessage)
	}

	return &chat.GetMessagesResponse{
		Messages: chatMessages,
	}, nil
}

// GetConversations 获取会话列表
func (s *ChatServer) GetConversations(ctx context.Context, req *chat.GetConversationsRequest) (*chat.GetConversationsResponse, error) {
	// 调用服务
	conversations, err := s.chatService.GetConversations(req.UserId)
	if err != nil {
		return &chat.GetConversationsResponse{
			Error: err.Error(),
		}, nil
	}

	// 转换响应
	var chatConversations []*chat.Conversation
	for _, conversation := range conversations {
		chatConversation := &chat.Conversation{
			Id:              conversation.ID,
			UserIds:         conversation.UserIDs,
			LastMessage:     conversation.LastMessage,
			LastMessageTime: conversation.LastMessageTime.Unix(),
			UnreadCount:     int32(conversation.UnreadCount),
			CreatedAt:       conversation.CreatedAt.Unix(),
			UpdatedAt:       conversation.UpdatedAt.Unix(),
		}
		chatConversations = append(chatConversations, chatConversation)
	}

	return &chat.GetConversationsResponse{
		Conversations: chatConversations,
	}, nil
}

// UpdateMessageStatus 更新消息状态
func (s *ChatServer) UpdateMessageStatus(ctx context.Context, req *chat.UpdateMessageStatusRequest) (*chat.UpdateMessageStatusResponse, error) {
	// 调用服务
	err := s.chatService.UpdateMessageStatus(req.MessageId, req.Status)
	if err != nil {
		return &chat.UpdateMessageStatusResponse{
			Error: err.Error(),
		}, nil
	}

	return &chat.UpdateMessageStatusResponse{
		Message: "Message status updated successfully",
	}, nil
}
