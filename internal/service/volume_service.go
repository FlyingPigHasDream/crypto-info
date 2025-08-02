package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"crypto-info/internal/config"
	"crypto-info/internal/model"
	"crypto-info/internal/pkg/database"
	"crypto-info/internal/pkg/logger"
)

// VolumeService 交易量服务接口
type VolumeService interface {
	GetVolumeAnalysis(ctx context.Context, symbol string, days int) (*model.VolumeAnalysisResponse, error)
	GetMarketVolumeFluctuation(ctx context.Context, symbol string, days int) (*model.VolumeAnalysisResponse, error)
	GetVolumeComparison(ctx context.Context, symbols []string, days int) (*model.VolumeComparisonResponse, error)
	GetTopVolumeCoins(ctx context.Context, days, limit int) (*model.TopVolumeCoinsResponse, error)
}

// volumeService 交易量服务实现
type volumeService struct {
	redisClient database.RedisClient
	config      *config.Config
	logger      logger.Logger
}

// NewVolumeService 创建交易量服务
func NewVolumeService(redisClient database.RedisClient, cfg *config.Config) VolumeService {
	return &volumeService{
		redisClient: redisClient,
		config:      cfg,
		logger:      logger.GetLogger(),
	}
}

// GetVolumeAnalysis 获取交易量分析
func (s *volumeService) GetVolumeAnalysis(ctx context.Context, symbol string, days int) (*model.VolumeAnalysisResponse, error) {
	// 参数验证
	if symbol == "" {
		symbol = s.config.Business.DefaultSymbol
	}
	if days <= 0 {
		days = s.config.Business.DefaultAnalysisDays
	}
	if days > s.config.Business.MaxAnalysisDays {
		days = s.config.Business.MaxAnalysisDays
	}

	// 检查是否支持该币种
	if !s.isSupportedSymbol(symbol) {
		return nil, fmt.Errorf("unsupported symbol: %s", symbol)
	}

	// 尝试从缓存获取
	if s.redisClient != nil {
		if cached, err := s.getVolumeFromCache(ctx, symbol, days); err == nil && cached != nil {
			s.logger.Debugf("Volume analysis cache hit for symbol: %s, days: %d", symbol, days)
			return cached, nil
		}
	}

	// 获取交易量数据
	analysis, err := s.fetchVolumeAnalysis(ctx, symbol, days)
	if err != nil {
		s.logger.Errorf("Failed to fetch volume analysis for %s: %v", symbol, err)
		return nil, err
	}

	// 缓存结果
	if s.redisClient != nil {
		if err := s.setVolumeCache(ctx, symbol, days, analysis); err != nil {
			s.logger.Warnf("Failed to cache volume analysis for %s: %v", symbol, err)
		}
	}

	return analysis, nil
}

// GetMarketVolumeFluctuation 获取市场交易量波动
func (s *volumeService) GetMarketVolumeFluctuation(ctx context.Context, symbol string, days int) (*model.VolumeAnalysisResponse, error) {
	return s.GetVolumeAnalysis(ctx, symbol, days)
}

// GetVolumeComparison 获取交易量对比
func (s *volumeService) GetVolumeComparison(ctx context.Context, symbols []string, days int) (*model.VolumeComparisonResponse, error) {
	if len(symbols) == 0 {
		symbols = []string{"BTC", "ETH", "LTC"}
	}
	if days <= 0 {
		days = s.config.Business.DefaultAnalysisDays
	}

	var comparison []model.VolumeAnalysisResponse
	for _, symbol := range symbols {
		analysis, err := s.GetVolumeAnalysis(ctx, symbol, days)
		if err != nil {
			s.logger.Warnf("Failed to get volume analysis for %s: %v", symbol, err)
			continue
		}
		comparison = append(comparison, *analysis)
	}

	return &model.VolumeComparisonResponse{
		Symbols:     symbols,
		Period:      fmt.Sprintf("%d days", days),
		Comparison:  comparison,
		GeneratedAt: time.Now().Format(time.RFC3339),
	}, nil
}

// GetTopVolumeCoins 获取交易量排行
func (s *volumeService) GetTopVolumeCoins(ctx context.Context, days, limit int) (*model.TopVolumeCoinsResponse, error) {
	if days <= 0 {
		days = s.config.Business.DefaultAnalysisDays
	}
	if limit <= 0 {
		limit = 10
	}

	// 获取所有支持的币种数据
	var topCoins []model.VolumeAnalysisResponse
	for i, symbol := range s.config.Business.SupportedSymbols {
		if i >= limit {
			break
		}
		analysis, err := s.GetVolumeAnalysis(ctx, symbol, days)
		if err != nil {
			s.logger.Warnf("Failed to get volume analysis for %s: %v", symbol, err)
			continue
		}
		topCoins = append(topCoins, *analysis)
	}

	return &model.TopVolumeCoinsResponse{
		Period:      fmt.Sprintf("%d days", days),
		Limit:       limit,
		TopCoins:    topCoins,
		GeneratedAt: time.Now().Format(time.RFC3339),
	}, nil
}

// fetchVolumeAnalysis 获取交易量分析数据
func (s *volumeService) fetchVolumeAnalysis(ctx context.Context, symbol string, days int) (*model.VolumeAnalysisResponse, error) {
	// 如果启用了模拟数据，返回模拟数据
	if s.config.Business.MockDataEnabled {
		return s.generateMockVolumeAnalysis(symbol, days), nil
	}

	// TODO: 实现真实的API调用
	return s.generateMockVolumeAnalysis(symbol, days), nil
}

// generateMockVolumeAnalysis 生成模拟交易量分析数据
func (s *volumeService) generateMockVolumeAnalysis(symbol string, days int) *model.VolumeAnalysisResponse {
	baseVolumes := map[string]float64{
		"BTC": 1000000000,
		"ETH": 500000000,
		"LTC": 100000000,
		"BCH": 80000000,
		"ADA": 200000000,
		"DOT": 150000000,
		"LINK": 120000000,
		"XRP": 300000000,
	}

	baseVolume, exists := baseVolumes[symbol]
	if !exists {
		baseVolume = 50000000
	}

	// 生成历史数据
	var data []model.VolumeData
	var totalVolume float64
	var maxVolume, minVolume float64

	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		
		// 添加随机波动
		variation := (time.Now().Unix() + int64(i)) % 200 - 100
		volume := baseVolume + float64(variation)*baseVolume*0.01
		amount := volume * (45000 + float64(variation)*100) // 假设价格

		data = append(data, model.VolumeData{
			Date:   date,
			Volume: volume,
			Amount: amount,
		})

		totalVolume += volume
		if i == days-1 || volume > maxVolume {
			maxVolume = volume
		}
		if i == days-1 || volume < minVolume {
			minVolume = volume
		}
	}

	avgVolume := totalVolume / float64(days)
	volatility := (maxVolume - minVolume) / avgVolume * 100

	// 简单趋势判断
	trend := "稳定"
	if len(data) >= 2 {
		firstHalf := data[:len(data)/2]
		secondHalf := data[len(data)/2:]
		
		var firstAvg, secondAvg float64
		for _, d := range firstHalf {
			firstAvg += d.Volume
		}
		firstAvg /= float64(len(firstHalf))
		
		for _, d := range secondHalf {
			secondAvg += d.Volume
		}
		secondAvg /= float64(len(secondHalf))
		
		if secondAvg > firstAvg*1.05 {
			trend = "上升"
		} else if secondAvg < firstAvg*0.95 {
			trend = "下降"
		}
	}

	return &model.VolumeAnalysisResponse{
		Symbol:      symbol,
		Period:      fmt.Sprintf("%d days", days),
		Data:        data,
		AvgVolume:   math.Round(avgVolume),
		MaxVolume:   math.Round(maxVolume),
		MinVolume:   math.Round(minVolume),
		Volatility:  math.Round(volatility*100) / 100,
		Trend:       trend,
		Source:      "Mock Data",
		GeneratedAt: time.Now().Format(time.RFC3339),
	}
}

// isSupportedSymbol 检查是否支持该币种
func (s *volumeService) isSupportedSymbol(symbol string) bool {
	for _, supported := range s.config.Business.SupportedSymbols {
		if supported == symbol {
			return true
		}
	}
	return false
}

// getVolumeFromCache 从缓存获取交易量数据
func (s *volumeService) getVolumeFromCache(ctx context.Context, symbol string, days int) (*model.VolumeAnalysisResponse, error) {
	cacheKey := fmt.Sprintf("volume:%s:%d", symbol, days)
	cachedData, err := s.redisClient.Get(ctx, cacheKey)
	if err != nil || cachedData == "" {
		return nil, fmt.Errorf("cache miss")
	}

	var analysis model.VolumeAnalysisResponse
	if err := json.Unmarshal([]byte(cachedData), &analysis); err != nil {
		return nil, err
	}

	return &analysis, nil
}

// setVolumeCache 设置交易量缓存
func (s *volumeService) setVolumeCache(ctx context.Context, symbol string, days int, analysis *model.VolumeAnalysisResponse) error {
	cacheKey := fmt.Sprintf("volume:%s:%d", symbol, days)
	data, err := json.Marshal(analysis)
	if err != nil {
		return err
	}

	return s.redisClient.Set(ctx, cacheKey, data, s.config.Cache.VolumeTTL)
}