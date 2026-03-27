package db

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DB 数据库实例
type DB struct {
	Client *mongo.Client
	Ctx    context.Context
}

// InitDB 初始化MongoDB连接
func InitDB(uri string) (*DB, error) {
	ctx := context.Background()
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
		Ctx:    ctx,
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

