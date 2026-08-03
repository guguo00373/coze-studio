# Coze Studio 部署集成指南

## 概述

本文档详细介绍了将 Coze Studio 集成到大系统中时需要关注的依赖、配置和注意事项。Coze Studio 采用微服务架构，包含多个基础设施组件，需要合理规划端口、网络、存储和资源。

## 系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        大系统 / 外部网络                         │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      coze-network (Bridge)                      │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐         │
│  │  coze-web   │───▶│ coze-server │───▶│    MySQL    │         │
│  │  (Nginx)    │    │   (Go)      │    │   8.4.5     │         │
│  │  Port:8888  │    │             │    │             │         │
│  └─────────────┘    └──────┬──────┘    └─────────────┘         │
│                            │                                    │
│         ┌──────────────────┼──────────────────┐                │
│         ▼                  ▼                  ▼                │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐         │
│  │    Redis    │    │Elasticsearch│    │    MinIO    │         │
│  │    8.0      │    │   8.18.0    │    │             │         │
│  └─────────────┘    └─────────────┘    └─────────────┘         │
│                                                                │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐         │
│  │   Milvus    │    │    etcd     │    │     NSQ     │         │
│  │  v2.5.10    │    │    3.5      │    │   v1.2.1    │         │
│  └─────────────┘    └─────────────┘    └─────────────┘         │
└─────────────────────────────────────────────────────────────────┘
```

## 服务依赖清单

### 应用服务

| 服务 | 镜像 | 说明 | 暴露端口 |
|------|------|------|----------|
| coze-web | cozedev/coze-studio-web:latest | 前端 + Nginx 代理 | 8888 |
| coze-server | cozedev/coze-studio-server:latest | Go 后端服务 | 内部 |

### 基础设施服务

| 服务 | 镜像 | 用途 | 内部端口 |
|------|------|------|----------|
| MySQL | mysql:8.4.5 | 结构化数据存储 | 3306 |
| Redis | bitnamilegacy/redis:8.0 | 缓存、会话管理 | 6379 |
| Elasticsearch | bitnamilegacy/elasticsearch:8.18.0 | 全文搜索、知识库 | 9200, 9300 |
| MinIO | minio/minio:RELEASE.2025-06-13T11-33-47Z-cpuv1 | 对象存储（文档、图标） | 9000, 9001 |
| Milvus | milvusdb/milvus:v2.5.10 | 向量数据库 | 19530, 9091 |
| etcd | bitnamilegacy/etcd:3.5 | 分布式配置中心 | 2379, 2380 |
| NSQ | nsqio/nsq:v1.2.1 | 消息队列 | 4150, 4160, 4161, 4171 |

## 端口规划

### 对外暴露端口

| 端口 | 服务 | 协议 | 说明 |
|------|------|------|------|
| 8888 | coze-web | HTTP | Web 界面访问入口 |

### 内部服务端口（默认不对外暴露）

| 端口 | 服务 | 说明 |
|------|------|------|
| 3306 | MySQL | 数据库 |
| 6379 | Redis | 缓存 |
| 9200 | Elasticsearch | HTTP API |
| 9300 | Elasticsearch | 集群通信 |
| 9000 | MinIO | S3 API |
| 9001 | MinIO | 控制台 |
| 19530 | Milvus | 向量数据库 |
| 2379 | etcd | 客户端通信 |
| 2380 | etcd | 集群通信 |

### 端口冲突处理

如果现有系统已占用上述端口，需要修改 `docker/.env` 文件：

```bash
# MySQL
MYSQL_PORT=3307

# Redis
REDIS_PORT_NUMBER=6380

# Elasticsearch
ES_HTTP_PORT=9201

# MinIO
MINIO_API_PORT=9002
MINIO_CONSOLE_PORT=9003

# Milvus
MILVUS_PORT=19531

# Web 访问端口
WEB_LISTEN_ADDR=8889
```

## 环境变量配置

### 必需配置项

创建 `docker/.env` 文件：

```bash
# ==================== 数据库配置 ====================
MYSQL_ROOT_PASSWORD=your_root_password
MYSQL_DATABASE=opencoze
MYSQL_USER=coze
MYSQL_PASSWORD=your_password

# ==================== Redis 配置 ====================
REDIS_AOF_ENABLED=no
ALLOW_EMPTY_PASSWORD=yes

# ==================== MinIO 配置 ====================
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=your_minio_password
STORAGE_BUCKET=opencoze

# ==================== AI 模型配置 ====================
# OpenAI
OPENAI_API_KEY=sk-your-api-key
OPENAI_BASE_URL=https://api.openai.com/v1

# 或者使用其他模型提供商（Ark, Claude, Gemini, Qwen, DeepSeek, Ollama）
# ARK_API_KEY=your-ark-key
# ARK_BASE_URL=https://ark.cn-beijing.volces.com/api/v3
```

### 模型配置

在 `backend/conf/model/` 目录下配置 AI 模型：

1. 复制模板文件：`backend/conf/model/template/`
2. 设置以下参数：
   - `id`: 模型唯一标识
   - `meta.conn_config.api_key`: API 密钥
   - `meta.conn_config.model`: 模型名称

支持的模型提供商：
- OpenAI
- Volcengine Ark
- Claude
- Gemini
- Qwen
- DeepSeek
- Ollama（本地部署）

## 数据持久化

### 持久化目录

所有数据存储在 `docker/data/` 目录下：

```
docker/data/
├── mysql/                    # MySQL 数据
├── bitnami/
│   ├── redis/               # Redis 数据
│   ├── elasticsearch/       # Elasticsearch 数据
│   └── etcd/               # etcd 数据
├── minio/                   # MinIO 对象存储
└── milvus/                  # Milvus 向量数据
```

### 备份策略

```bash
# 停止服务
make down

# 备份数据目录
tar -czf coze-data-backup-$(date +%Y%m%d).tar.gz docker/data/

# 恢复数据
tar -xzf coze-data-backup-20250101.tar.gz
make web
```

### 数据清理

```bash
# 停止容器并删除数据
make clean
```

## 网络要求

### Docker 网络

- 网络名称：`coze-network`
- 网络类型：`bridge`
- 容器间通信通过服务名称解析

### 外部访问

如需从外部系统访问 Coze Studio：

```bash
# 通过宿主机访问
curl http://localhost:8888

# 通过 Docker 网络访问（大系统在相同 Docker 网络）
curl http://coze-web:80
```

### 防火墙规则

如需对外暴露服务，需要开放以下端口：

```bash
# Web 界面
iptables -A INPUT -p tcp --dport 8888 -j ACCEPT

# 如需直接访问 API（不推荐）
iptables -A INPUT -p tcp --dport 8888 -j ACCEPT
```

## 资源需求

### 最低配置

| 资源 | 最低要求 | 推荐配置 |
|------|----------|----------|
| CPU | 2 核 | 4+ 核 |
| 内存 | 4 GB | 8+ GB |
| 磁盘 | 10 GB | 20+ GB |
| Docker | ≥ 20.10 | 最新稳定版 |
| Docker Compose | V2 (plugin) | 最新稳定版 |

### 服务资源分配

| 服务 | CPU 限制 | 内存限制 | 说明 |
|------|----------|----------|------|
| MySQL | 1-2 核 | 1-2 GB | 核心数据库 |
| Elasticsearch | 1-2 核 | 1-2 GB | 搜索引擎 |
| Milvus | 1-2 核 | 1-2 GB | 向量数据库 |
| MinIO | 0.5 核 | 512 MB | 对象存储 |
| Redis | 0.5 核 | 256 MB | 缓存 |
| coze-server | 1-2 核 | 1-2 GB | 后端服务 |
| coze-web | 0.5 核 | 256 MB | 前端代理 |

## 健康检查

### 服务健康状态

所有服务都配置了健康检查，可通过以下命令查看：

```bash
# 查看所有服务状态
docker compose -f ./docker/docker-compose.yml ps

# 查看特定服务日志
docker logs coze-server
docker logs coze-mysql
docker logs coze-elasticsearch
```

### 健康检查端点

| 服务 | 检查方式 | 超时时间 |
|------|----------|----------|
| MySQL | `mysqladmin ping` | 5s |
| Redis | `redis-cli ping` | 10s |
| Elasticsearch | `curl http://localhost:9200` | 10s |
| MinIO | `mc ready` | 10s |
| Milvus | `curl http://localhost:9091/healthz` | 10s |
| etcd | `etcdctl endpoint health` | 10s |

## 集成注意事项

### 1. 启动顺序

服务启动顺序由 `depends_on` 和健康检查决定：

```
MySQL/Redis/Elasticsearch/MinIO/Milvus/etcd/NSQ
         ↓ (全部健康)
    coze-server
         ↓
      coze-web
```

首次启动需要 5-15 分钟完成初始化。

### 2. 数据库迁移

Coze Studio 使用 Atlas 进行数据库迁移：

```bash
# 手动触发迁移
make sync_db

# 导出数据库结构
make dump_db

# 生成迁移文件哈希
make atlas-hash
```

### 3. Elasticsearch 插件

Elasticsearch 需要安装 `smartcn` 中文分词插件：

- 插件位置：`docker/volumes/elasticsearch/analysis-smartcn.zip`
- 自动安装：首次启动时自动安装

### 4. Milvus 依赖

Milvus 依赖 etcd 和 MinIO：

- 使用 etcd 存储元数据
- 使用 MinIO 存储向量数据
- 启动顺序：etcd → MinIO → Milvus

### 5. 与大系统集成

#### 方式一：API 集成

通过 HTTP API 与 Coze Studio 交互：

```bash
# 获取所有 Bot
curl http://localhost:8888/api/bots

# 创建对话
curl -X POST http://localhost:8888/api/conversations \
  -H "Content-Type: application/json" \
  -d '{"bot_id": "your-bot-id"}'
```

#### 方式二：SDK 集成

使用 CozeLoop SDK：

```python
from cozeloop import CozeLoop

loop = CozeLoop(
    base_url="http://localhost:8888",
    api_key="your-api-key"
)

# 创建 Trace
trace = loop.trace.create(
    bot_id="your-bot-id",
    user_id="user-123"
)
```

#### 方式三：共享基础设施

如需共享数据库或缓存：

1. **MySQL**: 创建独立数据库，避免表名冲突
2. **Redis**: 使用独立前缀，如 `coze:session:*`
3. **MinIO**: 使用独立 Bucket

### 6. 安全建议

- 修改所有默认密码
- 不对外暴露 MySQL、Redis 等内部服务
- 使用 SSL/TLS 加密通信
- 配置访问控制列表（ACL）
- 定期备份数据

### 7. 监控与日志

```bash
# 查看容器资源使用
docker stats

# 查看服务日志
docker compose logs -f coze-server

# 导出日志
docker compose logs coze-server > coze-server.log 2>&1
```

## 故障排查

### 常见问题

| 问题 | 原因 | 解决方案 |
|------|------|----------|
| 服务启动失败 | 端口冲突 | 检查端口占用，修改 `.env` |
| 数据库连接失败 | 密码错误 | 检查 `.env` 中的数据库配置 |
| Elasticsearch 不健康 | 内存不足 | 增加 Docker 内存限制 |
| Milvus 启动慢 | 依赖服务未就绪 | 等待 etcd 和 MinIO 健康 |
| 无法访问 Web | 防火墙 | 检查端口 8888 是否开放 |

### 日志查看

```bash
# 查看所有服务状态
docker compose ps

# 查看特定服务日志
docker compose logs --tail=100 coze-server

# 实时查看日志
docker compose logs -f coze-server
```

## 快速开始

```bash
# 1. 克隆仓库
git clone https://github.com/coze-dev/coze-studio.git
cd coze-studio

# 2. 配置环境
cp docker/.env.example docker/.env
# 编辑 docker/.env 修改密码和 API 密钥

# 3. 启动服务
make web

# 4. 访问系统
# 浏览器打开 http://localhost:8888
```

## 参考文档

- [Coze Studio 官方文档](https://www.coze.com/docs)
- [CozeLoop SDK 文档](https://www.coze.com/docs/cozeloop)
- [Docker Compose 文档](https://docs.docker.com/compose/)
- [Atlas 数据库迁移](https://atlasgo.io/)
