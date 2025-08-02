package dao

import "go-web-study/internal/models"

// CryptoPriceDAO 加密货币价格数据访问接口
type CryptoPriceDAO interface {
	// GetCryptoPriceFromAPI 从外部API获取指定加密货币价格
	GetCryptoPriceFromAPI(symbol string) (*models.CryptoPriceResponse, error)
	
	// GetBitcoinPriceFromAPI 从外部API获取比特币价格 (向后兼容)
	GetBitcoinPriceFromAPI() (*models.BitcoinPriceResponse, error)
	
	// GetCachedPrice 获取缓存的价格数据
	GetCachedPrice(symbol string) (*models.CryptoPriceResponse, error)
	
	// SavePriceToCache 保存价格数据到缓存
	SavePriceToCache(symbol string, price *models.CryptoPriceResponse) error
	
	// GetMockPrice 获取指定加密货币的模拟价格数据
	GetMockPrice(symbol string) *models.CryptoPriceResponse
}

// HealthDAO 健康检查数据访问接口
type HealthDAO interface {
	// CheckSystemHealth 检查系统健康状态
	CheckSystemHealth() *models.Response
	
	// CheckDatabaseHealth 检查数据库连接状态
	CheckDatabaseHealth() bool
	
	// CheckExternalAPIHealth 检查外部API连接状态
	CheckExternalAPIHealth() bool
}