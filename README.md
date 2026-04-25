# Evo Sonic

English | [中文](doc/README_ZH.md)

Evo Sonic is a Go blogging platform forked from [go-sonic/sonic](https://github.com/go-sonic/sonic) and evolved independently. It keeps Sonic's lightweight, high-performance, themeable, multi-database foundation while reorganizing the runtime, dependency injection, admin API, AI content features, health checks, and local development workflow.

> The Go module path is still `github.com/go-sonic/sonic` to preserve compatibility with the original package layout and existing imports.

## Positioning

Evo Sonic is built for self-hosted blogs, content sites, and personal knowledge bases. It includes posts, pages, journals, categories, tags, comments, attachments, menus, links, themes, statistics, and an admin console, with optional AI-assisted content workflows layered on top.

## Features

- Core blog management: posts, standalone pages, journals, categories, tags, comments, attachments, menus, links, statistics, and admin operations.
- Multiple databases: SQLite3, MySQL, and PostgreSQL. SQLite3 is enabled by default.
- Hertz HTTP runtime: the project uses CloudWeGo Hertz directly and no longer exposes a legacy runtime switch.
- GORM data access: database operations are organized around GORM and generated DAL query code.
- Uber FX dependency injection: configuration, logging, cache, database, event bus, services, and routes are assembled from one application entry point.
- Theme system: bundles `default-theme-anatole` and supports theme upload, fetching, activation, editing, and configuration from the admin console.
- Attachment storage: local filesystem storage by default, with MinIO and Aliyun OSS implementations retained.
- AI content features: summarization, tag suggestions, writing polish, SEO checks, streaming endpoints, related posts, and vector embeddings.
- AI providers: Anthropic, OpenAI, and OpenAI-compatible/Ollama paths.
- MCP integration: includes MCP handlers so external agents can call selected blog capabilities.
- Runtime checks: `/ping`, `/health`, and `/ready`.
- Safety and stability: JWT authentication, CSRF middleware, login rate limiting, AI endpoint rate limiting, request timeout, request ID, recovery middleware, and structured logging.

## Repository Layout

```text
.
├── main.go                 # Application entry point and FX wiring
├── conf/                   # Default and development configuration
├── config/                 # Configuration loading and path initialization
├── handler/                # HTTP server, routes, middleware, admin handlers, MCP handlers
├── service/                # Business services, including AI, storage, posts, themes, comments
├── dal/                    # Generated GORM query code
├── model/                  # entity, dto, vo, param, property, and related models
├── event/                  # Event bus and listeners
├── template/               # Template rendering and template extension functions
├── resources/              # Bundled admin assets and default theme
├── util/                   # Shared utilities
└── scripts/                # Database and helper scripts
```

## Requirements

- Go 1.25+
- SQLite3 works out of the box with the default configuration
- MySQL or PostgreSQL require a database instance and configuration changes
- Building SQLite-related functionality on Windows requires an available GCC toolchain

## Quick Start

Clone the project:

```bash
git clone <your-repo-url> evo-sonic
cd evo-sonic
```

Run directly:

```bash
go run main.go
```

Or specify a config file:

```bash
go run main.go -config conf/config.yaml
```

The default listening address is `0.0.0.0:8080`. On first startup, open:

```text
http://127.0.0.1:8080/admin#install
```

After installation, the admin console is available at:

```text
http://127.0.0.1:8080/admin
```

## Build

```bash
go build -o evo-sonic .
./evo-sonic -config conf/config.yaml
```

To verify that the current code compiles:

```bash
go build ./...
```

Run tests:

```bash
go test ./...
```

## Configuration

The default configuration file is [conf/config.yaml](conf/config.yaml). When `-config` is not provided, the application reads `./conf/config.yaml`.

Common settings:

- `server.host` / `server.port`: HTTP listen address and port.
- `sqlite3.enable`: enables SQLite3. Enabled by default.
- `sqlite3.file`: SQLite database file path. When omitted, `sonic.db` under the work directory is used.
- `mysql.dsn`: MySQL connection string.
- `postgres.dsn` or `postgres.host` and related fields: PostgreSQL connection settings.
- `database.max_idle_conns` / `database.max_open_conns`: database pool settings.
- `sonic.mode`: runtime mode. `development` enables development CORS.
- `sonic.work_dir`: work directory for the database, logs, templates, uploads, and related files.
- `sonic.log_dir`: log directory.
- `sonic.admin_url_path`: admin asset mount path. Defaults to `admin`.

The database compatibility priority remains:

```text
SQLite3 > MySQL > PostgreSQL
```

## AI Configuration

AI features are optional. Core blog functionality does not depend on AI when no API key is configured.

Configuration comes from two places:

- YAML startup defaults: `ai.api_key` and `ai.model` in `conf/config.dev.yaml`.
- Runtime admin configuration: saved through `/api/admin/ai/config` into system properties.

Main runtime properties:

- `ai_provider`: `anthropic`, `openai`, or `ollama`.
- `ai_api_key`: API key for the selected provider.
- `ai_model`: text generation model.
- `ai_base_url`: OpenAI-compatible endpoint, such as Ollama.
- `ai_embedding_model`: embedding model for related posts and semantic features.

See [.env.ai.example](.env.ai.example) for AI environment variable examples.

## Built-in Endpoints

Basic checks:

- `GET /ping`: returns `pong`.
- `GET /health`: process liveness check.
- `GET /ready`: database readiness check.

The admin API is mounted at:

```text
/api/admin
```

AI admin endpoints are mounted at:

```text
/api/admin/ai
```

They include configuration, summarization, tag suggestions, writing polish, SEO checks, streaming generation, and related-post endpoints.

## Themes

The default theme is bundled in this repository:

```text
resources/template/theme/default-theme-anatole
```

No git submodule initialization is required. The admin console supports theme listing, upload, remote fetching, activation, configuration, file browsing, and file editing.

## Relationship With go-sonic

Evo Sonic is forked from `go-sonic/sonic`. This repository keeps the stable blog capabilities from the original project and evolves independently on the current codebase:

- Uses Hertz as the fixed HTTP runtime.
- Organizes configuration and default resource paths.
- Adds AI content services, embedding services, and related event listeners.
- Adds MCP handlers.
- Improves health checks, rate limiting, request timeout, and logging middleware.
- Ships the default theme and admin resources with the repository.

## License

This project is released under the [MIT License](LICENSE.md).
