package models

// Response 定义API响应结构
type Response struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

// CryptoPriceResponse 定义加密货币价格响应结构
type CryptoPriceResponse struct {
	Symbol    string  `json:"symbol"`    // 加密货币符号 (BTC, ETH, etc.)
	Price     float64 `json:"price"`
	Source    string  `json:"source"`
	UpdatedAt string  `json:"updated_at"`
	Currency  string  `json:"currency"`
}

// BitcoinPriceResponse 为了向后兼容保留的比特币价格响应结构
type BitcoinPriceResponse = CryptoPriceResponse

// HuobiResponse 定义火币API响应结构
type HuobiResponse struct {
	Status string `json:"status"`
	Tick   struct {
		Close float64 `json:"close"`
	} `json:"tick"`
}

// CryptoSymbolMapping 定义加密货币符号映射
var CryptoSymbolMapping = map[string]string{
	"BTC": "btcusdt",
	"ETH": "ethusdt",
	"LTC": "ltcusdt",
	"BCH": "bchusdt",
}