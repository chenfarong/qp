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
		SenderID:   string(req.SenderId),
		ReceiverID: string(req.ReceiverId),
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
		Id:         1, // 临时值
		SenderId:   1, // 临时值
		ReceiverId: 2, // 临时值
		Content:    message.Message.Content,
		Type:       message.Message.Type,
		Status:     message.Message.Status,
		CreatedAt:  message.Message.CreatedAt.Unix(),
		UpdatedAt:  message.Message.UpdatedAt.Unix(),
	}

	return &chat.SendMessageResponse{
		Message: chatMessage,
	}, nil
}

// GetMessages 获取消息历史
func (s *ChatServer) GetMessages(ctx context.Context, req *chat.GetMessagesRequest) (*chat.GetMessagesResponse, error) {
	// 转换请求参数
	getReq := service.GetMessagesRequest{
		UserID:      string(req.UserId),
		OtherUserID: string(req.OtherUserId),
		Limit:       int(req.Limit),
		Offset:      int(req.Offset),
	}

	// 调用服务
	messages, err := s.chatService.GetMessages(getReq)
	if err != nil {
		return &chat.GetMessagesResponse{
			Error: err.Error(),
		}, nil
	}

	// 转换响应
	var chatMessages []*chat.Message
	for i, message := range messages {
		chatMessage := &chat.Message{
			Id:         uint32(i + 1), // 临时值
			SenderId:   1,             // 临时值
			ReceiverId: 2,             // 临时值
			Content:    message.Content,
			Type:       message.Type,
			Status:     message.Status,
			CreatedAt:  message.CreatedAt.Unix(),
			UpdatedAt:  message.UpdatedAt.Unix(),
		}
		chatMessages = append(chatMessages, chatMessage)
	}

	return &chat.GetMessagesResponse{
		Messages: chatMessages,
	}, nil
}

// GetConversations 获取会话列表
func (s *ChatServer) GetConversations(ctx context.Context, req *chat.GetConversationsRequest) (*chat.GetConversationsResponse, error) {
	// 转换请求参数
	getReq := service.GetConversationsRequest{
		UserID: string(req.UserId),
	}

	// 调用服务
	conversations, err := s.chatService.GetConversations(getReq)
	if err != nil {
		return &chat.GetConversationsResponse{
			Error: err.Error(),
		}, nil
	}

	// 转换响应
	var chatConversations []*chat.Conversation
	for i, conversation := range conversations {
		chatConversation := &chat.Conversation{
			Id:              uint32(i + 1),  // 临时值
			UserIds:         []uint32{1, 2}, // 临时值
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
	// 转换请求参数
	updateReq := service.UpdateMessageStatusRequest{
		MessageID: string(req.MessageId),
		Status:    req.Status,
	}

	// 调用服务
	err := s.chatService.UpdateMessageStatus(updateReq)
	if err != nil {
		return &chat.UpdateMessageStatusResponse{
			Error: err.Error(),
		}, nil
	}

	return &chat.UpdateMessageStatusResponse{
		Message: "Message status updated successfully",
	}, nil
}
