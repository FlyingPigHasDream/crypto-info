# MySQL Docker 配置说明

## 容器信息
- **容器名称**: `crypto-mysql`
- **镜像版本**: `mysql:8.0`
- **端口映射**: `3306:3306`
- **网络**: `crypto-network`

## 数据库配置
- **数据库名**: `crypto_info`
- **Root密码**: `crypto_root_2024`
- **用户名**: `crypto_user`
- **用户密码**: `crypto_pass_2024`
- **字符集**: `utf8mb4`
- **排序规则**: `utf8mb4_unicode_ci`

## 连接方式

### 1. 从宿主机连接
```bash
mysql -h 127.0.0.1 -P 3306 -u crypto_user -pcrypto_pass_2024 crypto_info
```

### 2. 从容器内连接
```bash
docker exec -it crypto-mysql mysql -u crypto_user -pcrypto_pass_2024 crypto_info
```

### 3. Root用户连接
```bash
docker exec -it crypto-mysql mysql -u root -pcrypto_root_2024
```

## 数据持久化
- **数据目录**: `mysql_data` volume -> `/var/lib/mysql`
- **配置目录**: `mysql_config` volume -> `/etc/mysql/conf.d`

## 健康检查
容器配置了健康检查，每30秒检查一次MySQL服务状态：
```bash
docker ps --filter name=crypto-mysql
```

## 管理命令

### 启动MySQL服务
```bash
cd deployments
docker-compose -f docker-compose-local.yml up mysql -d
```

### 停止MySQL服务
```bash
cd deployments
docker-compose -f docker-compose-local.yml stop mysql
```

### 查看MySQL日志
```bash
docker logs crypto-mysql
```

### 备份数据库
```bash
docker exec crypto-mysql mysqldump -u root -pcrypto_root_2024 crypto_info > backup.sql
```

### 恢复数据库
```bash
docker exec -i crypto-mysql mysql -u root -pcrypto_root_2024 crypto_info < backup.sql
```

## 应用程序配置

在应用程序中使用以下连接参数：
```yaml
database:
  mysql:
    host: "127.0.0.1"  # 或 "mysql" (如果应用也在容器中)
    port: 3306
    database: "crypto_info"
    username: "crypto_user"
    password: "crypto_pass_2024"
    charset: "utf8mb4"
```

## 注意事项

1. **安全性**: 生产环境中请修改默认密码
2. **性能**: 可根据需要调整MySQL配置参数
3. **备份**: 定期备份重要数据
4. **监控**: 建议配置MySQL性能监控

## 故障排除

### 连接失败
1. 检查容器状态: `docker ps --filter name=crypto-mysql`
2. 查看容器日志: `docker logs crypto-mysql`
3. 检查端口占用: `lsof -i :3306`

### 权限问题
```bash
# 进入容器修复权限
docker exec -it crypto-mysql mysql -u root -pcrypto_root_2024
GRANT ALL PRIVILEGES ON crypto_info.* TO 'crypto_user'@'%';
FLUSH PRIVILEGES;
```