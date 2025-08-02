package services

import (
	"encoding/json"
	"go-web-study/internal/models"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
)

// BitcoinService 比特币价格服务
type BitcoinService struct {
	client *http.Client
}

// NewBitcoinService 创建新的比特币服务实例
func NewBitcoinService() *BitcoinService {
	return &BitcoinService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// generateMockPrice 生成模拟的比特币价格（基于当前时间的波动）
func (s *BitcoinService) generateMockPrice() float64 {
	// 基础价格
	basePrice := 95000.0
	// 使用当前时间作为随机种子
	rand.Seed(time.Now().Unix())
	// 生成-2%到+2%的价格波动
	variation := (rand.Float64() - 0.5) * 0.04 * basePrice
	return basePrice + variation
}

// GetBitcoinPrice 获取比特币价格
func (s *BitcoinService) GetBitcoinPrice() models.BitcoinPriceResponse {
	// 尝试调用火币API获取比特币价格
	resp, err := s.client.Get("https://api.huobi.pro/market/detail/merged?symbol=btcusdt")
	if err != nil {
		// 如果API调用失败，返回模拟数据
		return models.BitcoinPriceResponse{
			Price:     s.generateMockPrice(),
			Source:    "Mock Data (Network Error - 网络连接失败，显示模拟数据)",
			UpdatedAt: time.Now().Format(time.RFC3339),
			Currency:  "USD",
		}
	}
	defer resp.Body.Close()

	// 检查HTTP响应状态码
	if resp.StatusCode != http.StatusOK {
		// 如果API返回错误状态码，返回模拟数据
		return models.BitcoinPriceResponse{
			Price:     s.generateMockPrice(),
			Source:    "Mock Data (API Error - API返回错误，显示模拟数据)",
			UpdatedAt: time.Now().Format(time.RFC3339),
			Currency:  "USD",
		}
	}

	// 读取响应体
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// 如果读取响应失败，返回模拟数据
		return models.BitcoinPriceResponse{
			Price:     s.generateMockPrice(),
			Source:    "Mock Data (Read Error - 数据读取失败，显示模拟数据)",
			UpdatedAt: time.Now().Format(time.RFC3339),
			Currency:  "USD",
		}
	}

	// 解析JSON响应
	var apiResponse models.HuobiResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		// 如果解析JSON失败，返回模拟数据
		return models.BitcoinPriceResponse{
			Price:     s.generateMockPrice(),
			Source:    "Mock Data (Parse Error - JSON解析失败，显示模拟数据)",
			UpdatedAt: time.Now().Format(time.RFC3339),
			Currency:  "USD",
		}
	}

	// 检查API响应状态
	if apiResponse.Status != "ok" {
		// 如果API状态不正常，返回模拟数据
		return models.BitcoinPriceResponse{
			Price:     s.generateMockPrice(),
			Source:    "Mock Data (API Status Error - API状态异常，显示模拟数据)",
			UpdatedAt: time.Now().Format(time.RFC3339),
			Currency:  "USD",
		}
	}

	// 构建响应
	return models.BitcoinPriceResponse{
		Price:     apiResponse.Tick.Close,
		Source:    "Huobi",
		UpdatedAt: time.Now().Format(time.RFC3339),
		Currency:  "USD",
	}
}