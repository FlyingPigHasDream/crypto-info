package service

import (
	"context"
	"crypto-info/internal/model"
	"crypto-info/internal/pkg/mq"
	"encoding/json"
	"fmt"
	"time"

	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/sirupsen/logrus"
)

// MessageService 消息服务
type MessageService struct {
	mqClient *mq.RocketMQClient
	logger   *logrus.Logger
}

// NewMessageService 创建消息服务
func NewMessageService(mqClient *mq.RocketMQClient, logger *logrus.Logger) *MessageService {
	return &MessageService{
		mqClient: mqClient,
		logger:   logger,
	}
}

// 消息主题常量
const (
	TopicPriceUpdate  = "crypto_price_update"
	TopicVolumeUpdate = "crypto_volume_update"
	TopicPriceAlert   = "crypto_price_alert"
	TopicSystemEvent  = "crypto_system_event"
)

// 消息标签常量
const (
	TagPriceChange    = "price_change"
	TagVolumeSpike    = "volume_spike"
	TagPriceAlert     = "price_alert"
	TagSystemStartup  = "system_startup"
	TagSystemShutdown = "system_shutdown"
)

// PriceUpdateMessage 价格更新消息
type PriceUpdateMessage struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Change    float64 `json:"change"`
	Timestamp int64   `json:"timestamp"`
	Source    string  `json:"source"`
}

// VolumeUpdateMessage 交易量更新消息
type VolumeUpdateMessage struct {
	Symbol     string  `json:"symbol"`
	Volume     float64 `json:"volume"`
	VolumeUSD  float64 `json:"volume_usd"`
	Change24h  float64 `json:"change_24h"`
	Timestamp  int64   `json:"timestamp"`
	Source     string  `json:"source"`
}

// PriceAlertMessage 价格警报消息
type PriceAlertMessage struct {
	Symbol      string  `json:"symbol"`
	CurrentPrice float64 `json:"current_price"`
	TargetPrice  float64 `json:"target_price"`
	AlertType    string  `json:"alert_type"` // "above", "below"
	Timestamp    int64   `json:"timestamp"`
	UserID       string  `json:"user_id,omitempty"`
}

// SystemEventMessage 系统事件消息
type SystemEventMessage struct {
	EventType   string                 `json:"event_type"`
	Message     string                 `json:"message"`
	Timestamp   int64                  `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Start 启动消息服务
func (s *MessageService) Start() error {
	if s.mqClient == nil {
		s.logger.Warn("RocketMQ client is not available, message service will not start")
		return nil
	}

	// 启动MQ客户端
	if err := s.mqClient.Start(); err != nil {
		return fmt.Errorf("failed to start MQ client: %w", err)
	}

	// 订阅消息主题
	if err := s.subscribeTopics(); err != nil {
		return fmt.Errorf("failed to subscribe topics: %w", err)
	}

	s.logger.Info("Message service started successfully")
	return nil
}

// Stop 停止消息服务
func (s *MessageService) Stop() error {
	if s.mqClient == nil {
		return nil
	}

	if err := s.mqClient.Stop(); err != nil {
		s.logger.Errorf("Failed to stop MQ client: %v", err)
		return err
	}

	s.logger.Info("Message service stopped")
	return nil
}

// subscribeTopics 订阅消息主题
func (s *MessageService) subscribeTopics() error {
	// 订阅价格更新消息
	if err := s.mqClient.Subscribe(TopicPriceUpdate, "*", s.handlePriceUpdate); err != nil {
		return fmt.Errorf("failed to subscribe price update topic: %w", err)
	}

	// 订阅交易量更新消息
	if err := s.mqClient.Subscribe(TopicVolumeUpdate, "*", s.handleVolumeUpdate); err != nil {
		return fmt.Errorf("failed to subscribe volume update topic: %w", err)
	}

	// 订阅价格警报消息
	if err := s.mqClient.Subscribe(TopicPriceAlert, "*", s.handlePriceAlert); err != nil {
		return fmt.Errorf("failed to subscribe price alert topic: %w", err)
	}

	// 订阅系统事件消息
	if err := s.mqClient.Subscribe(TopicSystemEvent, "*", s.handleSystemEvent); err != nil {
		return fmt.Errorf("failed to subscribe system event topic: %w", err)
	}

	return nil
}

// PublishPriceUpdate 发布价格更新消息
func (s *MessageService) PublishPriceUpdate(priceResp *model.PriceResponse) error {
	if s.mqClient == nil || !s.mqClient.IsStarted() {
		s.logger.Debug("MQ client not available, skipping price update message")
		return nil
	}

	msg := PriceUpdateMessage{
		Symbol:    priceResp.Symbol,
		Price:     priceResp.Price,
		Change:    0.0, // 暂时设为0，需要从其他地方获取变化率
		Timestamp: time.Now().Unix(),
		Source:    priceResp.Source,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal price update message: %w", err)
	}

	return s.mqClient.SendMessage(TopicPriceUpdate, TagPriceChange, body)
}

// PublishVolumeUpdate 发布交易量更新消息
func (s *MessageService) PublishVolumeUpdate(volumeResp *model.VolumeAnalysisResponse) error {
	if s.mqClient == nil || !s.mqClient.IsStarted() {
		s.logger.Debug("MQ client not available, skipping volume update message")
		return nil
	}

	msg := VolumeUpdateMessage{
		Symbol:     volumeResp.Symbol,
		Volume:     volumeResp.AvgVolume,
		VolumeUSD:  0.0, // 暂时设为0，需要从其他地方获取USD价值
		Change24h:  volumeResp.Volatility,
		Timestamp:  time.Now().Unix(),
		Source:     volumeResp.Source,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal volume update message: %w", err)
	}

	return s.mqClient.SendMessage(TopicVolumeUpdate, TagVolumeSpike, body)
}

// PublishPriceAlert 发布价格警报消息
func (s *MessageService) PublishPriceAlert(symbol string, currentPrice, targetPrice float64, alertType, userID string) error {
	if s.mqClient == nil || !s.mqClient.IsStarted() {
		s.logger.Debug("MQ client not available, skipping price alert message")
		return nil
	}

	msg := PriceAlertMessage{
		Symbol:       symbol,
		CurrentPrice: currentPrice,
		TargetPrice:  targetPrice,
		AlertType:    alertType,
		Timestamp:    time.Now().Unix(),
		UserID:       userID,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal price alert message: %w", err)
	}

	return s.mqClient.SendMessage(TopicPriceAlert, TagPriceAlert, body)
}

// PublishSystemEvent 发布系统事件消息
func (s *MessageService) PublishSystemEvent(eventType, message string, metadata map[string]interface{}) error {
	if s.mqClient == nil || !s.mqClient.IsStarted() {
		s.logger.Debug("MQ client not available, skipping system event message")
		return nil
	}

	msg := SystemEventMessage{
		EventType: eventType,
		Message:   message,
		Timestamp: time.Now().Unix(),
		Metadata:  metadata,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal system event message: %w", err)
	}

	tag := TagSystemStartup
	if eventType == "shutdown" {
		tag = TagSystemShutdown
	}

	return s.mqClient.SendMessage(TopicSystemEvent, tag, body)
}

// handlePriceUpdate 处理价格更新消息
func (s *MessageService) handlePriceUpdate(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	for _, msg := range msgs {
		var priceMsg PriceUpdateMessage
		if err := json.Unmarshal(msg.Body, &priceMsg); err != nil {
			s.logger.Errorf("Failed to unmarshal price update message: %v", err)
			continue
		}

		s.logger.Infof("Received price update: %s = $%.2f (%.2f%%)", 
			priceMsg.Symbol, priceMsg.Price, priceMsg.Change)

		// 这里可以添加价格更新的业务逻辑
		// 例如：更新缓存、触发警报、记录历史等
	}

	return consumer.ConsumeSuccess, nil
}

// handleVolumeUpdate 处理交易量更新消息
func (s *MessageService) handleVolumeUpdate(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	for _, msg := range msgs {
		var volumeMsg VolumeUpdateMessage
		if err := json.Unmarshal(msg.Body, &volumeMsg); err != nil {
			s.logger.Errorf("Failed to unmarshal volume update message: %v", err)
			continue
		}

		s.logger.Infof("Received volume update: %s volume = %.2f (%.2f%% change)", 
			volumeMsg.Symbol, volumeMsg.Volume, volumeMsg.Change24h)

		// 这里可以添加交易量更新的业务逻辑
	}

	return consumer.ConsumeSuccess, nil
}

// handlePriceAlert 处理价格警报消息
func (s *MessageService) handlePriceAlert(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	for _, msg := range msgs {
		var alertMsg PriceAlertMessage
		if err := json.Unmarshal(msg.Body, &alertMsg); err != nil {
			s.logger.Errorf("Failed to unmarshal price alert message: %v", err)
			continue
		}

		s.logger.Warnf("Price alert triggered: %s current price $%.2f %s target $%.2f", 
			alertMsg.Symbol, alertMsg.CurrentPrice, alertMsg.AlertType, alertMsg.TargetPrice)

		// 这里可以添加价格警报的业务逻辑
		// 例如：发送通知、邮件、短信等
	}

	return consumer.ConsumeSuccess, nil
}

// handleSystemEvent 处理系统事件消息
func (s *MessageService) handleSystemEvent(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	for _, msg := range msgs {
		var eventMsg SystemEventMessage
		if err := json.Unmarshal(msg.Body, &eventMsg); err != nil {
			s.logger.Errorf("Failed to unmarshal system event message: %v", err)
			continue
		}

		s.logger.Infof("System event: %s - %s", eventMsg.EventType, eventMsg.Message)

		// 这里可以添加系统事件的业务逻辑
		// 例如：记录审计日志、监控告警等
	}

	return consumer.ConsumeSuccess, nil
}