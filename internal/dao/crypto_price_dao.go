package dao

import (
	"encoding/json"
	"fmt"
	"go-web-study/internal/models"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// CryptoPriceDAOImpl 加密货币价格DAO实现
type CryptoPriceDAOImpl struct {
	client *http.Client
	cache  map[string]*models.CryptoPriceResponse
	mutex  sync.RWMutex
}

// NewCryptoPriceDAO 创建新的加密货币价格DAO实例
func NewCryptoPriceDAO() CryptoPriceDAO {
	return &CryptoPriceDAOImpl{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		cache: make(map[string]*models.CryptoPriceResponse),
	}
}

// GetCryptoPriceFromAPI 从火币API获取指定加密货币价格
func (d *CryptoPriceDAOImpl) GetCryptoPriceFromAPI(symbol string) (*models.CryptoPriceResponse, error) {
	// 获取火币API符号映射
	huobiSymbol, exists := models.CryptoSymbolMapping[symbol]
	if !exists {
		return nil, fmt.Errorf("unsupported cryptocurrency symbol: %s", symbol)
	}

	url := fmt.Sprintf("https://api.huobi.pro/market/detail/merged?symbol=%s", huobiSymbol)
	resp, err := d.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResponse models.HuobiResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, err
	}

	if apiResponse.Status != "ok" {
		return nil, fmt.Errorf("API response status not ok: %s", apiResponse.Status)
	}

	priceResponse := &models.CryptoPriceResponse{
		Symbol:    symbol,
		Price:     apiResponse.Tick.Close,
		Source:    "Huobi",
		UpdatedAt: time.Now().Format(time.RFC3339),
		Currency:  "USD",
	}

	// 保存到缓存
	d.SavePriceToCache(symbol, priceResponse)

	return priceResponse, nil
}

// GetBitcoinPriceFromAPI 从火币API获取比特币价格 (向后兼容)
func (d *CryptoPriceDAOImpl) GetBitcoinPriceFromAPI() (*models.BitcoinPriceResponse, error) {
	return d.GetCryptoPriceFromAPI("BTC")
}

// GetCachedPrice 获取缓存的价格数据
func (d *CryptoPriceDAOImpl) GetCachedPrice(symbol string) (*models.CryptoPriceResponse, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if price, exists := d.cache[symbol]; exists {
		// 检查缓存是否过期（5分钟）
		updatedAt, err := time.Parse(time.RFC3339, price.UpdatedAt)
		if err == nil && time.Since(updatedAt) < 5*time.Minute {
			return price, nil
		}
	}

	return nil, nil // 缓存不存在或已过期
}

// SavePriceToCache 保存价格数据到缓存
func (d *CryptoPriceDAOImpl) SavePriceToCache(symbol string, price *models.CryptoPriceResponse) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.cache[symbol] = price
	return nil
}

// generateMockPrice 生成模拟的加密货币价格
func (d *CryptoPriceDAOImpl) generateMockPrice(symbol string) float64 {
	// 不同加密货币的基础价格
	basePrices := map[string]float64{
		"BTC": 95000.0,
		"ETH": 3500.0,
		"LTC": 150.0,
		"BCH": 500.0,
	}
	
	basePrice, exists := basePrices[symbol]
	if !exists {
		basePrice = 1000.0 // 默认价格
	}
	
	rand.Seed(time.Now().Unix())
	variation := (rand.Float64() - 0.5) * 0.04 * basePrice
	return basePrice + variation
}

// GetMockPrice 获取指定加密货币的模拟价格数据
func (d *CryptoPriceDAOImpl) GetMockPrice(symbol string) *models.CryptoPriceResponse {
	return &models.CryptoPriceResponse{
		Symbol:    symbol,
		Price:     d.generateMockPrice(symbol),
		Source:    "Mock Data (DAO Layer - 数据访问层模拟数据)",
		UpdatedAt: time.Now().Format(time.RFC3339),
		Currency:  "USD",
	}
}