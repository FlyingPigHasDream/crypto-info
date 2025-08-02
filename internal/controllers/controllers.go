package controllers

import (
	"net/http"
	"strings"
	"go-web-study/internal/models"
	"go-web-study/internal/services"
	"github.com/gin-gonic/gin"
)

var (
	cryptoService  = services.NewCryptoService()
	bitcoinService = services.NewBitcoinService() // 向后兼容
	healthService  = services.NewHealthService()
)

// HomeController 处理主页请求
func HomeController(c *gin.Context) {
	response := models.Response{
		Message: "Welcome to Go Web Study!",
		Status:  http.StatusOK,
	}

	c.JSON(http.StatusOK, response)
}

// HealthController 处理健康检查请求
func HealthController(c *gin.Context) {
	response := healthService.GetSystemHealth()
	c.JSON(http.StatusOK, response)
}

// CryptoPriceController 处理加密货币价格查询请求 (适配器模式)
func CryptoPriceController(c *gin.Context) {
	// 从URL参数获取加密货币符号，默认为BTC
	symbol := c.DefaultQuery("symbol", "BTC")
	
	// 将符号转换为大写
	symbol = strings.ToUpper(symbol)
	
	priceResponse := cryptoService.GetCryptoPrice(symbol)
	c.JSON(http.StatusOK, priceResponse)
}

// BitcoinPriceController 处理比特币价格查询请求 (向后兼容)
func BitcoinPriceController(c *gin.Context) {
	priceResponse := bitcoinService.GetBitcoinPrice()
	c.JSON(http.StatusOK, priceResponse)
}
