package routes

import (
	"go-web-study/internal/controllers"
	"github.com/gin-gonic/gin"
)

// SetupBaseRoutes 设置基础路由
func SetupBaseRoutes(router *gin.Engine) {
	// 基础路由
	router.GET("/", controllers.HomeController)
	router.GET("/health", controllers.HealthController)
}