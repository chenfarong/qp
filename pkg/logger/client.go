package logger

import (
	"fmt"
	"net"
	"time"
)

// Client 日志客户端
type Client struct {
	udpAddr   *net.UDPAddr
	conn      *net.UDPConn
	ServerURI string
}

// NewClient 创建日志客户端
func NewClient(serverAddr string, serverURI string) (*Client, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, err
	}

	return &Client{
		udpAddr:   udpAddr,
		conn:      conn,
		ServerURI: serverURI,
	}, nil
}

// Close 关闭连接
func (c *Client) Close() error {
	return c.conn.Close()
}

// SendLog 发送日志
func (c *Client) SendLog(level string, message string) error {
	logMsg := fmt.Sprintf("[%s] [%s] [%s] %s",
		c.ServerURI,
		level,
		time.Now().Format("2006-01-02 15:04:05"),
		message)

	_, err := c.conn.Write([]byte(logMsg))
	return err
}

// Warn 发送警告日志
func (c *Client) Warn(message string) error {
	return c.SendLog("WARN", message)
}

// Error 发送错误日志
func (c *Client) Error(message string) error {
	return c.SendLog("ERROR", message)
}

// Fatal 发送致命错误日志
func (c *Client) Fatal(message string) error {
	return c.SendLog("FATAL", message)
}
