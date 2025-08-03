# Crypto Info Service Docker 部署指南

## 概述

本文档提供了加密货币信息服务的Docker部署方案。由于网络环境限制，我们提供了多种部署选项。

## 部署选项

### 选项1：混合部署（推荐）

**适用场景**：网络环境受限，无法拉取Docker镜像

```bash
# 1. 启动Redis容器
cd deployments
docker-compose -f docker-compose-local.yml up redis -d

# 2. 本地运行应用
cd ..
go run cmd/server/main_hertz.go
```

**优势**：
- Redis容器化，数据持久化
- 应用本地运行，开发调试方便
- 网络问题影响最小

### 选项2：完整容器化部署

**适用场景**：网络环境良好，可以拉取Docker镜像

```bash
# 使用简化配置（仅应用+Redis）
cd deployments
docker-compose -f docker-compose-simple.yml up --build -d

# 或使用完整配置（包含监控组件）
docker-compose up --build -d
```

### 选项3：本地开发模式

**适用场景**：开发测试环境

```bash
# 直接运行应用（使用本地Redis或配置文件中的Redis）
go run cmd/server/main_hertz.go
```

## 配置文件说明

### docker-compose-local.yml
- **用途**：混合部署，仅Redis容器化
- **特点**：应用使用host网络模式

### docker-compose-simple.yml
- **用途**：简化的完整容器化部署
- **包含**：应用服务 + Redis

### docker-compose.yml
- **用途**：完整的生产环境部署
- **包含**：应用 + Redis + Prometheus + Grafana + Jaeger + Nginx

## 网络问题解决方案

如果遇到Docker镜像拉取失败，可以尝试：

1. **配置Docker镜像源**（已创建daemon.json）：
```bash
# 将daemon.json复制到Docker配置目录
sudo cp deployments/daemon.json /etc/docker/daemon.json
sudo systemctl restart docker
```

2. **使用国内镜像源**：
- 已在Dockerfile中配置阿里云镜像源
- 已配置Go模块代理为goproxy.cn

## 服务验证

部署完成后，可以通过以下方式验证服务：

```bash
# 健康检查
curl http://localhost:8080/health

# 测试价格API（根据实际路由调整）
curl http://localhost:8080/btc-price

# 检查Redis连接
docker exec crypto-redis redis-cli ping
```

## 端口说明

- **8080**：HTTP API服务端口
- **9090**：gRPC服务端口
- **6379**：Redis端口
- **3000**：Grafana（完整部署）
- **9091**：Prometheus（完整部署）
- **16686**：Jaeger UI（完整部署）

## 故障排除

1. **容器启动失败**：检查端口占用和网络连接
2. **镜像拉取失败**：使用混合部署方案
3. **应用连接Redis失败**：确认Redis容器状态和网络配置

## 当前状态

✅ Redis容器已成功启动并运行  
✅ 应用服务本地运行正常  
✅ 健康检查端点可访问  
✅ 混合部署方案验证完成  

推荐使用**选项1（混合部署）**作为当前的部署方案。