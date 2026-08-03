# 应用测评与应用监测改造方案

## 目标

为 Coze Studio 增加两类能力：

- 应用测评：通过 API 批量调用智能体/工作流，评估回答质量、知识库命中、工具调用、延迟和失败率。
- 应用监测：持续统计线上应用运行数据，包括应用数、调用次数、Token、耗时、首包时长、满意度等。

本方案优先采用最小源码改造，先复用现有数据表，再补充必要埋点和监测表。

## 现有能力与数据源

当前 Coze 已有部分运行数据，可直接用于第一版统计。

| 数据源 | 可统计内容 |
| --- | --- |
| `single_agent_draft` / `single_agent_version` | 应用/智能体数量、应用配置 |
| `conversation` | 会话量、用户维度、应用维度 |
| `message` | 消息量、消息角色、消息状态、`run_id` |
| `run_record` | 每次智能体运行记录、状态、开始时间、完成时间、`usage` Token 信息 |
| `workflow_execution` | 工作流运行耗时、状态、Token、关联 `app_id` / `agent_id` |
| `node_execution` | 工作流节点级耗时、LLM 节点 Token |

现有数据可以支持：

- 应用总数
- 应用总调用次数
- 应用 Token 总量
- 平均调用时长
- 每应用对话量
- 每应用消息量
- 每应用失败率
- 工作流运行统计

现有数据不足以稳定支持：

- 平均首包时长
- 用户满意度
- 每次 LLM 调用明细
- 知识库/工具/插件的完整参与链路统计

## 应用测评方案

### 测评目标

通过 API 自动化调用应用，生成评测结果，支持以下指标：

- 回答准确性
- 知识库召回是否命中
- 工具/插件/工作流是否被调用
- 响应耗时
- 首包耗时
- Token 消耗
- 错误率
- 输出格式合规性

### 测评流程

```text
测试集 JSON/CSV
  -> 调用智能体 API 或工作流 API
  -> 保存请求、响应、耗时、Token、错误信息
  -> 自动评分或人工复核
  -> 输出测评报告
```

### 推荐测试集格式

```json
[
  {
    "case_id": "case_001",
    "app_id": "xxx",
    "agent_id": "xxx",
    "input": "帮我找一张现代灰色军舰在海上航行的图片",
    "expected_keywords": ["军舰", "海上", "052C"],
    "expected_knowledge_hit": true,
    "expected_tool_call": false,
    "tags": ["图片知识库", "召回测试"]
  }
]
```

### 测评结果表

建议新增表 `app_evaluation_result`。

```sql
CREATE TABLE app_evaluation_result (
  id BIGINT UNSIGNED NOT NULL PRIMARY KEY,
  eval_batch_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  case_id VARCHAR(128) NOT NULL DEFAULT '',
  app_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  agent_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  run_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  conversation_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  input MEDIUMTEXT NULL,
  output MEDIUMTEXT NULL,
  expected MEDIUMTEXT NULL,
  score DOUBLE NULL,
  passed BOOL NOT NULL DEFAULT 0,
  status VARCHAR(64) NOT NULL DEFAULT '',
  duration_ms BIGINT UNSIGNED NOT NULL DEFAULT 0,
  first_packet_ms BIGINT UNSIGNED NULL,
  input_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0,
  output_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0,
  total_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0,
  error_message TEXT NULL,
  extra JSON NULL,
  created_at BIGINT UNSIGNED NOT NULL DEFAULT 0,
  INDEX idx_batch(eval_batch_id),
  INDEX idx_agent_time(agent_id, created_at),
  INDEX idx_case(case_id)
);
```

### 最小实现方式

第一版可以不改核心 Agent 代码，只写一个独立测评脚本或服务：

```text
scripts/evaluation/
  run_eval.ts 或 run_eval.py
  cases.json
  report.md
```

脚本职责：

- 读取测试集
- 调用本地 Coze API
- 记录响应和耗时
- 从 API 响应或数据库补充 `run_id`、Token、状态
- 生成 Markdown/CSV/JSON 报告

## 应用监测方案

### 监测指标

总体统计：

- 应用总数
- 应用总调用次数
- 应用 LLM Token 总量
- 平均调用时长
- 应用平均首包时长
- 成功率 / 失败率

每个应用统计：

- 对话量
- 调用量
- Token 数
- 平均响应延迟
- 平均首包时长
- 用户满意度
- 错误数
- 错误率

### 分层统计设计

不要只在 LLM 调用层统计，因为一次 Agent 调用可能包含多次 LLM 调用、知识库召回、工具调用、工作流调用。

建议分两层：

```text
应用调用级统计 app_call_metrics
LLM 调用级统计 llm_call_metrics
```

第一期只实现 `app_call_metrics` 即可。

### 应用调用监测表

```sql
CREATE TABLE app_call_metrics (
  id BIGINT UNSIGNED NOT NULL PRIMARY KEY,
  app_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  agent_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  run_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  conversation_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  user_id VARCHAR(255) NOT NULL DEFAULT '',
  connector_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  status VARCHAR(64) NOT NULL DEFAULT '',
  start_at BIGINT UNSIGNED NOT NULL DEFAULT 0,
  first_packet_at BIGINT UNSIGNED NOT NULL DEFAULT 0,
  end_at BIGINT UNSIGNED NOT NULL DEFAULT 0,
  duration_ms BIGINT UNSIGNED NOT NULL DEFAULT 0,
  first_packet_ms BIGINT UNSIGNED NOT NULL DEFAULT 0,
  input_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0,
  output_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0,
  total_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0,
  workflow_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0,
  error_code VARCHAR(255) NOT NULL DEFAULT '',
  error_message TEXT NULL,
  extra JSON NULL,
  created_at BIGINT UNSIGNED NOT NULL DEFAULT 0,
  updated_at BIGINT UNSIGNED NOT NULL DEFAULT 0,
  INDEX idx_agent_time(agent_id, created_at),
  INDEX idx_app_time(app_id, created_at),
  INDEX idx_run(run_id),
  INDEX idx_conversation(conversation_id)
);
```

### LLM 调用监测表

第二期新增。

```sql
CREATE TABLE llm_call_metrics (
  id BIGINT UNSIGNED NOT NULL PRIMARY KEY,
  app_call_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  app_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  agent_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  run_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  model_id VARCHAR(128) NOT NULL DEFAULT '',
  model_name VARCHAR(255) NOT NULL DEFAULT '',
  provider VARCHAR(128) NOT NULL DEFAULT '',
  prompt_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0,
  completion_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0,
  total_tokens BIGINT UNSIGNED NOT NULL DEFAULT 0,
  start_at BIGINT UNSIGNED NOT NULL DEFAULT 0,
  end_at BIGINT UNSIGNED NOT NULL DEFAULT 0,
  duration_ms BIGINT UNSIGNED NOT NULL DEFAULT 0,
  status VARCHAR(64) NOT NULL DEFAULT '',
  error_message TEXT NULL,
  created_at BIGINT UNSIGNED NOT NULL DEFAULT 0,
  INDEX idx_app_call(app_call_id),
  INDEX idx_agent_time(agent_id, created_at),
  INDEX idx_run(run_id)
);
```

### 满意度反馈表

如果要统计满意度，需要补充反馈落库。

```sql
CREATE TABLE app_feedback (
  id BIGINT UNSIGNED NOT NULL PRIMARY KEY,
  app_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  agent_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  run_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  conversation_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  message_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  user_id VARCHAR(255) NOT NULL DEFAULT '',
  feedback_type TINYINT NOT NULL DEFAULT 0,
  detail_type VARCHAR(255) NOT NULL DEFAULT '',
  detail_content TEXT NULL,
  created_at BIGINT UNSIGNED NOT NULL DEFAULT 0,
  INDEX idx_agent_time(agent_id, created_at),
  INDEX idx_run(run_id),
  INDEX idx_message(message_id)
);
```

`feedback_type` 建议沿用现有 IDL：

```text
0 Default
1 Like
2 Unlike
```

## 最小源码改造点

### 新增监测组件

建议新增包：

```text
backend/domain/monitor/
backend/application/monitor/
```

核心接口：

```go
type Monitor interface {
    StartCall(ctx context.Context, event *StartCallEvent)
    FirstPacket(ctx context.Context, runID int64)
    EndCall(ctx context.Context, event *EndCallEvent)
}
```

业务代码只调用接口，不直接写 SQL。

### 打点位置 1：Agent 调用开始

在智能体运行入口创建监测记录：

```text
run_id
agent_id
conversation_id
user_id
connector_id
start_at
status=running
```

建议靠近 `run_record` 创建逻辑，避免新增独立 ID 关联困难。

### 打点位置 2：首包返回

在流式输出第一个 assistant chunk 时记录：

```text
first_packet_at
first_packet_ms = first_packet_at - start_at
```

可关注位置：

```text
backend/domain/agent/singleagent/internal/agentflow/callback_reply_chunk.go
```

### 打点位置 3：Agent 调用结束

在运行结束时更新：

```text
end_at
duration_ms
status
input_tokens
output_tokens
total_tokens
workflow_tokens
error_code
error_message
```

可以从 `run_record.usage` 或运行回调中的 usage 聚合得到。

## 统计 SQL 示例

### 应用总数

```sql
SELECT COUNT(*) AS app_count
FROM single_agent_draft
WHERE deleted_at IS NULL;
```

### 总调用次数

```sql
SELECT COUNT(*) AS total_calls
FROM app_call_metrics
WHERE created_at BETWEEN ? AND ?;
```

第一期未建表时可用：

```sql
SELECT COUNT(*) AS total_calls
FROM run_record
WHERE created_at BETWEEN ? AND ?;
```

### Token 总量

```sql
SELECT SUM(total_tokens) AS total_tokens
FROM app_call_metrics
WHERE created_at BETWEEN ? AND ?;
```

第一期未建表时可用：

```sql
SELECT SUM(JSON_EXTRACT(`usage`, '$.llm_total_tokens')) AS total_tokens
FROM run_record
WHERE created_at BETWEEN ? AND ?;
```

### 平均调用时长

```sql
SELECT AVG(duration_ms) AS avg_duration_ms
FROM app_call_metrics
WHERE created_at BETWEEN ? AND ?;
```

第一期未建表时可用：

```sql
SELECT AVG(completed_at - created_at) AS avg_duration_ms
FROM run_record
WHERE completed_at > 0
  AND created_at BETWEEN ? AND ?;
```

### 平均首包时长

```sql
SELECT AVG(first_packet_ms) AS avg_first_packet_ms
FROM app_call_metrics
WHERE first_packet_ms > 0
  AND created_at BETWEEN ? AND ?;
```

### 每应用使用分析

```sql
SELECT
  agent_id,
  COUNT(*) AS call_count,
  COUNT(DISTINCT conversation_id) AS conversation_count,
  SUM(total_tokens) AS total_tokens,
  AVG(duration_ms) AS avg_duration_ms,
  AVG(first_packet_ms) AS avg_first_packet_ms,
  SUM(status = 'success') AS success_count,
  SUM(status != 'success') AS fail_count
FROM app_call_metrics
WHERE created_at BETWEEN ? AND ?
GROUP BY agent_id;
```

### 每应用满意度

```sql
SELECT
  agent_id,
  COUNT(*) AS feedback_count,
  SUM(feedback_type = 1) AS like_count,
  SUM(feedback_type = 2) AS unlike_count,
  SUM(feedback_type = 1) / COUNT(*) AS satisfaction_rate
FROM app_feedback
WHERE created_at BETWEEN ? AND ?
GROUP BY agent_id;
```

## 实施阶段

### 阶段一：零改造统计

目标：快速出第一版监测报表。

工作内容：

- 基于 `run_record`、`conversation`、`message` 写 SQL 聚合。
- 输出总体统计和应用维度统计。
- 暂不支持首包时长和满意度。

### 阶段二：最小埋点

目标：补齐首包时长和更准确的调用级指标。

工作内容：

- 新增 `app_call_metrics` 表。
- 新增 `monitor` 组件。
- 在 Agent 开始、首包、结束三个点打点。

### 阶段三：满意度

目标：支持用户反馈和满意度分析。

工作内容：

- 新增 `app_feedback` 表。
- 增加消息点赞/点踩 API 或复用现有反馈结构。
- 前端接入点赞/点踩入口。

### 阶段四：LLM 明细

目标：支持模型调用级分析。

工作内容：

- 新增 `llm_call_metrics` 表。
- 在模型调用封装层记录每次 LLM 调用。
- 关联 `run_id` / `app_call_id`。

## 风险与注意事项

- 不要在主链路同步执行复杂统计，建议异步写入或轻量 upsert。
- 首包时长必须在流式输出第一个 chunk 时记录，结束后无法准确推导。
- 一次 Agent 调用可能包含多次 LLM 调用，应用级 Token 应聚合所有 LLM 调用和工作流 Token。
- 满意度不是运行数据，需要用户反馈入口和落库。
- 现有 `run_record.usage` 可先复用，但长期建议落到结构化字段，降低 JSON 查询成本。

## 推荐结论

最小可行路径：

```text
先用现有表做离线统计
  -> 新增 app_call_metrics
  -> 只在 Agent 开始/首包/结束三处打点
  -> 后续补 app_feedback 和 llm_call_metrics
```

这样改动最小、风险最低，同时能覆盖大部分应用测评和应用监测需求。
