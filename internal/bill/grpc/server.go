package grpc

import (
	"context"

	"github.com/aoyo/qp/internal/bill/model"
	"github.com/aoyo/qp/internal/bill/service"
	"github.com/aoyo/qp/pkg/proto/bill"
)

// BillServer 账单服务gRPC服务器
type BillServer struct {
	bill.UnimplementedBillServiceServer
	paymentService *service.PaymentService
}

// NewBillServer 创建账单服务gRPC服务器实例
func NewBillServer(paymentService *service.PaymentService) *BillServer {
	return &BillServer{
		paymentService: paymentService,
	}
}

// CreatePayment 创建支付
func (s *BillServer) CreatePayment(ctx context.Context, req *bill.CreatePaymentRequest) (*bill.CreatePaymentResponse, error) {
	// 转换请求参数
	payment := model.Payment{
		UserID:        string(req.UserId),
		Amount:        100,    // 临时值
		TokenAmount:   100,    // 临时值
		TokenType:     "gold", // 临时值
		PaymentMethod: req.PaymentMethod,
	}

	// 调用服务
	createdPayment, err := s.paymentService.CreatePayment(payment)
	if err != nil {
		return &bill.CreatePaymentResponse{
			Error: err.Error(),
		}, nil
	}

	// 转换响应
	billPayment := &bill.Payment{
		Id:            1,     // 临时值
		UserId:        1,     // 临时值
		ProductId:     "1",   // 临时值
		Amount:        "100", // 临时值
		PaymentMethod: createdPayment.PaymentMethod,
		Status:        createdPayment.Status,
		TransactionId: createdPayment.TransactionID,
		CreatedAt:     createdPayment.CreatedAt.Unix(),
		UpdatedAt:     createdPayment.UpdatedAt.Unix(),
	}

	return &bill.CreatePaymentResponse{
		Payment: billPayment,
	}, nil
}

// GetPayment 获取支付信息
func (s *BillServer) GetPayment(ctx context.Context, req *bill.GetPaymentRequest) (*bill.GetPaymentResponse, error) {
	// 调用服务
	payment, err := s.paymentService.GetPayment(string(req.PaymentId))
	if err != nil {
		return &bill.GetPaymentResponse{
			Error: err.Error(),
		}, nil
	}

	// 转换响应
	billPayment := &bill.Payment{
		Id:            1, // 临时值
		UserId:        1, // 临时值
		ProductId:     "1", // 临时值
		Amount:        "100", // 临时值
		PaymentMethod: payment.PaymentMethod,
		Status:        payment.Status,
		TransactionId: payment.TransactionID,
		CreatedAt:     payment.CreatedAt.Unix(),
		UpdatedAt:     payment.UpdatedAt.Unix(),
	}

	return &bill.GetPaymentResponse{
		Payment: billPayment,
	}, nil
}

// UpdatePaymentStatus 更新支付状态
func (s *BillServer) UpdatePaymentStatus(ctx context.Context, req *bill.UpdatePaymentStatusRequest) (*bill.UpdatePaymentStatusResponse, error) {
	// 调用服务
	err := s.paymentService.HandlePaymentCallback(string(req.PaymentId), "", req.Status)
	if err != nil {
		return &bill.UpdatePaymentStatusResponse{
			Error: err.Error(),
		}, nil
	}

	return &bill.UpdatePaymentStatusResponse{
		Message: "Payment status updated successfully",
	}, nil
}

// GetUserPayments 获取用户的所有支付记录
func (s *BillServer) GetUserPayments(ctx context.Context, req *bill.GetUserPaymentsRequest) (*bill.GetUserPaymentsResponse, error) {
	// 临时返回空列表
	var billPayments []*bill.Payment

	return &bill.GetUserPaymentsResponse{
		Payments: billPayments,
		Total:    0,
	}, nil
}
