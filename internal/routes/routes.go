package routes

import (
	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置所有路由
// 通过调用各个业务模块的路由设置函数来组织路由
func SetupRoutes(router *gin.Engine) {
	// 设置基础路由
	SetupBaseRoutes(router)
	
	// 设置加密货币相关路由
	SetupCryptoRoutes(router)
	
	// 设置API版本化路由
	SetupAPIRoutes(router)
	
	// 未来可以继续添加其他业务模块路由
	// SetupUserRoutes(router)
	// SetupAdminRoutes(router)
	// SetupPaymentRoutes(router)
}