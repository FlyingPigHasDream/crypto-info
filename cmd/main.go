package main

import (
	"fmt"
	"go-web-study/internal/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	// 创建Gin路由器
	r := gin.Default()

	// 设置路由
	routes.SetupRoutes(r)

	fmt.Println("Server starting on port 8080...")
	// 启动服务器
	r.Run(":8080")
}