package database

import (
	"context"
	"fmt"
	"time"

	"crypto-info/internal/config"
	"crypto-info/internal/pkg/logger"

	"github.com/redis/go-redis/v9"
)

// RedisClient Redis客户端接口
type RedisClient interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Del(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, keys ...string) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	HGet(ctx context.Context, key, field string) (string, error)
	HSet(ctx context.Context, key string, values ...interface{}) error
	HDel(ctx context.Context, key string, fields ...string) error
	HExists(ctx context.Context, key, field string) (bool, error)
	GetClient() *redis.Client
	Close() error
	Ping(ctx context.Context) error
}

// redisClient Redis客户端实现
type redisClient struct {
	client *redis.Client
	logger logger.Logger
}

// NewRedisClient 创建Redis客户端
func NewRedisClient(cfg *config.RedisConfig) (RedisClient, error) {
	log := logger.GetLogger()

	// 创建Redis客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		PoolTimeout:     cfg.PoolTimeout,
		ConnMaxIdleTime: cfg.IdleTimeout,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	log.Info("Redis connected successfully")

	return &redisClient{
		client: rdb,
		logger: log,
	}, nil
}

// Get 获取值
func (r *redisClient) Get(ctx context.Context, key string) (string, error) {
	result, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		r.logger.Errorf("Redis GET error for key %s: %v", key, err)
		return "", err
	}
	return result, nil
}

// Set 设置值
func (r *redisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	err := r.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		r.logger.Errorf("Redis SET error for key %s: %v", key, err)
		return err
	}
	return nil
}

// Del 删除键
func (r *redisClient) Del(ctx context.Context, keys ...string) error {
	err := r.client.Del(ctx, keys...).Err()
	if err != nil {
		r.logger.Errorf("Redis DEL error for keys %v: %v", keys, err)
		return err
	}
	return nil
}

// Exists 检查键是否存在
func (r *redisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	result, err := r.client.Exists(ctx, keys...).Result()
	if err != nil {
		r.logger.Errorf("Redis EXISTS error for keys %v: %v", keys, err)
		return 0, err
	}
	return result, nil
}

// Expire 设置过期时间
func (r *redisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	err := r.client.Expire(ctx, key, expiration).Err()
	if err != nil {
		r.logger.Errorf("Redis EXPIRE error for key %s: %v", key, err)
		return err
	}
	return nil
}

// TTL 获取过期时间
func (r *redisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	result, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		r.logger.Errorf("Redis TTL error for key %s: %v", key, err)
		return 0, err
	}
	return result, nil
}

// HGet 获取哈希字段值
func (r *redisClient) HGet(ctx context.Context, key, field string) (string, error) {
	result, err := r.client.HGet(ctx, key, field).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		r.logger.Errorf("Redis HGET error for key %s field %s: %v", key, field, err)
		return "", err
	}
	return result, nil
}

// HSet 设置哈希字段值
func (r *redisClient) HSet(ctx context.Context, key string, values ...interface{}) error {
	err := r.client.HSet(ctx, key, values...).Err()
	if err != nil {
		r.logger.Errorf("Redis HSET error for key %s: %v", key, err)
		return err
	}
	return nil
}

// HDel 删除哈希字段
func (r *redisClient) HDel(ctx context.Context, key string, fields ...string) error {
	err := r.client.HDel(ctx, key, fields...).Err()
	if err != nil {
		r.logger.Errorf("Redis HDEL error for key %s fields %v: %v", key, fields, err)
		return err
	}
	return nil
}

// HExists 检查哈希字段是否存在
func (r *redisClient) HExists(ctx context.Context, key, field string) (bool, error) {
	result, err := r.client.HExists(ctx, key, field).Result()
	if err != nil {
		r.logger.Errorf("Redis HEXISTS error for key %s field %s: %v", key, field, err)
		return false, err
	}
	return result, nil
}

// Close 关闭连接
func (r *redisClient) Close() error {
	err := r.client.Close()
	if err != nil {
		r.logger.Errorf("Redis close error: %v", err)
		return err
	}
	r.logger.Info("Redis connection closed")
	return nil
}

// GetClient 获取底层Redis客户端
func (r *redisClient) GetClient() *redis.Client {
	return r.client
}

// Ping 测试连接
func (r *redisClient) Ping(ctx context.Context) error {
	err := r.client.Ping(ctx).Err()
	if err != nil {
		r.logger.Errorf("Redis ping error: %v", err)
		return err
	}
	return nil
}