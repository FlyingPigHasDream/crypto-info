package session

import (
	"errors"
	"net/http"
	"time"

	"crypto-info/internal/config"
	"crypto-info/internal/pkg/logger"

	"github.com/gin-gonic/gin"
)

const (
	// SessionKey 在gin.Context中存储session的key
	SessionKey = "session"
	// SessionIDKey 在gin.Context中存储session ID的key
	SessionIDKey = "session_id"
)

// Middleware Session中间件
func Middleware(manager *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		if manager == nil {
			c.Next()
			return
		}

		config := manager.GetConfig()
		if !config.Enabled {
			c.Next()
			return
		}

		log := logger.GetLogger()

		// 从cookie中获取session ID
		sessionID, err := c.Cookie(config.CookieName)
		if err != nil || sessionID == "" {
			// 创建新session
			session, err := manager.CreateSession(c.Request.Context())
			if err != nil {
				log.Errorf("Failed to create session: %v", err)
				c.Next()
				return
			}
			sessionID = session.ID
			c.Set(SessionKey, session)
		} else {
			// 获取现有session
			session, err := manager.GetSession(c.Request.Context(), sessionID)
			if err != nil {
				// session不存在或已过期，创建新session
				session, err = manager.CreateSession(c.Request.Context())
				if err != nil {
					log.Errorf("Failed to create session: %v", err)
					c.Next()
					return
				}
				sessionID = session.ID
			}
			c.Set(SessionKey, session)
		}

		// 设置session ID到context
		c.Set(SessionIDKey, sessionID)

		// 设置cookie
		setSessionCookie(c, config, sessionID)

		// 处理请求
		c.Next()

		// 请求处理完成后保存session（如果有修改）
		if sessionValue, exists := c.Get(SessionKey); exists {
			if session, ok := sessionValue.(*Session); ok {
				if err := manager.SaveSession(c.Request.Context(), session); err != nil {
					log.Errorf("Failed to save session: %v", err)
				}
			}
		}
	}
}

// setSessionCookie 设置session cookie
func setSessionCookie(c *gin.Context, cfg *config.SessionConfig, sessionID string) {
	maxAge := int(cfg.MaxAge.Seconds())
	if maxAge <= 0 {
		maxAge = int((24 * time.Hour).Seconds()) // 默认24小时
	}

	sameSite := http.SameSiteDefaultMode
	switch cfg.SameSite {
	case "strict":
		sameSite = http.SameSiteStrictMode
	case "lax":
		sameSite = http.SameSiteLaxMode
	case "none":
		sameSite = http.SameSiteNoneMode
	}

	c.SetSameSite(sameSite)
	c.SetCookie(
		cfg.CookieName,
		sessionID,
		maxAge,
		cfg.Path,
		cfg.Domain,
		cfg.Secure,
		cfg.HttpOnly,
	)
}

// GetSession 从gin.Context获取session
func GetSession(c *gin.Context) (*Session, bool) {
	value, exists := c.Get(SessionKey)
	if !exists {
		return nil, false
	}

	session, ok := value.(*Session)
	return session, ok
}

// GetSessionID 从gin.Context获取session ID
func GetSessionID(c *gin.Context) (string, bool) {
	value, exists := c.Get(SessionIDKey)
	if !exists {
		return "", false
	}

	sessionID, ok := value.(string)
	return sessionID, ok
}

// SetSessionData 设置session数据
func SetSessionData(c *gin.Context, key string, value interface{}) error {
	session, exists := GetSession(c)
	if !exists {
		return errors.New("session not found")
	}

	session.Data[key] = value
	session.UpdatedAt = time.Now()
	return nil
}

// GetSessionData 获取session数据
func GetSessionData(c *gin.Context, key string) (interface{}, bool) {
	session, exists := GetSession(c)
	if !exists {
		return nil, false
	}

	value, exists := session.Data[key]
	return value, exists
}

// RemoveSessionData 移除session数据
func RemoveSessionData(c *gin.Context, key string) error {
	session, exists := GetSession(c)
	if !exists {
		return errors.New("session not found")
	}

	delete(session.Data, key)
	session.UpdatedAt = time.Now()
	return nil
}

// DestroySession 销毁session
func DestroySession(c *gin.Context, manager *Manager) error {
	sessionID, exists := GetSessionID(c)
	if !exists {
		return errors.New("session ID not found")
	}

	// 删除session数据
	if err := manager.DeleteSession(c.Request.Context(), sessionID); err != nil {
		return err
	}

	// 清除cookie
	cfg := manager.GetConfig()
	c.SetCookie(
		cfg.CookieName,
		"",
		-1,
		cfg.Path,
		cfg.Domain,
		cfg.Secure,
		cfg.HttpOnly,
	)

	// 清除context中的session
	c.Set(SessionKey, nil)
	c.Set(SessionIDKey, "")

	return nil
}