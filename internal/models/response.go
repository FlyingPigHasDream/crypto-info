package models

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