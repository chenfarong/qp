package service

import (
	"context"
	"errors"
	"time"

	"github.com/aoyo/qp/internal/bill/model"
	"github.com/aoyo/qp/pkg/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TokenService 代币管理服务
type TokenService struct {
	db     *db.DB
	dbName string
}

// NewTokenService 创建代币管理服务实例
func NewTokenService(db *db.DB, dbName string) *TokenService {
	return &TokenService{
		db:     db,
		dbName: dbName,
	}
}

// InitTokenType 初始化代币类型
func (s *TokenService) InitTokenType(tokenType model.TokenType) error {
	collection := s.db.GetCollection(s.dbName, "token_types")

	// 检查是否已存在
	var existingType model.TokenType
	err := collection.FindOne(context.Background(), bson.M{"symbol": tokenType.Symbol}).Decode(&existingType)
	if err == nil {
		return errors.New("token type already exists")
	}
	if err != mongo.ErrNoDocuments {
		return err
	}

	// 设置创建时间和更新时间
	tokenType.CreatedAt = time.Now()
	tokenType.UpdatedAt = time.Now()
	tokenType.IsActive = true

	_, err = collection.InsertOne(context.Background(), tokenType)
	return err
}

// GetTokenType 获取代币类型
func (s *TokenService) GetTokenType(symbol string) (*model.TokenType, error) {
	collection := s.db.GetCollection(s.dbName, "token_types")

	var tokenType model.TokenType
	err := collection.FindOne(context.Background(), bson.M{"symbol": symbol, "is_active": true}).Decode(&tokenType)
	if err != nil {
		return nil, err
	}

	return &tokenType, nil
}

// GetUserToken 获取用户代币余额
func (s *TokenService) GetUserToken(userID, tokenType string) (*model.UserToken, error) {
	collection := s.db.GetCollection(s.dbName, "user_tokens")

	var userToken model.UserToken
	err := collection.FindOne(context.Background(), bson.M{"user_id": userID, "token_type": tokenType}).Decode(&userToken)
	if err == mongo.ErrNoDocuments {
		// 如果不存在，返回一个余额为0的实例
		return &model.UserToken{
			UserID:    userID,
			TokenType: tokenType,
			Balance:   0,
			Locked:    0,
		}, nil
	}
	if err != nil {
		return nil, err
	}

	return &userToken, nil
}

// AddUserToken 增加用户代币
func (s *TokenService) AddUserToken(userID, tokenType string, amount int64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	collection := s.db.GetCollection(s.dbName, "user_tokens")

	// 尝试更新现有记录
	_, err := collection.UpdateOne(
		context.Background(),
		bson.M{"user_id": userID, "token_type": tokenType},
		bson.M{
			"$inc": bson.M{"balance": amount},
			"$set": bson.M{"updated_at": time.Now()},
			"$setOnInsert": bson.M{
				"created_at": time.Now(),
			},
		},
		options.Update().SetUpsert(true),
	)

	if err != nil {
		return err
	}

	// 记录操作日志
	// TODO: 添加操作日志

	return nil
}

// RemoveUserToken 减少用户代币
func (s *TokenService) RemoveUserToken(userID, tokenType string, amount int64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	// 先获取用户代币余额
	userToken, err := s.GetUserToken(userID, tokenType)
	if err != nil {
		return err
	}

	// 检查余额是否足够
	if userToken.Balance < amount {
		return errors.New("insufficient balance")
	}

	collection := s.db.GetCollection(s.dbName, "user_tokens")

	// 更新余额
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"user_id": userID, "token_type": tokenType},
		bson.M{
			"$inc": bson.M{"balance": -amount},
			"$set": bson.M{"updated_at": time.Now()},
		},
	)

	if err != nil {
		return err
	}

	// 记录操作日志
	// TODO: 添加操作日志

	return nil
}

// LockUserToken 锁定用户代币
func (s *TokenService) LockUserToken(userID, tokenType string, amount int64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	// 先获取用户代币余额
	userToken, err := s.GetUserToken(userID, tokenType)
	if err != nil {
		return err
	}

	// 检查余额是否足够
	if userToken.Balance < amount {
		return errors.New("insufficient balance")
	}

	collection := s.db.GetCollection(s.dbName, "user_tokens")

	// 更新锁定金额
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"user_id": userID, "token_type": tokenType},
		bson.M{
			"$inc": bson.M{
				"balance": -amount,
				"locked":  amount,
			},
			"$set": bson.M{"updated_at": time.Now()},
		},
	)

	if err != nil {
		return err
	}

	return nil
}

// UnlockUserToken 解锁用户代币
func (s *TokenService) UnlockUserToken(userID, tokenType string, amount int64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	// 先获取用户代币余额
	userToken, err := s.GetUserToken(userID, tokenType)
	if err != nil {
		return err
	}

	// 检查锁定金额是否足够
	if userToken.Locked < amount {
		return errors.New("insufficient locked amount")
	}

	collection := s.db.GetCollection(s.dbName, "user_tokens")

	// 更新锁定金额
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"user_id": userID, "token_type": tokenType},
		bson.M{
			"$inc": bson.M{
				"balance": amount,
				"locked":  -amount,
			},
			"$set": bson.M{"updated_at": time.Now()},
		},
	)

	if err != nil {
		return err
	}

	return nil
}
