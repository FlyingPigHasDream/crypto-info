package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"crypto-info/internal/config"
	"crypto-info/internal/pkg/database"
	"crypto-info/internal/pkg/logger"
	"crypto-info/internal/pkg/mq"
	"crypto-info/internal/server"
	"crypto-info/internal/service"
)

func main() {
	// 命令行参数
	var (
		enableHTTP    = flag.Bool("http", true, "Enable HTTP server (Gin)")
		enableHertz   = flag.Bool("hertz", false, "Enable Hertz server")
		enableGRPC    = flag.Bool("grpc", false, "Enable gRPC server (Kitex)")
		enableRocketMQ = flag.Bool("mq", false, "Enable RocketMQ message service")
		configPath    = flag.String("config", "configs/config.yaml", "Config file path")
	)
	flag.Parse()

	// 初始化配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logger.Init(&cfg.Log)
	appLogger := logger.GetLogger()

	// 初始化Redis客户端
	var redisClient database.RedisClient
	redisClient, err = database.NewRedisClient(&cfg.Database.Redis)
	if err != nil {
		appLogger.Warnf("Failed to connect to Redis: %v, continuing without cache", err)
		redisClient = nil
	}

	var wg sync.WaitGroup
	var servers []interface{ Shutdown(context.Context) error }

	// 初始化RocketMQ客户端
	var mqClient *mq.RocketMQClient
	var messageService *service.MessageService
	if *enableRocketMQ && cfg.RocketMQ.Enabled {
		// 获取logrus.Logger实例
		logrusLogger := logger.GetLogrusLogger()
		mqClient, err = mq.NewRocketMQClient(&cfg.RocketMQ, logrusLogger)
		if err != nil {
			appLogger.Warnf("Failed to create RocketMQ client: %v, continuing without MQ", err)
			mqClient = nil
		} else {
			// 启动RocketMQ客户端
			if err := mqClient.Start(); err != nil {
				appLogger.Warnf("Failed to start RocketMQ client: %v, continuing without MQ", err)
				mqClient = nil
			} else {
				// 初始化消息服务
				messageService = service.NewMessageService(mqClient, logrusLogger)
				if err := messageService.Start(); err != nil {
					appLogger.Warnf("Failed to start message service: %v", err)
				} else {
					appLogger.Info("RocketMQ message service started successfully")
				}
			}
		}
	}

	// 启动HTTP服务器 (Gin)
	if *enableHTTP {
		httpServer, err := server.NewHTTPServer(cfg, redisClient)
		if err != nil {
			appLogger.Fatalf("Failed to create HTTP server: %v", err)
		}
		servers = append(servers, httpServer)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := httpServer.Start(); err != nil {
				appLogger.Errorf("HTTP server error: %v", err)
			}
		}()
		appLogger.Info("HTTP server (Gin) started")
	}

	// 启动Hertz服务器
	if *enableHertz {
		// 修改端口避免冲突
		cfg.Server.HTTP.Port = 8081
		hertzServer := server.NewHertzServer(cfg, appLogger, redisClient)
		servers = append(servers, hertzServer)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := hertzServer.Start(); err != nil {
				appLogger.Errorf("Hertz server error: %v", err)
			}
		}()
		appLogger.Info("Hertz server started on port 8081")
	}

	// 启动gRPC服务器 (Kitex)
	if *enableGRPC {
		grpcServer := server.NewGRPCServer(cfg, appLogger, redisClient)
		servers = append(servers, grpcServer)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := grpcServer.Start(); err != nil {
				appLogger.Errorf("gRPC server error: %v", err)
			}
		}()
		appLogger.Infof("gRPC server started on port %d", cfg.Server.GRPC.Port)
	}

	if len(servers) == 0 {
		appLogger.Fatal("No servers enabled. Use -http, -hertz, or -grpc flags.")
	}

	appLogger.Info("Crypto Info Service started successfully")

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down servers...")

	// 优雅关闭所有服务器
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 关闭RocketMQ服务
	if messageService != nil {
		if err := messageService.Stop(); err != nil {
			appLogger.Errorf("Message service shutdown error: %v", err)
		}
	}
	if mqClient != nil {
		if err := mqClient.Stop(); err != nil {
			appLogger.Errorf("RocketMQ client shutdown error: %v", err)
		}
	}

	for _, srv := range servers {
		go func(s interface{ Shutdown(context.Context) error }) {
			if err := s.Shutdown(ctx); err != nil {
				appLogger.Errorf("Server shutdown error: %v", err)
			}
		}(srv)
	}

	// 等待所有服务器关闭
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		appLogger.Info("All servers exited gracefully")
	case <-ctx.Done():
		appLogger.Warn("Shutdown timeout, forcing exit")
	}
}
