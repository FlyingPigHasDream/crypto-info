# 项目架构说明

## MVC + DAO 架构概述

本项目采用标准的 MVC（Model-View-Controller）架构模式，并增加了 DAO（Data Access Object）层，结合 Gin 框架构建完整的分层架构。

## 目录结构

```
crypto-info/
├── cmd/                    # 应用程序入口
│   └── main.go            # 主程序文件
├── internal/              # 内部包
│   ├── controllers/       # 控制器层 (Controller)
│   │   └── controllers.go # 处理HTTP请求和响应
│   ├── dao/              # 数据访问层 (Data Access Object)
│   │   ├── interfaces.go # DAO接口定义
│   │   ├── crypto_price_dao.go # 加密货币价格数据访问
│   │   └── health_dao.go # 健康检查数据访问
│   ├── models/           # 模型层 (Model)
│   │   └── response.go   # 数据结构定义
│   ├── routes/           # 路由层
│   │   ├── routes.go     # 路由总入口
│   │   ├── base_routes.go # 基础路由
│   │   ├── crypto_routes.go # 加密货币路由
│   │   └── api_routes.go # API版本化路由
│   └── services/         # 服务层 (Business Logic)
│       ├── bitcoin_service.go # 比特币业务逻辑
│       └── health_service.go # 健康检查业务逻辑
├── configs/              # 配置文件
├── docs/                 # 文档
│   └── architecture.md  # 架构说明文档
├── go.mod               # Go模块文件
└── go.sum               # 依赖校验文件
```

## 架构层次说明

### 1. 路由层 (Routes)
- **主文件**: `internal/routes/routes.go` - 路由总入口
- **基础路由**: `internal/routes/base_routes.go` - 基础功能路由
- **加密货币路由**: `internal/routes/crypto_routes.go` - 加密货币相关路由
- **API路由**: `internal/routes/api_routes.go` - API版本化路由
- **职责**: 按业务模块分离路由配置
- **特点**: 模块化路由管理，高扩展性，支持业务分组

### 2. 控制器层 (Controllers)
- **文件**: `internal/controllers/controllers.go`
- **职责**: 处理HTTP请求，调用服务层，返回响应
- **特点**: 轻量级，只负责请求处理和响应格式化



### 4. 服务层 (Services)
- **比特币服务**: `internal/services/bitcoin_service.go` - 比特币相关业务逻辑
- **健康检查服务**: `internal/services/health_service.go` - 健康检查业务逻辑
- **职责**: 实现具体的业务逻辑，调用DAO层获取数据
- **特点**: 可复用的业务逻辑，独立于HTTP层和数据层

### 5. 数据访问层 (DAO)
- **接口定义**: `internal/dao/interfaces.go` - 定义数据访问接口
- **加密货币DAO**: `internal/dao/crypto_price_dao.go` - 加密货币数据访问实现
- **健康检查DAO**: `internal/dao/health_dao.go` - 健康检查数据访问实现
- **职责**: 封装所有数据访问逻辑，包括API调用、缓存管理
- **特点**: 数据访问抽象，支持缓存机制，易于测试和替换

### 6. 模型层 (Models)
- **文件**: `internal/models/response.go`
- **职责**: 定义数据结构和数据传输对象
- **特点**: 纯数据结构，无业务逻辑

## API 端点

### 基础路由 (base_routes.go)
- `GET /` - 首页
- `GET /health` - 健康检查

### 加密货币路由 (crypto_routes.go)
- `GET /crypto/btc-price` - 获取比特币价格 (新路径)
- `GET /btc-price` - 获取比特币价格 (向后兼容)

### API v1 路由组 (api_routes.go)
- `GET /api/v1/health` - 健康检查 (API版本)
- `GET /api/v1/btc-price` - 获取比特币价格 (API版本)

### 路由扩展性
每个路由文件都预留了扩展接口的注释，便于未来添加新功能：
- 加密货币模块可扩展：ETH价格、市值查询等
- API模块可扩展：用户管理、市场数据等
- 支持添加新的业务模块路由文件

## 架构优势

1. **关注点分离**: 每层都有明确的职责，数据访问与业务逻辑分离
2. **可维护性**: 代码结构清晰，层次分明，易于维护
3. **可测试性**: 各层可以独立测试，DAO层支持Mock测试
4. **可扩展性**: 易于添加新功能和修改现有功能
5. **代码复用**: 服务层和DAO层可以被多个控制器复用
6. **数据抽象**: DAO层提供统一的数据访问接口
7. **缓存支持**: 内置缓存机制，提高性能
8. **容错能力**: 多层降级策略，保证服务稳定性

## 数据流

```
请求 → 路由层 → 控制器层 → 服务层 → DAO层 → 外部API/缓存
响应 ← 路由层 ← 控制器层 ← 服务层 ← DAO层 ← 外部API/缓存
```

1. 客户端发送HTTP请求
2. 路由层根据URL路径分发到对应控制器
3. 控制器解析请求参数，调用相应服务
4. 服务层执行业务逻辑，调用DAO层获取数据
5. DAO层处理数据访问（缓存查询 → API调用 → 模拟数据）
6. 数据通过各层返回，最终格式化后响应给客户端

## DAO层特性

### 缓存机制
- 内存缓存，5分钟过期时间
- 优先返回缓存数据，提高响应速度
- 缓存失效时自动从API获取新数据

### 降级策略
- API调用失败时自动返回模拟数据
- 多层容错机制，确保服务可用性
- 明确标识数据来源（API/缓存/模拟）