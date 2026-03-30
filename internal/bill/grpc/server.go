package grpc

import (
	"context"

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
	createReq := service.CreatePaymentRequest{
		UserID:        req.UserId,
		ProductID:     req.ProductId,
		Amount:        req.Amount,
		PaymentMethod: req.PaymentMethod,
	}

	// 调用服务
	payment, err := s.paymentService.CreatePayment(createReq)
	if err != nil {
		return &bill.CreatePaymentResponse{
			Error: err.Error(),
		}, nil
	}

	// 转换响应
	billPayment := &bill.Payment{
		Id:            payment.ID,
		UserId:        payment.UserID,
		ProductId:     payment.ProductID,
		Amount:        payment.Amount,
		PaymentMethod: payment.PaymentMethod,
		Status:        payment.Status,
		TransactionId: payment.TransactionID,
		CreatedAt:     payment.CreatedAt.Unix(),
		UpdatedAt:     payment.UpdatedAt.Unix(),
	}

	return &bill.CreatePaymentResponse{
		Payment: billPayment,
	}, nil
}

// GetPayment 获取支付信息
func (s *BillServer) GetPayment(ctx context.Context, req *bill.GetPaymentRequest) (*bill.GetPaymentResponse, error) {
	// 调用服务
	payment, err := s.paymentService.GetPaymentByID(req.PaymentId)
	if err != nil {
		return &bill.GetPaymentResponse{
			Error: err.Error(),
		}, nil
	}

	// 转换响应
	billPayment := &bill.Payment{
		Id:            payment.ID,
		UserId:        payment.UserID,
		ProductId:     payment.ProductID,
		Amount:        payment.Amount,
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
	err := s.paymentService.UpdatePaymentStatus(req.PaymentId, req.Status)
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
	// 调用服务
	payments, total, err := s.paymentService.GetPaymentsByUserID(req.UserId, int(req.Page), int(req.PageSize))
	if err != nil {
		return &bill.GetUserPaymentsResponse{
			Error: err.Error(),
		}, nil
	}

	// 转换响应
	var billPayments []*bill.Payment
	for _, payment := range payments {
		billPayment := &bill.Payment{
			Id:            payment.ID,
			UserId:        payment.UserID,
			ProductId:     payment.ProductID,
			Amount:        payment.Amount,
			PaymentMethod: payment.PaymentMethod,
			Status:        payment.Status,
			TransactionId: payment.TransactionID,
			CreatedAt:     payment.CreatedAt.Unix(),
			UpdatedAt:     payment.UpdatedAt.Unix(),
		}
		billPayments = append(billPayments, billPayment)
	}

	return &bill.GetUserPaymentsResponse{
		Payments: billPayments,
		Total:    int32(total),
	}, nil
}
