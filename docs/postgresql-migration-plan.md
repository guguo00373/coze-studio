# PostgreSQL 迁移计划

## 当前状态

### Git 信息
- **恢复点 Commit**: `bc1c109a`
- **分支**: main
- **提交信息**: feat(helm): add K8s deployment charts for Coze Studio

### 回滚命令
```bash
git reset --hard bc1c109a
```

---

## 目标

将 Coze Studio 的数据库从内置 MySQL 切换到外部 PostgreSQL。

### 外部 PostgreSQL 信息
| 项目 | 值 |
|------|-----|
| 主机 | `postgres.ai-middleware.svc` |
| 端口 | `5432` |
| 超级用户 | `navigator` |
| 超级用户密码 | `ai-platform-pg-pwd-2026` |
| 现有数据库 | `platformservice`, `zitadel` |

---

## 问题分析

### 1. 后端代码只支持 MySQL

**证据**:
- `backend/infra/orm/impl/mysql/mysql.go` 硬编码使用 `gorm.io/driver/mysql`
- 环境变量固定为 `MYSQL_DSN`
- 整个代码库无 PostgreSQL 驱动

**需要修改的文件**:
- [x] `backend/infra/orm/impl/postgres/postgres.go` → 新建 postgres 版本
- [x] `backend/application/base/appinfra/app_infra.go` → 支持数据库类型切换
- [x] `backend/types/consts/consts.go` → 添加数据库类型常量
- [x] `go.mod` → 已有 PostgreSQL 驱动依赖 (gorm.io/driver/postgres v1.5.11)

### 2. Schema SQL 是 MySQL 语法

**MySQL 特有语法需要转换**:
| MySQL | PostgreSQL |
|-------|------------|
| `bigint unsigned` | `bigint` |
| `tinyint` | `smallint` |
| `json` | `jsonb` |
| `mediumtext` | `text` |
| `datetime(3)` | `timestamptz(3)` |
| `AUTO_INCREMENT` | `SERIAL` / `BIGSERIAL` |
| `ENGINE=InnoDB` | 删除 |
| `COMMENT 'xxx'` | `COMMENT ON COLUMN` |
| `INSERT ON DUPLICATE KEY UPDATE` | `INSERT ... ON CONFLICT DO UPDATE` |

**文件**: `helm/charts/opencoze-harbor-external-mysql/files/mysql/schema.sql` (778 行)

---

## 改造计划

### 阶段 1: 创建新分支
```bash
git checkout -b feat/postgresql-support
```

### 阶段 2: 修改后端代码
1. 添加 PostgreSQL 驱动依赖
2. 创建 `backend/infra/orm/impl/postgres/postgres.go`
3. 修改数据库初始化逻辑，支持环境变量切换数据库类型
4. 添加新的环境变量: `DATABASE_TYPE`, `DATABASE_URL`

### 阶段 3: 转换 Schema
1. 将 MySQL DDL 转换为 PostgreSQL 语法
2. 转换数据初始化 SQL（INSERT 语句）
3. 创建 PostgreSQL 用户和权限

### 阶段 4: 修改 Helm Chart
1. 创建新的 chart `opencoze-harbor-external-postgresql`
2. 移除 MySQL 相关模板
3. 添加 PostgreSQL 连接配置
4. 修改 coze-server 环境变量

### 阶段 5: 测试
1. 本地测试 PostgreSQL 连接
2. 部署到 K8s 测试
3. 验证所有功能正常

---

## 需要新建的数据库

```sql
-- 在外部 PostgreSQL 上执行
CREATE DATABASE opencoze;
CREATE USER opencoze_user WITH PASSWORD 'xxx';
GRANT ALL PRIVILEGES ON DATABASE opencoze TO opencoze_user;
```

---

## 进度记录

- [x] 2026-07-14: 提交当前 Helm Chart 代码 (bc1c109a)
- [x] 2026-07-14: 创建新分支 feat/postgresql-support
- [x] 2026-07-14: 修改后端代码支持 PostgreSQL
  - 创建 `backend/infra/orm/impl/postgres/postgres.go`
  - 修改 `app_infra.go` 支持 DATABASE_TYPE 环境变量
  - 添加数据库配置常量
- [x] 2026-07-14: 转换 Schema SQL (MySQL → PostgreSQL)
  - 创建 `files/postgres/schema.sql` (1116 行)
- [x] 2026-07-14: 创建新 Helm Chart
  - 创建 `opencoze-harbor-external-postgresql` chart
  - 移除 MySQL StatefulSet
  - 配置 PostgreSQL 连接
  - 更新部署文档
- [ ] 测试部署到 K8s

---

## 备注

- 保留 `opencoze-harbor-external-mysql` chart 作为参考
- 改造过程中如需回滚: `git reset --hard bc1c109a`
- PostgreSQL 版本的 Helm Chart: `helm/charts/opencoze-harbor-external-postgresql`
