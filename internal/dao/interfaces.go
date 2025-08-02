package dao

import "go-web-study/internal/models"

// CryptoPriceDAO 加密货币价格数据访问接口
type CryptoPriceDAO interface {
	// GetBitcoinPriceFromAPI 从外部API获取比特币价格
	GetBitcoinPriceFromAPI() (*models.BitcoinPriceResponse, error)
	
	// GetCachedPrice 获取缓存的价格数据
	GetCachedPrice(symbol string) (*models.BitcoinPriceResponse, error)
	
	// SavePriceToCache 保存价格数据到缓存
	SavePriceToCache(symbol string, price *models.BitcoinPriceResponse) error
	
	// GetMockPrice 获取模拟价格数据
	GetMockPrice() *models.BitcoinPriceResponse
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