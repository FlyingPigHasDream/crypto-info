package services

import (
	"go-web-study/internal/dao"
	"go-web-study/internal/models"
)

// CryptoService 加密货币价格服务
type CryptoService struct {
	cryptoPriceDAO dao.CryptoPriceDAO
}

// NewCryptoService 创建新的加密货币服务实例
func NewCryptoService() *CryptoService {
	return &CryptoService{
		cryptoPriceDAO: dao.NewCryptoPriceDAO(),
	}
}

// BitcoinService 为了向后兼容保留的比特币服务类型
type BitcoinService = CryptoService

// NewBitcoinService 创建新的比特币服务实例 (向后兼容)
func NewBitcoinService() *BitcoinService {
	return NewCryptoService()
}

// GetCryptoPrice 获取指定加密货币价格 (适配器模式)
func (s *CryptoService) GetCryptoPrice(symbol string) models.CryptoPriceResponse {
	// 首先尝试从缓存获取
	if cachedPrice, err := s.cryptoPriceDAO.GetCachedPrice(symbol); err == nil && cachedPrice != nil {
		return *cachedPrice
	}

	// 尝试从API获取实时数据
	if apiPrice, err := s.cryptoPriceDAO.GetCryptoPriceFromAPI(symbol); err == nil && apiPrice != nil {
		return *apiPrice
	}

	// 如果API调用失败，返回模拟数据
	mockPrice := s.cryptoPriceDAO.GetMockPrice(symbol)
	return *mockPrice
}

// GetBitcoinPrice 获取比特币价格 (向后兼容)
func (s *CryptoService) GetBitcoinPrice() models.BitcoinPriceResponse {
	return s.GetCryptoPrice("BTC")
}