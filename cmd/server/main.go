package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"crypto-info/internal/config"
	"crypto-info/internal/pkg/database"
	"crypto-info/internal/pkg/logger"
	"crypto-info/internal/server"

	"github.com/gin-gonic/gin"
)

var (
	// 构建信息，通过ldflags注入
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"

	// 命令行参数
	configPath = flag.String("config", "", "配置文件路径")
	version    = flag.Bool("version", false, "显示版本信息")
)

func main() {
	flag.Parse()

	// 显示版本信息
	if *version {
		fmt.Printf("crypto-info\n")
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		return
	}

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	logger.Init(&cfg.Log)
	log := logger.GetLogger()

	log.Infof("Starting crypto-info server, version: %s, build time: %s", Version, BuildTime)

	// 设置Gin模式
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// 初始化数据库连接
	var redisClient database.RedisClient
	if cfg.Database.Redis.Host != "" {
		redisClient, err = database.NewRedisClient(&cfg.Database.Redis)
		if err != nil {
			log.Errorf("Failed to connect to Redis: %v", err)
			// Redis连接失败不退出程序，使用内存缓存
		}
	}

	// 创建HTTP服务器
	httpServer, err := server.NewHTTPServer(cfg, redisClient)
	if err != nil {
		log.Fatalf("Failed to create HTTP server: %v", err)
	}

	// 启动HTTP服务器
	go func() {
		log.Infof("HTTP server starting on %s", cfg.GetHTTPAddr())
		if err := httpServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed to start: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 关闭HTTP服务器
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Errorf("HTTP server forced to shutdown: %v", err)
	} else {
		log.Info("HTTP server shutdown gracefully")
	}

	// 关闭数据库连接
	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			log.Errorf("Failed to close Redis connection: %v", err)
		}
	}

	log.Info("Server shutdown complete")
}