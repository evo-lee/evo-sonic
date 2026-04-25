# Evo Sonic

[English](../README.md) | 中文

Evo Sonic 是一个基于 [go-sonic/sonic](https://github.com/go-sonic/sonic) fork 后继续演进的 Go 博客系统。项目保留了 Sonic 的轻量、高性能、主题化和多数据库能力，同时对运行时、依赖注入、后台 API、AI 内容能力、健康检查和本地开发体验做了重新整理。

> 当前 Go module 仍保持为 `github.com/go-sonic/sonic`，以兼容原项目包路径和已有代码结构。

## 项目定位

Evo Sonic 面向希望自托管博客、内容站点或个人知识库的场景。它提供文章、页面、日志、分类、标签、评论、附件、菜单、友链、主题和后台管理能力，并在此基础上加入可选的 AI 内容辅助能力。

## 主要特性

- 博客核心功能：文章、独立页面、日志、分类、标签、评论、附件、菜单、友链、统计和后台管理。
- 多数据库支持：SQLite3、MySQL、PostgreSQL，默认配置使用 SQLite3。
- Hertz HTTP 运行时：项目固定使用 CloudWeGo Hertz，不再提供旧运行时切换配置。
- GORM 数据访问：使用 GORM 和生成的 DAL 查询代码组织数据库访问。
- Uber FX 依赖注入：入口统一装配配置、日志、缓存、数据库、事件总线、服务和路由。
- 主题系统：内置 `default-theme-anatole`，支持后台上传、拉取、激活、编辑和配置主题。
- 附件存储：默认本地文件存储，同时保留 MinIO、Aliyun OSS 等对象存储实现。
- AI 内容能力：支持摘要、标签建议、润色、SEO 检查、流式接口、相关文章和向量嵌入等能力。
- AI Provider：支持 Anthropic、OpenAI 以及 OpenAI-compatible/Ollama 路径。
- MCP 接入：提供 MCP handler，使外部 Agent 可以调用部分博客能力。
- 运行健康检查：提供 `/ping`、`/health` 和 `/ready`。
- 安全与稳定性：包含 JWT 鉴权、CSRF 中间件、登录限流、AI 接口限流、请求超时、请求 ID、恢复中间件和结构化日志。

## 目录结构

```text
.
├── main.go                 # 应用入口和 FX 装配
├── conf/                   # 默认配置和开发配置
├── config/                 # 配置加载与路径初始化
├── handler/                # HTTP server、路由、中间件、后台 handler、MCP handler
├── service/                # 业务服务，包含 AI、存储、文章、主题、评论等模块
├── dal/                    # GORM 生成查询代码
├── model/                  # entity、dto、vo、param、property 等模型
├── event/                  # 事件总线和监听器
├── template/               # 模板渲染和模板扩展函数
├── resources/              # 内置后台静态资源和默认主题
├── util/                   # 通用工具
└── scripts/                # 数据库脚本和辅助脚本
```

## 环境要求

- Go 1.25+
- SQLite3 默认可直接使用
- 如使用 MySQL 或 PostgreSQL，需要自行准备数据库并修改配置
- 在 Windows 上构建 SQLite 相关能力时，需要可用的 GCC 工具链

## 快速开始

克隆项目：

```bash
git clone <your-repo-url> evo-sonic
cd evo-sonic
```

直接运行：

```bash
go run main.go
```

或指定配置文件：

```bash
go run main.go -config conf/config.yaml
```

默认监听地址为 `0.0.0.0:8080`。首次启动后访问：

```text
http://127.0.0.1:8080/admin#install
```

完成初始化后，后台地址为：

```text
http://127.0.0.1:8080/admin
```

## 构建

```bash
go build -o evo-sonic .
./evo-sonic -config conf/config.yaml
```

如果只想确认当前代码可以编译：

```bash
go build ./...
```

运行测试：

```bash
go test ./...
```

## 配置

默认配置文件位于 [conf/config.yaml](../conf/config.yaml)。未指定 `-config` 时，程序会读取 `./conf/config.yaml`。

常用配置项：

- `server.host` / `server.port`：HTTP 监听地址和端口。
- `sqlite3.enable`：是否启用 SQLite3，默认开启。
- `sqlite3.file`：SQLite 数据库文件路径，未配置时使用工作目录下的 `sonic.db`。
- `mysql.dsn`：MySQL 连接字符串。
- `postgres.dsn` 或 `postgres.host` 等分项配置：PostgreSQL 连接信息。
- `database.max_idle_conns` / `database.max_open_conns`：数据库连接池配置。
- `sonic.mode`：运行模式，`development` 会启用开发 CORS。
- `sonic.work_dir`：工作目录，用于存放数据库、日志、模板、上传附件等。
- `sonic.log_dir`：日志目录。
- `sonic.admin_url_path`：后台静态资源挂载路径，默认 `admin`。

数据库兼容优先级保持为：

```text
SQLite3 > MySQL > PostgreSQL
```

## AI 配置

AI 功能是可选能力。未配置 API Key 时，博客核心功能不依赖 AI。

配置来源有两类：

- YAML 启动默认值：`conf/config.dev.yaml` 中的 `ai.api_key` 和 `ai.model`。
- 后台运行时配置：通过 `/api/admin/ai/config` 保存到系统属性中。

运行时使用的主要属性包括：

- `ai_provider`：`anthropic`、`openai` 或 `ollama`。
- `ai_api_key`：对应 provider 的 API Key。
- `ai_model`：文本生成模型。
- `ai_base_url`：OpenAI-compatible endpoint，例如 Ollama。
- `ai_embedding_model`：向量嵌入模型，用于相关文章和语义能力。

仓库提供 [.env.ai.example](../.env.ai.example) 作为 AI 环境变量参考。

## 内置接口

基础检查：

- `GET /ping`：返回 `pong`。
- `GET /health`：进程存活检查。
- `GET /ready`：数据库就绪检查。

后台 API 默认位于：

```text
/api/admin
```

其中 AI 管理接口位于：

```text
/api/admin/ai
```

包含配置、摘要、标签建议、润色、SEO 检查、流式生成和相关文章等接口。

## 主题

默认主题已经包含在仓库中：

```text
resources/template/theme/default-theme-anatole
```

不需要额外初始化 git submodule。后台支持主题列表、上传、远程拉取、激活、配置、文件查看和编辑等操作。

## 与 go-sonic 的关系

Evo Sonic fork 自 `go-sonic/sonic`。本仓库会继续保留原项目中稳定的博客能力，并在当前代码基础上进行独立演进：

- 固定 Hertz 运行时。
- 整理配置和默认资源路径。
- 加入 AI 内容服务、嵌入服务和相关事件监听器。
- 加入 MCP handler。
- 增强健康检查、限流、请求超时和日志中间件。
- 保持默认主题和后台资源随仓库分发。

## License

本项目源代码基于 [MIT License](../LICENSE.md) 发布。
