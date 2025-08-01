# Go Web Study

一个使用Go语言编写的Web应用程序示例。

## 项目结构

```
go-web-study/
├── cmd/                 # 主应用程序
├── configs/             # 配置文件
├── deployments/         # 部署相关文件
├── docs/                # 文档
├── internal/            # 私有应用程序和库代码
├── pkg/                 # 可被外部应用使用的库代码
├── scripts/             # 脚本文件
├── web/                 # Web资源文件
├── README.md            # 项目说明文件
└── go.mod               # Go模块文件
```

## 快速开始

1. 确保已安装Go环境（版本1.16+）
2. 运行以下命令启动服务器：
   ```bash
   go run cmd/main.go
   ```
3. 在浏览器中访问 `http://localhost:8080`

## API端点

- `GET /` - 主页
- `GET /health` - 健康检查端点
- `GET /btc-price` - 获取比特币实时价格

## 许可证

MIT# crypto-info
