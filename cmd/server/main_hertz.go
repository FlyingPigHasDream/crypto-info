package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"crypto-info/internal/config"
	"crypto-info/internal/pkg/database"
	"crypto-info/internal/pkg/logger"
	"crypto-info/internal/server"
)

func main() {
	// 初始化配置
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logger.Init(&cfg.Log)
	appLogger := logger.GetLogger()

	// 初始化Redis客户端
	var redisClient database.RedisClient
	// 假设Redis总是启用的，可以根据需要添加配置
	redisClient, err = database.NewRedisClient(&cfg.Database.Redis)
	if err != nil {
		appLogger.Warnf("Failed to connect to Redis: %v, continuing without cache", err)
		redisClient = nil
	}

	// 创建Hertz服务器
	hertzServer := server.NewHertzServer(cfg, appLogger, redisClient)

	// 启动服务器
	go func() {
		if err := hertzServer.Start(); err != nil {
			appLogger.Fatalf("Failed to start Hertz server: %v", err)
		}
	}()

	appLogger.Info("Crypto Info Service (Hertz) started successfully")

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := hertzServer.Shutdown(ctx); err != nil {
		appLogger.Errorf("Server forced to shutdown: %v", err)
	}

	appLogger.Info("Server exited")
}