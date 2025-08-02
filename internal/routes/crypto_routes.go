package routes

import (
	"go-web-study/internal/controllers"
	"github.com/gin-gonic/gin"
)

// SetupCryptoRoutes 设置加密货币相关路由
func SetupCryptoRoutes(router *gin.Engine) {
	// 加密货币路由组
	crypto := router.Group("/crypto")
	{
		// 通用加密货币价格查询 (适配器模式)
		crypto.GET("/price", controllers.CryptoPriceController)
		// 向后兼容的比特币价格路由
		crypto.GET("/btc-price", controllers.BitcoinPriceController)
		// 未来可以扩展更多加密货币相关接口
		// crypto.GET("/eth-price", controllers.EthereumPriceController)
		// crypto.GET("/prices", controllers.AllCryptoPricesController)
		// crypto.GET("/market-cap", controllers.MarketCapController)
	}

	// 保持向后兼容的旧路由
	router.GET("/btc-price", controllers.BitcoinPriceController)
}