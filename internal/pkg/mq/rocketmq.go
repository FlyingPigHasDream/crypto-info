package mq

import (
	"context"
	"crypto-info/internal/config"
	"fmt"
	"sync"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/sirupsen/logrus"
)

// RocketMQClient RocketMQ客户端
type RocketMQClient struct {
	config   *config.RocketMQ
	producer rocketmq.Producer
	consumer rocketmq.PushConsumer
	logger   *logrus.Logger
	mu       sync.RWMutex
	started  bool
}

// NewRocketMQClient 创建RocketMQ客户端
func NewRocketMQClient(cfg *config.RocketMQ, logger *logrus.Logger) (*RocketMQClient, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("RocketMQ is disabled")
	}

	client := &RocketMQClient{
		config: cfg,
		logger: logger,
	}

	// 初始化生产者
	if err := client.initProducer(); err != nil {
		return nil, fmt.Errorf("failed to init producer: %w", err)
	}

	// 初始化消费者
	if err := client.initConsumer(); err != nil {
		return nil, fmt.Errorf("failed to init consumer: %w", err)
	}

	return client, nil
}

// initProducer 初始化生产者
func (c *RocketMQClient) initProducer() error {
	p, err := rocketmq.NewProducer(
		producer.WithNameServer(c.config.NameServers),
		producer.WithGroupName(c.config.Producer.GroupName),
		producer.WithRetry(c.config.Producer.RetryTimes),
		producer.WithSendMsgTimeout(c.config.Producer.SendMsgTimeout),
	)
	if err != nil {
		return fmt.Errorf("failed to create producer: %w", err)
	}

	c.producer = p
	return nil
}

// initConsumer 初始化消费者
func (c *RocketMQClient) initConsumer() error {
	// 解析消费位置
	var consumeFromWhere consumer.ConsumeFromWhere
	switch c.config.Consumer.ConsumeFromWhere {
	case "CONSUME_FROM_LAST_OFFSET":
		consumeFromWhere = consumer.ConsumeFromLastOffset
	case "CONSUME_FROM_FIRST_OFFSET":
		consumeFromWhere = consumer.ConsumeFromFirstOffset
	case "CONSUME_FROM_TIMESTAMP":
		consumeFromWhere = consumer.ConsumeFromTimestamp
	default:
		consumeFromWhere = consumer.ConsumeFromLastOffset
	}

	consumer, err := rocketmq.NewPushConsumer(
		consumer.WithNameServer(c.config.NameServers),
		consumer.WithGroupName(c.config.Consumer.GroupName),
		consumer.WithConsumeFromWhere(consumeFromWhere),
		consumer.WithPullBatchSize(int32(c.config.Consumer.PullBatchSize)),
	)
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}

	c.consumer = consumer
	return nil
}

// Start 启动RocketMQ客户端
func (c *RocketMQClient) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.started {
		return nil
	}

	// 启动生产者
	if err := c.producer.Start(); err != nil {
		return fmt.Errorf("failed to start producer: %w", err)
	}

	// 启动消费者
	if err := c.consumer.Start(); err != nil {
		return fmt.Errorf("failed to start consumer: %w", err)
	}

	c.started = true
	c.logger.Info("RocketMQ client started successfully")
	return nil
}

// Stop 停止RocketMQ客户端
func (c *RocketMQClient) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.started {
		return nil
	}

	// 停止生产者
	if err := c.producer.Shutdown(); err != nil {
		c.logger.Errorf("Failed to shutdown producer: %v", err)
	}

	// 停止消费者
	if err := c.consumer.Shutdown(); err != nil {
		c.logger.Errorf("Failed to shutdown consumer: %v", err)
	}

	c.started = false
	c.logger.Info("RocketMQ client stopped")
	return nil
}

// SendMessage 发送消息
func (c *RocketMQClient) SendMessage(topic, tag string, body []byte) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.started {
		return fmt.Errorf("RocketMQ client is not started")
	}

	msg := &primitive.Message{
		Topic: topic,
		Body:  body,
	}

	if tag != "" {
		msg.WithTag(tag)
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.config.Producer.SendMsgTimeout)
	defer cancel()

	result, err := c.producer.SendSync(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	c.logger.Debugf("Message sent successfully: %+v", result)
	return nil
}

// SendAsyncMessage 异步发送消息
func (c *RocketMQClient) SendAsyncMessage(topic, tag string, body []byte, callback func(context.Context, *primitive.SendResult, error)) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.started {
		return fmt.Errorf("RocketMQ client is not started")
	}

	msg := &primitive.Message{
		Topic: topic,
		Body:  body,
	}

	if tag != "" {
		msg.WithTag(tag)
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.config.Producer.SendMsgTimeout)
	defer cancel()

	return c.producer.SendAsync(ctx, callback, msg)
}

// Subscribe 订阅主题
func (c *RocketMQClient) Subscribe(topic, selector string, handler func(context.Context, ...*primitive.MessageExt) (consumer.ConsumeResult, error)) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.consumer.Subscribe(topic, consumer.MessageSelector{
		Type:       consumer.TAG,
		Expression: selector,
	}, handler)
}

// IsStarted 检查客户端是否已启动
func (c *RocketMQClient) IsStarted() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.started
}

// GetProducer 获取生产者实例
func (c *RocketMQClient) GetProducer() rocketmq.Producer {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.producer
}

// GetConsumer 获取消费者实例
func (c *RocketMQClient) GetConsumer() rocketmq.PushConsumer {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.consumer
}