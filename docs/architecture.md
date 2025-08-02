# 项目架构说明

## MVC 架构概述

本项目采用标准的 MVC（Model-View-Controller）架构模式，结合 Gin 框架构建。

## 目录结构

```
crypto-info/
├── cmd/                    # 应用程序入口
│   └── main.go            # 主程序文件
├── internal/              # 内部包
│   ├── controllers/       # 控制器层 (Controller)
│   │   └── controllers.go # 处理HTTP请求和响应
│   ├── models/           # 模型层 (Model)
│   │   └── response.go   # 数据结构定义
│   ├── routes/           # 路由层
│   │   └── routes.go     # 路由配置
│   └── services/         # 服务层 (Business Logic)
│       └── bitcoin_service.go # 业务逻辑处理
├── configs/              # 配置文件
├── docs/                 # 文档
├── go.mod               # Go模块文件
└── go.sum               # 依赖校验文件
```

## 架构层次说明

### 1. 路由层 (Routes)
- **文件**: `internal/routes/routes.go`
- **职责**: 定义API路由和路径映射
- **特点**: 集中管理所有路由配置，支持路由分组

### 2. 控制器层 (Controllers)
- **文件**: `internal/controllers/controllers.go`
- **职责**: 处理HTTP请求，调用服务层，返回响应
- **特点**: 轻量级，只负责请求处理和响应格式化

### 3. 服务层 (Services)
- **文件**: `internal/services/bitcoin_service.go`
- **职责**: 实现具体的业务逻辑
- **特点**: 可复用的业务逻辑，独立于HTTP层

### 4. 模型层 (Models)
- **文件**: `internal/models/response.go`
- **职责**: 定义数据结构和数据传输对象
- **特点**: 纯数据结构，无业务逻辑

## API 端点

### 基础路由
- `GET /` - 首页
- `GET /health` - 健康检查
- `GET /btc-price` - 获取比特币价格

### API v1 路由组
- `GET /api/v1/health` - 健康检查 (API版本)
- `GET /api/v1/btc-price` - 获取比特币价格 (API版本)

## 架构优势

1. **关注点分离**: 每层都有明确的职责
2. **可维护性**: 代码结构清晰，易于维护
3. **可测试性**: 各层可以独立测试
4. **可扩展性**: 易于添加新功能和修改现有功能
5. **代码复用**: 服务层可以被多个控制器复用

## 数据流

```
请求 → 路由层 → 控制器层 → 服务层 → 模型层
响应 ← 路由层 ← 控制器层 ← 服务层 ← 模型层
```

1. 客户端发送HTTP请求
2. 路由层根据URL路径分发到对应控制器
3. 控制器解析请求参数，调用相应服务
4. 服务层执行业务逻辑，使用模型层数据结构
5. 结果通过控制器层格式化后返回给客户端