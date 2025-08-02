package services

import (
	"go-web-study/internal/dao"
	"go-web-study/internal/models"
)

// BitcoinService 比特币价格服务
type BitcoinService struct {
	cryptoPriceDAO dao.CryptoPriceDAO
}

// NewBitcoinService 创建新的比特币服务实例
func NewBitcoinService() *BitcoinService {
	return &BitcoinService{
		cryptoPriceDAO: dao.NewCryptoPriceDAO(),
	}
}

// GetBitcoinPrice 获取比特币价格
func (s *BitcoinService) GetBitcoinPrice() models.BitcoinPriceResponse {
	// 首先尝试从缓存获取
	if cachedPrice, err := s.cryptoPriceDAO.GetCachedPrice("BTC"); err == nil && cachedPrice != nil {
		return *cachedPrice
	}

	// 尝试从API获取实时数据
	if apiPrice, err := s.cryptoPriceDAO.GetBitcoinPriceFromAPI(); err == nil && apiPrice != nil {
		return *apiPrice
	}

	// 如果API调用失败，返回模拟数据
	mockPrice := s.cryptoPriceDAO.GetMockPrice()
	return *mockPrice
}