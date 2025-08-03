package session

import (
	"context"
	"errors"
	"sync"
	"time"

	"crypto-info/internal/config"
	"crypto-info/internal/pkg/logger"
)

// MemoryStore 内存会话存储
type MemoryStore struct {
	data   map[string]*Session
	mutex  sync.RWMutex
	config *config.SessionConfig
	logger logger.Logger
}

// NewMemoryStore 创建内存存储
func NewMemoryStore(cfg *config.SessionConfig, log logger.Logger) *MemoryStore {
	store := &MemoryStore{
		data:   make(map[string]*Session),
		config: cfg,
		logger: log,
	}

	// 启动清理协程
	go store.startCleanupRoutine()

	return store
}

// Get 获取会话
func (m *MemoryStore) Get(ctx context.Context, sessionID string) (*Session, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	session, exists := m.data[sessionID]
	if !exists {
		return nil, errors.New("session not found")
	}

	// 检查是否过期
	if session.ExpiresAt.Before(time.Now()) {
		m.mutex.RUnlock()
		m.mutex.Lock()
		delete(m.data, sessionID)
		m.mutex.Unlock()
		m.mutex.RLock()
		return nil, errors.New("session expired")
	}

	// 返回副本以避免并发修改
	sessionCopy := *session
	sessionCopy.Data = make(map[string]interface{})
	for k, v := range session.Data {
		sessionCopy.Data[k] = v
	}

	return &sessionCopy, nil
}

// Set 设置会话
func (m *MemoryStore) Set(ctx context.Context, session *Session) error {
	if session == nil {
		return errors.New("session is nil")
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 创建副本以避免外部修改
	sessionCopy := *session
	sessionCopy.Data = make(map[string]interface{})
	for k, v := range session.Data {
		sessionCopy.Data[k] = v
	}

	m.data[session.ID] = &sessionCopy
	m.logger.Debugf("Session %s saved to memory store", session.ID)
	return nil
}

// Delete 删除会话
func (m *MemoryStore) Delete(ctx context.Context, sessionID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.data, sessionID)
	m.logger.Debugf("Session %s deleted from memory store", sessionID)
	return nil
}

// Exists 检查会话是否存在
func (m *MemoryStore) Exists(ctx context.Context, sessionID string) (bool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	session, exists := m.data[sessionID]
	if !exists {
		return false, nil
	}

	// 检查是否过期
	if session.ExpiresAt.Before(time.Now()) {
		m.mutex.RUnlock()
		m.mutex.Lock()
		delete(m.data, sessionID)
		m.mutex.Unlock()
		return false, nil
	}

	return true, nil
}

// Cleanup 清理过期会话
func (m *MemoryStore) Cleanup(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	expiredSessions := make([]string, 0)

	for sessionID, session := range m.data {
		if session.ExpiresAt.Before(now) {
			expiredSessions = append(expiredSessions, sessionID)
		}
	}

	for _, sessionID := range expiredSessions {
		delete(m.data, sessionID)
	}

	if len(expiredSessions) > 0 {
		m.logger.Debugf("Cleaned up %d expired sessions from memory store", len(expiredSessions))
	}

	return nil
}

// Close 关闭存储
func (m *MemoryStore) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.data = nil
	m.logger.Debug("Memory store closed")
	return nil
}

// startCleanupRoutine 启动清理协程
func (m *MemoryStore) startCleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute) // 每5分钟清理一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := m.Cleanup(context.Background()); err != nil {
				m.logger.Errorf("Failed to cleanup expired sessions: %v", err)
			}
		}
	}
}