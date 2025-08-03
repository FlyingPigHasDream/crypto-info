package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"crypto-info/internal/config"
	"crypto-info/internal/handler"
	"crypto-info/internal/pkg/database"
	"crypto-info/internal/pkg/logger"
	"crypto-info/internal/pkg/middleware"
	"crypto-info/internal/pkg/session"
	"crypto-info/internal/service"

	"github.com/gin-gonic/gin"
)

// HTTPServer HTTP服务器
type HTTPServer struct {
	server         *http.Server
	config         *config.Config
	logger         logger.Logger
	sessionManager *session.Manager
}

// NewHTTPServer 创建HTTP服务器
func NewHTTPServer(cfg *config.Config, redisClient database.RedisClient) (*HTTPServer, error) {
	log := logger.GetLogger()

	// 创建session管理器
	var sessionManager *session.Manager
	if cfg.Security.Session.Enabled {
		var err error
		if redisClient != nil {
			sessionManager, err = session.NewManager(&cfg.Security.Session, redisClient.GetClient(), log)
		} else {
			sessionManager, err = session.NewManager(&cfg.Security.Session, nil, log)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to create session manager: %w", err)
		}
		log.Info("Session manager initialized")
	}

	// 创建Gin引擎
	router := gin.New()

	// 注册中间件
	setupMiddleware(router, cfg, sessionManager)

	// 注册路由
	setupRoutes(router, cfg, redisClient, sessionManager)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:           cfg.GetHTTPAddr(),
		Handler:        router,
		ReadTimeout:    cfg.Server.HTTP.ReadTimeout,
		WriteTimeout:   cfg.Server.HTTP.WriteTimeout,
		IdleTimeout:    cfg.Server.HTTP.IdleTimeout,
		MaxHeaderBytes: cfg.Server.HTTP.MaxHeaderBytes,
	}

	return &HTTPServer{
		server:         server,
		config:         cfg,
		logger:         log,
		sessionManager: sessionManager,
	}, nil
}

// Start 启动服务器
func (s *HTTPServer) Start() error {
	return s.server.ListenAndServe()
}

// Shutdown 关闭服务器
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// setupMiddleware 设置中间件
func setupMiddleware(router *gin.Engine, cfg *config.Config, sessionManager *session.Manager) {
	// 请求ID中间件
	router.Use(middleware.RequestID())

	// 日志中间件
	router.Use(middleware.Logger())

	// 恢复中间件
	router.Use(middleware.Recovery())

	// 安全头中间件
	router.Use(middleware.Security())

	// CORS中间件
	if cfg.Security.CORS.Enabled {
		router.Use(middleware.CORS(
			cfg.Security.CORS.AllowedOrigins,
			cfg.Security.CORS.AllowedMethods,
			cfg.Security.CORS.AllowedHeaders,
			cfg.Security.CORS.AllowCredentials,
			cfg.Security.CORS.MaxAge,
		))
	}

	// 健康检查中间件
	if cfg.Monitoring.HealthCheck.Enabled {
		router.Use(middleware.HealthCheck(cfg.Monitoring.HealthCheck.Path))
	}

	// Session中间件
	if sessionManager != nil {
		router.Use(middleware.Session(sessionManager))
	}

	// 超时中间件
	router.Use(middleware.Timeout(30 * time.Second))
}

// setupRoutes 设置路由
func setupRoutes(router *gin.Engine, cfg *config.Config, redisClient database.RedisClient, sessionManager *session.Manager) {
	// 创建服务层
	priceService := service.NewPriceService(redisClient, cfg)
	volumeService := service.NewVolumeService(redisClient, cfg)

	// 创建处理器
	priceHandler := handler.NewPriceHandler(priceService)
	volumeHandler := handler.NewVolumeHandler(volumeService)
	var sessionHandler *handler.SessionHandler
	if sessionManager != nil {
		sessionHandler = handler.NewSessionHandler(sessionManager)
	}

	// API v1 路由组
	v1 := router.Group("/api/v1")
	{
		// 加密货币路由组
		crypto := v1.Group("/crypto")
		{
			// 价格相关路由
			crypto.GET("/price", priceHandler.GetPrice)
			crypto.GET("/btc-price", priceHandler.GetBTCPrice)

			// 交易量相关路由
			volume := crypto.Group("/volume")
			{
				volume.GET("/analysis", volumeHandler.GetVolumeAnalysis)
				volume.GET("/fluctuation", volumeHandler.GetMarketVolumeFluctuation)
				volume.GET("/comparison", volumeHandler.GetVolumeComparison)
				volume.GET("/top", volumeHandler.GetTopVolumeCoins)
			}
		}

		// Session相关路由
		if sessionHandler != nil {
			session := v1.Group("/session")
			{
				session.GET("/info", sessionHandler.GetSession)
				session.GET("/status", sessionHandler.SessionStatus)
				session.POST("/data", sessionHandler.SetSessionData)
				session.GET("/data/:key", sessionHandler.GetSessionData)
				session.DELETE("/data/:key", sessionHandler.RemoveSessionData)
				session.POST("/refresh", sessionHandler.RefreshSession)
				session.DELETE("/destroy", sessionHandler.DestroySession)
			}
		}
	}

	// 兼容旧版路由
	router.GET("/crypto/price", priceHandler.GetPrice)
	router.GET("/btc-price", priceHandler.GetBTCPrice)
	router.GET("/crypto/volume/analysis", volumeHandler.GetVolumeAnalysis)
	router.GET("/crypto/volume/fluctuation", volumeHandler.GetMarketVolumeFluctuation)
	router.GET("/crypto/volume/comparison", volumeHandler.GetVolumeComparison)
	router.GET("/crypto/volume/top", volumeHandler.GetTopVolumeCoins)

	// 根路径
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "crypto-info",
			"version": cfg.App.Version,
			"status":  "running",
		})
	})
}