package routes

import (
	"go-web-study/internal/controllers"
	"github.com/gin-gonic/gin"
)

// SetupAPIRoutes 设置API版本化路由
func SetupAPIRoutes(router *gin.Engine) {
	// API v1 路由组
	v1 := router.Group("/api/v1")
	{
		// 基础接口
		v1.GET("/health", controllers.HealthController)
		
		// 加密货币接口
		// 通用加密货币价格查询 (适配器模式)
		v1.GET("/crypto/price", controllers.CryptoPriceController)
		// 向后兼容的比特币价格路由
		v1.GET("/btc-price", controllers.BitcoinPriceController)
		
		// 未来可以扩展更多API接口
		// v1.GET("/user/profile", controllers.UserProfileController)
		// v1.POST("/user/login", controllers.UserLoginController)
		// v1.GET("/market/summary", controllers.MarketSummaryController)
	}

	// 未来可以添加 v2 版本
	// v2 := router.Group("/api/v2")
	// {
	//     v2.GET("/health", controllers.HealthV2Controller)
	//     v2.GET("/crypto/prices", controllers.CryptoPricesV2Controller)
	// }
}