package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"crypto-info/internal/config"
	"crypto-info/internal/model"
	"crypto-info/internal/pkg/database"
	"crypto-info/internal/pkg/logger"
)

// PriceService 价格服务接口
type PriceService interface {
	GetPrice(ctx context.Context, symbol string) (*model.PriceResponse, error)
	GetBTCPrice(ctx context.Context) (*model.PriceResponse, error)
}

// priceService 价格服务实现
type priceService struct {
	redisClient database.RedisClient
	config      *config.Config
	logger      logger.Logger
	bscService  BSCService
}

// NewPriceService 创建价格服务
func NewPriceService(redisClient database.RedisClient, cfg *config.Config, bscService BSCService) PriceService {
	return &priceService{
		redisClient: redisClient,
		config:      cfg,
		logger:      logger.GetLogger(),
		bscService:  bscService,
	}
}

// GetPrice 获取加密货币价格
func (s *priceService) GetPrice(ctx context.Context, symbol string) (*model.PriceResponse, error) {
	// 参数验证
	if symbol == "" {
		symbol = s.config.Business.DefaultSymbol
	}

	// 检查是否支持该币种
	if !s.isSupportedSymbol(symbol) {
		return nil, fmt.Errorf("unsupported symbol: %s", symbol)
	}

	// 尝试从缓存获取
	if s.redisClient != nil {
		if cached, err := s.getPriceFromCache(ctx, symbol); err == nil && cached != nil {
			s.logger.Debugf("Price cache hit for symbol: %s", symbol)
			return cached, nil
		}
	}

	// 获取价格数据
	price, err := s.fetchPrice(ctx, symbol)
	if err != nil {
		s.logger.Errorf("Failed to fetch price for %s: %v", symbol, err)
		return nil, err
	}

	// 缓存结果
	if s.redisClient != nil {
		if err := s.setPriceCache(ctx, symbol, price); err != nil {
			s.logger.Warnf("Failed to cache price for %s: %v", symbol, err)
		}
	}

	return price, nil
}

// GetBTCPrice 获取BTC价格
func (s *priceService) GetBTCPrice(ctx context.Context) (*model.PriceResponse, error) {
	return s.GetPrice(ctx, "BTC")
}

// fetchPrice 获取价格数据
func (s *priceService) fetchPrice(ctx context.Context, symbol string) (*model.PriceResponse, error) {
	// 如果启用了模拟数据，返回模拟价格
	if s.config.Business.MockDataEnabled {
		return s.generateMockPrice(symbol), nil
	}

	// 优先使用BSC链上流动性数据计算价格
	if s.bscService != nil && s.config.BSC.Enabled {
		price, err := s.bscService.GetTokenPriceInUSDT(ctx, symbol)
		if err == nil {
			priceFloat, _ := price.Float64()
			return &model.PriceResponse{
				Symbol:    symbol,
				Price:     priceFloat,
				Currency:  "USDT",
				UpdatedAt: time.Now().Format(time.RFC3339),
				Source:    "BSC_Liquidity",
			}, nil
		}
		s.logger.Warnf("Failed to get price from BSC for %s: %v, falling back to mock data", symbol, err)
	}

	// 如果BSC服务不可用，回退到模拟数据
	return s.generateMockPrice(symbol), nil
}

// generateMockPrice 生成模拟价格数据
func (s *priceService) generateMockPrice(symbol string) *model.PriceResponse {
	prices := map[string]float64{
		"BTC": 45000.0,
		"ETH": 3000.0,
		"LTC": 150.0,
		"BCH": 400.0,
		"ADA": 0.5,
		"DOT": 25.0,
		"LINK": 20.0,
		"XRP": 0.6,
	}

	basePrice, exists := prices[symbol]
	if !exists {
		basePrice = 100.0
	}

	// 添加一些随机波动
	variation := (time.Now().Unix() % 100) - 50
	finalPrice := basePrice + float64(variation)*basePrice*0.001

	return &model.PriceResponse{
		Symbol:    symbol,
		Price:     finalPrice,
		Source:    "Mock Data",
		UpdatedAt: time.Now().Format(time.RFC3339),
		Currency:  "USD",
	}
}

// isSupportedSymbol 检查是否支持该币种
func (s *priceService) isSupportedSymbol(symbol string) bool {
	for _, supported := range s.config.Business.SupportedSymbols {
		if supported == symbol {
			return true
		}
	}
	return false
}

// getPriceFromCache 从缓存获取价格
func (s *priceService) getPriceFromCache(ctx context.Context, symbol string) (*model.PriceResponse, error) {
	cacheKey := fmt.Sprintf("price:%s", symbol)
	cachedData, err := s.redisClient.Get(ctx, cacheKey)
	if err != nil || cachedData == "" {
		return nil, fmt.Errorf("cache miss")
	}

	var price model.PriceResponse
	if err := json.Unmarshal([]byte(cachedData), &price); err != nil {
		return nil, err
	}

	return &price, nil
}

// setPriceCache 设置价格缓存
func (s *priceService) setPriceCache(ctx context.Context, symbol string, price *model.PriceResponse) error {
	cacheKey := fmt.Sprintf("price:%s", symbol)
	data, err := json.Marshal(price)
	if err != nil {
		return err
	}

	return s.redisClient.Set(ctx, cacheKey, data, s.config.Cache.PriceTTL)
}