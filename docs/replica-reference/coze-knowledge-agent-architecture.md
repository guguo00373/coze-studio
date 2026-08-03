# Coze Knowledge And Agent Architecture Reference

本文档按当前项目代码结构梳理 Coze Studio 的知识库和单 Agent 实现方式，作为后续复刻一个简化版 Coze 的工程参考。

代码参考基于当前仓库：

- 知识库主目录：`backend/domain/knowledge`
- Agent 主目录：`backend/domain/agent/singleagent`
- API Handler：`backend/api/handler/coze`
- 应用层：`backend/application`
- 基础设施：`backend/infra`

## 1. 总体分层

后端整体是典型 DDD 分层：

```text
api/handler
  -> application
    -> domain
      -> repository/interface
      -> internal/dal
      -> infra
```

关键职责：

- `api/handler/coze`：HTTP API 入口，负责请求响应对象转换。
- `application/*`：应用服务，处理权限、上下文、资源事件、跨领域编排。
- `domain/*`：核心业务逻辑，知识库、Agent、Workflow、Plugin 等都在这里。
- `infra/*`：基础设施，包括存储、搜索、向量库、文档解析、消息队列、关系库等。
- `crossdomain/*`：跨领域接口模型，降低 domain 之间的直接耦合。

复刻时建议保留这个分层，但可以先简化为：

```text
handler -> service -> repository -> infra
```

## 2. 关键基础设施

当前项目依赖：

- MySQL：业务元数据，知识库、文档、切片、Agent 草稿、发布记录等。
- Redis：缓存、进度条、临时索引状态。
- MinIO：文件和图片对象存储。
- Elasticsearch：全文检索。
- Milvus：向量检索。
- NSQ 或 RocketMQ：异步索引、资源事件。
- LLM Provider：模型调用、图片 caption、query rewrite、rerank、Agent 对话。

复刻 MVP 可以先用：

- PostgreSQL 或 MySQL 存元数据。
- 本地文件或 S3/MinIO 存文件。
- pgvector 或 Milvus 存向量。
- Redis Stream 或简单任务表代替 MQ。
- OpenAI compatible API 做 embedding/chat。

## 3. 知识库模块

### 3.1 目录结构

核心目录：`backend/domain/knowledge`

```text
domain/knowledge
  entity/                 # 领域实体：Knowledge、Document、Slice、Event、Strategy
  service/                # 知识库核心服务
    knowledge.go          # 知识库、文档、切片 CRUD
    event_handle.go       # 异步索引事件处理
    retrieve.go           # RAG 检索链路
    convert.go            # 文档、切片、搜索文档转换
    interface.go          # 领域服务接口和请求响应结构
  processor/              # 创建文档时的处理器
  repository/             # 仓储接口
  internal/dal            # GORM DAO、model、query
  internal/events         # MQ 事件构造
  internal/convert        # 表格、parser、RDB 转换
```

应用层：

```text
backend/application/knowledge
```

API 层：

```text
backend/api/handler/coze/knowledge_service.go
```

### 3.2 核心数据模型

知识库相关核心表：

```text
knowledge
knowledge_document
knowledge_document_slice
knowledge_document_review
```

`knowledge`：知识库元信息。

关键字段：

- `id`：知识库 ID。
- `name`：名称。
- `space_id`：空间。
- `app_id`：项目或应用绑定关系。
- `status`：知识库状态，启用后才参与检索。
- `format_type`：知识库类型，文本、表格、图片。

`knowledge_document`：知识库里的文档或图片。

关键字段：

- `id`：文档 ID。
- `knowledge_id`：所属知识库。
- `document_type`：文本、表格、图片。
- `uri`：MinIO/TOS 对象路径。
- `status`：文档状态。
- `parse_rule`：解析和切片策略。
- `table_info`：表格知识库元信息。
- `slice_count`、`size`、`char_count`：统计信息。

文档状态定义在 `domain/knowledge/entity/const.go`：

```text
-1 init
0 uploading
1 enable
2 disable
3 deleted
4 chunking
9 failed
```

`knowledge_document_slice`：切片。

关键字段：

- `id`：切片 ID，同时也是搜索库文档 ID。
- `knowledge_id`：知识库 ID。
- `document_id`：文档 ID。
- `content`：切片文本。图片知识库里是图片标注/OCR/caption 文本。
- `status`：切片状态。
- `hit`：命中次数。

切片状态定义在 `domain/knowledge/internal/dal/model/progress.go`：

```text
0 processing
1 done
2 failed
```

跨域模型里也有一套对外状态：

```text
SliceStatusInit = 0
SliceStatusFinishStore = 1
SliceStatusFailed = 9
```

复刻时要注意：文档状态和切片状态是两套概念。

### 3.3 创建知识库

入口大致是：

```text
HTTP API
  -> application/knowledge
    -> domain/knowledge/service.CreateKnowledge
      -> knowledgeRepo.Create
      -> publish resource event
```

核心代码：

- `application/knowledge/knowledge.go`
- `domain/knowledge/service/knowledge.go`

创建知识库只写元数据，不会马上创建向量库集合。当前代码注释里说明：向量库初始化由文档索引触发。

### 3.4 上传文档或图片

入口大致是：

```text
CreateDocument API
  -> application/knowledge
    -> domain/knowledge/service.CreateDocument
      -> processor.BuildDBModel
      -> processor.InsertDBModel
      -> send index_documents event
```

核心代码：

- `domain/knowledge/service/knowledge.go`
- `domain/knowledge/processor/impl/base.go`
- `domain/knowledge/internal/events/events.go`

文本、表格、图片共用文档模型，但处理细节不同。

图片知识库特殊点：

- 上传图片时会先创建一个空切片。
- 初始切片 `status=0`。
- 这个空切片用于承载后续图片标注文本。
- 图片索引时需要复用这个切片 ID，确保 MySQL slice ID 和向量库 document ID 一致。

这也是本次修复关注的点。

### 3.5 异步索引事件

索引由 MQ 异步处理，核心文件：

```text
domain/knowledge/service/event_handle.go
```

事件类型定义在：

```text
domain/knowledge/entity/event.go
```

主要事件：

```text
index_documents        # 批量文档索引入口，会拆成单文档事件
index_document         # 单文档解析、切片、入搜索库
index_slice            # 单切片更新后重新索引
delete_knowledge_data  # 删除搜索库数据
document_review        # 预览切片
```

`indexDocuments` 的职责：

```text
遍历 documents
  -> 为每个 document 发送 index_document event
```

`indexDocument` 的职责：

```text
校验知识库和文档可写
设置文档状态为 chunking
读取对象存储文件
调用 parser 解析
缓存解析结果
清理旧切片和旧搜索库数据
切片分批处理
写 MySQL slice
写 Elasticsearch
写 Milvus
设置 slice status=done
设置 document status=enable
更新文档切片统计
清理缓存
```

简化流程图：

```text
Upload Document
  -> MySQL document/slice init
  -> MQ index_documents
  -> MQ index_document
  -> parser.Parse
  -> schema.Document[]
  -> entity.Slice[]
  -> knowledge_document_slice
  -> searchstore.Store ES
  -> searchstore.Store VectorDB
  -> status done
```

### 3.6 文档解析和切片

解析入口：

```text
k.parseManager.GetParser(convert.DocumentToParseConfig(doc))
docParser.Parse(...)
```

相关代码：

- `domain/knowledge/internal/convert/parser.go`
- `infra/document/parser`

解析结果统一转为：

```go
[]*schema.Document
```

然后通过 `d2sMapping` 转成业务切片：

```text
domain/knowledge/service/convert.go
```

核心映射：

- 文本：`schema.Document.Content -> SliceContentTypeText`
- 表格：列数据转 `SliceContentTypeTable`
- 图片：图片 caption/OCR 文本转 `SliceContentTypeText`

### 3.7 搜索库字段映射

搜索库字段在 `service/convert.go`：

```text
fMapping   # 文档类型 -> 搜索字段定义
s2dMapping # Slice -> schema.Document
d2sMapping # schema.Document -> Slice
```

文本知识库字段：

```text
id
creator_id
document_id
text_content indexed
```

表格知识库字段：

```text
id
creator_id
document_id
indexed columns 或 compact text_content
```

图片知识库字段：

```text
id
creator_id
document_id
text_content indexed
```

图片不是按二进制图像向量检索，而是按图片生成的文本 caption/OCR/手动标注检索。

### 3.8 图片知识库逻辑

图片知识库当前行为：

1. 上传图片。
2. DB 创建 `knowledge_document`。
3. DB 创建一个空 `knowledge_document_slice`，`status=0`。
4. 异步索引读取图片。
5. parser 生成图片文本描述。
6. 更新已有图片 slice 的 `content`。
7. 将 slice 写入 ES 和 Milvus。
8. 设置 slice `status=1`。
9. 设置 document `status=1`。

关键代码：

- 图片预创建切片：`processor/impl/base.go`
- 图片索引复用 slice ID：`service/event_handle.go`
- 图片 slice 转搜索文档：`service/convert.go`
- 图片召回展示：`service/retrieve.go`

注意：如果图片 `content` 为空，代码会跳过写搜索库：

```go
if doc.Type == knowledge.DocumentTypeImage && len(ssDocs) == 1 && len(ssDocs[0].Content) == 0 {
    return nil
}
```

所以复刻时要决定：

- 是否必须手动标注。
- 是否自动 OCR。
- 是否调用 VLM 生成图片 caption。
- 空 caption 时是否允许索引文件名、标签、用户描述。

建议复刻版图片知识库至少存这些检索文本：

```text
文件名
用户手动标签
OCR 文本
VLM caption
业务分类标签
```

### 3.9 检索链路 RAG

核心文件：

```text
domain/knowledge/service/retrieve.go
```

入口：

```go
Retrieve(ctx, request)
```

检索流程：

```text
newRetrieveContext
  -> 过滤启用知识库
  -> 过滤 status=1 的文档
queryRewriteNode
  -> 可选，多轮对话改写 query
parallel
  -> vectorRetrieveNode
  -> esRetrieveNode
  -> nl2SqlRetrieveNode
  -> passRequestContext
reRankNode
  -> 合并多路召回
  -> rerank
  -> topK/minScore 过滤
packResults
  -> 回表查 slice/document/knowledge
  -> 组装 RetrieveSlice
  -> increment hit
```

检索策略：

```text
SearchTypeSemantic = 0 # 向量检索
SearchTypeFullText = 1 # 全文检索
SearchTypeHybrid   = 2 # 混合检索
```

关键过滤：

- 知识库必须启用。
- 文档必须 `status=1`。
- 搜索库召回结果必须能回表找到 slice。
- rerank 分数必须大于 `min_score`。
- 最终数量受 `top_k` 控制。

hit 字段：

```text
hit 是命中次数，不是启用状态。
只有 packResults 成功返回某个 slice 后，才会 hit + 1。
```

### 3.10 表格知识库

表格知识库同时使用：

- MySQL 元数据。
- RDB 物理表保存原始行列数据。
- ES/VectorDB 保存用于召回的文本字段。

关键代码：

- `service/rdb.go`
- `service/sheet.go`
- `service/retrieve.go` 的 `nl2SqlRetrieveNode`

检索时可以走：

- 向量/全文召回行。
- NL2SQL 查询表格物理表。

复刻 MVP 可以先不做 NL2SQL，只把表格每行转成文本切片。

### 3.11 知识库复刻建议

MVP 表结构：

```sql
knowledge(id, name, type, status, space_id, created_at, updated_at)
knowledge_document(id, knowledge_id, name, type, uri, status, parse_rule, created_at, updated_at)
knowledge_slice(id, knowledge_id, document_id, content, status, hit, metadata, created_at, updated_at)
```

MVP 服务：

```text
CreateKnowledge
UploadDocument
IndexDocumentJob
UpdateSlice
DeleteDocument
Retrieve
```

MVP 索引：

```text
text -> chunk -> embedding -> vector db
image -> OCR/VLM caption -> embedding -> vector db
table -> row text -> embedding -> vector db
```

MVP 检索：

```text
query -> embedding -> vector search -> topK -> DB hydrate -> prompt context
```

## 4. Agent 模块

### 4.1 目录结构

核心目录：

```text
backend/domain/agent/singleagent
```

结构：

```text
singleagent
  entity/                  # Agent 实体、发布实体、事件
  service/                 # Agent 草稿、发布、查询
  repository/              # 仓储接口
  internal/dal             # single_agent_draft/version/publish DAO
  internal/agentflow       # 运行时 Agent Graph
```

最关键的是：

```text
internal/agentflow/agent_flow_builder.go
internal/agentflow/agent_flow_runner.go
internal/agentflow/node_retriever.go
internal/agentflow/node_tool_knowledge.go
```

### 4.2 Agent 数据模型

核心表大致是：

```text
single_agent_draft
single_agent_version
single_agent_publish
```

职责：

- `single_agent_draft`：草稿配置。
- `single_agent_version`：发布版本快照。
- `single_agent_publish`：发布渠道和发布记录。

Agent 配置实体包含：

```text
Prompt/Persona
ModelInfo
Knowledge
Plugin
Workflow
Database
Variables
SuggestReply
```

复刻时可以先抽象为：

```json
{
  "id": "agent id",
  "name": "agent name",
  "prompt": "system prompt",
  "model": {},
  "knowledge_ids": [],
  "tools": [],
  "variables": {}
}
```

### 4.3 Agent 创建、编辑、发布

Agent 管理和运行入口主要在：

- `domain/agent/singleagent/service`
- `api/handler/coze/playground_service.go`
- `api/handler/coze/agent_run_service.go`

典型流程：

```text
创建 Agent 草稿
  -> 写 single_agent_draft
编辑 Agent 配置
  -> 更新 draft JSON
发布 Agent
  -> 生成 version
  -> 写 publish
  -> 发布资源事件
运行 Agent
  -> 读取 draft 或发布版本
  -> BuildAgent
  -> StreamExecute
```

### 4.4 Agent 运行时 Graph

Agent 运行时使用 Eino Graph/Chain 编排。

核心构建函数：

```go
BuildAgent(ctx, conf)
```

在 `agent_flow_builder.go` 里，主要节点：

```text
persona_render
prompt_variables
knowledge_retriever
knowledge_retriever_pack
tools_pre_retriever
prompt_template
react_agent 或 llm
suggest_graph 可选
```

Graph 边关系：

```text
START -> persona_render -> prompt_template
START -> prompt_variables -> prompt_template
START -> knowledge_retriever -> knowledge_retriever_pack -> prompt_template
START -> tools_pre_retriever -> prompt_template
prompt_template -> react_agent/llm -> END
```

如果有工具，使用 ReAct Agent：

```text
react.NewAgent
  -> ToolCallingModel
  -> ToolsNode
  -> ToolReturnDirectly
```

如果没有工具，直接调用 ChatModel。

### 4.5 Prompt 组装

Agent 的 prompt 不是单一字符串，而是多个来源合并：

```text
persona/system prompt
variables
knowledge retrieve result
tools pre retrieve result
chat history
user input
```

知识库召回结果会被格式化为：

```text
---
recall slice 1: xxx
---
recall slice 2: yyy
```

对应代码：

```go
PackRetrieveResultInfo
```

### 4.6 Agent 知识库调用

Agent 自动知识库检索代码：

```text
domain/agent/singleagent/internal/agentflow/node_retriever.go
```

流程：

```text
读取 Agent.Knowledge.KnowledgeInfo
  -> 提取 knowledge ids
  -> genKnowledgeRequest
  -> crossknowledge.DefaultSVC().Retrieve
  -> convertDocument
  -> PackRetrieveResultInfo
  -> 塞入 prompt placeholder
```

`genKnowledgeRequest` 会将 Agent 配置转成知识库检索策略：

```text
top_k
min_score
auto/on-demand
semantic/fulltext/hybrid
query rewrite
rerank
nl2sql
```

Agent 知识库工具调用代码：

```text
node_tool_knowledge.go
```

这个模式下，知识库作为工具暴露给模型，由模型决定是否调用。

### 4.7 Agent 工具体系

工具节点分散在：

```text
node_tool_plugin.go
node_tool_workflow.go
node_tool_database.go
node_tool_variables.go
node_tool_knowledge.go
```

工具来源：

- Plugin tool：调用插件能力。
- Workflow tool：把 workflow 包装成 tool。
- Database tool：数据库查询工具。
- Variable tool：读写 Agent 变量。
- Knowledge tool：按需检索知识库。

如果任意工具存在：

```text
isReActAgent = true
requireCheckpoint = true
```

并要求模型支持 function call：

```go
if modelInfo.Capability != nil && !modelInfo.Capability.GetFunctionCall() {
    return nil, fmt.Errorf("model does not support function call")
}
```

复刻 MVP 可先只做：

```text
knowledge retrieval as pre-context
plugin/function tools as OpenAI tools
```

### 4.8 Agent 流式执行

核心文件：

```text
agent_flow_runner.go
```

入口：

```go
StreamExecute(ctx, req)
```

流程：

```text
创建 executeID
创建 callback stream
如果包含 workflow tool，创建 workflow message pipe
注册 callbacks
如果需要 checkpoint，设置 checkpoint id
goroutine 执行 runner.Stream
返回 AgentEvent StreamReader
```

事件输出通过：

```text
callback_reply_chunk.go
```

支持：

- 普通模型 token 流。
- 工具调用事件。
- workflow 中间回答。
- 直接返回工具结果。
- 中断和恢复。

### 4.9 多模态输入处理

`agent_flow_runner.go` 里会预处理用户输入：

```text
ImageURL
FileURL
AudioURL
```

如果模型支持对应能力，会把本地 URL 转成 base64 或保留可访问 URL。

这和图片知识库不是同一件事：

- 图片知识库：图片作为知识源，被 caption/OCR 后用于检索。
- Agent 多模态输入：用户本轮消息带图片，直接给多模态模型理解。

### 4.10 Agent 复刻建议

MVP Agent 运行链路：

```text
LoadAgentConfig
  -> RetrieveKnowledge(query)
  -> BuildPrompt(system, history, retrieved_context, user_input)
  -> ChatModel.Stream
  -> SSE/WebSocket response
```

工具版：

```text
LoadAgentConfig
  -> Build tools schema
  -> ChatModel with tools
  -> Tool call loop
  -> Final answer
```

建议先实现两种模式：

1. 无工具普通 Chat Agent。
2. 带知识库 RAG 的 Agent。

再扩展：

3. Plugin tools。
4. Workflow tools。
5. Database tools。
6. 变量记忆。
7. 发布版本和渠道。

## 5. 知识库和 Agent 的交互链路

Agent 绑定知识库配置形如：

```json
{
  "knowledge": {
    "min_score": 0.01,
    "top_k": 5,
    "auto": true,
    "search_strategy": 2,
    "recall_strategy": {
      "use_rerank": true,
      "use_rewrite": true,
      "use_nl2sql": true
    },
    "knowledge_info": [
      {"id": "knowledge id", "name": "knowledge name"}
    ]
  }
}
```

运行时：

```text
User message
  -> AgentRunner
  -> knowledgeRetriever.Retrieve
  -> knowledge.Retrieve
  -> ES/Milvus/NL2SQL
  -> RetrieveSlice[]
  -> schema.Document[]
  -> packed knowledge prompt
  -> ChatModel/ReActAgent
  -> stream answer
```

如果图片知识库被召回，内容会类似：

```html
<img src="signed-url" data-tos-key="object-key">图片描述文本
```

是否展示图片取决于 Agent prompt 和前端渲染策略。模型拿到的是上下文文本和 `<img>` 标签，不一定会主动输出图片。

## 6. 复刻项目推荐框架

可以按下面结构实现一个简化 Coze：

```text
backend/
  api/
    handlers/
      knowledge_handler.go
      agent_handler.go
      chat_handler.go
  application/
    knowledge_app.go
    agent_app.go
  domain/
    knowledge/
      entity.go
      service.go
      retrieve.go
      indexer.go
      repository.go
    agent/
      entity.go
      service.go
      runner.go
      prompt.go
      tools.go
  infra/
    db/
    object_storage/
    vectorstore/
    fulltext/
    llm/
    mq/
    parser/
  cmd/server/main.go
```

数据库表：

```text
users
spaces
agents
agent_versions
agent_publishes
agent_conversations
agent_messages
knowledge
knowledge_documents
knowledge_slices
index_jobs
tool_configs
```

核心接口：

```text
POST /api/knowledge
POST /api/knowledge/:id/documents
GET  /api/knowledge/:id/documents
POST /api/knowledge/:id/retrieve
POST /api/agent
PUT  /api/agent/:id
POST /api/agent/:id/publish
POST /api/agent/:id/chat
```

## 7. 最小可行版本路线

第一阶段：知识库基础

```text
创建知识库
上传文本文件
切片
embedding
向量检索
```

第二阶段：Agent RAG

```text
创建 Agent
绑定知识库
聊天时召回 topK
拼 prompt
流式回答
```

第三阶段：图片知识库

```text
上传图片
OCR 或 VLM caption
图片描述入向量库
召回后返回图片 URL + 描述
```

第四阶段：工具调用

```text
OpenAI function calling
插件工具
数据库查询工具
工作流工具
```

第五阶段：生产化

```text
异步任务重试
索引状态机
权限
发布版本
审计日志
监控告警
多租户隔离
```

## 8. 复刻时最容易踩坑的点

1. 文档状态和切片状态不要混用。
2. 搜索库 document ID 必须和 DB slice ID 对齐。
3. 删除或重切片时必须同步删除 ES/Milvus 旧数据。
4. 图片知识库必须有可检索文本，否则不会召回。
5. `hit` 是命中次数，不是是否可用。
6. Agent 自动 RAG 和知识库工具调用是两种模式，不要混在一起。
7. `top_k` 太小和 `min_score` 太高都会导致看起来“检索不到”。
8. K8s 更新镜像时要改 tag，不要一直用 `latest` 且 `IfNotPresent`。
9. Harbor 推新镜像后要确认 Deployment 实际镜像 tag。
10. 多模态输入和图片知识库不是同一条链路。

## 9. 推荐默认参数

知识库检索：

```text
search_strategy = hybrid
top_k = 5
min_score = 0.01
query_rewrite = false for simple test
rerank = true after basic recall works
```

图片知识库：

```text
caption required = true
index fields = filename + manual tags + OCR + VLM caption
empty caption fallback = filename + user label
```

Agent：

```text
temperature = 0.3 - 0.7
stream = true
history window = recent 10 - 20 messages
knowledge context max length = 4k - 8k tokens
```

## 10. 当前项目对应关键文件清单

知识库：

```text
backend/domain/knowledge/service/interface.go
backend/domain/knowledge/service/knowledge.go
backend/domain/knowledge/service/event_handle.go
backend/domain/knowledge/service/retrieve.go
backend/domain/knowledge/service/convert.go
backend/domain/knowledge/processor/impl/base.go
backend/domain/knowledge/internal/dal/model/knowledge.gen.go
backend/domain/knowledge/internal/dal/model/knowledge_document.gen.go
backend/domain/knowledge/internal/dal/model/knowledge_document_slice.gen.go
backend/application/knowledge/knowledge.go
backend/api/handler/coze/knowledge_service.go
```

Agent：

```text
backend/domain/agent/singleagent/service/single_agent.go
backend/domain/agent/singleagent/service/single_agent_impl.go
backend/domain/agent/singleagent/service/publish.go
backend/domain/agent/singleagent/internal/agentflow/agent_flow_builder.go
backend/domain/agent/singleagent/internal/agentflow/agent_flow_runner.go
backend/domain/agent/singleagent/internal/agentflow/node_retriever.go
backend/domain/agent/singleagent/internal/agentflow/node_tool_knowledge.go
backend/domain/agent/singleagent/internal/agentflow/node_tool_plugin.go
backend/domain/agent/singleagent/internal/agentflow/node_tool_workflow.go
backend/domain/agent/singleagent/internal/agentflow/node_chat_prompt.go
backend/api/handler/coze/agent_run_service.go
backend/api/handler/coze/playground_service.go
```
