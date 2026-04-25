# AI Develop Plan — sonic-evo

> 基于 `ai-change` 分支现有进度，结合重构建议，制定的后续开发路线。
> 更新日期：2026-04-12

---

## 现状速览

### 已完成（阶段一部分）

| 模块                       | 文件                             | 状态              |
| ------------------------ | ------------------------------ | --------------- |
| Provider 抽象接口            | `service/ai/provider.go`       | ✅ 完成            |
| ContentService 接口        | `service/ai/content.go`        | ✅ 完成            |
| Anthropic provider       | `service/ai/impl/anthropic.go` | ✅ 完成（含流式）       |
| OpenAI / Ollama provider | `service/ai/impl/openai.go`    | ✅ 完成（含流式）       |
| 运行时可配置 provider          | `service/ai/impl/factory.go`   | ✅ 完成（DB 读取，热切换） |
| 配置属性定义                   | `model/property/ai.go`         | ✅ 完成（4 个配置项）    |
| Admin HTTP 接口            | `handler/admin/ai.go`          | ✅ 完成（7 个端点）     |
| 路由注册                     | `handler/router.go:311-318`    | ✅ 完成            |
| FX 注入                    | `injection/`                   | ✅ 完成            |

### 接口清单

```
GET  /api/admin/ai/config          读取配置（key 脱敏）
POST /api/admin/ai/config          保存配置
POST /api/admin/ai/summarize       生成摘要（同步）
POST /api/admin/ai/suggest-tags    建议标签（同步）
POST /api/admin/ai/polish          润色文章（同步）
POST /api/admin/ai/stream/summarize  生成摘要（SSE 流式）
POST /api/admin/ai/stream/polish     润色文章（SSE 流式）
```

---

## 阶段一：补全基础 AI 能力（当前阶段）

> 目标：阶段一后端 100% 完成，前端具备可用的 AI 侧边栏。

### 1.1 补全流式接口

- [ ] `POST /api/admin/ai/stream/suggest-tags` — 当前标签推荐只有同步版本，补充 SSE 流式版本
- [ ] handler 层统一错误格式：流式错误目前直接写 JSON，需与非流式 `xerr` 风格对齐

### 1.2 AI 调用安全护栏

- [ ] **Token 预算上限**：在 `configurableProvider` 中增加每次请求的 `MaxTokens` 硬上限（可配置，默认 4096），防止费用失控
- [ ] **单独限流**：在 `/api/admin/ai/` 路由组加专属 rate limiter（与业务流量隔离），参考现有 `handler/middleware/ratelimit.go`
- [ ] **noopProvider 错误信息优化**：当前错误只说"未配置"，改为提示具体配置路径

### 1.3 测试覆盖

- [ ] `service/ai/impl/` 单元测试（mock Provider 接口）
- [ ] `handler/admin/ai.go` handler 测试（验证入参校验、响应格式）
- [ ] `configurableProvider` 测试：覆盖 provider 切换逻辑和 noopProvider 降级

### 1.4 管理后台 AI 侧边栏（前端）

当前管理后台为 Go Template，渐进替换策略：

- [ ] **编辑器页面注入 AI 面板**（`resources/` 模板层）：不替换整个前端，仅在文章编辑器页侧边注入一个轻量 JS 组件
- [ ] 侧边栏功能：
  - "生成摘要" 按钮 → 调 `/stream/summarize`，实时显示结果
  - "润色内容" 按钮 → 调 `/stream/polish`，diff 展示修改
  - "推荐标签" 按钮 → 调 `/suggest-tags`，一键追加到标签输入框
- [ ] AI 配置页：在管理后台设置页新增 AI 提供商配置表单（调 `GET/POST /api/admin/ai/config`）

---

## 阶段二：知识化（语义搜索 + 向量化）

> 目标：文章内容可被 AI 语义检索，支撑推荐和问答场景。

### 2.1 向量化基础设施

- [ ] **选型决策**：
  - PostgreSQL 部署场景 → 启用 `pgvector` 扩展，复用现有连接
  - SQLite / 轻量部署场景 → 嵌入式向量库（如 `chromem-go` 纯 Go 实现）
  - 建议：通过接口抽象隔离，与 provider 模式一致

```go
// service/ai/embedding.go
type EmbeddingProvider interface {
    Embed(ctx context.Context, text string) ([]float32, error)
}

type VectorStore interface {
    Upsert(ctx context.Context, id string, vec []float32, meta map[string]string) error
    Search(ctx context.Context, vec []float32, topK int) ([]SearchResult, error)
}
```

- [ ] 新增 `service/ai/impl/embedding_openai.go`（复用 OpenAI Embedding API）
- [ ] 新增 `dal/vector/` — 向量存储实现（pgvector / 内存）

### 2.2 事件驱动向量化

利用现有事件系统，零侵入发布流程：

- [ ] 新增 `event/listener/embedding_listener.go`
  - 监听 `PostPublished` / `PostUpdated` 事件
  - 异步（goroutine）提取文章正文 → 调 EmbeddingProvider → 写入 VectorStore
  - 监听 `PostDeleted` → 从向量库删除对应向量

```go
type EmbeddingListener struct {
    embeddingProvider ai.EmbeddingProvider
    vectorStore       dal.VectorStore
}
```

### 2.3 语义搜索 API

- [ ] `GET /api/content/search/semantic?q=<query>&limit=10`
  - query → Embed → VectorStore.Search → 返回匹配文章列表
- [ ] `GET /api/admin/ai/related-posts?postId=<id>&limit=5`
  - 基于目标文章向量，检索最相似文章（用于编辑器"相关推荐"面板）

### 2.4 数据库 Schema

- [ ] 新增 `post_embeddings` 表（`post_id`, `model`, `vector`, `updated_at`）
- [ ] GORM model 定义 + DAL 查询生成

---

## 阶段三：Agent 化（MCP Server + 工作流）

> 目标：将 sonic-evo 接入 AI Agent 生态，成为可调用节点。

### 3.1 MCP Server

将博客核心操作暴露为 MCP Tools，让 Claude / Cursor 等 Agent 可直接操作：

- [ ] 新增 `handler/mcp/server.go` — MCP 协议实现（基于 JSON-RPC over HTTP/SSE）
- [ ] 新增 `handler/mcp/tools.go` — 工具注册：

```
create_post    创建草稿文章
update_post    更新文章内容
publish_post   发布文章
list_posts     列出文章（支持分页 / 状态过滤）
search_posts   全文 + 语义搜索
get_post       读取单篇文章（含 Markdown 原文）
list_tags      列出所有标签
```

- [ ] 新增 `handler/mcp/resources.go` — 暴露资源：`posts://`, `drafts://`, `tags://`
- [ ] MCP 鉴权：复用现有 JWT 中间件，`Authorization: Bearer <token>`
- [ ] 路由注册：`/mcp/v1/...`

### 3.2 AI 工作流引擎

- [ ] 新增 `service/ai/workflow.go` — 工作流定义接口
- [ ] 内置工作流实现：
  - **自动摘要**：新文章发布后自动生成并写入 `post.summary` 字段（若为空）
  - **自动翻译**（可选）：发布后触发翻译任务，创建多语言副本草稿
  - **评论情感分析**：新评论入库后异步分析情感，标记负面评论供人工审核
  - **SEO 建议**：发布前检查标题长度、摘要完整性、标签数量，返回建议列表
- [ ] 工作流调度：利用 Go `time.AfterFunc` / ticker 实现轻量 Cron，或接入 NATS（异步解耦）

### 3.3 多 Agent 协作示例

```
用户描述主题
  → Agent 调 MCP: search_posts（检索已有相关内容）
  → Agent 调 MCP: create_post（创建草稿）
  → Agent 调 /api/admin/ai/stream/polish（润色）
  → Agent 调 MCP: publish_post（发布）
```

---

## 技术债 & 横切关注点

### 异步任务基础设施

当前事件总线是同步的，AI 任务（向量化、翻译等）不能阻塞 HTTP 请求：

- [ ] 在 `event/` 层增加异步分发选项（goroutine 池，channel 缓冲，带超时 cancel）
- [ ] 或引入 NATS（轻量，单二进制，不依赖外部服务）用于 AI 任务队列

### 部署支持

- [ ] 新增 `docker-compose.ai.yml`：包含 sonic-evo + pgvector + Ollama（本地推理），一键启动完整 AI 栈
- [ ] `.env.ai.example`：列出所有 AI 相关环境变量（`AI_PROVIDER`, `AI_API_KEY`, `AI_MODEL`, `AI_BASE_URL`）
- [ ] `conf/config.dev.yaml` 增加 AI 配置示例块

---

## 执行优先级

```
P0（当前 Sprint）
  ├─ 1.2 安全护栏（限流 + token 上限）
  ├─ 1.3 测试覆盖
  └─ 1.4 管理后台 AI 侧边栏

P1（下一 Sprint）
  ├─ 2.1 向量化基础设施选型 + 接口定义
  ├─ 2.2 EmbeddingListener
  └─ 2.3 语义搜索 API

P2（阶段三启动条件：阶段二稳定上线后）
  ├─ 3.1 MCP Server（tools + resources）
  ├─ 3.2 自动摘要 / SEO 检查工作流
  └─ 异步任务基础设施
```

---

## 关键原则（来自架构建议）

1. **渐进增强**：AI 功能作为可选层叠加，不动现有稳定路径
2. **接口优先**：所有 AI 能力通过 Go interface 隔离，provider 可热切换
3. **本地可用**：Ollama 路径确保敏感场景无需外部 API
4. **费用可控**：token 预算上限 + 独立限流是上线前的前置条件，不是可选项
5. **不重写**：现有 CSRF、JWT、bcrypt、多数据库适配等安全基础设施保留，AI 能力是叠加层

