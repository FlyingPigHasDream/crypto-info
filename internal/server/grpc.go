package server

import (
	"context"
	"fmt"
	"net"

	"crypto-info/internal/config"
	"crypto-info/internal/grpc"
	"crypto-info/internal/pkg/database"
	"crypto-info/internal/pkg/logger"
	"crypto-info/internal/service"
	cryptov1 "crypto-info/kitex_gen/crypto/v1/cryptopriceservice"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
)

// GRPCServer Kitex gRPC服务器
type GRPCServer struct {
	server server.Server
	config *config.Config
	logger logger.Logger
}

// NewGRPCServer 创建新的gRPC服务器
func NewGRPCServer(cfg *config.Config, log logger.Logger, redisClient database.RedisClient) *GRPCServer {
	// 创建服务层
	priceService := service.NewPriceService(redisClient, cfg)

	// 创建gRPC服务实现
	priceServiceImpl := grpc.NewCryptoPriceService(priceService)

	// 创建Kitex服务器
	addr, _ := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", cfg.Server.GRPC.Host, cfg.Server.GRPC.Port))
	svr := cryptov1.NewServer(
		priceServiceImpl,
		server.WithServiceAddr(addr),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: "crypto-price-service",
			Method:      "",
			Tags:        map[string]string{"env": cfg.App.Env},
		}),
	)

	return &GRPCServer{
		server: svr,
		config: cfg,
		logger: log,
	}
}

// Start 启动gRPC服务器
func (s *GRPCServer) Start() error {
	s.logger.Info(fmt.Sprintf("gRPC server starting on %s:%d", s.config.Server.GRPC.Host, s.config.Server.GRPC.Port))
	return s.server.Run()
}

// Stop 停止gRPC服务器
func (s *GRPCServer) Stop() error {
	s.logger.Info("gRPC server stopping...")
	return s.server.Stop()
}

// Shutdown 优雅关闭gRPC服务器
func (s *GRPCServer) Shutdown(ctx context.Context) error {
	s.logger.Info("gRPC server shutting down...")
	// Kitex服务器没有Shutdown方法，使用Stop
	return s.server.Stop()
}