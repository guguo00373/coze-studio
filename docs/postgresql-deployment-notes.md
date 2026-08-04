# Coze Studio PostgreSQL 版部署问题记录

> 环境：Windows 11 + Docker Desktop（Docker 29.6.2 / Compose v5.3.1）
> 分支：`postgresql-support`（fork: guguo00373/coze-studio）
> 日期：2026-08-03

## 部署方式

```powershell
cd D:\OpenCode\coze-studio\docker
copy .env.postgres .env   # 然后填入真实 API key
docker compose -f docker-compose.postgres.yml up -d --build
```

---

## 问题 1：首次启动报 `container coze-elasticsearch is unhealthy`（假失败）

**现象**

```
✘ Container coze-elasticsearch  Error dependency elasticsearch failed to start
dependency failed to start: container coze-elasticsearch is unhealthy
```

**原因**

Elasticsearch 首次启动需要安装 smartcn 中文分词插件、加载索引模板和 ILM 策略，耗时较长（本机约 100s+），超过了 compose 健康检查的 `start_period`（10s）+ 重试窗口，compose 误判依赖失败并中止启动。此时 ES 本身没有任何错误，只是慢。

**解决**

无需处理，直接再跑一次（不要加 `--build`）：

```powershell
docker compose -f docker-compose.postgres.yml up -d
```

---

## 问题 2：`.env` 里配置的模型 key 不生效（compose 硬编码覆盖）

**现象**

`.env` 中已填写真实 API key，但容器内仍是 `sk-your-key`。

**原因**

`docker-compose.postgres.yml` 的 `coze-server.environment:` 块硬编码了 16 行模型配置（含 `MODEL_API_KEY_0: "sk-your-key"` 等）。优先级规则：

- compose `environment:` > compose `env_file:` > 容器内挂载的 `.env`
- 后端 `main.go` 用的是 `godotenv.Load()`（非 Overload），**不会覆盖已存在的环境变量**，所以硬编码值会永久生效

**解决（已提交到本仓库）**

删除 `docker-compose.postgres.yml` 中 `coze-server.environment:` 下所有模型/Embedding 配置行，改由 `env_file` 从 `.env` 统一注入。数据库连接配置（`DATABASE_*`）保留在 compose 中（与容器网络名绑定，且默认值与 `.env` 一致）。

**验证方法**

```powershell
docker compose -f docker-compose.postgres.yml config | Select-String 'MODEL_API_KEY'
# 渲染结果应显示 .env 中的真实 key，而不是 sk-your-key
```

---

## 问题 3：`.env` 修改"已保存"但实际未生效

**现象**

用户确认已修改 key，但 `docker compose config` 渲染结果仍是 `sk-your-key`。

**原因**

实际改的是模板文件 `.env.postgres`，或其他目录下的 `.env`，而非 `docker\.env` 本身（或编辑器未保存）。

**教训**

改完后用 `docker compose ... config` 渲染验证，不要凭记忆确认。渲染输出里 `MODEL_API_KEY_0` 等以 `_0` 结尾的变量不受简单 `KEY:` 掩码正则影响，检查时要直接看值。

---

## 问题 4：访问 `/admin` 报 `missing session_key in cookie`

**现象**

直接访问 `http://localhost:8888/admin/#model-management` 报错 `missing session_key in cookie`。

**原因**

不是故障。后端 `SessionAuthMW` 中间件要求除登录/注册接口外必须携带 `session_key` Cookie，未登录直接访问 admin 页面会被拒。

**解决**

1. 先访问 `http://localhost:8888/sign`，用 `.env` 中 `ALLOW_REGISTRATION_EMAIL` 指定的邮箱（默认 `admin@qq.com`）注册，该邮箱注册后自动拥有管理员权限
2. 注册后自动登录，再访问 `/admin` 即可

**注意**：Cookie 按域名隔离，`localhost:8888` 和 `127.0.0.1:8888` 不要混用，否则会话丢失。

---

## 问题 5：登录后接口 500，报 `no such index [project_draft]`

**现象**

登录成功，但首页/项目列表接口报 500：

```
status: 404, failed: [index_not_found_exception], reason: no such index [project_draft]
POST /api/intelligence_api/search/get_draft_intelligence_list
```

**原因**

ES 容器启动时会执行 `setup_es.sh` 注册索引模板（`project_draft`、`coze_resource`）并创建索引。问题 1 中首次启动被中止时，初始化流程被连带打断，但 compose 命令里初始化脚本是后台子 shell 执行、`touch /tmp/es_init_complete` 与脚本用 `;` 分隔——**脚本失败也会创建就绪标记**，导致健康检查通过、索引实际未创建，留下隐患。

**解决**

手动重跑 ES 初始化脚本（幂等，已存在的模板/索引会跳过）：

```powershell
docker exec coze-elasticsearch sh -c "sed 's/\r$//' /setup_es.sh > /tmp/setup_fixed.sh && sh /tmp/setup_fixed.sh --index-dir /es_index_schema"
```

**验证**

```powershell
docker exec coze-elasticsearch curl -s "localhost:9200/_cat/indices?v&h=health,index,docs.count"
# 应看到 project_draft 和 coze_resource 两个 green 索引
```

之后刷新浏览器页面即可（前端会缓存之前的失败状态）。

---

## 问题 6（顺带修复）：Embedding 维度配置不一致

compose 硬编码 `OPENAI_EMBEDDING_DIMS: "2048"`，而 `.env` 中是 `1024`（与 `Qwen/Qwen3-Embedding-8B` 及 `OPENAI_EMBEDDING_REQUEST_DIMS=1024` 匹配）。随问题 2 一并修复，现统一以 `.env` 为准。

---

## 重新部署注意事项

执行 `docker compose -f docker-compose.postgres.yml down -v` 清空数据重新部署时：

1. 问题 1（ES 假失败）和问题 5（索引缺失）可能复现，按上面命令重跑 `up -d` 和 ES 初始化脚本即可
2. PostgreSQL 表结构（55 张表）由 postgres 容器首次启动自动执行 `volumes/postgres/schema.sql` 完成，无需人工干预
3. `--build` 只有首次或后端代码变更后才需要

## 验证清单（部署后）

```powershell
# 1. 所有容器 healthy
docker compose -f docker-compose.postgres.yml ps

# 2. PG 表已建（应为 55）
docker exec coze-postgres psql -U opencoze -d opencoze -c "SELECT count(*) FROM information_schema.tables WHERE table_schema='public';"

# 3. ES 索引已建
docker exec coze-elasticsearch curl -s "localhost:9200/_cat/indices?v&h=health,index"

# 4. 模型 key 已注入
docker compose -f docker-compose.postgres.yml config | Select-String 'MODEL_API_KEY_0'

# 5. 前端可访问
# 浏览器打开 http://localhost:8888/sign 注册登录
```
