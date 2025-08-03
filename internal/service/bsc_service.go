package service

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"crypto-info/internal/config"
	"crypto-info/internal/model"
	"crypto-info/internal/pkg/database"
	"crypto-info/internal/pkg/logger"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
)

// BSCService BSC链上数据监控服务接口
type BSCService interface {
	// 启动监控
	Start(ctx context.Context) error
	// 停止监控
	Stop() error
	// 获取监控状态
	GetStatus() *model.BSCMonitoringResponse
	// 获取最新区块信息
	GetLatestBlock(ctx context.Context) (*model.BSCBlock, error)
	// 获取交易信息
	GetTransactions(ctx context.Context, blockNumber *big.Int, page, pageSize int) (*model.BSCTransactionResponse, error)
	// 获取代币转账记录
	GetTokenTransfers(ctx context.Context, tokenAddress common.Address, page, pageSize int) (*model.BSCTokenTransferResponse, error)
	// 获取交换事件
	GetSwapEvents(ctx context.Context, pairAddress common.Address, page, pageSize int) (*model.BSCSwapEventResponse, error)
	// 获取交易对信息
	GetPairInfo(ctx context.Context, pairAddress common.Address) (*model.BSCPairInfo, error)
	// 通过流动性池计算代币价格
	GetTokenPriceFromLiquidity(ctx context.Context, tokenAddress common.Address) (decimal.Decimal, error)
	// 获取代币对USDT的价格
	GetTokenPriceInUSDT(ctx context.Context, tokenSymbol string) (decimal.Decimal, error)
}

// bscService BSC服务实现
type bscService struct {
	client      *ethclient.Client
	wsClient    *ethclient.Client
	config      *config.BSC
	redisClient database.RedisClient
	logger      logger.Logger
	stats       *model.BSCMonitoringStats
	statsMutex  sync.RWMutex
	running     bool
	runMutex    sync.RWMutex
	cancel      context.CancelFunc
}

// NewBSCService 创建BSC服务
func NewBSCService(cfg *config.Config, redisClient database.RedisClient) (BSCService, error) {
	if !cfg.BSC.Enabled {
		return &bscService{
			config: &cfg.BSC,
			logger: logger.GetLogger(),
			stats: &model.BSCMonitoringStats{
				Status: "disabled",
			},
		}, nil
	}

	// 连接BSC节点
	client, err := ethclient.Dial(cfg.BSC.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to BSC RPC: %w", err)
	}

	// 连接WebSocket（可选）
	var wsClient *ethclient.Client
	if cfg.BSC.WebSocketURL != "" {
		wsClient, err = ethclient.Dial(cfg.BSC.WebSocketURL)
		if err != nil {
			logger.GetLogger().Warnf("Failed to connect to BSC WebSocket: %v", err)
		}
	}

	return &bscService{
		client:      client,
		wsClient:    wsClient,
		config:      &cfg.BSC,
		redisClient: redisClient,
		logger:      logger.GetLogger(),
		stats: &model.BSCMonitoringStats{
			StartTime: time.Now(),
			Status:    "initialized",
		},
	}, nil
}

// Start 启动BSC监控
func (s *bscService) Start(ctx context.Context) error {
	if !s.config.Enabled || !s.config.Monitoring.Enabled {
		return fmt.Errorf("BSC monitoring is disabled")
	}

	s.runMutex.Lock()
	defer s.runMutex.Unlock()

	if s.running {
		return fmt.Errorf("BSC monitoring is already running")
	}

	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.running = true

	s.updateStats(func(stats *model.BSCMonitoringStats) {
		stats.Status = "running"
		stats.StartTime = time.Now()
	})

	s.logger.Info("Starting BSC monitoring service")

	// 启动区块监控
	go s.monitorBlocks(ctx)

	// 如果有WebSocket连接，启动实时事件监控
	if s.wsClient != nil {
		go s.monitorEvents(ctx)
	}

	return nil
}

// Stop 停止BSC监控
func (s *bscService) Stop() error {
	s.runMutex.Lock()
	defer s.runMutex.Unlock()

	if !s.running {
		return fmt.Errorf("BSC monitoring is not running")
	}

	if s.cancel != nil {
		s.cancel()
	}

	s.running = false
	s.updateStats(func(stats *model.BSCMonitoringStats) {
		stats.Status = "stopped"
	})

	s.logger.Info("BSC monitoring service stopped")
	return nil
}

// GetStatus 获取监控状态
func (s *bscService) GetStatus() *model.BSCMonitoringResponse {
	s.statsMutex.RLock()
	defer s.statsMutex.RUnlock()

	return &model.BSCMonitoringResponse{
		Enabled: s.config.Enabled,
		Stats:   *s.stats,
		Message: "BSC monitoring service status",
	}
}

// GetLatestBlock 获取最新区块信息
func (s *bscService) GetLatestBlock(ctx context.Context) (*model.BSCBlock, error) {
	if s.client == nil {
		return nil, fmt.Errorf("BSC client not initialized")
	}

	header, err := s.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block header: %w", err)
	}

	block, err := s.client.BlockByNumber(ctx, header.Number)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block: %w", err)
	}

	return &model.BSCBlock{
		Number:       block.Number(),
		Hash:         block.Hash(),
		ParentHash:   block.ParentHash(),
		Timestamp:    block.Time(),
		GasUsed:      block.GasUsed(),
		GasLimit:     block.GasLimit(),
		Transactions: len(block.Transactions()),
		Miner:        block.Coinbase(),
	}, nil
}

// GetTransactions 获取交易信息
func (s *bscService) GetTransactions(ctx context.Context, blockNumber *big.Int, page, pageSize int) (*model.BSCTransactionResponse, error) {
	if s.client == nil {
		return nil, fmt.Errorf("BSC client not initialized")
	}

	block, err := s.client.BlockByNumber(ctx, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	txs := block.Transactions()
	total := len(txs)

	// 分页处理
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= total {
		return &model.BSCTransactionResponse{
			Transactions: []model.BSCTransaction{},
			Total:        total,
			Page:         page,
			PageSize:     pageSize,
		}, nil
	}
	if end > total {
		end = total
	}

	var transactions []model.BSCTransaction
	for i := start; i < end; i++ {
		tx := txs[i]
		receipt, err := s.client.TransactionReceipt(ctx, tx.Hash())
		if err != nil {
			continue
		}

		transactions = append(transactions, model.BSCTransaction{
			Hash:        tx.Hash(),
			BlockNumber: block.Number(),
			From:        s.getTransactionSender(tx),
			To:          tx.To(),
			Value:       tx.Value(),
			GasPrice:    tx.GasPrice(),
			GasUsed:     receipt.GasUsed,
			Status:      receipt.Status,
			Timestamp:   time.Unix(int64(block.Time()), 0),
		})
	}

	return &model.BSCTransactionResponse{
		Transactions: transactions,
		Total:        total,
		Page:         page,
		PageSize:     pageSize,
	}, nil
}

// GetTokenTransfers 获取代币转账记录
func (s *bscService) GetTokenTransfers(ctx context.Context, tokenAddress common.Address, page, pageSize int) (*model.BSCTokenTransferResponse, error) {
	// 这里应该从缓存或数据库中获取代币转账记录
	// 为了演示，返回空结果
	return &model.BSCTokenTransferResponse{
		Transfers: []model.BSCTokenTransfer{},
		Total:     0,
		Page:      page,
		PageSize:  pageSize,
	}, nil
}

// GetSwapEvents 获取交换事件
func (s *bscService) GetSwapEvents(ctx context.Context, pairAddress common.Address, page, pageSize int) (*model.BSCSwapEventResponse, error) {
	// 这里应该从缓存或数据库中获取交换事件
	// 为了演示，返回空结果
	return &model.BSCSwapEventResponse{
		Swaps:    []model.BSCSwapEvent{},
		Total:    0,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetPairInfo 获取交易对信息
func (s *bscService) GetPairInfo(ctx context.Context, pairAddress common.Address) (*model.BSCPairInfo, error) {
	// 这里应该调用PancakeSwap合约获取交易对信息
	// 为了演示，返回模拟数据
	return &model.BSCPairInfo{
		Address:     pairAddress,
		Token0:      common.HexToAddress(s.config.Contracts.WBNB),
		Token1:      common.HexToAddress(s.config.Contracts.USDT),
		Reserve0:    big.NewInt(1000000),
		Reserve1:    big.NewInt(300000000),
		TotalSupply: big.NewInt(17320508),
		Price0:      decimal.NewFromFloat(300.0),
		Price1:      decimal.NewFromFloat(0.00333),
		UpdatedAt:   time.Now(),
	}, nil
}

// monitorBlocks 监控区块
func (s *bscService) monitorBlocks(ctx context.Context) {
	ticker := time.NewTicker(s.config.Monitoring.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.processLatestBlocks(ctx); err != nil {
				s.logger.Errorf("Failed to process latest blocks: %v", err)
			}
		}
	}
}

// monitorEvents 监控事件
func (s *bscService) monitorEvents(ctx context.Context) {
	// 实现WebSocket事件监控
	s.logger.Info("Starting real-time event monitoring")
	
	// 这里可以实现具体的事件监控逻辑
	// 例如监控Transfer、Swap等事件
}

// processLatestBlocks 处理最新区块
func (s *bscService) processLatestBlocks(ctx context.Context) error {
	header, err := s.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to get latest block header: %w", err)
	}

	s.updateStats(func(stats *model.BSCMonitoringStats) {
		stats.LatestBlock = header.Number
		stats.ProcessedBlocks++
		stats.LastUpdateTime = time.Now()
	})

	return nil
}

// updateStats 更新统计信息
func (s *bscService) updateStats(updateFunc func(*model.BSCMonitoringStats)) {
	s.statsMutex.Lock()
	defer s.statsMutex.Unlock()
	updateFunc(s.stats)
}

// getTransactionSender 获取交易发送者
func (s *bscService) getTransactionSender(tx *types.Transaction) common.Address {
	// 这里需要实现交易签名验证来获取发送者地址
	// 为了简化，返回零地址
	return common.Address{}
}

// ERC20 ABI for token operations
const erc20ABI = `[
	{
		"constant": true,
		"inputs": [],
		"name": "name",
		"outputs": [{"name": "", "type": "string"}],
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "symbol",
		"outputs": [{"name": "", "type": "string"}],
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "decimals",
		"outputs": [{"name": "", "type": "uint8"}],
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "totalSupply",
		"outputs": [{"name": "", "type": "uint256"}],
		"type": "function"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "from", "type": "address"},
			{"indexed": true, "name": "to", "type": "address"},
			{"indexed": false, "name": "value", "type": "uint256"}
		],
		"name": "Transfer",
		"type": "event"
	}
]`

// getERC20ABI 获取ERC20 ABI
func getERC20ABI() (abi.ABI, error) {
	return abi.JSON(strings.NewReader(erc20ABI))
}

// GetTokenPriceFromLiquidity 通过流动性池计算代币价格
func (s *bscService) GetTokenPriceFromLiquidity(ctx context.Context, tokenAddress common.Address) (decimal.Decimal, error) {
	if s.client == nil {
		return decimal.Zero, fmt.Errorf("BSC client not initialized")
	}

	// 查找代币对应的流动性池
	// 这里假设使用PancakeSwap V2的工厂合约来查找交易对
	// 实际实现中需要调用PancakeSwap Factory合约的getPair方法
	
	// 为演示目的，返回模拟价格计算
	// 实际应该通过以下步骤：
	// 1. 调用PancakeSwap Factory合约获取token/USDT交易对地址
	// 2. 调用交易对合约获取储备量(reserves)
	// 3. 根据储备量计算价格: price = reserve_usdt / reserve_token
	
	// 模拟价格数据
	prices := map[string]float64{
		"0x55d398326f99059fF775485246999027B3197955": 1.0,    // USDT
		"0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c": 300.0,  // WBNB
		"0x2170Ed0880ac9A755fd29B2688956BD959F933F8": 3000.0, // ETH
		"0x7130d2A12B9BCbFAe4f2634d864A1Ee1Ce3Ead9c": 45000.0, // BTCB
	}
	
	tokenAddressStr := tokenAddress.Hex()
	if price, exists := prices[tokenAddressStr]; exists {
		return decimal.NewFromFloat(price), nil
	}
	
	// 默认返回1.0作为未知代币的价格
	return decimal.NewFromFloat(1.0), nil
}

// GetTokenPriceInUSDT 获取代币对USDT的价格
func (s *bscService) GetTokenPriceInUSDT(ctx context.Context, tokenSymbol string) (decimal.Decimal, error) {
	if s.client == nil {
		return decimal.Zero, fmt.Errorf("BSC client not initialized")
	}

	// 代币符号到合约地址的映射
	tokenAddresses := map[string]string{
		"USDT": "0x55d398326f99059fF775485246999027B3197955",
		"BNB":  "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c", // WBNB
		"ETH":  "0x2170Ed0880ac9A755fd29B2688956BD959F933F8",
		"BTC":  "0x7130d2A12B9BCbFAe4f2634d864A1Ee1Ce3Ead9c", // BTCB
		"ADA":  "0x3EE2200Efb3400fAbB9AacF31297cBdD1d435D47",
		"DOT":  "0x7083609fCE4d1d8Dc0C979AAb8c869Ea2C873402",
		"LINK": "0xF8A0BF9cF54Bb92F17374d9e9A321E6a111a51bD",
		"LTC":  "0x4338665CBB7B2485A8855A139b75D5e34AB0DB94",
		"BCH":  "0x8fF795a6F4D97E7887C79beA79aba5cc76444aDf",
	}

	addressStr, exists := tokenAddresses[tokenSymbol]
	if !exists {
		return decimal.Zero, fmt.Errorf("unsupported token symbol: %s", tokenSymbol)
	}

	tokenAddress := common.HexToAddress(addressStr)
	return s.GetTokenPriceFromLiquidity(ctx, tokenAddress)
}