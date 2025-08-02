package dao

import (
	"encoding/json"
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
	cache  map[string]*models.BitcoinPriceResponse
	mutex  sync.RWMutex
}

// NewCryptoPriceDAO 创建新的加密货币价格DAO实例
func NewCryptoPriceDAO() CryptoPriceDAO {
	return &CryptoPriceDAOImpl{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		cache: make(map[string]*models.BitcoinPriceResponse),
	}
}

// GetBitcoinPriceFromAPI 从火币API获取比特币价格
func (d *CryptoPriceDAOImpl) GetBitcoinPriceFromAPI() (*models.BitcoinPriceResponse, error) {
	resp, err := d.client.Get("https://api.huobi.pro/market/detail/merged?symbol=btcusdt")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, err
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
		return nil, err
	}

	priceResponse := &models.BitcoinPriceResponse{
		Price:     apiResponse.Tick.Close,
		Source:    "Huobi",
		UpdatedAt: time.Now().Format(time.RFC3339),
		Currency:  "USD",
	}

	// 保存到缓存
	d.SavePriceToCache("BTC", priceResponse)

	return priceResponse, nil
}

// GetCachedPrice 获取缓存的价格数据
func (d *CryptoPriceDAOImpl) GetCachedPrice(symbol string) (*models.BitcoinPriceResponse, error) {
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
func (d *CryptoPriceDAOImpl) SavePriceToCache(symbol string, price *models.BitcoinPriceResponse) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.cache[symbol] = price
	return nil
}

// generateMockPrice 生成模拟的比特币价格
func (d *CryptoPriceDAOImpl) generateMockPrice() float64 {
	basePrice := 95000.0
	rand.Seed(time.Now().Unix())
	variation := (rand.Float64() - 0.5) * 0.04 * basePrice
	return basePrice + variation
}

// GetMockPrice 获取模拟价格数据
func (d *CryptoPriceDAOImpl) GetMockPrice() *models.BitcoinPriceResponse {
	return &models.BitcoinPriceResponse{
		Price:     d.generateMockPrice(),
		Source:    "Mock Data (DAO Layer - 数据访问层模拟数据)",
		UpdatedAt: time.Now().Format(time.RFC3339),
		Currency:  "USD",
	}
}