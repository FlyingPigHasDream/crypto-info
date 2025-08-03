package grpc

import (
	"context"

	cryptov1 "crypto-info/kitex_gen/crypto/v1"
	"crypto-info/internal/service"
	"crypto-info/internal/pkg/logger"
)

// CryptoPriceServiceImpl Kitex gRPC价格服务实现
type CryptoPriceServiceImpl struct {
	priceService service.PriceService
	logger       logger.Logger
}

// NewCryptoPriceService 创建价格服务实现
func NewCryptoPriceService(priceService service.PriceService) *CryptoPriceServiceImpl {
	return &CryptoPriceServiceImpl{
		priceService: priceService,
		logger:       logger.GetLogger(),
	}
}

// GetPrice 获取加密货币价格
func (s *CryptoPriceServiceImpl) GetPrice(ctx context.Context, req *cryptov1.GetPriceRequest) (*cryptov1.GetPriceResponse, error) {
	s.logger.Infof("gRPC GetPrice called with symbol: %s", req.Symbol)

	// 调用业务服务
	priceResp, err := s.priceService.GetPrice(ctx, req.Symbol)
	if err != nil {
		s.logger.Errorf("Failed to get price: %v", err)
		return &cryptov1.GetPriceResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 转换响应
	return &cryptov1.GetPriceResponse{
		Symbol:    priceResp.Symbol,
		Price:     priceResp.Price,
		Currency:  priceResp.Currency,
		Timestamp: 0, // TODO: 需要在model中添加timestamp字段或使用UpdatedAt
		Source:    priceResp.Source,
		Success:   true,
		Message:   "success",
	}, nil
}

// GetBTCPrice 获取BTC价格
func (s *CryptoPriceServiceImpl) GetBTCPrice(ctx context.Context, req *cryptov1.GetBTCPriceRequest) (*cryptov1.GetPriceResponse, error) {
	s.logger.Info("gRPC GetBTCPrice called")

	// 调用业务服务
	priceResp, err := s.priceService.GetBTCPrice(ctx)
	if err != nil {
		s.logger.Errorf("Failed to get BTC price: %v", err)
		return &cryptov1.GetPriceResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 转换响应
	return &cryptov1.GetPriceResponse{
		Symbol:    priceResp.Symbol,
		Price:     priceResp.Price,
		Currency:  priceResp.Currency,
		Timestamp: 0, // TODO: 需要在model中添加timestamp字段或使用UpdatedAt
		Source:    priceResp.Source,
		Success:   true,
		Message:   "success",
	}, nil
}