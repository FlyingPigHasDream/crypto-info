package handler

import (
	"net/http"
	"time"

	"crypto-info/internal/pkg/session"

	"github.com/gin-gonic/gin"
)

// SessionHandler Session处理器
type SessionHandler struct {
	manager *session.Manager
}

// NewSessionHandler 创建Session处理器
func NewSessionHandler(manager *session.Manager) *SessionHandler {
	return &SessionHandler{
		manager: manager,
	}
}

// GetSession 获取session信息
func (h *SessionHandler) GetSession(c *gin.Context) {
	sess, exists := session.GetSession(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Session not found",
			"message": "会话不存在",
			"code":    404,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": sess.ID,
		"data":       sess.Data,
		"created_at": sess.CreatedAt,
		"updated_at": sess.UpdatedAt,
		"expires_at": sess.ExpiresAt,
	})
}

// SetSessionData 设置session数据
func (h *SessionHandler) SetSessionData(c *gin.Context) {
	var req struct {
		Key   string      `json:"key" binding:"required"`
		Value interface{} `json:"value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": "请求参数无效: " + err.Error(),
			"code":    400,
		})
		return
	}

	if err := session.SetSessionData(c, req.Key, req.Value); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to set session data",
			"message": "设置会话数据失败: " + err.Error(),
			"code":    500,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Session data set successfully",
		"key":     req.Key,
		"value":   req.Value,
	})
}

// GetSessionData 获取session数据
func (h *SessionHandler) GetSessionData(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Key is required",
			"message": "键名是必需的",
			"code":    400,
		})
		return
	}

	value, exists := session.GetSessionData(c, key)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Session data not found",
			"message": "会话数据不存在",
			"code":    404,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"key":   key,
		"value": value,
	})
}

// RemoveSessionData 移除session数据
func (h *SessionHandler) RemoveSessionData(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Key is required",
			"message": "键名是必需的",
			"code":    400,
		})
		return
	}

	if err := session.RemoveSessionData(c, key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to remove session data",
			"message": "移除会话数据失败: " + err.Error(),
			"code":    500,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Session data removed successfully",
		"key":     key,
	})
}

// DestroySession 销毁session
func (h *SessionHandler) DestroySession(c *gin.Context) {
	if err := session.DestroySession(c, h.manager); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to destroy session",
			"message": "销毁会话失败: " + err.Error(),
			"code":    500,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Session destroyed successfully",
	})
}

// RefreshSession 刷新session过期时间
func (h *SessionHandler) RefreshSession(c *gin.Context) {
	sessionID, exists := session.GetSessionID(c)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Session not found",
			"message": "会话不存在",
			"code":    404,
		})
		return
	}

	if err := h.manager.RefreshSession(c.Request.Context(), sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to refresh session",
			"message": "刷新会话失败: " + err.Error(),
			"code":    500,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Session refreshed successfully",
		"session_id": sessionID,
		"expires_at": time.Now().Add(h.manager.GetConfig().MaxAge),
	})
}

// SessionStatus 获取session状态
func (h *SessionHandler) SessionStatus(c *gin.Context) {
	sess, exists := session.GetSession(c)
	if !exists {
		c.JSON(http.StatusOK, gin.H{
			"session_exists": false,
			"message":        "No active session",
		})
		return
	}

	sessionID, _ := session.GetSessionID(c)
	timeToExpire := time.Until(sess.ExpiresAt)

	c.JSON(http.StatusOK, gin.H{
		"session_exists":   true,
		"session_id":       sessionID,
		"created_at":       sess.CreatedAt,
		"updated_at":       sess.UpdatedAt,
		"expires_at":       sess.ExpiresAt,
		"time_to_expire":   timeToExpire.String(),
		"data_count":       len(sess.Data),
		"is_expired":       timeToExpire <= 0,
	})
}