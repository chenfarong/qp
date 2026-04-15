package protocol

import (
	"bytes"
	"context"
	"log"
	"net"
	"strings"

	"github.com/aoyo/qp/internal/gamelogic"
	"github.com/aoyo/qp/pkg/protocol"
)

// Server 协议服务器
type Server struct {
	app *gamelogic.App
}

// NewServer 创建协议服务器实例
func NewServer(app *gamelogic.App) *Server {
	return &Server{
		app: app,
	}
}

// Start 启动协议服务器
func (s *Server) Start(address string) error {
	// 监听TCP端口
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.Printf("Game Logic protocol server starting on %s...", address)

	// 处理连接
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		// 处理连接
		go s.handleConnection(conn)
	}
}

// handleConnection 处理连接
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	// 读取数据
	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Printf("Error reading from connection: %v", err)
		return
	}

	// 检查数据长度
	if n < 32 { // 32字节角色编号 + 协议数据
		log.Printf("Data too short: %d bytes", n)
		return
	}

	// 提取32字节角色编号
	characterID := string(buffer[:32])
	// 去除空白字符
	characterID = strings.TrimSpace(characterID)

	log.Printf("Received character ID: %s", characterID)

	// 解码协议数据包
	packet, err := protocol.Decode(bytes.NewReader(buffer[32:n]))
	if err != nil {
		log.Printf("Error decoding protocol packet: %v", err)
		return
	}

	log.Printf("Received packet: Type=%c, Compress=%c, ID=%d, Data length=%d",
		packet.MessageType, packet.CompressFlag, packet.MessageID, len(packet.Data))

	// 生成上下文
	ctx := context.WithValue(context.Background(), "characterID", characterID)

	// 处理消息
	s.handleMessageWithContext(ctx, packet)

	// 发送响应
	response := []byte("Message received")
	encodedResponse, err := protocol.Encode(
		protocol.MessageTypeResponse,
		protocol.CompressFlagNone,
		packet.MessageID,
		response,
	)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		return
	}

	_, err = conn.Write(encodedResponse)
	if err != nil {
		log.Printf("Error writing response: %v", err)
		return
	}
}

// handleMessage 处理消息
func (s *Server) handleMessage(packet *protocol.Packet) {
	// 默认上下文
	ctx := context.Background()
	s.handleMessageWithContext(ctx, packet)
}

// handleMessageWithContext 带上下文处理消息
func (s *Server) handleMessageWithContext(ctx context.Context, packet *protocol.Packet) {
	// 根据消息类型处理
	switch packet.MessageType {
	case protocol.MessageTypeRequest:
		s.handleRequestWithContext(ctx, packet)
	case protocol.MessageTypeNotify:
		s.handleNotifyWithContext(ctx, packet)
	default:
		log.Printf("Unknown message type: %c", packet.MessageType)
	}
}

// handleRequest 处理请求消息
func (s *Server) handleRequest(packet *protocol.Packet) {
	// 默认上下文
	ctx := context.Background()
	s.handleRequestWithContext(ctx, packet)
}

// handleRequestWithContext 带上下文处理请求消息
func (s *Server) handleRequestWithContext(ctx context.Context, packet *protocol.Packet) {
	// 获取角色编号
	characterID, ok := ctx.Value("characterID").(string)
	if !ok {
		characterID = ""
	}

	log.Printf("Handling request with ID: %d for character: %s", packet.MessageID, characterID)

	// 根据消息号处理不同的请求
	switch packet.MessageID {
	case 101: // MSG_TYPE_GAME_CREATE_CHARACTER
		// 处理创建角色请求
		s.handleCreateCharacter(ctx, packet)
	case 102: // MSG_TYPE_GAME_GET_CHARACTERS
		// 处理获取角色列表请求
		s.handleGetCharacters(ctx, packet)
	case 103: // MSG_TYPE_GAME_GET_CHARACTER
		// 处理获取角色详情请求
		s.handleGetCharacter(ctx, packet)
	case 104: // MSG_TYPE_GAME_UPDATE_CHARACTER_STATUS
		// 处理更新角色状态请求
		s.handleUpdateCharacterStatus(ctx, packet)
	case 105: // MSG_TYPE_GAME_BATTLE
		// 处理战斗请求
		s.handleBattle(ctx, packet)
	case 106: // MSG_TYPE_GAME_GET_BAG
		// 处理获取背包请求
		s.handleGetBag(ctx, packet)
	default:
		log.Printf("Unknown message ID: %d", packet.MessageID)
	}
}

// handleNotify 处理通知消息
func (s *Server) handleNotify(packet *protocol.Packet) {
	// 默认上下文
	ctx := context.Background()
	s.handleNotifyWithContext(ctx, packet)
}

// handleNotifyWithContext 带上下文处理通知消息
func (s *Server) handleNotifyWithContext(ctx context.Context, packet *protocol.Packet) {
	// 获取角色编号
	characterID, ok := ctx.Value("characterID").(string)
	if !ok {
		characterID = ""
	}

	log.Printf("Handling notify with ID: %d for character: %s", packet.MessageID, characterID)
	// 实际处理逻辑...
}

// handleCreateCharacter 处理创建角色请求
func (s *Server) handleCreateCharacter(ctx context.Context, packet *protocol.Packet) {
	// 获取角色编号
	characterID, _ := ctx.Value("characterID").(string)
	log.Printf("Handling create character request for character: %s", characterID)
	// 实际处理逻辑...
}

// handleGetCharacters 处理获取角色列表请求
func (s *Server) handleGetCharacters(ctx context.Context, packet *protocol.Packet) {
	// 获取角色编号
	characterID, _ := ctx.Value("characterID").(string)
	log.Printf("Handling get characters request for character: %s", characterID)
	// 实际处理逻辑...
}

// handleGetCharacter 处理获取角色详情请求
func (s *Server) handleGetCharacter(ctx context.Context, packet *protocol.Packet) {
	// 获取角色编号
	characterID, _ := ctx.Value("characterID").(string)
	log.Printf("Handling get character request for character: %s", characterID)
	// 实际处理逻辑...
}

// handleUpdateCharacterStatus 处理更新角色状态请求
func (s *Server) handleUpdateCharacterStatus(ctx context.Context, packet *protocol.Packet) {
	// 获取角色编号
	characterID, _ := ctx.Value("characterID").(string)
	log.Printf("Handling update character status request for character: %s", characterID)
	// 实际处理逻辑...
}

// handleBattle 处理战斗请求
func (s *Server) handleBattle(ctx context.Context, packet *protocol.Packet) {
	// 获取角色编号
	characterID, _ := ctx.Value("characterID").(string)
	log.Printf("Handling battle request for character: %s", characterID)
	// 实际处理逻辑...
}

// handleGetBag 处理获取背包请求
func (s *Server) handleGetBag(ctx context.Context, packet *protocol.Packet) {
	// 获取角色编号
	characterID, _ := ctx.Value("characterID").(string)
	log.Printf("Handling get bag request for character: %s", characterID)
	// 实际处理逻辑...
}
