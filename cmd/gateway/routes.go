package main

import (
	"context"
	"fmt"

	"github.com/aoyo/qp/pkg/proto/game"
	"github.com/aoyo/qp/pkg/proto/proto"
	"github.com/aoyo/qp/pkg/router"
)

// 全局gamelogic gRPC客户端
var gamelogicClient game.GameServiceClient

// createErrorResponse 创建错误响应
func createErrorResponse(code int32, message string) *proto.Message {
	return &proto.Message{
		Type: proto.MessageType_MSG_TYPE_RESPONSE,
		Data: &proto.Message_Response{
			Response: &proto.Response{
				Code:    code,
				Message: message,
			},
		},
	}
}

// setupRouter 设置路由
func setupRouter() *router.Router {
	r := router.NewRouter()

	// 注册Auth协议处理
	r.Register(proto.MessageType_MSG_TYPE_AUTH_REGISTER, handleAuthRegister)
	r.Register(proto.MessageType_MSG_TYPE_AUTH_LOGIN, handleAuthLogin)
	r.Register(proto.MessageType_MSG_TYPE_AUTH_VALIDATE, handleAuthValidate)

	// 注册Game协议处理
	r.Register(proto.MessageType_MSG_TYPE_GAME_CREATE_CHARACTER, handleGameCreateCharacter)
	r.Register(proto.MessageType_MSG_TYPE_GAME_GET_CHARACTERS, handleGameGetCharacters)
	r.Register(proto.MessageType_MSG_TYPE_GAME_GET_CHARACTER, handleGameGetCharacter)
	r.Register(proto.MessageType_MSG_TYPE_GAME_UPDATE_CHARACTER_STATUS, handleGameUpdateCharacterStatus)
	r.Register(proto.MessageType_MSG_TYPE_GAME_BATTLE, handleGameBattle)
	r.Register(proto.MessageType_MSG_TYPE_GAME_GET_BAG, handleGameGetBag)

	// 注册Bill协议处理
	r.Register(proto.MessageType_MSG_TYPE_BILL_CREATE_PAYMENT, handleBillCreatePayment)
	r.Register(proto.MessageType_MSG_TYPE_BILL_GET_PAYMENT, handleBillGetPayment)
	r.Register(proto.MessageType_MSG_TYPE_BILL_UPDATE_PAYMENT_STATUS, handleBillUpdatePaymentStatus)
	r.Register(proto.MessageType_MSG_TYPE_BILL_GET_USER_PAYMENTS, handleBillGetUserPayments)

	// 注册Chat协议处理
	r.Register(proto.MessageType_MSG_TYPE_CHAT_SEND_MESSAGE, handleChatSendMessage)
	r.Register(proto.MessageType_MSG_TYPE_CHAT_GET_MESSAGES, handleChatGetMessages)
	r.Register(proto.MessageType_MSG_TYPE_CHAT_GET_CONVERSATIONS, handleChatGetConversations)
	r.Register(proto.MessageType_MSG_TYPE_CHAT_UPDATE_MESSAGE_STATUS, handleChatUpdateMessageStatus)

	return r
}

// 处理函数实现
func handleAuthRegister(msg *proto.Message) *proto.Message {
	// 实现认证注册处理
	if req := msg.GetAuthRegister(); req != nil {
		// 这里应该调用ssoauth服务
		return &proto.Message{
			Type: proto.MessageType_MSG_TYPE_RESPONSE,
			Data: &proto.Message_Response{
				Response: &proto.Response{
					Code:    200,
					Message: "Auth register handled",
					Data: &proto.Response_AuthResponse{
						AuthResponse: &proto.AuthResponse{
							Token:    "test-token",
							UserId:   "1",
							Username: req.Username,
							Nickname: req.Nickname,
						},
					},
				},
			},
		}
	}
	return createErrorResponse(400, "Invalid auth register request")
}

func handleAuthLogin(msg *proto.Message) *proto.Message {
	// 实现认证登录处理
	if req := msg.GetAuthLogin(); req != nil {
		// 这里应该调用ssoauth服务
		return &proto.Message{
			Type: proto.MessageType_MSG_TYPE_RESPONSE,
			Data: &proto.Message_Response{
				Response: &proto.Response{
					Code:    200,
					Message: "Auth login handled",
					Data: &proto.Response_AuthResponse{
						AuthResponse: &proto.AuthResponse{
							Token:    "test-token",
							UserId:   "1",
							Username: req.Username,
							Nickname: req.Username,
						},
					},
				},
			},
		}
	}
	return createErrorResponse(400, "Invalid auth login request")
}

func handleAuthValidate(msg *proto.Message) *proto.Message {
	// 实现认证验证处理
	if req := msg.GetAuthValidate(); req != nil {
		// 这里应该调用ssoauth服务
		return &proto.Message{
			Type: proto.MessageType_MSG_TYPE_RESPONSE,
			Data: &proto.Message_Response{
				Response: &proto.Response{
					Code:    200,
					Message: "Auth validate handled",
					Data: &proto.Response_AuthResponse{
						AuthResponse: &proto.AuthResponse{
							Token:    req.Token,
							UserId:   "1",
							Username: "test",
							Nickname: "Test User",
						},
					},
				},
			},
		}
	}
	return createErrorResponse(400, "Invalid auth validate request")
}

func handleGameCreateCharacter(msg *proto.Message) *proto.Message {
	// 实现游戏创建角色处理
	if req := msg.GetGameCreateCharacter(); req != nil {
		// 检查gamelogic客户端是否初始化
		if gamelogicClient == nil {
			return createErrorResponse(500, "Gamelogic service not available")
		}

		// 构建gRPC请求
		grpcReq := &game.CreateCharacterRequest{
			UserId: req.UserId,
			Name:   req.Name,
		}

		// 调用gamelogic gRPC服务
		grpcResp, err := gamelogicClient.CreateCharacter(context.Background(), grpcReq)
		if err != nil {
			return createErrorResponse(500, fmt.Sprintf("Failed to call gamelogic service: %v", err))
		}

		// 检查是否有错误
		if grpcResp.Error != "" {
			return createErrorResponse(400, grpcResp.Error)
		}

		// 构建响应
		return &proto.Message{
			Type: proto.MessageType_MSG_TYPE_RESPONSE,
			Data: &proto.Message_Response{
				Response: &proto.Response{
					Code:    200,
					Message: "Game create character handled",
					Data: &proto.Response_GameResponse{
						GameResponse: &proto.GameResponse{
							Data: &proto.GameResponse_Character{
								Character: &proto.Character{
									Id:           grpcResp.Character.Id,
									UserId:       grpcResp.Character.UserId,
									Name:         grpcResp.Character.Name,
									Level:        grpcResp.Character.Level,
									Exp:          grpcResp.Character.Exp,
									Hp:           grpcResp.Character.Hp,
									Mp:           grpcResp.Character.Mp,
									Strength:     grpcResp.Character.Strength,
									Agility:      grpcResp.Character.Agility,
									Intelligence: grpcResp.Character.Intelligence,
									Gold:         grpcResp.Character.Gold,
									Status:       grpcResp.Character.Status,
								},
							},
						},
					},
				},
			},
		}
	}
	return createErrorResponse(400, "Invalid game create character request")
}

func handleGameGetCharacters(msg *proto.Message) *proto.Message {
	// 实现游戏获取角色列表处理
	if req := msg.GetGameGetCharacters(); req != nil {
		// 检查gamelogic客户端是否初始化
		if gamelogicClient == nil {
			return createErrorResponse(500, "Gamelogic service not available")
		}

		// 构建gRPC请求
		grpcReq := &game.GetCharactersRequest{
			UserId: req.UserId,
		}

		// 调用gamelogic gRPC服务
		grpcResp, err := gamelogicClient.GetCharacters(context.Background(), grpcReq)
		if err != nil {
			return createErrorResponse(500, fmt.Sprintf("Failed to call gamelogic service: %v", err))
		}

		// 检查是否有错误
		if grpcResp.Error != "" {
			return createErrorResponse(400, grpcResp.Error)
		}

		// 构建响应
		characters := make([]*proto.Character, len(grpcResp.Characters))
		for i, c := range grpcResp.Characters {
			characters[i] = &proto.Character{
				Id:           c.Id,
				UserId:       c.UserId,
				Name:         c.Name,
				Level:        c.Level,
				Exp:          c.Exp,
				Hp:           c.Hp,
				Mp:           c.Mp,
				Strength:     c.Strength,
				Agility:      c.Agility,
				Intelligence: c.Intelligence,
				Gold:         c.Gold,
				Status:       c.Status,
			}
		}

		return &proto.Message{
			Type: proto.MessageType_MSG_TYPE_RESPONSE,
			Data: &proto.Message_Response{
				Response: &proto.Response{
					Code:    200,
					Message: "Game get characters handled",
					Data: &proto.Response_GameResponse{
						GameResponse: &proto.GameResponse{
							Data: &proto.GameResponse_Characters{
								Characters: &proto.CharacterList{
									Items: characters,
								},
							},
						},
					},
				},
			},
		}
	}
	return createErrorResponse(400, "Invalid game get characters request")
}

func handleGameGetCharacter(msg *proto.Message) *proto.Message {
	// 实现游戏获取角色详情处理
	if req := msg.GetGameGetCharacter(); req != nil {
		// 检查gamelogic客户端是否初始化
		if gamelogicClient == nil {
			return createErrorResponse(500, "Gamelogic service not available")
		}

		// 构建gRPC请求
		grpcReq := &game.GetCharacterRequest{
			CharacterId: req.CharacterId,
		}

		// 调用gamelogic gRPC服务
		grpcResp, err := gamelogicClient.GetCharacter(context.Background(), grpcReq)
		if err != nil {
			return createErrorResponse(500, fmt.Sprintf("Failed to call gamelogic service: %v", err))
		}

		// 检查是否有错误
		if grpcResp.Error != "" {
			return createErrorResponse(400, grpcResp.Error)
		}

		// 构建响应
		return &proto.Message{
			Type: proto.MessageType_MSG_TYPE_RESPONSE,
			Data: &proto.Message_Response{
				Response: &proto.Response{
					Code:    200,
					Message: "Game get character handled",
					Data: &proto.Response_GameResponse{
						GameResponse: &proto.GameResponse{
							Data: &proto.GameResponse_Character{
								Character: &proto.Character{
									Id:           grpcResp.Character.Id,
									UserId:       grpcResp.Character.UserId,
									Name:         grpcResp.Character.Name,
									Level:        grpcResp.Character.Level,
									Exp:          grpcResp.Character.Exp,
									Hp:           grpcResp.Character.Hp,
									Mp:           grpcResp.Character.Mp,
									Strength:     grpcResp.Character.Strength,
									Agility:      grpcResp.Character.Agility,
									Intelligence: grpcResp.Character.Intelligence,
									Gold:         grpcResp.Character.Gold,
									Status:       grpcResp.Character.Status,
								},
							},
						},
					},
				},
			},
		}
	}
	return createErrorResponse(400, "Invalid game get character request")
}

func handleGameUpdateCharacterStatus(msg *proto.Message) *proto.Message {
	// 实现游戏更新角色状态处理
	if req := msg.GetGameUpdateCharacterStatus(); req != nil {
		// 检查gamelogic客户端是否初始化
		if gamelogicClient == nil {
			return createErrorResponse(500, "Gamelogic service not available")
		}

		// 构建gRPC请求
		grpcReq := &game.UpdateCharacterStatusRequest{
			CharacterId: req.CharacterId,
			Status:      req.Status,
		}

		// 调用gamelogic gRPC服务
		grpcResp, err := gamelogicClient.UpdateCharacterStatus(context.Background(), grpcReq)
		if err != nil {
			return createErrorResponse(500, fmt.Sprintf("Failed to call gamelogic service: %v", err))
		}

		// 检查是否有错误
		if grpcResp.Error != "" {
			return createErrorResponse(400, grpcResp.Error)
		}

		// 构建响应
		return &proto.Message{
			Type: proto.MessageType_MSG_TYPE_RESPONSE,
			Data: &proto.Message_Response{
				Response: &proto.Response{
					Code:    200,
					Message: "Game update character status handled",
				},
			},
		}
	}
	return createErrorResponse(400, "Invalid game update character status request")
}

func handleGameBattle(msg *proto.Message) *proto.Message {
	// 实现游戏战斗处理
	if req := msg.GetGameBattle(); req != nil {
		// 检查gamelogic客户端是否初始化
		if gamelogicClient == nil {
			return createErrorResponse(500, "Gamelogic service not available")
		}

		// 构建gRPC请求
		grpcReq := &game.BattleRequest{
			CharacterId: req.CharacterId,
			EnemyLevel:  1, // 临时值，实际应该从请求中获取
		}

		// 调用gamelogic gRPC服务
		grpcResp, err := gamelogicClient.Battle(context.Background(), grpcReq)
		if err != nil {
			return createErrorResponse(500, fmt.Sprintf("Failed to call gamelogic service: %v", err))
		}

		// 检查是否有错误
		if grpcResp.Error != "" {
			return createErrorResponse(400, grpcResp.Error)
		}

		// 构建响应
		return &proto.Message{
			Type: proto.MessageType_MSG_TYPE_RESPONSE,
			Data: &proto.Message_Response{
				Response: &proto.Response{
					Code:    200,
					Message: "Game battle handled",
					Data: &proto.Response_GameResponse{
						GameResponse: &proto.GameResponse{
							Data: &proto.GameResponse_BattleResult{
								BattleResult: &proto.BattleResult{
									Victory:    grpcResp.Victory,
									ExpGained:  grpcResp.ExpGained,
									GoldGained: grpcResp.GoldGained,
									Message:    grpcResp.Message,
								},
							},
						},
					},
				},
			},
		}
	}
	return createErrorResponse(400, "Invalid game battle request")
}

func handleGameGetBag(msg *proto.Message) *proto.Message {
	// 实现游戏获取背包处理
	if req := msg.GetGameGetBag(); req != nil {
		// 这里应该调用gamelogic服务
		return &proto.Message{
			Type: proto.MessageType_MSG_TYPE_RESPONSE,
			Data: &proto.Message_Response{
				Response: &proto.Response{
					Code:    200,
					Message: "Game get bag handled",
					Data: &proto.Response_GameResponse{
						GameResponse: &proto.GameResponse{
							Data: &proto.GameResponse_BagItems{
								BagItems: &proto.BagItemList{
									Items: []*proto.BagItemData{
										{
											ItemId:    1,
											ItemCfgId: 1001,
											Num:       1,
										},
										{
											ItemId:    2,
											ItemCfgId: 1002,
											Num:       5,
										},
									},
								},
							},
						},
					},
				},
			},
		}
	}
	return createErrorResponse(400, "Invalid game get bag request")
}

func handleBillCreatePayment(msg *proto.Message) *proto.Message {
	// 实现账单创建支付处理
	if req := msg.GetBillCreatePayment(); req != nil {
		// 这里应该调用bill服务
		return &proto.Message{
			Type: proto.MessageType_MSG_TYPE_RESPONSE,
			Data: &proto.Message_Response{
				Response: &proto.Response{
					Code:    200,
					Message: "Bill create payment handled",
					Data: &proto.Response_BillResponse{
						BillResponse: &proto.BillResponse{
							Data: &proto.BillResponse_Payment{
								Payment: &proto.Payment{
									Id:            1,
									UserId:        req.UserId,
									ProductId:     req.ProductId,
									Amount:        req.Amount,
									PaymentMethod: req.PaymentMethod,
									Status:        "pending",
									TransactionId: "test-transaction",
									CreatedAt:     1620000000,
									UpdatedAt:     1620000000,
								},
							},
						},
					},
				},
			},
		}
	}
	return createErrorResponse(400, "Invalid bill create payment request")
}

func handleBillGetPayment(msg *proto.Message) *proto.Message {
	// 实现账单获取支付处理
	if req := msg.GetBillGetPayment(); req != nil {
		// 这里应该调用bill服务
		return &proto.Message{
			Type: proto.MessageType_MSG_TYPE_RESPONSE,
			Data: &proto.Message_Response{
				Response: &proto.Response{
					Code:    200,
					Message: "Bill get payment handled",
					Data: &proto.Response_BillResponse{
						BillResponse: &proto.BillResponse{
							Data: &proto.BillResponse_Payment{
								Payment: &proto.Payment{
									Id:            req.PaymentId,
									UserId:        1,
									ProductId:     "prod_001",
									Amount:        "100",
									PaymentMethod: "alipay",
									Status:        "completed",
									TransactionId: "test-transaction",
									CreatedAt:     1620000000,
									UpdatedAt:     1620000000,
								},
							},
						},
					},
				},
			},
		}
	}
	return createErrorResponse(400, "Invalid bill get payment request")
}

func handleBillUpdatePaymentStatus(msg *proto.Message) *proto.Message {
	// 实现账单更新支付状态处理
	if req := msg.GetBillUpdatePaymentStatus(); req != nil {
		// 这里应该调用bill服务
		return &proto.Message{
			Type: proto.MessageType_MSG_TYPE_RESPONSE,
			Data: &proto.Message_Response{
				Response: &proto.Response{
					Code:    200,
					Message: "Bill update payment status handled",
				},
			},
		}
	}
	return createErrorResponse(400, "Invalid bill update payment status request")
}

func handleBillGetUserPayments(msg *proto.Message) *proto.Message {
	// 实现账单获取用户支付记录处理
	if req := msg.GetBillGetUserPayments(); req != nil {
		// 这里应该调用bill服务
		return &proto.Message{
			Type: proto.MessageType_MSG_TYPE_RESPONSE,
			Data: &proto.Message_Response{
				Response: &proto.Response{
					Code:    200,
					Message: "Bill get user payments handled",
					Data: &proto.Response_BillResponse{
						BillResponse: &proto.BillResponse{
							Data: &proto.BillResponse_Payments{
								Payments: &proto.PaymentList{
									Items: []*proto.Payment{
										{
											Id:            1,
											UserId:        req.UserId,
											ProductId:     "prod_001",
											Amount:        "100",
											PaymentMethod: "alipay",
											Status:        "completed",
											TransactionId: "test-transaction-1",
											CreatedAt:     1620000000,
											UpdatedAt:     1620000000,
										},
										{
											Id:            2,
											UserId:        req.UserId,
											ProductId:     "prod_002",
											Amount:        "200",
											PaymentMethod: "wechat",
											Status:        "completed",
											TransactionId: "test-transaction-2",
											CreatedAt:     1620000000,
											UpdatedAt:     1620000000,
										},
									},
								},
							},
						},
					},
				},
			},
		}
	}
	return createErrorResponse(400, "Invalid bill get user payments request")
}

func handleChatSendMessage(msg *proto.Message) *proto.Message {
	// 实现聊天发送消息处理
	if req := msg.GetChatSendMessage(); req != nil {
		// 这里应该调用chat服务
		return &proto.Message{
			Type: proto.MessageType_MSG_TYPE_RESPONSE,
			Data: &proto.Message_Response{
				Response: &proto.Response{
					Code:    200,
					Message: "Chat send message handled",
					Data: &proto.Response_ChatResponse{
						ChatResponse: &proto.ChatResponse{
							// 暂时不使用ChatMessage，因为我们的protobuf定义中可能有问题
						},
					},
				},
			},
		}
	}
	return createErrorResponse(400, "Invalid chat send message request")
}

func handleChatGetMessages(msg *proto.Message) *proto.Message {
	// 实现聊天获取消息历史处理
	if req := msg.GetChatGetMessages(); req != nil {
		// 这里应该调用chat服务
		return &proto.Message{
			Type: proto.MessageType_MSG_TYPE_RESPONSE,
			Data: &proto.Message_Response{
				Response: &proto.Response{
					Code:    200,
					Message: "Chat get messages handled",
					Data: &proto.Response_ChatResponse{
						ChatResponse: &proto.ChatResponse{
							// 暂时不使用ChatMessages，因为我们的protobuf定义中可能有问题
						},
					},
				},
			},
		}
	}
	return createErrorResponse(400, "Invalid chat get messages request")
}

func handleChatGetConversations(msg *proto.Message) *proto.Message {
	// 实现聊天获取会话列表处理
	if req := msg.GetChatGetConversations(); req != nil {
		// 这里应该调用chat服务
		return &proto.Message{
			Type: proto.MessageType_MSG_TYPE_RESPONSE,
			Data: &proto.Message_Response{
				Response: &proto.Response{
					Code:    200,
					Message: "Chat get conversations handled",
					Data: &proto.Response_ChatResponse{
						ChatResponse: &proto.ChatResponse{
							Data: &proto.ChatResponse_Conversations{
								Conversations: &proto.ConversationList{
									Items: []*proto.Conversation{
										{
											Id:              1,
											UserIds:         []uint32{req.UserId, 2},
											LastMessage:     "Hello!",
											LastMessageTime: 1620000000,
											UnreadCount:     0,
											CreatedAt:       1620000000,
											UpdatedAt:       1620000000,
										},
									},
								},
							},
						},
					},
				},
			},
		}
	}
	return createErrorResponse(400, "Invalid chat get conversations request")
}

func handleChatUpdateMessageStatus(msg *proto.Message) *proto.Message {
	// 实现聊天更新消息状态处理
	if req := msg.GetChatUpdateMessageStatus(); req != nil {
		// 这里应该调用chat服务
		return &proto.Message{
			Type: proto.MessageType_MSG_TYPE_RESPONSE,
			Data: &proto.Message_Response{
				Response: &proto.Response{
					Code:    200,
					Message: "Chat update message status handled",
				},
			},
		}
	}
	return createErrorResponse(400, "Invalid chat update message status request")
}
