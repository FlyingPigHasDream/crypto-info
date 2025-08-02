package routes

import (
	"go-web-study/internal/controllers"
	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置所有路由
func SetupRoutes(router *gin.Engine) {
	// API v1 路由组
	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", controllers.HealthController)
		v1.GET("/btc-price", controllers.BitcoinPriceController)
	}

	// 根路由
	router.GET("/", controllers.HomeController)
	router.GET("/health", controllers.HealthController)
	router.GET("/btc-price", controllers.BitcoinPriceController)
}