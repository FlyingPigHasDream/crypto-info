package model

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

// BSCBlock BSC区块信息
type BSCBlock struct {
	Number       *big.Int    `json:"number"`
	Hash         common.Hash `json:"hash"`
	ParentHash   common.Hash `json:"parent_hash"`
	Timestamp    uint64      `json:"timestamp"`
	GasUsed      uint64      `json:"gas_used"`
	GasLimit     uint64      `json:"gas_limit"`
	Transactions int         `json:"transactions"`
	Miner        common.Address `json:"miner"`
}

// BSCTransaction BSC交易信息
type BSCTransaction struct {
	Hash        common.Hash    `json:"hash"`
	BlockNumber *big.Int       `json:"block_number"`
	From        common.Address `json:"from"`
	To          *common.Address `json:"to"`
	Value       *big.Int       `json:"value"`
	GasPrice    *big.Int       `json:"gas_price"`
	GasUsed     uint64         `json:"gas_used"`
	Status      uint64         `json:"status"`
	Timestamp   time.Time      `json:"timestamp"`
}

// BSCTokenTransfer 代币转账事件
type BSCTokenTransfer struct {
	TxHash      common.Hash    `json:"tx_hash"`
	BlockNumber *big.Int       `json:"block_number"`
	LogIndex    uint           `json:"log_index"`
	Token       common.Address `json:"token"`
	From        common.Address `json:"from"`
	To          common.Address `json:"to"`
	Amount      *big.Int       `json:"amount"`
	Timestamp   time.Time      `json:"timestamp"`
}

// BSCSwapEvent 交换事件
type BSCSwapEvent struct {
	TxHash       common.Hash    `json:"tx_hash"`
	BlockNumber  *big.Int       `json:"block_number"`
	LogIndex     uint           `json:"log_index"`
	Pair         common.Address `json:"pair"`
	Sender       common.Address `json:"sender"`
	To           common.Address `json:"to"`
	Amount0In    *big.Int       `json:"amount0_in"`
	Amount1In    *big.Int       `json:"amount1_in"`
	Amount0Out   *big.Int       `json:"amount0_out"`
	Amount1Out   *big.Int       `json:"amount1_out"`
	Timestamp    time.Time      `json:"timestamp"`
}

// BSCLiquidityEvent 流动性事件
type BSCLiquidityEvent struct {
	TxHash      common.Hash    `json:"tx_hash"`
	BlockNumber *big.Int       `json:"block_number"`
	LogIndex    uint           `json:"log_index"`
	Pair        common.Address `json:"pair"`
	Sender      common.Address `json:"sender"`
	To          common.Address `json:"to"`
	Amount0     *big.Int       `json:"amount0"`
	Amount1     *big.Int       `json:"amount1"`
	Liquidity   *big.Int       `json:"liquidity"`
	EventType   string         `json:"event_type"` // "mint" or "burn"
	Timestamp   time.Time      `json:"timestamp"`
}

// BSCTokenInfo 代币信息
type BSCTokenInfo struct {
	Address  common.Address `json:"address"`
	Symbol   string         `json:"symbol"`
	Name     string         `json:"name"`
	Decimals uint8          `json:"decimals"`
	TotalSupply *big.Int    `json:"total_supply"`
}

// BSCPairInfo 交易对信息
type BSCPairInfo struct {
	Address    common.Address `json:"address"`
	Token0     common.Address `json:"token0"`
	Token1     common.Address `json:"token1"`
	Reserve0   *big.Int       `json:"reserve0"`
	Reserve1   *big.Int       `json:"reserve1"`
	TotalSupply *big.Int      `json:"total_supply"`
	Price0     decimal.Decimal `json:"price0"`
	Price1     decimal.Decimal `json:"price1"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// BSCMonitoringStats BSC监控统计
type BSCMonitoringStats struct {
	LatestBlock      *big.Int  `json:"latest_block"`
	ProcessedBlocks  uint64    `json:"processed_blocks"`
	TotalTransactions uint64   `json:"total_transactions"`
	TotalTransfers   uint64    `json:"total_transfers"`
	TotalSwaps       uint64    `json:"total_swaps"`
	TotalLiquidity   uint64    `json:"total_liquidity"`
	StartTime        time.Time `json:"start_time"`
	LastUpdateTime   time.Time `json:"last_update_time"`
	Status           string    `json:"status"`
}

// BSCMonitoringResponse BSC监控响应
type BSCMonitoringResponse struct {
	Enabled bool                `json:"enabled"`
	Stats   BSCMonitoringStats  `json:"stats"`
	Message string              `json:"message"`
}

// BSCTransactionResponse BSC交易查询响应
type BSCTransactionResponse struct {
	Transactions []BSCTransaction `json:"transactions"`
	Total        int              `json:"total"`
	Page         int              `json:"page"`
	PageSize     int              `json:"page_size"`
}

// BSCTokenTransferResponse BSC代币转账查询响应
type BSCTokenTransferResponse struct {
	Transfers []BSCTokenTransfer `json:"transfers"`
	Total     int                `json:"total"`
	Page      int                `json:"page"`
	PageSize  int                `json:"page_size"`
}

// BSCSwapEventResponse BSC交换事件查询响应
type BSCSwapEventResponse struct {
	Swaps    []BSCSwapEvent `json:"swaps"`
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// BSCPairInfoResponse BSC交易对信息响应
type BSCPairInfoResponse struct {
	Pairs []BSCPairInfo `json:"pairs"`
	Total int           `json:"total"`
}