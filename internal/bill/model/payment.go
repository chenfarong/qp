package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Payment 支付交易模型
type Payment struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CreatedAt     time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time         `bson:"updated_at" json:"updated_at"`
	UserID        string            `bson:"user_id" json:"user_id"`               // 用户ID
	OrderID       string            `bson:"order_id" json:"order_id"`             // 订单ID
	Amount        int64             `bson:"amount" json:"amount"`                 // 支付金额（分）
	Currency      string            `bson:"currency" json:"currency"`             // 货币类型
	TokenType     string            `bson:"token_type" json:"token_type"`         // 代币类型
	TokenAmount   int64             `bson:"token_amount" json:"token_amount"`     // 代币数量
	PaymentMethod string            `bson:"payment_method" json:"payment_method"` // 支付方式
	Status        string            `bson:"status" json:"status"`                 // 支付状态
	TransactionID string            `bson:"transaction_id" json:"transaction_id"` // 交易ID
	CallbackData  string            `bson:"callback_data" json:"callback_data"`   // 回调数据
	ExpiredAt     time.Time         `bson:"expired_at" json:"expired_at"`         // 过期时间
}
