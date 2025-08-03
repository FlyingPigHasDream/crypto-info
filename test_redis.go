package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"crypto-info/internal/config"
	"crypto-info/internal/pkg/database"
	"crypto-info/internal/pkg/logger"
)

func main() {
	// 初始化配置
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logger.Init(&cfg.Log)
	appLogger := logger.GetLogger()

	// 初始化Redis客户端
	redisClient, err := database.NewRedisClient(&cfg.Database.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	ctx := context.Background()

	// 测试基本操作
	fmt.Println("=== Redis连接测试 ===")

	// 测试Ping
	if err := redisClient.Ping(ctx); err != nil {
		fmt.Printf("❌ Ping失败: %v\n", err)
	} else {
		fmt.Println("✅ Ping成功")
	}

	// 测试Set操作
	testKey := "crypto:test:" + fmt.Sprintf("%d", time.Now().Unix())
	testValue := "Hello from crypto-info service!"
	if err := redisClient.Set(ctx, testKey, testValue, 60*time.Second); err != nil {
		fmt.Printf("❌ Set操作失败: %v\n", err)
	} else {
		fmt.Printf("✅ Set操作成功: %s = %s\n", testKey, testValue)
	}

	// 测试Get操作
	if value, err := redisClient.Get(ctx, testKey); err != nil {
		fmt.Printf("❌ Get操作失败: %v\n", err)
	} else {
		fmt.Printf("✅ Get操作成功: %s = %s\n", testKey, value)
	}

	// 测试TTL
	if ttl, err := redisClient.TTL(ctx, testKey); err != nil {
		fmt.Printf("❌ TTL查询失败: %v\n", err)
	} else {
		fmt.Printf("✅ TTL查询成功: %s TTL = %v\n", testKey, ttl)
	}

	// 测试Hash操作
	hashKey := "crypto:hash:test"
	if err := redisClient.HSet(ctx, hashKey, "field1", "value1", "field2", "value2"); err != nil {
		fmt.Printf("❌ HSet操作失败: %v\n", err)
	} else {
		fmt.Println("✅ HSet操作成功")
	}

	if value, err := redisClient.HGet(ctx, hashKey, "field1"); err != nil {
		fmt.Printf("❌ HGet操作失败: %v\n", err)
	} else {
		fmt.Printf("✅ HGet操作成功: %s.field1 = %s\n", hashKey, value)
	}

	// 清理测试数据
	if err := redisClient.Del(ctx, testKey, hashKey); err != nil {
		fmt.Printf("❌ 清理测试数据失败: %v\n", err)
	} else {
		fmt.Println("✅ 测试数据清理成功")
	}

	fmt.Println("\n=== Redis连接测试完成 ===")
	appLogger.Info("Redis connection test completed successfully")
}