# Python 复刻 Coze 的知识库与 Agent 设计参考

这份文档不按当前项目源码解释，而是抽象 Coze Studio 这类产品的核心结构、功能、实现方式和可学习设计点，目标是指导你用 Python 复刻一个简化版 Coze。

## 1. 产品本质

Coze 类产品本质是一个“AI 应用搭建平台”，核心能力不是单次聊天，而是把这些能力组合起来：

- 创建智能体：配置角色、模型、开场白、变量、工具、知识库。
- 管理知识库：上传文档、图片、表格，解析、切片、向量化、检索。
- 工具调用：插件、HTTP API、数据库查询、工作流。
- 对话运行：根据用户消息、历史上下文、知识库召回、工具结果生成回答。
- 发布和版本：草稿调试、发布版本、线上运行隔离。
- 多租户资源：用户、空间、项目、权限、资源列表。

如果用 Python 复刻，建议不要一开始追求完整 Coze，而是先实现：

```text
知识库 RAG + 智能体配置 + 流式对话 + 工具调用 + 发布版本
```

## 2. 推荐 Python 项目结构

推荐用 FastAPI + SQLAlchemy + Celery/RQ/Arq + PostgreSQL/pgvector 或 Milvus。

```text
app/
  main.py
  api/
    knowledge.py
    agent.py
    chat.py
    files.py
    tools.py
  core/
    config.py
    security.py
    logging.py
    errors.py
  db/
    session.py
    models.py
    migrations/
  services/
    knowledge/
      service.py
      parser.py
      chunker.py
      indexer.py
      retriever.py
      schemas.py
    agent/
      service.py
      runner.py
      prompt.py
      memory.py
      schemas.py
    tools/
      registry.py
      http_tool.py
      database_tool.py
      workflow_tool.py
    llm/
      client.py
      embedding.py
      rerank.py
    storage/
      minio.py
      local.py
    search/
      vector_store.py
      fulltext_store.py
  workers/
    celery_app.py
    knowledge_jobs.py
  frontend/ 或独立前端项目
```

最重要的边界：

- `api` 只处理 HTTP 请求响应。
- `services/knowledge` 只管知识库生命周期和检索。
- `services/agent` 只管智能体配置和运行。
- `services/llm` 统一封装模型、embedding、rerank。
- `workers` 处理耗时任务，例如解析、OCR、向量化。

## 3. 核心数据库设计

### 3.1 用户和空间

```sql
users(id, username, email, created_at)
spaces(id, name, owner_id, created_at)
space_members(id, space_id, user_id, role)
```

作用：资源隔离、多用户协作。

MVP 可以先只有一个默认空间。

### 3.2 知识库

```sql
knowledge_bases(
  id,
  space_id,
  name,
  description,
  type,          -- text/table/image
  status,        -- enabled/disabled/deleted
  created_by,
  created_at,
  updated_at
)
```

知识库只是一组文档的容器。真正被检索的是文档切片。

### 3.3 文档

```sql
knowledge_documents(
  id,
  knowledge_base_id,
  name,
  file_type,
  object_key,
  status,        -- uploading/indexing/enabled/failed
  parse_config,
  chunk_config,
  error_message,
  created_at,
  updated_at
)
```

文档状态建议：

```text
uploading -> indexing -> enabled
                    -> failed
```

### 3.4 切片

```sql
knowledge_chunks(
  id,
  knowledge_base_id,
  document_id,
  content,
  metadata,
  status,        -- pending/indexed/failed
  hit_count,
  created_at,
  updated_at
)
```

注意：切片 ID 最好和向量库里的 document ID 一致。这样删除、更新、命中统计都简单。

### 3.5 Agent

```sql
agents(
  id,
  space_id,
  name,
  description,
  draft_config,
  status,
  created_by,
  created_at,
  updated_at
)

agent_versions(
  id,
  agent_id,
  version,
  config_snapshot,
  created_at
)

agent_publishes(
  id,
  agent_id,
  version_id,
  channel,
  status,
  published_at
)
```

值得学习的设计：草稿和发布版本分离。

- 草稿用于编辑和调试。
- 发布版本用于线上稳定运行。
- 发布后用户再改草稿，不影响线上版本。

### 3.6 对话

```sql
conversations(id, agent_id, user_id, title, created_at, updated_at)
messages(id, conversation_id, role, content, metadata, created_at)
```

消息 metadata 可以存：

- 召回的知识库 chunks。
- 工具调用记录。
- token 使用量。
- 模型名称。

## 4. 知识库实现

### 4.1 知识库实现了什么

知识库模块要实现这些功能：

- 创建知识库。
- 上传文档、图片、表格。
- 解析文档内容。
- 文本切片。
- 图片 OCR 或图片描述生成。
- embedding 向量化。
- 写入向量库和全文检索库。
- 检索召回。
- 更新、删除、重新索引。
- 查询索引进度。

### 4.2 上传到索引的流程

推荐异步实现，不要在上传接口里同步完成解析和向量化。

```text
用户上传文件
  -> 保存到 MinIO/S3
  -> 创建 document，status=uploading
  -> 创建 index job
  -> worker 异步处理
  -> document status=indexing
  -> 解析文件
  -> 切片
  -> 生成 embedding
  -> 写 vector store
  -> 写 fulltext store
  -> chunks status=indexed
  -> document status=enabled
```

Python 伪代码：

```python
@router.post("/knowledge/{kb_id}/documents")
async def upload_document(kb_id: int, file: UploadFile):
    object_key = await storage.put(file)
    doc = document_repo.create(
        knowledge_base_id=kb_id,
        name=file.filename,
        object_key=object_key,
        status="uploading",
    )
    index_document.delay(doc.id)
    return {"document_id": doc.id}


@celery.task
def index_document(document_id: int):
    doc = document_repo.get(document_id)
    document_repo.update_status(doc.id, "indexing")

    raw = storage.get(doc.object_key)
    text_items = parser.parse(doc.file_type, raw)
    chunks = chunker.split(text_items, doc.chunk_config)

    for chunk in chunks:
        chunk_id = id_generator.next_id()
        chunk_repo.create(id=chunk_id, content=chunk.text, status="pending")
        vector = embedding.embed(chunk.text)
        vector_store.upsert(id=str(chunk_id), vector=vector, metadata={"document_id": doc.id})
        fulltext_store.upsert(id=str(chunk_id), content=chunk.text, metadata={"document_id": doc.id})
        chunk_repo.update_status(chunk_id, "indexed")

    document_repo.update_status(doc.id, "enabled")
```

### 4.3 为什么要异步索引

文档解析、OCR、embedding、向量库写入都可能很慢。异步化可以带来：

- 上传接口快速返回。
- 前端可以轮询进度。
- 失败可以重试。
- 消费者可以限流，避免模型和向量库被打爆。
- 大文件可以分批处理。

推荐加一张任务表：

```sql
index_jobs(id, document_id, status, retry_count, error_message, created_at, updated_at)
```

即使用 Celery，也建议保留任务表，便于排查和重试。

### 4.4 文本知识库

文本知识库处理流程：

```text
文件 -> 文本提取 -> 清洗 -> 切片 -> embedding -> vector store
```

切片策略：

- 按字符数切，例如 500-1000 字。
- 按段落切，保留自然语义。
- chunk overlap 50-150 字。
- 保存标题、页码、原文件信息。

可学习点：切片不要只看长度，要尽量保留语义边界。

### 4.5 图片知识库

图片知识库不是直接拿图片二进制做 RAG，通常是把图片转成可检索文本。

来源可以是：

- 文件名。
- 用户手动标签。
- OCR 文本。
- VLM 生成 caption。
- 业务分类标签。

推荐图片索引文本格式：

```text
文件名：坦克.png
图片描述：一辆主战坦克在道路上行驶，炮塔朝前。
OCR文字：...
标签：军事装备，坦克，履带车辆
```

图片知识库数据结构建议：

```sql
image_annotations(
  id,
  document_id,
  caption,
  ocr_text,
  manual_tags,
  generated_tags,
  updated_at
)
```

写入 chunk 时把这些字段拼成 `content`。

可学习点：

- 图片检索效果主要取决于 caption 质量。
- 如果 caption 为空，就算状态是成功，也很难召回。
- 图片返回给 Agent 时，最好同时返回图片 URL 和文本描述。

检索结果可以是：

```json
{
  "content": "图片描述：一辆坦克...",
  "image_url": "https://...",
  "document_name": "tanke.png"
}
```

### 4.6 表格知识库

表格知识库有两种实现方式。

简单版：每一行转成文本。

```text
商品=苹果, 价格=5元, 库存=100
商品=香蕉, 价格=3元, 库存=80
```

高级版：保留结构化表，并支持 NL2SQL。

```text
用户问题 -> LLM 生成 SQL -> 查询表格 -> 返回结果
```

复刻时建议先做简单版。NL2SQL 后面再做。

### 4.7 检索怎么做

一次完整检索推荐分几步：

```text
query
  -> 可选 query rewrite
  -> vector search
  -> fulltext search
  -> merge
  -> rerank
  -> min_score filter
  -> top_k
  -> 回表补全文档信息
```

三种检索策略：

- 语义检索：适合问法和文档表述不完全一致。
- 全文检索：适合关键词、编号、名称。
- 混合检索：推荐默认方式。

Python 接口设计：

```python
class KnowledgeRetriever:
    async def retrieve(self, query: str, kb_ids: list[int], top_k: int = 5):
        vector_hits = await self.vector_store.search(query, kb_ids, top_k=top_k * 3)
        text_hits = await self.fulltext_store.search(query, kb_ids, top_k=top_k * 3)
        merged = self.merge_hits(vector_hits, text_hits)
        reranked = await self.reranker.rerank(query, merged)
        return reranked[:top_k]
```

### 4.8 Rerank 是否必要

推荐做，但不要一开始依赖它。

原因：

- 向量库召回负责“粗筛”。
- rerank 负责“精排”。
- top_k 小的时候，rerank 能明显提高质量。

MVP 可以先不用 rerank，后面接：

- bge-reranker
- Cohere rerank
- OpenAI compatible rerank
- 自己用 LLM 打分，但成本高

## 5. Agent 实现

### 5.1 Agent 实现了什么

Agent 模块要实现：

- Agent 配置管理。
- 模型配置。
- Prompt/persona 管理。
- 绑定知识库。
- 绑定工具。
- 变量和记忆。
- 草稿调试。
- 发布版本。
- 流式对话。
- 工具调用循环。

本质上 Agent 是一个运行时编排器：

```text
用户输入 + 历史消息 + Agent 配置 + 知识库召回 + 工具能力 -> 模型回答
```

### 5.2 Agent 配置结构

推荐配置 JSON：

```json
{
  "name": "客服助手",
  "model": {
    "provider": "openai",
    "model": "gpt-4o-mini",
    "temperature": 0.5
  },
  "prompt": "你是一个专业客服...",
  "knowledge": {
    "enabled": true,
    "knowledge_base_ids": [1, 2],
    "top_k": 5,
    "min_score": 0.01,
    "search_type": "hybrid"
  },
  "tools": [
    {"type": "http", "name": "query_order"}
  ],
  "variables": {
    "company_name": "xxx"
  }
}
```

### 5.3 Agent 运行流程

不带工具的简单 RAG Agent：

```text
加载 Agent 配置
读取历史消息
检索知识库
组装 system prompt
调用 LLM stream
保存 assistant 消息
返回流式响应
```

Python 伪代码：

```python
class AgentRunner:
    async def stream_chat(self, agent_id: int, user_id: int, message: str):
        agent = await self.agent_repo.get(agent_id)
        history = await self.message_repo.list_recent(agent_id, user_id)

        retrieved = []
        if agent.knowledge.enabled:
            retrieved = await self.knowledge.retrieve(
                query=message,
                kb_ids=agent.knowledge.knowledge_base_ids,
                top_k=agent.knowledge.top_k,
            )

        prompt = self.prompt_builder.build(
            persona=agent.prompt,
            variables=agent.variables,
            retrieved_chunks=retrieved,
            history=history,
            user_message=message,
        )

        async for token in self.llm.stream(prompt):
            yield token
```

### 5.4 Prompt 设计

推荐 prompt 由几部分组成：

```text
System Persona
Agent Variables
Knowledge Context
Tool Instructions
Chat History
User Message
```

知识库上下文格式：

```text
以下是可能相关的知识库内容：

[知识 1]
来源：xxx.pdf 第 3 页
内容：...

[知识 2]
来源：xxx.png
图片： https://...
描述：...

请优先基于知识库回答。如果知识库不足，请说明无法确定。
```

可学习点：

- 不要直接把召回内容无格式塞给模型。
- 要告诉模型如何使用知识库。
- 要保留来源，方便回答引用和调试。
- 图片知识库要给图片 URL 和 caption。

### 5.5 自动 RAG 和工具 RAG 的区别

自动 RAG：每次用户提问都先检索知识库，再把结果塞进 prompt。

优点：

- 简单稳定。
- 用户无感。
- 适合大多数知识问答。

缺点：

- 每次都检索，可能浪费。
- 模型不能主动决定检索哪些知识库。

工具 RAG：把知识库检索做成工具，让模型决定是否调用。

优点：

- 更灵活。
- 适合多工具复杂 Agent。

缺点：

- 要求模型 function calling 能力好。
- 实现更复杂。
- 调用失败和循环控制更麻烦。

建议 Python 复刻顺序：

```text
先做自动 RAG，再做知识库工具。
```

### 5.6 工具调用实现

工具调用核心是一个循环：

```text
LLM 生成 tool_call
  -> 后端执行工具
  -> 工具结果追加到 messages
  -> 再次调用 LLM
  -> 直到模型输出最终回答
```

Python 伪代码：

```python
async def run_with_tools(messages, tools):
    for _ in range(MAX_TOOL_ROUNDS):
        resp = await llm.chat(messages=messages, tools=tools)
        if not resp.tool_calls:
            return resp.content

        messages.append(resp.assistant_message)
        for call in resp.tool_calls:
            result = await tool_registry.execute(call.name, call.arguments)
            messages.append({
                "role": "tool",
                "tool_call_id": call.id,
                "content": json.dumps(result, ensure_ascii=False),
            })

    raise RuntimeError("too many tool calls")
```

工具类型建议：

- HTTP API 工具。
- 数据库查询工具。
- 知识库检索工具。
- 工作流工具。

### 5.7 工作流工具

Coze 里 Workflow 可以作为 Agent 的工具。复刻时可以简化为：

```text
Workflow = 一个可配置 DAG
Agent Tool = 调用某个 workflow 并传入参数
```

MVP 可以先不做可视化 DAG，只做“后端注册函数”。

### 5.8 变量和记忆

变量分两类：

- Agent 配置变量：例如公司名、品牌名。
- 用户长期记忆：例如用户偏好、姓名、历史选择。

建议表：

```sql
agent_variables(agent_id, key, value)
user_memories(agent_id, user_id, key, value, updated_at)
```

变量进入 prompt：

```text
当前变量：
company_name = xxx
user_level = vip
```

### 5.9 流式输出

推荐用 SSE：

```http
Content-Type: text/event-stream
```

事件类型：

```text
message.delta
message.done
tool.call
tool.result
error
```

前端可以逐字显示，也能展示工具调用进度。

## 6. 发布与版本设计

这是 Coze 类产品很值得学的设计。

不要让线上 Agent 直接读草稿配置。

正确方式：

```text
draft_config   # 编辑态
publish        # 发布动作
version_config # 发布快照
runtime        # 运行时读取 version_config
```

优点：

- 线上稳定。
- 可以回滚。
- 可以审计谁发布了什么。
- 调试和生产隔离。

发布流程：

```text
用户点击发布
  -> 校验模型/工具/知识库是否可用
  -> 创建 agent_version
  -> 更新 agent_publish 当前版本
  -> 线上服务使用新 version
```

## 7. 文件和对象存储

所有上传文件都不要直接存在数据库。

推荐：

```text
MinIO/S3 保存原文件
DB 保存 object_key
需要访问时生成 signed URL
```

图片知识库召回时返回：

```json
{
  "image_url": "signed url",
  "caption": "...",
  "source": "xxx.png"
}
```

## 8. 任务和状态机

知识库最容易出问题的是状态不一致。

推荐明确状态机：

文档：

```text
uploading -> indexing -> enabled
uploading -> failed
indexing -> failed
enabled -> indexing # 重新索引
enabled -> deleted
```

切片：

```text
pending -> indexed
pending -> failed
indexed -> pending # 更新切片重新索引
indexed -> deleted
```

任务：

```text
queued -> running -> success
queued -> running -> failed -> retrying
```

可学习点：

- 状态更新要尽量和实际动作对应。
- 搜索库写入成功后再把 slice 标为 indexed。
- document 要等所有 slice 成功后再 enabled。
- 失败要记录 error_message。

## 9. 检索调试能力

复刻时一定要做调试接口，否则很难排查。

建议提供：

```text
POST /knowledge/retrieve_debug
```

返回：

```json
{
  "query": "用户问题",
  "rewritten_query": "改写后问题",
  "vector_hits": [],
  "fulltext_hits": [],
  "rerank_results": [],
  "final_chunks": []
}
```

这样可以明确问题发生在哪一层：

- 没有写入向量库。
- 向量召回不到。
- 全文召回不到。
- rerank 过滤掉。
- top_k 太小。
- Agent 没把结果放进 prompt。

## 10. 值得学习的设计点

### 10.1 异步索引

上传和索引解耦，是知识库系统必须有的设计。

### 10.2 统一切片模型

文本、表格、图片最终都转成 chunk，然后统一检索。

```text
Text -> Chunk
Table Row -> Chunk
Image Caption -> Chunk
```

这样 Agent 不需要关心原始文件类型。

### 10.3 搜索库 ID 和数据库 ID 对齐

向量库里的 ID 使用 `knowledge_chunks.id`，排查、删除、更新都简单。

### 10.4 混合检索

只做向量检索不够。关键词、名称、编号更适合全文检索。

默认推荐：

```text
hybrid = vector + fulltext + rerank
```

### 10.5 草稿和发布版本分离

这是平台型产品必须有的能力。

### 10.6 Agent Graph 思维

Agent 不只是一次 LLM 调用，而是多个节点的组合：

```text
变量 -> 知识库 -> 工具 -> Prompt -> LLM -> 后处理
```

Python 可以不用 Eino，但要保留这种“节点化”思想。

### 10.7 召回内容进入 Prompt 前要格式化

知识库结果要带来源、标题、内容、图片 URL。否则模型很难稳定使用。

### 10.8 命中统计

`hit_count` 很有用，可以用来判断知识库是否真的被使用。

## 11. 不建议一开始做的东西

复刻 MVP 阶段先别做：

- 完整 Workflow 可视化编排。
- 插件市场。
- 多渠道发布。
- 复杂权限体系。
- NL2SQL。
- 多模型管理后台。
- 复杂前端 IDE。

先把这条链路跑通：

```text
上传文件 -> 索引成功 -> Agent 绑定知识库 -> 提问 -> 召回 -> 回答
```

## 12. Python 技术选型建议

后端：

```text
FastAPI
SQLAlchemy 2.x
Alembic
Pydantic v2
Celery + Redis 或 Arq
PostgreSQL
pgvector 或 Milvus
Elasticsearch 或 OpenSearch
MinIO
```

LLM：

```text
OpenAI compatible client
litellm 可选
sentence-transformers 可选本地 embedding
```

文档解析：

```text
pypdf
pdfplumber
python-docx
openpyxl
pandas
pytesseract 或 PaddleOCR
```

图片 caption：

```text
多模态 LLM，例如 Qwen-VL、GPT-4o、Gemini、InternVL
```

前端：

```text
React + Ant Design/Semi Design
或先用简单管理后台
```

## 13. 推荐开发顺序

第一周：

```text
FastAPI 项目搭建
用户/空间简化
Agent CRUD
普通 Chat
```

第二周：

```text
知识库 CRUD
文件上传 MinIO
文本解析和切片
embedding + pgvector
检索接口
```

第三周：

```text
Agent 绑定知识库
RAG Prompt
流式回答
检索调试页面
```

第四周：

```text
图片知识库
OCR/VLM caption
图片召回展示
hit_count 统计
```

第五周：

```text
工具调用
HTTP API tool
发布版本
线上/草稿隔离
```

## 14. 最小闭环示例

最小闭环只需要实现这些接口：

```text
POST /knowledge-bases
POST /knowledge-bases/{id}/documents
GET  /documents/{id}/status
POST /knowledge-bases/{id}/retrieve
POST /agents
PUT  /agents/{id}
POST /agents/{id}/chat
```

Agent chat 内部：

```text
保存用户消息
检索知识库
组装 prompt
调用 LLM stream
保存助手消息
返回 SSE
```

这就是复刻 Coze 的核心骨架。
