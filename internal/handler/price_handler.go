package handler

import (
	"net/http"

	"crypto-info/internal/model"
	"crypto-info/internal/pkg/logger"
	"crypto-info/internal/service"

	"github.com/gin-gonic/gin"
)

// PriceHandler 价格处理器
type PriceHandler struct {
	priceService service.PriceService
	logger       logger.Logger
}

// NewPriceHandler 创建价格处理器
func NewPriceHandler(priceService service.PriceService) *PriceHandler {
	return &PriceHandler{
		priceService: priceService,
		logger:       logger.GetLogger(),
	}
}

// GetPrice 获取加密货币价格
// @Summary 获取加密货币价格
// @Description 根据符号获取加密货币的当前价格
// @Tags 价格
// @Accept json
// @Produce json
// @Param symbol query string false "加密货币符号" default(BTC)
// @Success 200 {object} model.PriceResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/crypto/price [get]
func (h *PriceHandler) GetPrice(c *gin.Context) {
	symbol := c.Query("symbol")
	requestID := c.GetString("request_id")

	h.logger.WithField("request_id", requestID).Infof("Getting price for symbol: %s", symbol)

	price, err := h.priceService.GetPrice(c.Request.Context(), symbol)
	if err != nil {
		h.logger.WithField("request_id", requestID).Errorf("Failed to get price: %v", err)
		h.respondWithError(c, http.StatusInternalServerError, "获取价格失败", err.Error())
		return
	}

	h.respondWithSuccess(c, price)
}

// GetBTCPrice 获取BTC价格
// @Summary 获取BTC价格
// @Description 获取比特币的当前价格
// @Tags 价格
// @Accept json
// @Produce json
// @Success 200 {object} model.PriceResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/crypto/btc-price [get]
func (h *PriceHandler) GetBTCPrice(c *gin.Context) {
	requestID := c.GetString("request_id")

	h.logger.WithField("request_id", requestID).Info("Getting BTC price")

	price, err := h.priceService.GetBTCPrice(c.Request.Context())
	if err != nil {
		h.logger.WithField("request_id", requestID).Errorf("Failed to get BTC price: %v", err)
		h.respondWithError(c, http.StatusInternalServerError, "获取BTC价格失败", err.Error())
		return
	}

	h.respondWithSuccess(c, price)
}

// respondWithSuccess 成功响应
func (h *PriceHandler) respondWithSuccess(c *gin.Context, data interface{}) {
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
func (h *PriceHandler) respondWithError(c *gin.Context, statusCode int, message, detail string) {
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