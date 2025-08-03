package handler

import (
	"math/big"
	"net/http"
	"strconv"

	"crypto-info/internal/model"
	"crypto-info/internal/pkg/logger"
	"crypto-info/internal/service"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

// BSCHandler BSC处理器
type BSCHandler struct {
	bscService service.BSCService
	logger     logger.Logger
}

// NewBSCHandler 创建BSC处理器
func NewBSCHandler(bscService service.BSCService) *BSCHandler {
	return &BSCHandler{
		bscService: bscService,
		logger:     logger.GetLogger(),
	}
}

// GetStatus 获取BSC监控状态
// @Summary 获取BSC监控状态
// @Description 获取BSC链上数据监控服务的当前状态
// @Tags BSC
// @Accept json
// @Produce json
// @Success 200 {object} model.BSCMonitoringResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/bsc/status [get]
func (h *BSCHandler) GetStatus(c *gin.Context) {
	requestID := c.GetString("request_id")

	h.logger.WithField("request_id", requestID).Info("Getting BSC monitoring status")

	status := h.bscService.GetStatus()
	h.respondWithSuccess(c, status)
}

// GetLatestBlock 获取最新区块信息
// @Summary 获取最新区块信息
// @Description 获取BSC链上的最新区块信息
// @Tags BSC
// @Accept json
// @Produce json
// @Success 200 {object} model.BSCBlock
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/bsc/block/latest [get]
func (h *BSCHandler) GetLatestBlock(c *gin.Context) {
	requestID := c.GetString("request_id")

	h.logger.WithField("request_id", requestID).Info("Getting latest BSC block")

	block, err := h.bscService.GetLatestBlock(c.Request.Context())
	if err != nil {
		h.logger.WithField("request_id", requestID).Errorf("Failed to get latest block: %v", err)
		h.respondWithError(c, http.StatusInternalServerError, "获取最新区块失败", err.Error())
		return
	}

	h.respondWithSuccess(c, block)
}

// GetTransactions 获取区块交易信息
// @Summary 获取区块交易信息
// @Description 获取指定区块的交易信息
// @Tags BSC
// @Accept json
// @Produce json
// @Param block_number query string true "区块号"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} model.BSCTransactionResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/bsc/transactions [get]
func (h *BSCHandler) GetTransactions(c *gin.Context) {
	requestID := c.GetString("request_id")
	blockNumberStr := c.Query("block_number")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	if blockNumberStr == "" {
		h.respondWithError(c, http.StatusBadRequest, "参数错误", "block_number is required")
		return
	}

	blockNumber, ok := new(big.Int).SetString(blockNumberStr, 10)
	if !ok {
		h.respondWithError(c, http.StatusBadRequest, "参数错误", "invalid block_number")
		return
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	h.logger.WithField("request_id", requestID).Infof("Getting transactions for block %s, page %d, pageSize %d", blockNumberStr, page, pageSize)

	transactions, err := h.bscService.GetTransactions(c.Request.Context(), blockNumber, page, pageSize)
	if err != nil {
		h.logger.WithField("request_id", requestID).Errorf("Failed to get transactions: %v", err)
		h.respondWithError(c, http.StatusInternalServerError, "获取交易信息失败", err.Error())
		return
	}

	h.respondWithSuccess(c, transactions)
}

// GetTokenTransfers 获取代币转账记录
// @Summary 获取代币转账记录
// @Description 获取指定代币的转账记录
// @Tags BSC
// @Accept json
// @Produce json
// @Param token_address query string true "代币合约地址"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} model.BSCTokenTransferResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/bsc/token/transfers [get]
func (h *BSCHandler) GetTokenTransfers(c *gin.Context) {
	requestID := c.GetString("request_id")
	tokenAddressStr := c.Query("token_address")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	if tokenAddressStr == "" {
		h.respondWithError(c, http.StatusBadRequest, "参数错误", "token_address is required")
		return
	}

	if !common.IsHexAddress(tokenAddressStr) {
		h.respondWithError(c, http.StatusBadRequest, "参数错误", "invalid token_address")
		return
	}

	tokenAddress := common.HexToAddress(tokenAddressStr)

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	h.logger.WithField("request_id", requestID).Infof("Getting token transfers for %s, page %d, pageSize %d", tokenAddressStr, page, pageSize)

	transfers, err := h.bscService.GetTokenTransfers(c.Request.Context(), tokenAddress, page, pageSize)
	if err != nil {
		h.logger.WithField("request_id", requestID).Errorf("Failed to get token transfers: %v", err)
		h.respondWithError(c, http.StatusInternalServerError, "获取代币转账记录失败", err.Error())
		return
	}

	h.respondWithSuccess(c, transfers)
}

// GetSwapEvents 获取交换事件
// @Summary 获取交换事件
// @Description 获取指定交易对的交换事件
// @Tags BSC
// @Accept json
// @Produce json
// @Param pair_address query string true "交易对合约地址"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} model.BSCSwapEventResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/bsc/swap/events [get]
func (h *BSCHandler) GetSwapEvents(c *gin.Context) {
	requestID := c.GetString("request_id")
	pairAddressStr := c.Query("pair_address")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	if pairAddressStr == "" {
		h.respondWithError(c, http.StatusBadRequest, "参数错误", "pair_address is required")
		return
	}

	if !common.IsHexAddress(pairAddressStr) {
		h.respondWithError(c, http.StatusBadRequest, "参数错误", "invalid pair_address")
		return
	}

	pairAddress := common.HexToAddress(pairAddressStr)

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	h.logger.WithField("request_id", requestID).Infof("Getting swap events for %s, page %d, pageSize %d", pairAddressStr, page, pageSize)

	swaps, err := h.bscService.GetSwapEvents(c.Request.Context(), pairAddress, page, pageSize)
	if err != nil {
		h.logger.WithField("request_id", requestID).Errorf("Failed to get swap events: %v", err)
		h.respondWithError(c, http.StatusInternalServerError, "获取交换事件失败", err.Error())
		return
	}

	h.respondWithSuccess(c, swaps)
}

// GetPairInfo 获取交易对信息
// @Summary 获取交易对信息
// @Description 获取指定交易对的详细信息
// @Tags BSC
// @Accept json
// @Produce json
// @Param pair_address query string true "交易对合约地址"
// @Success 200 {object} model.BSCPairInfo
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/bsc/pair/info [get]
func (h *BSCHandler) GetPairInfo(c *gin.Context) {
	requestID := c.GetString("request_id")
	pairAddressStr := c.Query("pair_address")

	if pairAddressStr == "" {
		h.respondWithError(c, http.StatusBadRequest, "参数错误", "pair_address is required")
		return
	}

	if !common.IsHexAddress(pairAddressStr) {
		h.respondWithError(c, http.StatusBadRequest, "参数错误", "invalid pair_address")
		return
	}

	pairAddress := common.HexToAddress(pairAddressStr)

	h.logger.WithField("request_id", requestID).Infof("Getting pair info for %s", pairAddressStr)

	pairInfo, err := h.bscService.GetPairInfo(c.Request.Context(), pairAddress)
	if err != nil {
		h.logger.WithField("request_id", requestID).Errorf("Failed to get pair info: %v", err)
		h.respondWithError(c, http.StatusInternalServerError, "获取交易对信息失败", err.Error())
		return
	}

	h.respondWithSuccess(c, pairInfo)
}

// StartMonitoring 启动BSC监控
// @Summary 启动BSC监控
// @Description 启动BSC链上数据监控服务
// @Tags BSC
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/bsc/monitoring/start [post]
func (h *BSCHandler) StartMonitoring(c *gin.Context) {
	requestID := c.GetString("request_id")

	h.logger.WithField("request_id", requestID).Info("Starting BSC monitoring")

	err := h.bscService.Start(c.Request.Context())
	if err != nil {
		h.logger.WithField("request_id", requestID).Errorf("Failed to start BSC monitoring: %v", err)
		h.respondWithError(c, http.StatusInternalServerError, "启动BSC监控失败", err.Error())
		return
	}

	h.respondWithSuccess(c, map[string]interface{}{
		"message": "BSC monitoring started successfully",
		"status":  "running",
	})
}

// StopMonitoring 停止BSC监控
// @Summary 停止BSC监控
// @Description 停止BSC链上数据监控服务
// @Tags BSC
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/bsc/monitoring/stop [post]
func (h *BSCHandler) StopMonitoring(c *gin.Context) {
	requestID := c.GetString("request_id")

	h.logger.WithField("request_id", requestID).Info("Stopping BSC monitoring")

	err := h.bscService.Stop()
	if err != nil {
		h.logger.WithField("request_id", requestID).Errorf("Failed to stop BSC monitoring: %v", err)
		h.respondWithError(c, http.StatusInternalServerError, "停止BSC监控失败", err.Error())
		return
	}

	h.respondWithSuccess(c, map[string]interface{}{
		"message": "BSC monitoring stopped successfully",
		"status":  "stopped",
	})
}

// respondWithSuccess 成功响应
func (h *BSCHandler) respondWithSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// respondWithError 错误响应
func (h *BSCHandler) respondWithError(c *gin.Context, statusCode int, errorType, message string) {
	c.JSON(statusCode, model.ErrorResponse{
		Error:   errorType,
		Message: message,
		Code:    statusCode,
	})
}