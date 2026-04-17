package client

import (
	"context"
	"fmt"
	"time"

	"zagame/common/logger"
	gateway "zagame/inside/gamelogic/grpc/gateway"
	"zagame/proto"

	grpcclient "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client gRPC客户端
type Client struct {
	conn        *grpcclient.ClientConn
	client      gateway.GatewayServiceClient
	address     string
	serverID    string
	serverName  string
	serverAddr  string
	serverPort  int32
	isConnected bool
}

// NewClient 创建gRPC客户端
func NewClient(address string) (*Client, error) {
	client := &Client{
		address:     address,
		isConnected: false,
	}

	// 尝试连接
	err := client.connect()
	if err != nil {
		// 连接失败，启动自动重连
		go client.reconnect()
		return client, fmt.Errorf("初始连接失败，已启动自动重连: %v", err)
	}

	return client, nil
}

// connect 连接到gateway服务
func (c *Client) connect() error {
	// 连接到gateway服务
	conn, err := grpcclient.NewClient(c.address, grpcclient.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	// 创建客户端
	client := gateway.NewGatewayServiceClient(conn)

	c.conn = conn
	c.client = client
	c.isConnected = true

	logger.Infof("成功连接到gateway服务: %s", c.address)
	return nil
}

// reconnect 自动重连到gateway服务
func (c *Client) reconnect() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if c.isConnected {
			continue
		}

		logger.Warnf("尝试重新连接到gateway服务: %s", c.address)
		err := c.connect()
		if err != nil {
			logger.Errorf("重新连接失败: %v", err)
			continue
		}

		// 重新连接成功，重新注册服务器
		if c.serverID != "" {
			err = c.RegisterServer(c.serverID, c.serverName, c.serverAddr, c.serverPort)
			if err != nil {
				logger.Errorf("重新注册服务器失败: %v", err)
			}
		}

		// 连接成功，退出重连循环
		break
	}
}

// Close 关闭连接
func (c *Client) Close() error {
	c.isConnected = false
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// RegisterServer 注册服务器
func (c *Client) RegisterServer(serverID, serverName, address string, port int32) error {
	// 保存服务器信息，用于重连后重新注册
	c.serverID = serverID
	c.serverName = serverName
	c.serverAddr = address
	c.serverPort = port

	// 检查连接状态
	if !c.isConnected {
		return fmt.Errorf("未连接到gateway服务")
	}

	// 获取消息ID范围
	startMsgID, endMsgID := proto.GetMessageIDRange()

	// 创建注册请求
	req := &gateway.RegisterServerRequest{
		ServerInfo: &gateway.ServerInfo{
			ServerId:   serverID,
			ServerName: serverName,
			StartMsgId: startMsgID,
			EndMsgId:   endMsgID,
			Address:    address,
			Port:       port,
		},
	}

	// 设置上下文超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 发送注册请求
	resp, err := c.client.RegisterServer(ctx, req)
	if err != nil {
		c.isConnected = false
		// 启动自动重连
		go c.reconnect()
		return err
	}

	// 检查注册是否成功
	if !resp.Success {
		return fmt.Errorf("注册失败: %s", resp.Message)
	}

	// 打印注册成功日志
	logger.Infof("服务器注册成功: serverID=%s, serverName=%s, startMsgID=%d, endMsgID=%d, address=%s, port=%d",
		serverID, serverName, startMsgID, endMsgID, address, port)

	return nil
}
