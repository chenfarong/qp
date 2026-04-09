package etcd

import (
	"context"
	"log"
	"time"

	"go.etcd.io/etcd/client/v3"
)

// Client etcd 客户端管理

type Client struct {
	client *clientv3.Client
}

// NewClient 创建 etcd 客户端实例
func NewClient(endpoints []string) (*Client, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = cli.Status(ctx, endpoints[0])
	if err != nil {
		return nil, err
	}

	log.Println("etcd client connected successfully")
	return &Client{client: cli}, nil
}

// Put 存储键值对
func (c *Client) Put(ctx context.Context, key, value string) error {
	_, err := c.client.Put(ctx, key, value)
	return err
}

// Get 获取键值
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	resp, err := c.client.Get(ctx, key)
	if err != nil {
		return "", err
	}

	if len(resp.Kvs) == 0 {
		return "", nil
	}

	return string(resp.Kvs[0].Value), nil
}

// Delete 删除键
func (c *Client) Delete(ctx context.Context, key string) error {
	_, err := c.client.Delete(ctx, key)
	return err
}

// Close 关闭客户端连接
func (c *Client) Close() error {
	return c.client.Close()
}

// RegisterService 注册服务
func (c *Client) RegisterService(serviceName, serviceAddress string) error {
	key := "/services/" + serviceName
	value := serviceAddress

	// 设置租约，确保服务健康检查
	resp, err := c.client.Grant(context.Background(), 10)
	if err != nil {
		return err
	}

	// 注册服务
	_, err = c.client.Put(context.Background(), key, value, clientv3.WithLease(resp.ID))
	if err != nil {
		return err
	}

	// 续约
	ch, err := c.client.KeepAlive(context.Background(), resp.ID)
	if err != nil {
		return err
	}

	// 处理续约响应
	go func() {
		for range ch {
			// 续约成功，无需处理
		}
	}()

	log.Printf("Service %s registered at %s", serviceName, serviceAddress)
	return nil
}

// DiscoverService 发现服务
func (c *Client) DiscoverService(serviceName string) (string, error) {
	key := "/services/" + serviceName
	value, err := c.Get(context.Background(), key)
	if err != nil {
		return "", err
	}

	return value, nil
}