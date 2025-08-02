package handlers

import (
	"encoding/json"
	"net/http"
	"time"
	"io/ioutil"
	"math/rand"
)

// Response 定义API响应结构
type Response struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

// BitcoinPriceResponse 定义比特币价格响应结构
type BitcoinPriceResponse struct {
	Price     float64 `json:"price"`
	Source    string  `json:"source"`
	UpdatedAt string  `json:"updated_at"`
	Currency  string  `json:"currency"`
}

// HuobiResponse 定义火币API响应结构
type HuobiResponse struct {
	Status string `json:"status"`
	Tick   struct {
		Close float64 `json:"close"`
	} `json:"tick"`
}

// generateMockPrice 生成模拟的比特币价格（基于当前时间的波动）
func generateMockPrice() float64 {
	// 基础价格
	basePrice := 95000.0
	// 使用当前时间作为随机种子
	rand.Seed(time.Now().Unix())
	// 生成-2%到+2%的价格波动
	variation := (rand.Float64() - 0.5) * 0.04 * basePrice
	return basePrice + variation
}

// HomeHandler 处理主页请求
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	response := Response{
		Message: "Welcome to Go Web Study!",
		Status:  http.StatusOK,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HealthHandler 处理健康检查请求
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	response := Response{
		Message: "OK",
		Status:  http.StatusOK,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// BitcoinPriceHandler 处理比特币价格查询请求
func BitcoinPriceHandler(w http.ResponseWriter, r *http.Request) {
	// 创建一个超时时间
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	// 尝试调用火币API获取比特币价格
	resp, err := client.Get("https://api.huobi.pro/market/detail/merged?symbol=btcusdt")
	if err != nil {
		// 如果API调用失败，返回模拟数据
		priceResponse := BitcoinPriceResponse{
			Price:     generateMockPrice(),
			Source:    "Mock Data (Network Error - 网络连接失败，显示模拟数据)",
			UpdatedAt: time.Now().Format(time.RFC3339),
			Currency:  "USD",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(priceResponse)
		return
	}
	defer resp.Body.Close()
	
	// 检查HTTP响应状态码
	if resp.StatusCode != http.StatusOK {
		// 如果API返回错误状态码，返回模拟数据
		priceResponse := BitcoinPriceResponse{
			Price:     generateMockPrice(),
			Source:    "Mock Data (API Error - API返回错误，显示模拟数据)",
			UpdatedAt: time.Now().Format(time.RFC3339),
			Currency:  "USD",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(priceResponse)
		return
	}
	
	// 读取响应体
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// 如果读取响应失败，返回模拟数据
		priceResponse := BitcoinPriceResponse{
			Price:     generateMockPrice(),
			Source:    "Mock Data (Read Error - 数据读取失败，显示模拟数据)",
			UpdatedAt: time.Now().Format(time.RFC3339),
			Currency:  "USD",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(priceResponse)
		return
	}
	
	// 解析JSON响应
	var apiResponse HuobiResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		// 如果解析JSON失败，返回模拟数据
		priceResponse := BitcoinPriceResponse{
			Price:     generateMockPrice(),
			Source:    "Mock Data (Parse Error - JSON解析失败，显示模拟数据)",
			UpdatedAt: time.Now().Format(time.RFC3339),
			Currency:  "USD",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(priceResponse)
		return
	}
	
	// 检查API响应状态
	if apiResponse.Status != "ok" {
		// 如果API状态不正常，返回模拟数据
		priceResponse := BitcoinPriceResponse{
			Price:     generateMockPrice(),
			Source:    "Mock Data (API Status Error - API状态异常，显示模拟数据)",
			UpdatedAt: time.Now().Format(time.RFC3339),
			Currency:  "USD",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(priceResponse)
		return
	}
	
	// 构建响应
	priceResponse := BitcoinPriceResponse{
		Price:     apiResponse.Tick.Close,
		Source:    "Huobi",
		UpdatedAt: time.Now().Format(time.RFC3339),
		Currency:  "USD",
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(priceResponse)
}