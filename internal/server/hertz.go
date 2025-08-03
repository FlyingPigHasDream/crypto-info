package server

import (
	"context"
	"fmt"
	"time"

	"crypto-info/internal/config"
	"crypto-info/internal/handler"
	"crypto-info/internal/pkg/database"
	"crypto-info/internal/pkg/logger"
	"crypto-info/internal/pkg/session"
	"crypto-info/internal/service"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/gin-gonic/gin"
)

// HertzServer Hertz HTTP服务器
type HertzServer struct {
	server         *server.Hertz
	config         *config.Config
	logger         logger.Logger
	sessionManager *session.Manager
}

// NewHertzServer 创建新的Hertz服务器
func NewHertzServer(cfg *config.Config, log logger.Logger, redisClient database.RedisClient) *HertzServer {
	// 创建Hertz服务器实例
	h := server.Default(
		server.WithHostPorts(fmt.Sprintf("%s:%d", cfg.Server.HTTP.Host, cfg.Server.HTTP.Port)),
		server.WithReadTimeout(30*time.Second),
		server.WithWriteTimeout(30*time.Second),
		server.WithIdleTimeout(60*time.Second),
	)

	// 使用默认Hertz日志配置
	// TODO: 后续可以创建适配器来集成现有的logger

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
			log.Errorf("Failed to create session manager: %v", err)
		} else {
			log.Info("Session manager initialized for Hertz")
		}
	}

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

	// 设置中间件
	setupHertzMiddleware(h, cfg, log)

	// 设置路由
	setupHertzRoutes(h, priceHandler, volumeHandler, sessionHandler)

	return &HertzServer{
		server:         h,
		config:         cfg,
		logger:         log,
		sessionManager: sessionManager,
	}
}

// Start 启动服务器
func (s *HertzServer) Start() error {
	s.logger.Info(fmt.Sprintf("Hertz server starting on %s:%d", s.config.Server.HTTP.Host, s.config.Server.HTTP.Port))
	return s.server.Run()
}

// Shutdown 优雅关闭服务器
func (s *HertzServer) Shutdown(ctx context.Context) error {
	s.logger.Info("Hertz server shutting down...")
	return s.server.Shutdown(ctx)
}

// setupHertzMiddleware 设置Hertz中间件
func setupHertzMiddleware(h *server.Hertz, cfg *config.Config, log logger.Logger) {
	// 请求ID中间件
	h.Use(func(ctx context.Context, c *app.RequestContext) {
		requestID := string(c.GetHeader("X-Request-ID"))
		if requestID == "" {
			requestID = fmt.Sprintf("%d", time.Now().UnixNano())
		}
		c.Header("X-Request-ID", requestID)
		c.Next(ctx)
	})

	// 日志中间件
	h.Use(func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		c.Next(ctx)
		latency := time.Since(start)

		log.WithFields(map[string]interface{}{
			"method":     string(c.Method()),
			"path":       string(c.Path()),
			"status":     c.Response.StatusCode(),
			"latency":    latency,
			"request_id": string(c.GetHeader("X-Request-ID")),
		}).Info("HTTP Request")
	})

	// 恢复中间件
	h.Use(func(ctx context.Context, c *app.RequestContext) {
		defer func() {
			if err := recover(); err != nil {
				log.WithFields(map[string]interface{}{
					"error": err,
					"path":  string(c.Path()),
				}).Error("Panic recovered")
				c.JSON(consts.StatusInternalServerError, map[string]interface{}{
					"error":   "Internal Server Error",
					"message": "服务器内部错误",
					"code":    500,
				})
				c.Abort()
			}
		}()
		c.Next(ctx)
	})

	// CORS中间件
	h.Use(func(ctx context.Context, c *app.RequestContext) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		c.Header("Access-Control-Max-Age", "86400")

		if string(c.Method()) == "OPTIONS" {
			c.Status(consts.StatusNoContent)
			c.Abort()
			return
		}
		c.Next(ctx)
	})
}

// setupHertzRoutes 设置Hertz路由
func setupHertzRoutes(h *server.Hertz, priceHandler *handler.PriceHandler, volumeHandler *handler.VolumeHandler, sessionHandler *handler.SessionHandler) {
	// 健康检查
	h.GET("/health", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(consts.StatusOK, map[string]interface{}{
			"status":  "ok",
			"message": "Crypto Info Service is running",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// 根路径
	h.GET("/", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(consts.StatusOK, map[string]interface{}{
			"message": "Welcome to Crypto Info Service (Hertz)",
			"version": "v1.0.0",
			"powered": "CloudWeGo Hertz",
		})
	})

	// API v1 路由组
	v1 := h.Group("/api/v1")
	{
		// 价格相关API
		v1.GET("/crypto/price", adaptHertzHandler(priceHandler.GetPrice))
		v1.GET("/crypto/btc-price", adaptHertzHandler(priceHandler.GetBTCPrice))

		// 交易量相关API
		v1.GET("/crypto/volume/analysis", adaptHertzHandler(volumeHandler.GetVolumeAnalysis))
		v1.GET("/crypto/volume/fluctuation", adaptHertzHandler(volumeHandler.GetMarketVolumeFluctuation))
		v1.GET("/crypto/volume/comparison", adaptHertzHandler(volumeHandler.GetVolumeComparison))
		v1.GET("/crypto/volume/top", adaptHertzHandler(volumeHandler.GetTopVolumeCoins))

		// Session相关API
		if sessionHandler != nil {
			v1.GET("/session/info", adaptHertzHandler(sessionHandler.GetSession))
			v1.GET("/session/status", adaptHertzHandler(sessionHandler.SessionStatus))
			v1.POST("/session/data", adaptHertzHandler(sessionHandler.SetSessionData))
			v1.GET("/session/data/:key", adaptHertzHandler(sessionHandler.GetSessionData))
			v1.DELETE("/session/data/:key", adaptHertzHandler(sessionHandler.RemoveSessionData))
			v1.POST("/session/refresh", adaptHertzHandler(sessionHandler.RefreshSession))
			v1.DELETE("/session/destroy", adaptHertzHandler(sessionHandler.DestroySession))
		}
	}

	// 兼容旧路由
	h.GET("/crypto/price", adaptHertzHandler(priceHandler.GetPrice))
	h.GET("/btc-price", adaptHertzHandler(priceHandler.GetBTCPrice))
	h.GET("/crypto/volume/analysis", adaptHertzHandler(volumeHandler.GetVolumeAnalysis))
	h.GET("/crypto/volume/fluctuation", adaptHertzHandler(volumeHandler.GetMarketVolumeFluctuation))
	h.GET("/crypto/volume/comparison", adaptHertzHandler(volumeHandler.GetVolumeComparison))
	h.GET("/crypto/volume/top", adaptHertzHandler(volumeHandler.GetTopVolumeCoins))
}

// adaptHertzHandler 适配Gin处理器到Hertz
func adaptHertzHandler(ginHandler func(*gin.Context)) func(context.Context, *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		// 创建一个简单的适配器，暂时返回占位响应
		// TODO: 需要重构Handler层以支持Hertz的RequestContext
		c.JSON(consts.StatusOK, map[string]interface{}{
			"message": "API endpoint available (Hertz)",
			"path":    string(c.Path()),
			"method":  string(c.Method()),
			"note":    "Handler adaptation in progress",
		})
	}
}