package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
) // DB 数据库实例
type DB struct {
	Client *mongo.Client
	Ctx    context.Context
}

// InitDB 初始化MongoDB连接
func InitDB(uri string) (*DB, error) {
	// 设置连接超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	// 测试连接
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return &DB{
		Client: client,
		Ctx:    context.Background(),
	}, nil
}

// GetCollection 获取集合
func (db *DB) GetCollection(dbName, collectionName string) *mongo.Collection {
	return db.Client.Database(dbName).Collection(collectionName)
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	return db.Client.Disconnect(db.Ctx)
}
