package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/aoyo/qp/internal/bill/model"
	"github.com/aoyo/qp/pkg/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PaymentService 支付处理服务
type PaymentService struct {
	db           *db.DB
	dbName       string
	tokenService *TokenService
}

// NewPaymentService 创建支付处理服务实例
func NewPaymentService(db *db.DB, dbName string, tokenService *TokenService) *PaymentService {
	return &PaymentService{
		db:           db,
		dbName:       dbName,
		tokenService: tokenService,
	}
}

// CreatePayment 创建支付订单
func (s *PaymentService) CreatePayment(payment model.Payment) (*model.Payment, error) {
	// 生成订单ID
	if payment.OrderID == "" {
		payment.OrderID = s.generateOrderID()
	}

	// 设置默认值
	payment.CreatedAt = time.Now()
	payment.UpdatedAt = time.Now()
	payment.Status = "pending"
	payment.ExpiredAt = time.Now().Add(30 * time.Minute) // 30分钟过期

	// 验证代币类型是否存在
	tokenType, err := s.tokenService.GetTokenType(payment.TokenType)
	if err != nil {
		return nil, errors.New("invalid token type")
	}

	// 验证支付金额和代币数量
	if payment.Amount <= 0 || payment.TokenAmount <= 0 {
		return nil, errors.New("invalid amount")
	}

	// 保存支付订单
	collection := s.db.GetCollection(s.dbName, "payments")
	result, err := collection.InsertOne(context.Background(), payment)
	if err != nil {
		return nil, err
	}

	// 设置ID
	payment.ID = result.InsertedID.(primitive.ObjectID)

	return &payment, nil
}

// GetPayment 获取支付订单
func (s *PaymentService) GetPayment(orderID string) (*model.Payment, error) {
	collection := s.db.GetCollection(s.dbName, "payments")

	var payment model.Payment
	err := collection.FindOne(context.Background(), bson.M{"order_id": orderID}).Decode(&payment)
	if err != nil {
		return nil, err
	}

	return &payment, nil
}

// HandlePaymentCallback 处理支付回调
func (s *PaymentService) HandlePaymentCallback(orderID, transactionID, status string) error {
	collection := s.db.GetCollection(s.dbName, "payments")

	// 查找支付订单
	var payment model.Payment
	err := collection.FindOne(context.Background(), bson.M{"order_id": orderID}).Decode(&payment)
	if err != nil {
		return err
	}

	// 检查订单状态
	if payment.Status != "pending" {
		return errors.New("order status is not pending")
	}

	// 更新支付状态
	payment.Status = status
	payment.TransactionID = transactionID
	payment.UpdatedAt = time.Now()

	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"order_id": orderID},
		bson.M{"$set": payment},
	)
	if err != nil {
		return err
	}

	// 如果支付成功，发放代币
	if status == "success" {
		err = s.tokenService.AddUserToken(payment.UserID, payment.TokenType, payment.TokenAmount)
		if err != nil {
			return err
		}
	}

	return nil
}

// QueryPaymentStatus 查询支付状态
func (s *PaymentService) QueryPaymentStatus(orderID string) (string, error) {
	payment, err := s.GetPayment(orderID)
	if err != nil {
		return "", err
	}

	// 检查是否过期
	if time.Now().After(payment.ExpiredAt) && payment.Status == "pending" {
		// 更新为过期状态
		collection := s.db.GetCollection(s.dbName, "payments")
		_, err = collection.UpdateOne(
			context.Background(),
			bson.M{"order_id": orderID},
			bson.M{"$set": bson.M{"status": "expired", "updated_at": time.Now()}},
		)
		if err != nil {
			return "", err
		}
		return "expired", nil
	}

	return payment.Status, nil
}

// GeneratePaymentURL 生成支付URL
func (s *PaymentService) GeneratePaymentURL(orderID, paymentMethod string) (string, error) {
	payment, err := s.GetPayment(orderID)
	if err != nil {
		return "", err
	}

	// 根据支付方式生成支付URL
	switch paymentMethod {
	case "wechat":
		// 生成微信支付URL
		return fmt.Sprintf("https://wx.tenpay.com/pay?order_id=%s&amount=%d", orderID, payment.Amount), nil
	case "alipay":
		// 生成支付宝支付URL
		return fmt.Sprintf("https://openapi.alipay.com/gateway.do?order_id=%s&amount=%d", orderID, payment.Amount), nil
	default:
		return "", errors.New("unsupported payment method")
	}
}

// 生成订单ID
func (s *PaymentService) generateOrderID() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("ORD%d%d", time.Now().Unix(), rand.Intn(10000))
}
