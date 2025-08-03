package middleware

import (
	"net/http"
	"strconv"
	"time"

	"crypto-info/internal/pkg/logger"
	"crypto-info/internal/pkg/session"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID 请求ID中间件
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		log := logger.GetLogger()
		
		fields := map[string]interface{}{
			"timestamp":    param.TimeStamp.Format(time.RFC3339),
			"status":       param.StatusCode,
			"latency":      param.Latency.String(),
			"client_ip":    param.ClientIP,
			"method":       param.Method,
			"path":         param.Path,
			"user_agent":   param.Request.UserAgent(),
			"request_size": param.Request.ContentLength,
		}
		
		if requestID := param.Request.Header.Get("X-Request-ID"); requestID != "" {
			fields["request_id"] = requestID
		}
		
		if param.ErrorMessage != "" {
			fields["error"] = param.ErrorMessage
			log.WithFields(fields).Error("HTTP request completed with error")
		} else {
			log.WithFields(fields).Info("HTTP request completed")
		}
		
		return ""
	})
}

// Recovery 恢复中间件
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log := logger.GetLogger()
		
		fields := map[string]interface{}{
			"panic":     recovered,
			"method":    c.Request.Method,
			"path":      c.Request.URL.Path,
			"client_ip": c.ClientIP(),
		}
		
		if requestID := c.GetString("request_id"); requestID != "" {
			fields["request_id"] = requestID
		}
		
		log.WithFields(fields).Error("Panic recovered")
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "服务器内部错误",
			"code":    500,
		})
	})
}

// CORS 跨域中间件
func CORS(allowedOrigins []string, allowedMethods []string, allowedHeaders []string, allowCredentials bool, maxAge int) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// 检查允许的源
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}
		
		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		
		c.Header("Access-Control-Allow-Methods", joinStrings(allowedMethods, ", "))
		c.Header("Access-Control-Allow-Headers", joinStrings(allowedHeaders, ", "))
		c.Header("Access-Control-Max-Age", strconv.Itoa(maxAge))
		
		if allowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	}
}

// RateLimit 限流中间件
type RateLimiter interface {
	Allow(key string) bool
}

func RateLimit(limiter RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		
		if !limiter.Allow(key) {
			log := logger.GetLogger()
			log.WithField("client_ip", key).Warn("Rate limit exceeded")
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too Many Requests",
				"message": "请求过于频繁，请稍后再试",
				"code":    429,
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// Timeout 超时中间件
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置超时上下文
		ctx := c.Request.Context()
		select {
		case <-time.After(timeout):
			c.JSON(http.StatusRequestTimeout, gin.H{
				"error":   "Request Timeout",
				"message": "请求超时",
				"code":    408,
			})
			c.Abort()
			return
		case <-ctx.Done():
			c.Abort()
			return
		default:
			c.Next()
		}
	}
}

// Security 安全头中间件
func Security() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	}
}

// HealthCheck 健康检查中间件
func HealthCheck(path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == path {
			c.JSON(http.StatusOK, gin.H{
				"status":    "ok",
				"timestamp": time.Now().Unix(),
				"service":   "crypto-info",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// Session Session中间件
func Session(manager *session.Manager) gin.HandlerFunc {
	return session.Middleware(manager)
}

// joinStrings 辅助函数
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}
	
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}