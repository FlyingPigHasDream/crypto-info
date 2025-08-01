package handlers

import (
	"encoding/json"
	"net/http"
	"time"
	"io/ioutil"
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

// CoinGeckoResponse 定义CoinGecko API响应结构
type CoinGeckoResponse struct {
	Bitcoin struct {
		USD float64 `json:"usd"`
	} `json:"bitcoin"`
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
	
	// 尝试调用CoinGecko API获取比特币价格
	resp, err := client.Get("https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&vs_currencies=usd")
	if err != nil {
		// 如果API调用失败，返回模拟数据
		priceResponse := BitcoinPriceResponse{
			Price:     95000.00, // 模拟价格
			Source:    "Mock Data (API Unavailable)",
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
			Price:     95000.00, // 模拟价格
			Source:    "Mock Data (API Error)",
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
			Price:     95000.00, // 模拟价格
			Source:    "Mock Data (Read Error)",
			UpdatedAt: time.Now().Format(time.RFC3339),
			Currency:  "USD",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(priceResponse)
		return
	}
	
	// 解析JSON响应
	var apiResponse CoinGeckoResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		// 如果解析JSON失败，返回模拟数据
		priceResponse := BitcoinPriceResponse{
			Price:     95000.00, // 模拟价格
			Source:    "Mock Data (Parse Error)",
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
		Price:     apiResponse.Bitcoin.USD,
		Source:    "CoinGecko",
		UpdatedAt: time.Now().Format(time.RFC3339),
		Currency:  "USD",
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(priceResponse)
}