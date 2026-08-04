# Coze Studio Docker 启动指南 (PostgreSQL 版本)

## 快速开始

### 1. 复制环境变量文件

```bash
cd D:\claudecode\Coze_studio\coze-studio\docker
copy .env.postgres .env
```

### 2. 修改配置

编辑 `.env` 文件，配置以下必填项：

```bash
# AI 模型配置 (必填)
export MODEL_API_KEY_0="your-api-key"
export MODEL_BASE_URL_0="https://api.openai.com/v1"
export MODEL_ID_0="gpt-4"

# Embedding 配置 (知识库功能需要)
export OPENAI_EMBEDDING_API_KEY="your-api-key"
export OPENAI_EMBEDDING_BASE_URL="https://api.openai.com/v1/embeddings"
export OPENAI_EMBEDDING_MODEL="text-embedding-3-large"
```

### 3. 启动服务

```bash
# 使用 PostgreSQL 版本启动
docker compose -f docker-compose.postgres.yml up -d

# 查看日志
docker compose -f docker-compose.postgres.yml logs -f

# 查看特定服务日志
docker compose -f docker-compose.postgres.yml logs -f coze-server
```

### 4. 访问服务

- **前端**: http://localhost:8888
- **MinIO 控制台**: http://localhost:9001

---

## 常用命令

```bash
# 启动服务
docker compose -f docker-compose.postgres.yml up -d

# 停止服务
docker compose -f docker-compose.postgres.yml down

# 停止并删除数据
docker compose -f docker-compose.postgres.yml down -v

# 重新构建并启动
docker compose -f docker-compose.postgres.yml up -d --build

# 查看服务状态
docker compose -f docker-compose.postgres.yml ps

# 查看日志
docker compose -f docker-compose.postgres.yml logs -f coze-server
```

---

## 故障排查

> 实际部署中遇到的问题及解决方案汇总，见 [docs/postgresql-deployment-notes.md](../docs/postgresql-deployment-notes.md)。

### 1. PostgreSQL 连接失败

```bash
# 检查 PostgreSQL 是否启动
docker compose -f docker-compose.postgres.yml ps postgres

# 查看 PostgreSQL 日志
docker compose -f docker-compose.postgres.yml logs postgres

# 测试连接
docker compose -f docker-compose.postgres.yml exec postgres psql -U opencoze -d opencoze
```

### 2. coze-server 启动失败

```bash
# 查看日志
docker compose -f docker-compose.postgres.yml logs coze-server

# 检查环境变量
docker compose -f docker-compose.postgres.yml exec coze-server env | grep DATABASE
```

### 3. Schema 初始化问题

```bash
# 进入 PostgreSQL 容器
docker compose -f docker-compose.postgres.yml exec postgres psql -U opencoze -d opencoze

# 检查表是否创建
\dt

# 退出
\q
```

---

## 与 MySQL 版本的区别

| 项目 | MySQL 版本 | PostgreSQL 版本 |
|------|-----------|-----------------|
| 数据库 | MySQL 8.4.5 | PostgreSQL 15 |
| 启动文件 | docker-compose.yml | docker-compose.postgres.yml |
| 环境文件 | .env.example | .env.postgres |
| Schema | volumes/mysql/schema.sql | volumes/postgres/schema.sql |
| 连接配置 | MYSQL_DSN | DATABASE_HOST/PORT/USER/PASSWORD/NAME |

---

## 数据库管理

### 连接到 PostgreSQL

```bash
# 使用 docker exec
docker compose -f docker-compose.postgres.yml exec postgres psql -U opencoze -d opencoze

# 使用本地 psql
psql -h localhost -p 5432 -U opencoze -d opencoze
```

### 常用 SQL 命令

```sql
-- 查看所有表
\dt

-- 查看表结构
\d table_name

-- 查看数据
SELECT * FROM single_agent_draft LIMIT 10;

-- 退出
\q
```

---

## 切换回 MySQL 版本

如果需要切换回 MySQL 版本：

```bash
# 停止 PostgreSQL 版本
docker compose -f docker-compose.postgres.yml down

# 使用 MySQL 版本启动
docker compose up -d
```
