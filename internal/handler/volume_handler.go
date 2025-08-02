package handler

import (
	"net/http"
	"strconv"
	"strings"

	"crypto-info/internal/model"
	"crypto-info/internal/pkg/logger"
	"crypto-info/internal/service"

	"github.com/gin-gonic/gin"
)

// VolumeHandler 交易量处理器
type VolumeHandler struct {
	volumeService service.VolumeService
	logger        logger.Logger
}

// NewVolumeHandler 创建交易量处理器
func NewVolumeHandler(volumeService service.VolumeService) *VolumeHandler {
	return &VolumeHandler{
		volumeService: volumeService,
		logger:        logger.GetLogger(),
	}
}

// GetVolumeAnalysis 获取交易量分析
// @Summary 获取交易量分析
// @Description 获取指定加密货币的交易量分析数据
// @Tags 交易量
// @Accept json
// @Produce json
// @Param symbol query string false "加密货币符号" default(BTC)
// @Param days query int false "分析天数" default(10)
// @Success 200 {object} model.VolumeAnalysisResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/crypto/volume/analysis [get]
func (h *VolumeHandler) GetVolumeAnalysis(c *gin.Context) {
	symbol := c.Query("symbol")
	daysStr := c.Query("days")
	requestID := c.GetString("request_id")

	days := 10 // 默认值
	if daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil {
			days = parsedDays
		}
	}

	h.logger.WithField("request_id", requestID).Infof("Getting volume analysis for symbol: %s, days: %d", symbol, days)

	analysis, err := h.volumeService.GetVolumeAnalysis(c.Request.Context(), symbol, days)
	if err != nil {
		h.logger.WithField("request_id", requestID).Errorf("Failed to get volume analysis: %v", err)
		h.respondWithError(c, http.StatusInternalServerError, "获取交易量分析失败", err.Error())
		return
	}

	h.respondWithSuccess(c, analysis)
}

// GetMarketVolumeFluctuation 获取市场交易量波动
// @Summary 获取市场交易量波动
// @Description 获取指定加密货币的市场交易量波动数据
// @Tags 交易量
// @Accept json
// @Produce json
// @Param symbol query string false "加密货币符号" default(BTC)
// @Param days query int false "分析天数" default(10)
// @Success 200 {object} model.VolumeAnalysisResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/crypto/volume/fluctuation [get]
func (h *VolumeHandler) GetMarketVolumeFluctuation(c *gin.Context) {
	symbol := c.Query("symbol")
	daysStr := c.Query("days")
	requestID := c.GetString("request_id")

	days := 10 // 默认值
	if daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil {
			days = parsedDays
		}
	}

	h.logger.WithField("request_id", requestID).Infof("Getting market volume fluctuation for symbol: %s, days: %d", symbol, days)

	fluctuation, err := h.volumeService.GetMarketVolumeFluctuation(c.Request.Context(), symbol, days)
	if err != nil {
		h.logger.WithField("request_id", requestID).Errorf("Failed to get market volume fluctuation: %v", err)
		h.respondWithError(c, http.StatusInternalServerError, "获取市场交易量波动失败", err.Error())
		return
	}

	h.respondWithSuccess(c, fluctuation)
}

// GetVolumeComparison 获取交易量对比
// @Summary 获取交易量对比
// @Description 获取多个加密货币的交易量对比数据
// @Tags 交易量
// @Accept json
// @Produce json
// @Param symbols query string false "加密货币符号列表，逗号分隔" default(BTC,ETH,LTC)
// @Param days query int false "分析天数" default(10)
// @Success 200 {object} model.VolumeComparisonResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/crypto/volume/comparison [get]
func (h *VolumeHandler) GetVolumeComparison(c *gin.Context) {
	symbolsStr := c.Query("symbols")
	daysStr := c.Query("days")
	requestID := c.GetString("request_id")

	var symbols []string
	if symbolsStr != "" {
		symbols = strings.Split(symbolsStr, ",")
		// 清理空格
		for i, symbol := range symbols {
			symbols[i] = strings.TrimSpace(symbol)
		}
	} else {
		symbols = []string{"BTC", "ETH", "LTC"}
	}

	days := 10 // 默认值
	if daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil {
			days = parsedDays
		}
	}

	h.logger.WithField("request_id", requestID).Infof("Getting volume comparison for symbols: %v, days: %d", symbols, days)

	comparison, err := h.volumeService.GetVolumeComparison(c.Request.Context(), symbols, days)
	if err != nil {
		h.logger.WithField("request_id", requestID).Errorf("Failed to get volume comparison: %v", err)
		h.respondWithError(c, http.StatusInternalServerError, "获取交易量对比失败", err.Error())
		return
	}

	h.respondWithSuccess(c, comparison)
}

// GetTopVolumeCoins 获取交易量排行
// @Summary 获取交易量排行
// @Description 获取交易量排行榜数据
// @Tags 交易量
// @Accept json
// @Produce json
// @Param days query int false "分析天数" default(10)
// @Param limit query int false "返回数量限制" default(10)
// @Success 200 {object} model.TopVolumeCoinsResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/crypto/volume/top [get]
func (h *VolumeHandler) GetTopVolumeCoins(c *gin.Context) {
	daysStr := c.Query("days")
	limitStr := c.Query("limit")
	requestID := c.GetString("request_id")

	days := 10 // 默认值
	if daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil {
			days = parsedDays
		}
	}

	limit := 10 // 默认值
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	h.logger.WithField("request_id", requestID).Infof("Getting top volume coins for days: %d, limit: %d", days, limit)

	topCoins, err := h.volumeService.GetTopVolumeCoins(c.Request.Context(), days, limit)
	if err != nil {
		h.logger.WithField("request_id", requestID).Errorf("Failed to get top volume coins: %v", err)
		h.respondWithError(c, http.StatusInternalServerError, "获取交易量排行失败", err.Error())
		return
	}

	h.respondWithSuccess(c, topCoins)
}

// respondWithSuccess 成功响应
func (h *VolumeHandler) respondWithSuccess(c *gin.Context, data interface{}) {
	response := model.APIResponse{
		Success: true,
		Data:    data,
		Meta: &model.Meta{
			RequestID: c.GetString("request_id"),
			Timestamp: c.GetTime("timestamp"),
			Version:   "v1",
		},
	}

	c.JSON(http.StatusOK, response)
}

// respondWithError 错误响应
func (h *VolumeHandler) respondWithError(c *gin.Context, statusCode int, message, detail string) {
	errorResp := &model.ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
		Code:    statusCode,
	}

	response := model.APIResponse{
		Success: false,
		Error:   errorResp,
		Meta: &model.Meta{
			RequestID: c.GetString("request_id"),
			Timestamp: c.GetTime("timestamp"),
			Version:   "v1",
		},
	}

	c.JSON(statusCode, response)
}