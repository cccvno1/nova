# Docker 测试环境

## 快速启动

```bash
# 方式1: 使用脚本
./scripts/test-env.sh up

# 方式2: 使用 Makefile
make -f Makefile.docker test-up

# 方式3: 使用 docker-compose
docker-compose -f docker-compose.test.yml up -d
```

## 服务信息

- **后端服务**: http://localhost:8080
- **PostgreSQL**: localhost:5432 (用户: postgres, 密码: postgres, 数据库: nova)
- **Redis**: localhost:6379

## 常用命令

### 使用脚本
```bash
./scripts/test-env.sh up       # 启动
./scripts/test-env.sh down     # 停止
./scripts/test-env.sh restart  # 重启
./scripts/test-env.sh logs     # 查看日志
./scripts/test-env.sh status   # 查看状态
./scripts/test-env.sh build    # 构建镜像
./scripts/test-env.sh rebuild  # 重新构建并启动
./scripts/test-env.sh clean    # 清理所有数据
```

### 使用 Makefile
```bash
make -f Makefile.docker test-up              # 启动
make -f Makefile.docker test-down            # 停止
make -f Makefile.docker test-restart         # 重启
make -f Makefile.docker test-logs            # 查看所有日志
make -f Makefile.docker test-logs-server     # 查看服务日志
make -f Makefile.docker test-logs-postgres   # 查看数据库日志
make -f Makefile.docker test-logs-redis      # 查看Redis日志
make -f Makefile.docker test-ps              # 查看容器状态
make -f Makefile.docker test-build           # 构建镜像
make -f Makefile.docker test-rebuild         # 重新构建并启动
make -f Makefile.docker test-clean           # 清理所有数据
```

### 进入容器
```bash
make -f Makefile.docker test-exec-server     # 进入后端容器
make -f Makefile.docker test-exec-postgres   # 进入数据库容器（自动连接数据库）
make -f Makefile.docker test-exec-redis      # 进入Redis容器（自动连接Redis）
```

## 配置文件

测试环境使用独立的配置文件 `configs/config.test.yaml`，可根据需要修改。

环境变量优先级高于配置文件，格式: `NOVA_<SECTION>_<KEY>`

示例:
```bash
export NOVA_DATABASE_HOST=custom-postgres
export NOVA_REDIS_HOST=custom-redis
export NOVA_SERVER_PORT=9090
```

## 数据持久化

数据存储在 Docker volumes 中:
- `nova_postgres_data`: PostgreSQL 数据
- `nova_redis_data`: Redis 数据

删除数据:
```bash
./scripts/test-env.sh clean
# 或
make -f Makefile.docker test-clean
```

## 健康检查

服务启动后会自动进行健康检查，确保服务就绪后再启动应用。

查看健康状态:
```bash
docker-compose -f docker-compose.test.yml ps
```

## 故障排查

### 查看日志
```bash
# 所有服务
docker-compose -f docker-compose.test.yml logs -f

# 特定服务
docker-compose -f docker-compose.test.yml logs -f nova-server
docker-compose -f docker-compose.test.yml logs -f postgres
docker-compose -f docker-compose.test.yml logs -f redis
```

### 重新构建
```bash
./scripts/test-env.sh rebuild
```

### 完全重置
```bash
./scripts/test-env.sh clean
./scripts/test-env.sh up
```
