# Crypto Info - 加密货币信息服务

基于 ByteDance Go 代码规范的高性能加密货币信息服务，提供价格查询和交易量分析功能。

## 🏗️ 项目架构

本项目采用现代化的微服务架构设计，遵循 ByteDance Go 开发最佳实践：

```
crypto-info/
├── api/                    # API定义
│   └── crypto/v1/         # gRPC/Protobuf定义
├── build/                  # 构建相关文件
│   └── Dockerfile         # Docker构建文件
├── cmd/                    # 应用程序入口
│   └── server/            # 服务器主程序
├── configs/                # 配置文件
│   ├── config.yaml        # 主配置文件
│   ├── development.yaml   # 开发环境配置
│   └── production.yaml    # 生产环境配置
├── deployments/            # 部署配置
│   ├── docker-compose.yml # Docker Compose
│   └── k8s/               # Kubernetes配置
├── internal/               # 内部代码
│   ├── config/            # 配置管理
│   ├── handler/           # HTTP处理器
│   ├── model/             # 数据模型
│   ├── pkg/               # 公共包
│   │   ├── database/      # 数据库连接
│   │   ├── logger/        # 日志管理
│   │   └── middleware/    # 中间件
│   ├── server/            # 服务器实现
│   └── service/           # 业务逻辑
└── Makefile               # 构建脚本
```

## 🚀 核心特性

### 技术栈
- **框架**: Gin (HTTP) + CloudWeGo Hertz/Kitex (微服务)
- **配置管理**: Viper
- **日志**: Logrus + Lumberjack
- **缓存**: Redis
- **监控**: Prometheus + Grafana + Jaeger
- **部署**: Docker + Kubernetes

### 业务功能
- 🔍 **价格查询**: 支持多种加密货币实时价格查询
- 📊 **交易量分析**: 提供详细的交易量统计和趋势分析
- 📈 **市场波动**: 实时监控市场交易量波动情况
- 🔄 **多币种对比**: 支持多个加密货币的交易量对比
- 🏆 **排行榜**: 交易量排行榜功能

### 架构特点
- 🏗️ **分层架构**: Handler -> Service -> Model 清晰分层
- 🔧 **依赖注入**: 基于接口的依赖注入设计
- 📝 **配置驱动**: 支持多环境配置管理
- 🚀 **高性能**: Redis缓存 + 连接池优化
- 🛡️ **安全性**: CORS、限流、安全头等中间件
- 📊 **可观测性**: 完整的日志、监控、链路追踪

## 🛠️ 快速开始

### 环境要求
- Go 1.21+
- Redis 6.0+
- Docker & Docker Compose (可选)

### 本地开发

1. **克隆项目**
```bash
git clone <repository-url>
cd crypto-info
```

2. **安装依赖**
```bash
make deps
```

3. **启动Redis**
```bash
docker run -d --name redis -p 6379:6379 redis:7-alpine
```

4. **运行服务**
```bash
make dev
# 或者
go run cmd/server/main.go
```

5. **测试API**
```bash
# 获取BTC价格
curl http://localhost:8080/api/v1/crypto/price?symbol=BTC

# 获取交易量分析
curl http://localhost:8080/api/v1/crypto/volume/analysis?symbol=BTC&days=7
```

### Docker部署

```bash
# 构建并启动所有服务
docker-compose -f deployments/docker-compose.yml up -d

# 查看服务状态
docker-compose -f deployments/docker-compose.yml ps
```

### Kubernetes部署

```bash
# 创建命名空间
kubectl create namespace crypto-info

# 部署应用
kubectl apply -f deployments/k8s/

# 查看部署状态
kubectl get pods -n crypto-info
```

## 📚 API文档

### 价格相关API

| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/v1/crypto/price` | GET | 获取加密货币价格 |
| `/api/v1/crypto/btc-price` | GET | 获取BTC价格 |

### 交易量相关API

| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/v1/crypto/volume/analysis` | GET | 获取交易量分析 |
| `/api/v1/crypto/volume/fluctuation` | GET | 获取交易量波动 |
| `/api/v1/crypto/volume/comparison` | GET | 获取交易量对比 |
| `/api/v1/crypto/volume/top` | GET | 获取交易量排行 |

### 请求参数

- `symbol`: 加密货币符号 (BTC, ETH, LTC等)
- `days`: 分析天数 (默认10天，最大365天)
- `symbols`: 多个币种符号，逗号分隔
- `limit`: 返回数量限制

## 🔧 配置说明

### 环境配置

项目支持多环境配置：
- `configs/config.yaml`: 基础配置
- `configs/development.yaml`: 开发环境
- `configs/production.yaml`: 生产环境

### 环境变量

支持通过环境变量覆盖配置：
```bash
CRYPTO_APP_ENV=production
CRYPTO_DATABASE_REDIS_HOST=redis-cluster
CRYPTO_DATABASE_REDIS_PASSWORD=your-password
CRYPTO_LOG_LEVEL=info
```

## 🧪 测试

```bash
# 运行所有测试
make test

# 运行基准测试
make bench

# 生成测试覆盖率报告
make test
open coverage.html
```

## 📊 监控

### 健康检查
```bash
curl http://localhost:8080/health
```

### Prometheus指标
```bash
curl http://localhost:9091/metrics
```

### Grafana仪表板
访问 http://localhost:3000 (admin/admin123)

### Jaeger链路追踪
访问 http://localhost:16686

## 🔨 开发工具

### Makefile命令

```bash
make help          # 显示帮助信息
make build         # 构建应用
make test          # 运行测试
make lint          # 代码检查
make fmt           # 格式化代码
make docker-build  # 构建Docker镜像
make deploy-dev    # 部署到开发环境
```

### 代码生成

```bash
# 生成API文档
make swag

# 生成代码
make generate
```

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

### 代码规范

- 遵循 [ByteDance Go 代码规范](https://github.com/bytedance/gopkg)
- 使用 `golangci-lint` 进行代码检查
- 保持测试覆盖率 > 80%
- 添加适当的注释和文档

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

- [ByteDance](https://github.com/bytedance) - Go 开发最佳实践
- [CloudWeGo](https://github.com/cloudwego) - 高性能微服务框架
- [Gin](https://github.com/gin-gonic/gin) - HTTP Web框架
- [Redis](https://redis.io/) - 内存数据库

## 📞 联系方式

如有问题或建议，请通过以下方式联系：

- 提交 [Issue](../../issues)
- 发送邮件到 [your-email@example.com]
- 加入讨论群组

---

**注意**: 本项目目前使用模拟数据，生产环境请配置真实的加密货币API接口。
