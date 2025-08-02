package model

import "time"

// PriceResponse 价格响应结构
type PriceResponse struct {
	Symbol    string  `json:"symbol"`     // 加密货币符号
	Price     float64 `json:"price"`      // 价格
	Source    string  `json:"source"`     // 数据源
	UpdatedAt string  `json:"updated_at"` // 更新时间
	Currency  string  `json:"currency"`   // 货币单位
}

// VolumeData 交易量数据结构
type VolumeData struct {
	Date   string  `json:"date"`   // 日期
	Volume float64 `json:"volume"` // 交易量
	Amount float64 `json:"amount"` // 交易额
}

// VolumeAnalysisResponse 交易量分析响应结构
type VolumeAnalysisResponse struct {
	Symbol      string        `json:"symbol"`       // 加密货币符号
	Period      string        `json:"period"`       // 分析周期
	Data        []VolumeData  `json:"data"`         // 历史交易量数据
	AvgVolume   float64       `json:"avg_volume"`   // 平均交易量
	MaxVolume   float64       `json:"max_volume"`   // 最大交易量
	MinVolume   float64       `json:"min_volume"`   // 最小交易量
	Volatility  float64       `json:"volatility"`   // 波动率
	Trend       string        `json:"trend"`        // 趋势
	Source      string        `json:"source"`       // 数据源
	GeneratedAt string        `json:"generated_at"` // 生成时间
}

// VolumeComparisonResponse 交易量对比响应结构
type VolumeComparisonResponse struct {
	Symbols     []string                 `json:"symbols"`      // 对比的加密货币符号
	Period      string                   `json:"period"`       // 分析周期
	Comparison  []VolumeAnalysisResponse `json:"comparison"`   // 对比数据
	GeneratedAt string                   `json:"generated_at"` // 生成时间
}

// TopVolumeCoinsResponse 交易量排行响应结构
type TopVolumeCoinsResponse struct {
	Period      string                   `json:"period"`       // 分析周期
	Limit       int                      `json:"limit"`        // 返回数量限制
	TopCoins    []VolumeAnalysisResponse `json:"top_coins"`    // 排行数据
	GeneratedAt string                   `json:"generated_at"` // 生成时间
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Error   string `json:"error"`   // 错误类型
	Message string `json:"message"` // 错误消息
	Code    int    `json:"code"`    // 错误代码
}

// HealthResponse 健康检查响应结构
type HealthResponse struct {
	Status    string `json:"status"`    // 状态
	Timestamp int64  `json:"timestamp"` // 时间戳
	Service   string `json:"service"`   // 服务名称
}

// APIResponse 通用API响应结构
type APIResponse struct {
	Success bool        `json:"success"`   // 是否成功
	Data    interface{} `json:"data"`      // 数据
	Error   *ErrorResponse `json:"error,omitempty"` // 错误信息
	Meta    *Meta       `json:"meta,omitempty"`  // 元数据
}

// Meta 元数据结构
type Meta struct {
	RequestID string    `json:"request_id,omitempty"` // 请求ID
	Timestamp time.Time `json:"timestamp"`            // 时间戳
	Version   string    `json:"version,omitempty"`    // API版本
}

// PaginationMeta 分页元数据
type PaginationMeta struct {
	Page       int `json:"page"`        // 当前页
	PageSize   int `json:"page_size"`   // 每页大小
	Total      int `json:"total"`       // 总数
	TotalPages int `json:"total_pages"` // 总页数
}

// CacheInfo 缓存信息
type CacheInfo struct {
	Cached    bool      `json:"cached"`     // 是否来自缓存
	CachedAt  time.Time `json:"cached_at"`  // 缓存时间
	ExpiresAt time.Time `json:"expires_at"` // 过期时间
}