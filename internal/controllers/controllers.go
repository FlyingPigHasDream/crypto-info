package controllers

import (
	"net/http"
	"go-web-study/internal/models"
	"go-web-study/internal/services"
	"github.com/gin-gonic/gin"
)

var (
	bitcoinService = services.NewBitcoinService()
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
	response := models.Response{
		Message: "OK",
		Status:  http.StatusOK,
	}

	c.JSON(http.StatusOK, response)
}

// BitcoinPriceController 处理比特币价格查询请求
func BitcoinPriceController(c *gin.Context) {
	priceResponse := bitcoinService.GetBitcoinPrice()
	c.JSON(http.StatusOK, priceResponse)
}
