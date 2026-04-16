package client

import (
	"context"
	"fmt"
	"log"
	"time"

	gateway "zagame/inside/gamelogic/grpc/gateway"
	"zagame/proto"

	grpcclient "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client gRPC客户端
type Client struct {
	conn   *grpcclient.ClientConn
	client gateway.GatewayServiceClient
}

// NewClient 创建gRPC客户端
func NewClient(address string) (*Client, error) {
	// 连接到gateway服务
	conn, err := grpcclient.NewClient(address, grpcclient.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	// 创建客户端
	client := gateway.NewGatewayServiceClient(conn)

	return &Client{
		conn:   conn,
		client: client,
	}, nil
}

// Close 关闭连接
func (c *Client) Close() error {
	return c.conn.Close()
}

// RegisterServer 注册服务器
func (c *Client) RegisterServer(serverID, serverName, address string, port int32) error {
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
		return err
	}

	// 检查注册是否成功
	if !resp.Success {
		return fmt.Errorf("注册失败: %s", resp.Message)
	}

	// 打印注册成功日志
	log.Printf("服务器注册成功: serverID=%s, serverName=%s, startMsgID=%d, endMsgID=%d, address=%s, port=%d\n",
		serverID, serverName, startMsgID, endMsgID, address, port)

	return nil
}
