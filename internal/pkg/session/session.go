package session

import (
	"context"
	"errors"
	"time"

	"crypto-info/internal/config"
	"crypto-info/internal/pkg/logger"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Session 会话数据结构
type Session struct {
	ID        string                 `json:"id"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	ExpiresAt time.Time              `json:"expires_at"`
}

// Store Session存储接口
type Store interface {
	// Get 获取会话
	Get(ctx context.Context, sessionID string) (*Session, error)
	// Set 设置会话
	Set(ctx context.Context, session *Session) error
	// Delete 删除会话
	Delete(ctx context.Context, sessionID string) error
	// Exists 检查会话是否存在
	Exists(ctx context.Context, sessionID string) (bool, error)
	// Cleanup 清理过期会话
	Cleanup(ctx context.Context) error
	// Close 关闭存储
	Close() error
}

// Manager Session管理器
type Manager struct {
	store  Store
	config *config.SessionConfig
	logger logger.Logger
}

// NewManager 创建Session管理器
func NewManager(cfg *config.SessionConfig, redisClient *redis.Client, log logger.Logger) (*Manager, error) {
	if cfg == nil {
		return nil, errors.New("session config is required")
	}

	var store Store

	switch cfg.Store {
	case "redis":
		if redisClient == nil {
			return nil, errors.New("redis client is required for redis store")
		}
		store = NewRedisStore(redisClient, cfg, log)
	case "memory":
		store = NewMemoryStore(cfg, log)
	default:
		return nil, errors.New("unsupported session store type: " + cfg.Store)
	}

	return &Manager{
		store:  store,
		config: cfg,
		logger: log,
	}, nil
}

// CreateSession 创建新会话
func (m *Manager) CreateSession(ctx context.Context) (*Session, error) {
	sessionID := uuid.New().String()
	now := time.Now()

	session := &Session{
		ID:        sessionID,
		Data:      make(map[string]interface{}),
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(m.config.MaxAge),
	}

	if err := m.store.Set(ctx, session); err != nil {
		m.logger.Errorf("Failed to create session: %v", err)
		return nil, err
	}

	m.logger.Debugf("Created new session: %s", sessionID)
	return session, nil
}

// GetSession 获取会话
func (m *Manager) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	if sessionID == "" {
		return nil, errors.New("session ID is required")
	}

	session, err := m.store.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// 检查会话是否过期
	if session.ExpiresAt.Before(time.Now()) {
		_ = m.store.Delete(ctx, sessionID)
		return nil, errors.New("session expired")
	}

	return session, nil
}

// SaveSession 保存会话
func (m *Manager) SaveSession(ctx context.Context, session *Session) error {
	if session == nil {
		return errors.New("session is required")
	}

	session.UpdatedAt = time.Now()
	return m.store.Set(ctx, session)
}

// DeleteSession 删除会话
func (m *Manager) DeleteSession(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return errors.New("session ID is required")
	}

	return m.store.Delete(ctx, sessionID)
}

// RefreshSession 刷新会话过期时间
func (m *Manager) RefreshSession(ctx context.Context, sessionID string) error {
	session, err := m.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	session.ExpiresAt = time.Now().Add(m.config.MaxAge)
	return m.SaveSession(ctx, session)
}

// SetSessionData 设置会话数据
func (m *Manager) SetSessionData(ctx context.Context, sessionID string, key string, value interface{}) error {
	session, err := m.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	session.Data[key] = value
	return m.SaveSession(ctx, session)
}

// GetSessionData 获取会话数据
func (m *Manager) GetSessionData(ctx context.Context, sessionID string, key string) (interface{}, error) {
	session, err := m.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	value, exists := session.Data[key]
	if !exists {
		return nil, errors.New("session data not found")
	}

	return value, nil
}

// RemoveSessionData 移除会话数据
func (m *Manager) RemoveSessionData(ctx context.Context, sessionID string, key string) error {
	session, err := m.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	delete(session.Data, key)
	return m.SaveSession(ctx, session)
}

// Cleanup 清理过期会话
func (m *Manager) Cleanup(ctx context.Context) error {
	return m.store.Cleanup(ctx)
}

// Close 关闭Session管理器
func (m *Manager) Close() error {
	if m.store != nil {
		return m.store.Close()
	}
	return nil
}

// GetConfig 获取Session配置
func (m *Manager) GetConfig() *config.SessionConfig {
	return m.config
}