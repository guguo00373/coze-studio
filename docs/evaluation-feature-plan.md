# Coze Studio 评测（Evaluation）功能实施计划

> 本文档是给后续实现 agent 的唯一参考。目标：在 **coze-studio 仓库内原生实现**一套精简的 LLM 评测子系统，评测对象支持 **智能体（agent）、工作流（workflow）、对话流（chatflow）**，**不包含 app**。不引入 coze-loop、不引入 RocketMQ / ClickHouse / FaaS，全部复用 coze-studio 现有基础设施（PostgreSQL + GORM + Redis + eino + Atlas 迁移 + 前端路由）。

---

## 1. 目标与范围

### 1.1 用户诉求

在 coze-studio 中：
1. 准备 **评测集**（一批带字段的测试数据，如问题+期望答案）
2. 准备 **评估器**（给模型/智能体的输出打分，v1 只做 prompt 型：用一个 LLM 按规则打分）
3. 新建 **实验**：选择评测集 + 选择评测对象（自己创建的 agent / workflow / chatflow）+ 选择评估器，运行后得到逐条结果与聚合分数

### 1.2 范围

| 项 | 支持 | 说明 |
|---|---|---|
| 评测对象 | ✅ agent（bot） | 走 `conversation.ConversationOpenAPISVC.OpenapiAgentRunSync` |
| 评测对象 | ✅ workflow | 走 `workflow.ApplicationService.OpenAPIRun` |
| 评测对象 | ✅ chatflow | 走 `workflow.ApplicationService.OpenAPIChatFlowRun` |
| 评测对象 | ❌ app | 本期不做 |
| 评估器 | ✅ prompt 型 | LLM 打分（复用 eino + modelmgr 配置的模型） |
| 评估器 | ❌ code 型 | 后续再做（JS/Python 脚本） |
| 执行方式 | ✅ 进程内 | 实验提交后在进程内并发跑，不引入消息队列 |
| 存储 | ✅ PostgreSQL | Atlas 迁移文件建表 |
| 前端 | ✅ 测评页面 | 评测集 / 评估器 / 实验三个页面 |

### 1.3 不做的事

- 不移植 coze-loop 的代码（它的评测模块强依赖 RocketMQ / ClickHouse / MinIO / FaaS / 自己的 workspace 权限模型，移植代价高且无必要）
- 不部署独立的 coze-loop 服务
- 不做 code 型评估器、不做 app 评测对象、不做在线实验（实时流量评测）、不做人工标注

---

## 2. 背景：coze-loop 的评测模型（概念借鉴，代码不移植）

coze-loop（`/opt/coze-loop`，可参考其设计但不拷贝实现）的核心概念是 **三元组**：

- **评测集（EvalSet）**：一个有 schema 的表格式数据集，每行是一个 item（如：`question` + `expected_answer`），每个 item 可含多个 turn（多轮对话）。
- **评测对象（Target）**：被测对象。coze-loop 支持 bot / workflow / prompt / agent 等。执行时把 item 的字段喂进去，拿到 **actual_output**（实际输出）。
- **评估器（Evaluator）**：打分器。输入 = 评测集字段 + actual_output，输出 = **score（分数）+ reasoning（理由）**。coze-loop 支持 prompt 型（LLM 打分）和 code 型。
- **实验（Experiment）**：把三者绑定起来跑。一个实验 = 评测集版本 + 评测对象版本 + 一组评估器 + 运行配置（并发度/重试）。结果有 item 粒度结果和聚合分数。

本方案在 coze-studio 里实现这个模型的最小子集，全部原生落地。

---

## 3. 关键决策记录

| 决策 | 选择 | 理由 |
|---|---|---|
| D1 实现方式 | 原生实现于 coze-studio | 不增加部署镜像、不引入新基础设施，贴合用户"系统不显繁重"的要求 |
| D2 评测对象 | agent / workflow / chatflow | 用户明确要求，不用 app |
| D3 评估器 | 仅 prompt 型 | code 型依赖 FaaS 运行时，v1 不做 |
| D4 执行引擎 | 进程内 goroutine + 并发限制 | 评测是小批量离线任务，用 HTTP 请求触发 + 内存执行，避免引入 MQ |
| D5 存储 | PostgreSQL + GORM | 与 coze-studio 现有 ORM 一致（`backend/infra/orm` 支持 postgres） |
| D6 建表 | Atlas 迁移 | coze-studio 统一用 `docker/atlas/migrations/*.sql` |
| D7 权限 | 按用户隔离（created_by） | 评测资源属于创建用户，不做跨空间分享（v1） |

---

## 4. 现状盘点（实现时可复用的已有能力）

> 以下路径均为实现时的精确参考，务必先读对应文件再动手。

### 4.1 后端技术栈

- Web 框架：Hertz（`github.com/cloudwego/hertz`），路由文件在 `backend/api/router/coze/api.go`（生成），手动路由挂在 `backend/api/router/register.go` 的 `//INSERT_POINT`。
- ORM：`gorm.io/gorm` + gorm/gen（`backend/infra/orm`），PostgreSQL driver 已支持（`backend/infra/orm/impl/postgres/postgres.go`）。
- DI：手写装配，入口在 `backend/application/application.go` 的 `Init()` 链条（`initBasicServices` → `initPrimaryServices` → `initComplexServices`）；跨模块访问用 `backend/crossdomain/<module>` 的 `DefaultSVC()/SetDefaultSVC()` 模式。
- 建表：Atlas，迁移 SQL 在 `docker/atlas/migrations/`，最新全量 schema 在 `docker/atlas/opencoze_latest_schema.hcl`。
- 用户上下文：`backend/application/base/ctxutil/session.go` 的 `GetUserSessionFromCtx(ctx)`。

### 4.2 评测对象调用入口（重点）

| 评测对象 | 调用入口（application 层） | 请求类型 |
|---|---|---|
| agent | `conversation.ConversationOpenAPISVC.OpenapiAgentRunSync(ctx, *run.ChatV3Request)`（`backend/application/conversation/openapi_agent_run.go:525`） | `backend/api/model/conversation/run` 的 `ChatV3Request`（`BotID`、`User`、`Stream=false`、`AdditionalMessages []*EnterMessage`、`Parameters`） |
| workflow | `workflow.ApplicationService.OpenAPIRun(ctx, *workflow.OpenAPIRunFlowRequest)`（`backend/application/workflow/workflow.go:1711`） | `OpenAPIRunFlowRequest`（`WorkflowID`、`Parameters *string`、`IsAsync`、`ExecuteMode`） |
| chatflow | `workflow.ApplicationService.OpenAPIChatFlowRun(ctx, *workflow.ChatFlowRunRequest)`（`backend/application/workflow/chatflow.go:509`） | `ChatFlowRunRequest` |

> HTTP 等价接口（便于联调）：`/v3/chat`（agent）、`/v1/workflow/run`（workflow）、`/v1/workflows/chat`（chatflow），对应 handler 在 `backend/api/handler/coze/workflow_service.go:864`、`workflow_service.go:1090`、`agent_run_service.go:91`。

### 4.3 LLM 能力（评估器打分用）

- coze-studio 基于 `github.com/cloudwego/eino` 调用大模型（见 `backend/application/workflow/workflow.go`、`backend/application/conversation/agent_run.go` 的用法）。
- 模型配置来源：modelmgr（`backend/application/modelmgr`）。评估器创建时让用户选一个已配置模型，保存 model_id/model_name，执行时构造 eino ChatModel 调用。
- 实现时参考 `backend/application/singleagent/init.go` 里如何用 eino 初始化模型客户端。

### 4.4 领域模块代码范式（照抄结构）

参考最简模块 `backend/domain/shortcutcmd/`：

```
backend/domain/shortcutcmd/
├── entity/                          # DO 实体
├── internal/dal/model/*.gen.go      # gorm/gen 生成的 model（手写后跑 gorm/gen 生成）
├── internal/dal/query/*.gen.go      # gorm/gen 生成的 query
├── internal/dal/dao.go              # DAO 封装（db 注入）
├── repository/repository.go         # repository 接口 + 实现
└── service/*.go                     # 领域服务（接口 + impl）
```

- application 层：`backend/application/<module>/`，`InitService(...)` 装配。
- handler 层：`backend/api/handler/coze/<module>_service.go`。
- 本模块的跨模块依赖（调 agent / workflow / LLM）走 `crossdomain` 或直接注入对应 application service。

### 4.5 前端路由

- 路由文件：`frontend/apps/coze-studio/src/routes/index.tsx`。
- 空间子菜单：`SpaceSubModuleEnum`（见 `frontend/apps/coze-studio/src/routes/index.tsx`，agent/app/plugin/workflow 等子菜单的注册方式）。
- 页面开发用 lazy + async-components（`frontend/apps/coze-studio/src/routes/async-components.tsx`）。
- 前端工程是 Rush monorepo（`frontend/packages/*`），新增页面包时参考 `frontend/packages/studio` 下既有包的结构；也可以直接放 `apps/coze-studio/src/pages`。

---

## 5. 总体架构

```
┌────────────────────────── coze-studio 前端 ──────────────────────────┐
│  路由 /space/:space_id/evaluation/...                                │
│  页面：评测集管理 | 评估器管理 | 实验列表 | 实验结果                    │
└──────────────┬───────────────────────────────────────────────────────┘
               │ HTTP /api/evaluation/...
┌──────────────▼───────────────────────────────────────────────────────┐
│  backend/api/handler/coze/evaluation_service.go   （Hertz handler）  │
│  backend/application/evaluation/                   （应用服务）        │
│  backend/domain/evaluation/                        （领域层）          │
│    ├─ eval_set：评测集 + item 管理                                    │
│    ├─ evaluator：评估器（prompt 型）管理 + 执行                        │
│    └─ experiment：实验管理 + 执行引擎（进程内并发）                     │
│         │  执行每个 item：                                           │
│         │   1. item 字段 → 拼评测对象入参（agent/workflow/chatflow）  │
│         │   2. 调用评测对象适配器 → actual_output                    │
│         │   3. 每个评估器：评测集字段 + actual_output → LLM → score    │
│         │   4. 写 experiment_item_result                            │
│         └  全部完成 → 聚合 → experiment_aggr_result                  │
└──────┬──────────────────┬──────────────────────┬────────────────────┘
       │                  │                      │
  PostgreSQL(GORM)   Redis(可选:实验状态/并发锁)   eino LLM（评估器打分）
   Atlas 迁移建表
```

---

## 6. 数据模型（PostgreSQL 表设计）

### 6.1 表清单

| 表名 | 用途 |
|---|---|
| `eval_set` | 评测集元数据 + schema |
| `eval_set_item` | 评测集条目（一行一条） |
| `evaluator` | 评估器（prompt 型） |
| `experiment` | 实验（绑定三元组） |
| `experiment_item_result` | 每条 item 的执行结果（含各评估器得分） |
| `experiment_aggr_result` | 实验聚合结果 |

### 6.2 DDL（PostgreSQL，字段全部 nullable/有默认，方便演进）

```sql
-- 评测集
CREATE TABLE eval_set (
    id           BIGSERIAL PRIMARY KEY,
    space_id     BIGINT,                       -- 所属空间（可为空，v1 按用户隔离）
    name         VARCHAR(255) NOT NULL,
    description  TEXT,
    schema_json  JSONB,                        -- 字段定义：[{"key":"question","type":"text","desc":"..."}]
    item_count   BIGINT NOT NULL DEFAULT 0,
    status       INT NOT NULL DEFAULT 0,       -- 0 草稿 / 1 可用
    created_by   VARCHAR(64) NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 评测集条目
CREATE TABLE eval_set_item (
    id         BIGSERIAL PRIMARY KEY,
    eval_set_id BIGINT NOT NULL,
    data_json  JSONB NOT NULL,                 -- 一行数据：{"question":"...","expected_answer":"..."}
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_eval_set_item_set ON eval_set_item(eval_set_id);

-- 评估器（v1 仅 prompt 型）
CREATE TABLE evaluator (
    id          BIGSERIAL PRIMARY KEY,
    space_id    BIGINT,
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    type        INT NOT NULL DEFAULT 1,        -- 1 prompt
    model_id    VARCHAR(255) NOT NULL,         -- 打分用的模型（modelmgr 配置）
    prompt      TEXT NOT NULL,                 -- 打分 prompt 模板，含变量占位
    temperature FLOAT NOT NULL DEFAULT 0,
    status      INT NOT NULL DEFAULT 0,        -- 0 草稿 / 1 可用
    created_by  VARCHAR(64) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 实验
CREATE TABLE experiment (
    id                  BIGSERIAL PRIMARY KEY,
    space_id            BIGINT,
    name                VARCHAR(255) NOT NULL,
    description         TEXT,
    eval_set_id         BIGINT NOT NULL,
    -- 评测对象：类型 + 标识
    target_type         INT NOT NULL,          -- 1 agent / 2 workflow / 3 chatflow
    target_id           VARCHAR(255) NOT NULL, -- agent: bot_id; workflow/chatflow: workflow_id
    target_config_json  JSONB,                 -- 额外参数（如 agent 的 user、workflow 的 parameters 模板）
    evaluator_ids       BIGINT[] NOT NULL,     -- 评估器 id 列表
    concurrency         INT NOT NULL DEFAULT 1, -- 并发条数
    status              INT NOT NULL DEFAULT 0, -- 0 待运行 / 1 运行中 / 2 成功 / 3 失败 / 4 部分成功
    run_stats_json      JSONB,                 -- {total, success, failed, running}
    error_msg           TEXT,
    started_at          TIMESTAMPTZ,
    finished_at         TIMESTAMPTZ,
    created_by          VARCHAR(64) NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 实验条目结果
CREATE TABLE experiment_item_result (
    id                 BIGSERIAL PRIMARY KEY,
    experiment_id      BIGINT NOT NULL,
    eval_set_item_id   BIGINT NOT NULL,
    status             INT NOT NULL DEFAULT 0,  -- 0 待运行 /1 成功 /2 失败
    input_json         JSONB,                   -- 喂给评测对象的入参快照
    actual_output      TEXT,                    -- 评测对象实际输出
    output_json        JSONB,                   -- 评测对象结构化输出（usage 等）
    target_error_msg   TEXT,
    -- 各评估器结果：{"<evaluatorId>":{"score":0.9,"reason":"..."}}
    evaluator_results_json JSONB,
    tokens_used_json   JSONB,                   -- {input_tokens, output_tokens}
    latency_ms         BIGINT,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_expt_item_result_expt ON experiment_item_result(experiment_id);

-- 实验聚合结果
CREATE TABLE experiment_aggr_result (
    id             BIGSERIAL PRIMARY KEY,
    experiment_id  BIGINT NOT NULL,
    -- 每个评估器的平均分/通过率
    evaluator_scores_json JSONB,               -- [{"evaluator_id":..,"avg_score":0.85,"pass_rate":0.9,"count":10}]
    total_item_count     BIGINT NOT NULL,
    success_item_count   BIGINT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_expt_aggr_expt ON experiment_aggr_result(experiment_id);
```

> **建表方式**：把上面 DDL 拆成一个 Atlas migration 文件放 `docker/atlas/migrations/`（参考既有 migration 的格式），并同步更新 `docker/atlas/opencoze_latest_schema.hcl`。提交后 CI / 部署会自动应用。**不要用 gorm AutoMigrate**。

### 6.3 GORM model

在 `backend/domain/evaluation/internal/dal/model/` 下按 gorm/gen 约定手写 model（`eval_set.gen.go` 等），再跑 `go generate` 生成 query（参考 `backend/domain/shortcutcmd/internal/dal/` 的生成产物结构）。

---

## 7. 后端模块实现

### 7.1 目录结构（遵循 coze-studio 范式）

```
backend/domain/evaluation/
├── entity/                     # DO：EvalSet / EvalSetItem / Evaluator / Experiment / ItemResult / AggrResult
├── internal/dal/model/         # gorm/gen model
├── internal/dal/query/         # gorm/gen query
├── internal/dal/dao.go         # DAO（注入 infra.DB）
├── repository/
│   ├── eval_set_repo.go        # 评测集 CRUD + item 分页
│   ├── evaluator_repo.go
│   └── experiment_repo.go      # 实验 CRUD + item 结果 + 聚合
└── service/
    ├── eval_set_service.go(+impl)
    ├── evaluator_service.go(+impl)      # 含 LLM 打分执行
    ├── target_adapter.go               # 评测对象适配器接口
    ├── target_adapter_agent.go         # agent 实现
    ├── target_adapter_workflow.go      # workflow 实现
    ├── target_adapter_chatflow.go      # chatflow 实现
    └── experiment_service.go(+impl)    # 实验编排 + 执行引擎

backend/application/evaluation/
├── eval_set_app.go
├── evaluator_app.go
├── experiment_app.go
├── evaluation.go               # InitService 装配
└── wire 装配（在 backend/application/application.go 的 init 链中挂上）

backend/api/model/evaluation/     # API DTO（struct，可不用 thrift）
backend/api/handler/coze/evaluation_service.go
```

### 7.2 DI 装配

- 在 `backend/application/application.go`：
  - `initBasicServices` / `initPrimaryServices` 之后新增 `evaluationSVC := evaluation.InitService(infra.DB, infra.IDGenSVC, deps...)`。
  - deps 包括：`infra.DB`、`infra.IDGenSVC`、评测对象调用需要的 `conversation.ConversationOpenAPISVC` 与 `workflow` 的 ApplicationService（复用 `complexServices` 里的实例，或通过 `crossdomain` 的 `DefaultSVC()` 拿）。
- 新增 `backend/crossdomain/evaluation/`（可选）：如果其他模块未来要用，提供 `DefaultSVC()`；v1 不需要，直接在 application 层引用。

### 7.3 路由注册

在 `backend/api/router/register.go` 的 `//INSERT_POINT` 之后追加一个 `RegisterEvaluationRouter(r)`（不要动 `router/coze/api.go` 生成文件），路由前缀：

```
POST   /api/evaluation/v1/eval_sets/create
GET    /api/evaluation/v1/eval_sets/list
GET    /api/evaluation/v1/eval_sets/detail
POST   /api/evaluation/v1/eval_sets/update
DELETE /api/evaluation/v1/eval_sets/delete
POST   /api/evaluation/v1/eval_sets/items/batch_create     # 支持批量导入
POST   /api/evaluation/v1/eval_sets/items/batch_delete
POST   /api/evaluation/v1/evaluators/create
GET    /api/evaluation/v1/evaluators/list
GET    /api/evaluation/v1/evaluators/detail
POST   /api/evaluation/v1/evaluators/update
DELETE /api/evaluation/v1/evaluators/delete
POST   /api/evaluation/v1/experiments/create
POST   /api/evaluation/v1/experiments/run
POST   /api/evaluation/v1/experiments/:expt_id/cancel
GET    /api/evaluation/v1/experiments/list
GET    /api/evaluation/v1/experiments/detail
GET    /api/evaluation/v1/experiments/result          # 分页 item 结果
GET    /api/evaluation/v1/experiments/aggr            # 聚合结果
GET    /api/evaluation/v1/targets/list                # 评测对象候选列表（agent/workflow/chatflow）
```

`/targets/list` 返回当前用户可选的：agent（bot 列表）、workflow 列表、chatflow 列表，分别复用对应 domain 的查询能力。

---

## 8. 评测对象适配器（关键设计）

统一接口（`target_adapter.go`）：

```go
type TargetType int // 1 agent / 2 workflow / 3 chatflow

type TargetAdapter interface {
    Type() TargetType
    // Run 同步执行一次评测对象，返回实际输出 + token 用量
    Run(ctx context.Context, target *entity.Experiment, item *entity.EvalSetItem,
        fieldMapping map[string]string) (*TargetRunResult, error)
}

type TargetRunResult struct {
    ActualOutput string            // 智能体回复 / 工作流输出
    ExtOutput    map[string]string // 结构化输出（可选）
    InputTokens  int64
    OutputTokens int64
    LatencyMS    int64
}
```

### 8.1 agent 适配器

- 构造 `run.ChatV3Request{ BotID: <target_id>, User: <固定评测账号或当前用户>, Stream: ptr(false), AdditionalMessages: [user 消息], Parameters }`。
- `fieldMapping`：评测集字段 → user 消息内容。默认把 `question` 字段作为消息；若评测集有多个字段，可拼接成一条 user 消息（如 `问题：{question}\n参考资料：{context}`）。
- 调用 `conversation.ConversationOpenAPISVC.OpenapiAgentRunSync`，从返回中取 `content`/`answer` 作为 `ActualOutput`，取 usage 的 token。
- 注意：评测用固定的 `user_id`（如 `eval_bot_<space_id>`），避免污染真实对话历史；`auto_save_history=false`。

### 8.2 workflow 适配器

- 构造 `workflow.OpenAPIRunFlowRequest{ WorkflowID: <target_id>, Parameters: <json>, IsAsync: false, ExecuteMode: "DEBUG"(可选) }`。
- `Parameters` 由 `fieldMapping` 生成 JSON（评测集字段 → 工作流入参变量）。
- 调用 `workflow.ApplicationService.OpenAPIRun`，输出取返回结果中的文本字段。

### 8.3 chatflow 适配器

- 构造 `workflow.ChatFlowRunRequest`，字段映射方式同 agent（消息 → 对话流）。
- 调用 `workflow.ApplicationService.OpenAPIChatFlowRun`，输出取回答文本。

---

## 9. 评估器实现（prompt 型）

### 9.1 数据结构

```go
type EvaluatorResult struct {
    EvaluatorID int64
    Score       float64 // 0~1
    Reasoning   string  // 打分理由
}
```

### 9.2 打分逻辑

1. 渲染打分 prompt：`evaluator.Prompt` 模板，替换变量：
   - `{{input}}` → 评测集 item 的字段（JSON）
   - `{{expected}}` → 期望答案（若评测集有该字段）
   - `{{output}}` → 评测对象的 `actual_output`
2. 用 eino 构造 ChatModel（根据 `evaluator.ModelID` 从 modelmgr 拿模型配置）。
3. 要求模型输出 JSON：`{"score": 0.85, "reasoning": "..."}`，解析后校验 `score ∈ [0,1]`。
4. 超时、解析失败、模型报错 → 该评估器记 fail，不影响整条 item 其它评估器。

### 9.3 模型配置来源

- 评估器创建/编辑时，前端调用现有 modelmgr 接口拉模型列表让用户选一个。
- 执行时 model_id → eino ChatModel 的初始化参考 `backend/application/singleagent/init.go`。

---

## 10. 实验执行引擎（进程内）

### 10.1 生命周期

```
experiment.create → status=0(待运行)
experiment.run   → status=1(运行中)
                   ├─ 分页加载 eval_set_item
                   ├─ 并发池（concurrency 上限，建议 ≤10）逐条处理：
                   │    item → 目标适配器.Run → actual_output
                   │         → 各评估器打分 → evaluator_results_json
                   │         → 写 experiment_item_result（成功/失败）
                   │         → 更新 run_stats_json
                   └─ 全部结束 → 聚合 → experiment_aggr_result → status=2/3/4
experiment.cancel → 置 status=3(失败)，终止剩余任务
```

### 10.2 实现要点

- **并发控制**：用 `errgroup.WithContext` + `semaphore`（或简单 worker 池），限制并发条数。
- **幂等**：`experiment.run` 只在 status=0 时可执行；运行中重复调用直接返回当前状态（用 Redis 或 DB 乐观锁 `UPDATE experiment SET status=1 WHERE id=? AND status=0`）。
- **失败策略**：单条 item 失败不中断整批，标记该条失败并继续；评估器失败只影响该评估器得分。
- **取消**：context cancel + 检查 `ctx.Done()`，中断剩余 item。
- **结果一致性**：item 结果逐条写库；聚合在全部 item 写完后计算。
- **不引入 MQ**：`run` 接口同步返回（如果实验大，可先返回 run_id 并异步 goroutine 执行，用 Redis 记录运行中状态；v1 建议直接异步 goroutine + 轮询状态，接口返回后前端轮询 detail/result）。

### 10.3 聚合逻辑

- 对每个评估器：`avg_score = Σscore / 有分条目数`，`pass_rate = 得分≥阈值(默认0.6)的条数/有分条目数`。
- `experiment_aggr_result` 记录每个评估器的聚合 + 总条数/成功条数。

---

## 11. API 接口清单

> 统一响应格式参考现有 handler：`{code, msg, data}`（见 `backend/api/internal/httputil`）。参数从 body 读 JSON；当前用户从 `ctxutil.GetUserSessionFromCtx` 取。

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/api/evaluation/v1/eval_sets/create` | 创建评测集（带 schema_json） |
| GET | `/api/evaluation/v1/eval_sets/list` | 分页列表 |
| GET | `/api/evaluation/v1/eval_sets/detail` | 详情（含 schema + item 分页） |
| POST | `/api/evaluation/v1/eval_sets/update` | 更新元数据/schema |
| DELETE | `/api/evaluation/v1/eval_sets/delete` | 删除 |
| POST | `/api/evaluation/v1/eval_sets/items/batch_create` | 批量导入 item（JSON 数组，含自动建 schema 或校验字段） |
| POST | `/api/evaluation/v1/eval_sets/items/batch_delete` | 批量删除 item |
| POST | `/api/evaluation/v1/evaluators/create` | 创建评估器（选模型 + 打分 prompt） |
| GET | `/api/evaluation/v1/evaluators/list` | 分页列表 |
| GET | `/api/evaluation/v1/evaluators/detail` | 详情 |
| POST | `/api/evaluation/v1/evaluators/update` | 更新 |
| DELETE | `/api/evaluation/v1/evaluators/delete` | 删除 |
| GET | `/api/evaluation/v1/targets/list` | 评测对象候选（agent/workflow/chatflow 各自一组） |
| POST | `/api/evaluation/v1/experiments/create` | 创建实验（选评测集+对象+评估器） |
| POST | `/api/evaluation/v1/experiments/run` | 运行实验（返回 run 状态） |
| POST | `/api/evaluation/v1/experiments/:expt_id/cancel` | 取消 |
| GET | `/api/evaluation/v1/experiments/list` | 分页列表（含 run_stats 状态） |
| GET | `/api/evaluation/v1/experiments/detail` | 实验详情 |
| GET | `/api/evaluation/v1/experiments/result` | item 结果分页（含每个评估器 score/reason） |
| GET | `/api/evaluation/v1/experiments/aggr` | 聚合结果 |

---

## 12. 前端页面

### 12.1 路由与菜单

- 在 `frontend/apps/coze-studio/src/routes/index.tsx` 的 `space/:space_id` 下新增子路由 `evaluation`，并在 `SpaceSubModuleEnum` 增加 `EVALUATION` 子菜单项（参考 `AGENT`/`WORKFLOW` 的注册方式）。
- 新增 lazy 组件：`frontend/apps/coze-studio/src/routes/async-components.tsx`。

### 12.2 页面清单

| 页面 | 路由 | 内容 |
|---|---|---|
| 评测集列表 | `space/:space_id/evaluation/dataset` | 列表 + 新建 + 删除 + 导入 |
| 评测集详情 | `space/:space_id/evaluation/dataset/:id` | schema 编辑 + item 表格（批量新增/删除） |
| 评估器列表 | `space/:space_id/evaluation/evaluator` | 列表 + 新建（选模型 + 写打分 prompt，支持变量占位说明） |
| 实验列表 | `space/:space_id/evaluation/experiment` | 列表 + 新建（选评测集 + 选评测对象 + 选评估器） + 运行/取消 |
| 实验结果 | `space/:space_id/evaluation/experiment/:id` | 进度/状态、聚合分数、item 结果表（含每评估器 score/reason）、失败原因 |

### 12.3 交互要点

- 新建实验三步：① 选评测集 → ② 选评测对象（agent/workflow/chatflow tab，分别调 `/targets/list`）→ ③ 选评估器 + 并发度。
- 实验运行中前端轮询 detail（2s）；完成后展示聚合与逐条结果。

---

## 13. 分阶段实施计划

### Phase 0：地基（半天）
- 读本文档「4. 现状盘点」列出的所有参考文件，确认调用入口签名。
- 建 `backend/domain/evaluation/`、`backend/application/evaluation/`、`backend/api/model/evaluation/` 骨架。

### Phase 1：数据层（1 天）
- 写 Atlas 迁移 SQL + 更新 `opencoze_latest_schema.hcl`。
- 写 gorm/gen model + DAO + repository（评测集 / 评估器 / 实验 CRUD）。

### Phase 2：评测集 + 评估器管理（1 天）
- 领域服务 + 应用服务 + handler + 路由。
- 评估器含"用 eino 跑一次打分"的单测（mock LLM 或本地模型）。

### Phase 3：评测对象适配器 + 实验执行（1~2 天）
- agent / workflow / chatflow 三个适配器（先各写一个 `Run`，用手工构造的 item 冒烟）。
- 实验执行引擎（并发、状态机、聚合）。
- `/targets/list`：复用 agent / workflow 的列表查询。

### Phase 4：前端（2 天）
- 路由 + 菜单 + 五个页面。

### Phase 5：联调与验收（1 天）
- 端到端：建评测集（导入 N 条）→ 建评估器 → 建实验选 agent → 运行 → 看聚合与逐条结果。
- workflow / chatflow 各跑一遍。

**预估总工期：5~6 人天。**

---

## 14. 风险与注意事项

1. **评测对象调用是同步长耗时**（agent 可能几十秒）：实验运行接口务必异步返回，前端轮询；并发度别默认开太大（建议 1~3，可配）。
2. **agent 的 `User` 字段**：评测要用独立账号或固定 user，避免把评测请求混进真实用户会话；`auto_save_history` 关掉。
3. **workflow `OpenAPIRun` 的 `ExecuteMode`**：默认"正式运行"，调试用 `"DEBUG"`；评测默认正式运行。
4. **eino 模型初始化失败**：评估器创建时做一次连通性校验（可选），执行时失败要落库不崩实验。
5. **打分 prompt 要约束输出 JSON**：在 prompt 里明确要求只输出 JSON；解析失败时该评估器记 fail，reason 写解析错误。
6. **PostgreSQL 迁移**：只在 `docker/atlas/migrations/` 加文件，不要用 AutoMigrate；本地开发用 `make` 或 atlas 命令应用迁移。
7. **权限隔离**：v1 所有资源按 `created_by` 过滤，跨用户不可见；不实现分享。
8. **gorm/gen 生成**：改 model 后跑生成命令，别手改 query 文件。
9. **不要动生成文件**：`backend/api/router/coze/api.go` 是 hertz 生成的，手动路由挂 `register.go` 的 INSERT_POINT。

---

## 15. 代码导航表（给实现者）

| 想做什么 | 看哪里 |
|---|---|
| 找 agent 同步调用入口 | `backend/application/conversation/openapi_agent_run.go:525`，请求模型 `backend/api/model/conversation/run` 的 `ChatV3Request` |
| 找 workflow 同步调用入口 | `backend/application/workflow/workflow.go:1711`，请求模型 `backend/api/model/workflow` 的 `OpenAPIRunFlowRequest` |
| 找 chatflow 调用入口 | `backend/application/workflow/chatflow.go:509`，请求模型 `ChatFlowRunRequest` |
| 看 DDD 模块代码范式 | `backend/domain/shortcutcmd/` |
| 看 DI 装配 | `backend/application/application.go`（init 链）+ `backend/crossdomain/` |
| 看 gorm/gen model/query 产物 | `backend/domain/shortcutcmd/internal/dal/` |
| 看 eino 初始化 LLM | `backend/application/singleagent/init.go`、`backend/application/workflow/workflow.go` |
| 看 Atlas 迁移 | `docker/atlas/migrations/`、`docker/atlas/opencoze_latest_schema.hcl` |
| 加后端路由 | `backend/api/router/register.go` 的 `//INSERT_POINT` |
| 加前端路由/菜单 | `frontend/apps/coze-studio/src/routes/index.tsx`、`async-components.tsx`、`SpaceSubModuleEnum` |
| 用户上下文 | `backend/application/base/ctxutil/session.go` 的 `GetUserSessionFromCtx` |
| 统一响应工具 | `backend/api/internal/httputil` |
| 概念参考（不移植） | `/opt/coze-loop` 的 evaluation 模块（评测集/评估器/实验三元组设计） |

---

> 本计划由人工评估产出，后续实现如遇与代码现状不符，以代码为准并回填更新本文档。

---

## 16. 实施进度（2026-08-06 回填）

> 以下记录实际落地情况，与上文计划不一致处以实际实现为准。**已完成：后端全链路（Phase 0-3）+ 数据库建表 + E2E 验证。**

### 16.1 已完成的后端实现

| 阶段 | 内容 | 状态 |
|---|---|---|
| Phase 0 骨架 | `backend/domain/evaluation/{entity, internal/dal/model, internal/dal, repository}` | ✅ |
| Phase 1 数据层 | 6 张表 model + DAO + repository（手写，不用 gorm/gen 生成） | ✅ |
| Phase 2 管理服务 | 评测集 / 评估器 / 实验 CRUD（domain service + application + handler） | ✅ |
| Phase 3 执行引擎 | agent/workflow/chatflow 三适配器 + eino LLM 打分 + 实验异步并发引擎 + 聚合 | ✅ |
| 数据库 | `docker/volumes/postgres/schema.sql` 追加 6 表 DDL，已应用到运行中的 coze-postgres 容器 | ✅ |
| DI 装配 | `backend/application/application.go` `initComplexServices` 挂 `evaluation.InitService` | ✅ |
| 路由 | `backend/api/router/coze/api.go` 新增 `/api/evaluation_api/*` 分组（20 个 handler） | ✅ |
| E2E | 镜像重建 + server 重启，用 admin@qq.com session 实测 CRUD 全通过 | ✅ |

### 16.2 实际实现与计划的差异（以代码为准）

1. **路由前缀**：计划 `/api/evaluation/v1/*`，实际为 `/api/evaluation_api/*`（`backend/api/router/coze/api.go:348` 新分组 `_evaluation_api`），沿用 coze/api.go 生成文件的 `_xxxMw()` 分组范式，未用 register.go 的 INSERT_POINT。
2. **建表方式**：计划 Atlas 迁移，实际直接写 `docker/volumes/postgres/schema.sql`（PostgreSQL 环境，Atlas 迁移是 MySQL 语法，不适配；schema.sql 由容器 bind-mount 为 init.sql，`CREATE TABLE IF NOT EXISTS` 幂等可重复执行）。
3. **gorm/gen**：生成工具连 MySQL，PostgreSQL 环境不可用；改为**手写 model（.gen.go 命名保留）+ 原生 gorm.DB 链式 DAO**（`db.Model().Where()`），不依赖 query 层。
4. **JSONB 列**：统一用 `gorm.io/datatypes` 的 `datatypes.JSON` 映射。
5. **target 适配器签名**：简化为 `Run(ctx, targetID string, inputJSON string) (*entity.TargetRunResult, error)`，不做 fieldMapping 参数（v1 直接整体入参，模板变量由评估器 prompt 渲染）。
6. **鉴权**：三个评测对象入口都经 `ctxutil.GetApiAuthFromCtx(ctx)` 取 userID/connectorID。进程内调用时用 `ctxcache.Store(ctx, consts.OpenapiAuthKeyInCtx, &authEntity.ApiKey{UserID: 当前用户ID, ConnectorID: 0})` 注入（与 `api/middleware/openapi_auth.go:154` 一致）。`connectorID==0` 时 agent 走 `GetSingleAgent`（正式版，跳过发布检查）、workflow 走 `checkUserSpace`。
7. **agent 输出提取**：`OpenapiAgentRunSync` 返回的 `ChatDetail` 只有元数据（无 content），实际回答通过 `MessageDomainSVC.GetByRunIDs(ctx, conversationID, [chatID])` 取最后一条 `MessageTypeAnswer` 消息的 `Content`。
8. **workflow/chatflow 输出**：workflow 取 `OpenAPIRunFlowResponse.Data`；chatflow 消费流式 `StreamReader`，匹配 `conversation.message.completed` 事件解析 `MessageDetail.Content`。
9. **评估器打分**：复用 `bizpkg/llm/modelbuilder.BuildModelByID(ctx, modelID, params)` 构造 eino ChatModel（自动适配 OpenAI/Ark/Claude/DeepSeek/Gemini/Qwen/Ollama），`Generate` 后 JSON 解析 `{"score":0-100,"reasoning":"..."}`。**score 范围按 0~100 实现**（不同于计划中的 0~1）。
10. **并发引擎**：用 `pkg/taskgroup.NewTaskGroup(ctx, concurrency)`（errgroup+信号量）控制并发；实验运行是异步 goroutine，接口立即返回，前端轮询。
11. **实验删除**：初始实现为不物理删除、置 `status=3(failed)`（避免悬空结果）；后续改为**物理删除 + 级联清理** `experiment_item_result` / `experiment_aggr_result`（见 16.5）。
12. **go.mod**：`gorm.io/datatypes`、`gorm.io/driver/postgres` 由 indirect 提升为直接依赖，并补齐 `github.com/jackc/pgx/v5` 等依赖。

### 16.3 后端 API 清单（实际）

所有接口前缀 `/api/evaluation_api`，Session 鉴权（cookie `session_key`）。

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/create_eval_set` | 创建评测集 |
| POST | `/update_eval_set` | 更新评测集 |
| POST | `/delete_eval_set` | 删除评测集 |
| GET | `/get_eval_set` | 评测集详情 |
| GET | `/list_eval_sets` | 评测集分页列表 |
| POST | `/add_eval_set_items` | 批量新增 item |
| POST | `/update_eval_set_item` | 更新单条 item（data_json） |
| POST | `/delete_eval_set_items` | 批量删除 item |
| GET | `/list_eval_set_items` | item 分页列表 |
| POST | `/create_evaluator` | 创建评估器 |
| POST | `/update_evaluator` | 更新评估器 |
| POST | `/delete_evaluator` | 删除评估器 |
| GET | `/get_evaluator` | 评估器详情 |
| GET | `/list_evaluators` | 评估器分页列表 |
| GET | `/list_evaluator_templates` | 内置评估器模板列表（14 个 prompt 模板，中英双语） |
| POST | `/parse_eval_set_file` | 解析导入文件（CSV/XLSX/XLS），返回表头 + 数据行 |
| POST | `/create_experiment` | 创建实验 |
| POST | `/update_experiment` | 更新实验 |
| POST | `/delete_experiment` | 删除实验 |
| GET | `/get_experiment` | 实验详情 |
| GET | `/list_experiments` | 实验分页列表 |
| POST | `/start_experiment` | 运行实验（异步，立即返回） |
| GET | `/get_experiment_detail` | 实验详情 + item 结果分页 |
| GET | `/list_targets` | 评测对象候选列表（当前支持 Agent，返回已发布 agent 的 id + 名称） |

> **待办（Phase 4 前端）**：~~路由/菜单 + 评测集/评估器/实验/结果五个页面；`/targets/list` 接口~~ 均已完成（见 16.4 / 16.5）。剩余：workflow/chatflow 评测对象下拉、草稿智能体评测。

### 16.4 前端完成状态（2026-08-06）

**已完成**
- 前端页面位于 `frontend/apps/coze-studio/src/pages/evaluation/`：
  - `types.ts`：接口与枚举（ExperimentStatus/ItemResultStatus/TargetType 等）
  - `api.ts`：fetch 封装（`/api/evaluation_api` 前缀，same-origin 鉴权）+ `getErrorMessage`
  - `constants.tsx`：状态/目标类型映射、`parseRunStats`/`parseEvaluatorResults`、`formatDateTime` 等
  - `hooks.ts`：自写 `useFetch`/`useMutation`（替代 ahooks，因 `@coze-studio/app` 不依赖 ahooks）
  - `EvaluationLayout.tsx`：评测区 Tab（评测集/评估器/实验）
  - `EvalSetListPage.tsx` / `EvalSetDetailPage.tsx`：评测集列表 + 创建/编辑 + item 批量添加/删除
  - `EvaluatorListPage.tsx`：评估器列表 + 创建/编辑（模型选择走 `/api/admin/config/model/list`）
  - `ExperimentListPage.tsx` / `ExperimentDetailPage.tsx`：实验列表/创建/运行 + 结果表格（运行中 3s 轮询）+ 明细弹窗
- 路由注册：`routes/async-components.tsx`（懒加载）+ `routes/index.tsx`（`/evaluation`、`/evaluation/sets/:id`、`/evaluation/experiments/:id`）
- 菜单：`packages/foundation/global-adapter/.../global-layout-composed/index.tsx` 新增「评测」（`layout_evaluation-button`）
- 部署：`docker build -f frontend/Dockerfile -t cozedev/coze-studio-web:latest .` + `docker compose up -d --force-recreate coze-web`（端口 8889）

**E2E 实测结果（curl + session_key）**
- 评测集/评估器/实验全部 CRUD 接口正常；模型列表需 admin 权限（已在 `kv_entries` 写入 `basic_config`，`admin_emails` 设为 `admin@qq.com`）
- 完整实验链路通过：创建评测集 → 加 items → 建评估器（model_id 100001）→ 建实验 → 运行 → status=2，item 全成功，打分 100/100，`run_stats_json` 含 `scores:{<evaluatorId>:{average,count}}`

**后端本轮修复的 3 个 bug（E2E 发现）**
1. 统计与 item 状态不一致：评估器失败把 item 降级为失败后，成功/失败计数未随之更新 → 计数改为按最终状态结算
2. agent 回答取不到：`OpenapiAgentRunSync` 在 run 创建后即返回（后台流式），答案消息尚未落库 → 改为轮询等待答案消息
3. 评估器输出解析脆弱 + 无占位符 prompt：
   - `parseEvaluatorOutput` 增加 `extractJSONObject`（在散文/```json 包裹中提取首个平衡 JSON 对象）
   - `renderPrompt` 当自定义 prompt 不含 `{{input}}/{{expected}}/{{output}}` 时自动追加 case 数据与 JSON 输出指令
   - `TargetRunResult.Input/Expected` 由 `parseCaseData` 从 item `data_json` 填充（runner 之前未填充）
   - token 用量从回答消息 `ext` 的 `input_tokens/output_tokens` 读取（run record 的 usage 在 run 完成前不可用）
- 新增单元测试 `evaluator_executor_test.go`（JSON 解析健壮性 + prompt 渲染），`go test ./domain/evaluation/...` 通过

**已知限制**
- 评测对象下拉（`/list_targets`）目前仅支持**已发布 Agent**；工作流/对话流类型暂无候选列表，实验表单回退为手填 ID（见 16.5）
- 草稿（未发布）智能体无法评测：评测执行复用对外发布链路 `OpenapiAgentRunSync`（`IsDraft=false`），且草稿试跑接口（`submit_bot_task`）后端未实现
- evaluator 的 model 选择依赖管理员权限的模型列表接口；非管理员需后端放开或配置默认模型

### 16.5 增量迭代记录（2026-08-06 之后）

> 前端通过 `scripts/dev-sync-backend.sh` / `scripts/dev-sync-frontend.sh` 快速迭代（编译产物直接 `docker cp` 进运行中容器，无需重建镜像；`dev-sync-frontend.sh` 用 tar 管道同步源码并覆盖 global-adapter / layout 包）。

**前端体验优化**
- 三个列表页（评测集/评估器/实验）新增按钮移到**右上角**，并重命名「新增××」
- 评测集创建/编辑表单改为 coze-loop 风格：**基本信息（名称 0/50、描述 0/200）+ 配置列**（动态列：名称/数据类型/必填/描述，默认 `input` + `reference_output`），`schema_json` 序列化为 `{field_schemas:[...]}`
- 评测集详情页每条 item 增加「编辑」（按配置列智能渲染输入框，无列时回退 JSON 文本）；新增「导入用例」入口
- 评估器表单模型选择显示**模型名**（如 `Qwen3.6-27B (100001)`），列表「模型」列同样显示名称
- 实验表单「评测对象」改为**下拉选择**（来自 `/list_targets`），无候选时回退手填
- 主菜单新增「配置管理」（`layout_admin-button`，`IconCozSetting`），点击整页跳转 `/admin#model-management`；为此给 `LayoutMenuItem` 增加 `onClick` 支持
- 修复前端表单 bug：coze-design 的 `Form` 是 Semi 实现，`formProps={{...}}` 不生效导致 `getFormApi` 未调用、点击确定无反应 → 改为直接传 `initValues/onValueChange/getFormApi`；`api.ts request()` 增加对响应体 `{code != 0}` 业务错误的检测（此前 HTTP 200 业务错误被误判为成功）

**后端新增/修复**
- **内置评估器模板**：`backend/domain/evaluation/entity/evaluator_template.go`（14 个 prompt 模板，参考 coze-loop `evaluator_template_conf`，中英双语），接口 `GET /list_evaluator_templates`
- **文件导入解析**：`backend/domain/evaluation/service/import_parser.go`（CSV 用 `encoding/csv`、XLSX 用 `xuri/excelize/v2`、XLS 用 `extrame/xls`，均已在 go.mod），接口 `POST /parse_eval_set_file`；前端导入弹窗做**列映射**（未配置映射的列不导入）+ 预览
- **item 更新**：新增 `POST /update_eval_set_item`（DAO/Repo/Service/Application/Handler 全链路）
- **实验真删**：`DeleteExperiment` 由置 failed 改为物理删除 + 级联清理 item/aggr 结果
- **评估器 prompt 可选**：`CreateEvaluator` 不再要求 prompt 非空（executor 自动用默认模板兜底）
- **评测对象列表**：`GET /list_targets`（当前仅 Agent，查 `single_agent_publish` 关联 `single_agent_version` 取名；PostgreSQL 需 GROUP BY 全列，避免 `p.created_at` 未聚合报错）
- `parseCaseData` 支持 `reference_output`（兼容旧 `expected`）
- 新增单元测试 `helpers_test.go`、`evaluator_executor_test.go`（renderPrompt 无占位符追加上下文等），`go test ./domain/evaluation/...` 通过
