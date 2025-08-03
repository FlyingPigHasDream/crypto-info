package session

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"crypto-info/internal/config"
	"crypto-info/internal/pkg/logger"

	"github.com/redis/go-redis/v9"
)

// RedisStore Redis会话存储
type RedisStore struct {
	client *redis.Client
	config *config.SessionConfig
	logger logger.Logger
	prefix string
}

// NewRedisStore 创建Redis存储
func NewRedisStore(client *redis.Client, cfg *config.SessionConfig, log logger.Logger) *RedisStore {
	return &RedisStore{
		client: client,
		config: cfg,
		logger: log,
		prefix: "session:",
	}
}

// Get 获取会话
func (r *RedisStore) Get(ctx context.Context, sessionID string) (*Session, error) {
	key := r.getKey(sessionID)
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session from redis: %w", err)
	}

	var session Session
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	return &session, nil
}

// Set 设置会话
func (r *RedisStore) Set(ctx context.Context, session *Session) error {
	key := r.getKey(session.ID)
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		ttl = r.config.MaxAge
	}

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set session in redis: %w", err)
	}

	r.logger.Debugf("Session %s saved to Redis with TTL %v", session.ID, ttl)
	return nil
}

// Delete 删除会话
func (r *RedisStore) Delete(ctx context.Context, sessionID string) error {
	key := r.getKey(sessionID)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete session from redis: %w", err)
	}

	r.logger.Debugf("Session %s deleted from Redis", sessionID)
	return nil
}

// Exists 检查会话是否存在
func (r *RedisStore) Exists(ctx context.Context, sessionID string) (bool, error) {
	key := r.getKey(sessionID)
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check session existence in redis: %w", err)
	}

	return count > 0, nil
}

// Cleanup 清理过期会话
func (r *RedisStore) Cleanup(ctx context.Context) error {
	// Redis会自动清理过期的key，这里不需要手动清理
	r.logger.Debug("Redis store cleanup completed (automatic expiration)")
	return nil
}

// Close 关闭存储
func (r *RedisStore) Close() error {
	// Redis客户端由外部管理，这里不关闭
	return nil
}

// getKey 获取Redis key
func (r *RedisStore) getKey(sessionID string) string {
	return r.prefix + sessionID
}