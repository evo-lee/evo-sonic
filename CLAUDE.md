# [CLAUDE.md](http://CLAUDE.md)

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Sonic is a high-performance blogging platform written in Go, using Cloudwego Hertz as its HTTP framework, GORM for database access, and Uber FX for dependency injection.

## Common Commands

```bash
# Build
go build ./...

# Run
go run main.go

# Run tests
go test ./...
go test ./path/to/package/... -run TestName  # single test

# Lint
golangci-lint run --config=.golangci.yml
```

## Architecture

### Request Flow

```
HTTP Request
→ Hertz Router (handler/router.go)
→ Global Middleware (timeout, locale, request ID)
→ Route-specific Middleware (auth, CSRF, rate limit)
→ Handler (admin/ or content/)
→ Service (service/impl/)
→ DAL (dal/) via GORM
→ Response
```

### Key Layers

- **`handler/`** — HTTP handlers and routing
  - `handler/admin/` — Admin console API (authenticated endpoints)
  - `handler/content/` — Public-facing content API
  - `handler/middleware/` — Auth, CSRF, rate limiting, timeout, logging
  - `handler/web/` — Hertz adapter/abstraction (request/response helpers)
  - `handler/binding/` — Request parsing and validation
- **`service/impl/`** — Business logic implementations
- **`service/ai/`** — AI provider abstraction and content service; `impl/` subdirectory holds Anthropic, OpenAI, and Ollama backends plus the configurable factory
- **`dal/`** — Data access layer; GORM query generation (generated code lives here)
- **`model/`** — Data structures: `entity/` (GORM models), `dto/` (request/response), `vo/` (value objects), `param/`, `projection/`
- **`injection/`** — Uber FX dependency injection wiring
- **`template/extension/`** — Custom Go template functions for content rendering
- **`event/`** + **`event/listener/`** — Synchronous event bus for system events
- **`config/`** — Configuration structs and loading (Viper-based)
- **`resources/`** — Bundled static assets, admin HTML templates, default theme

### Database

Supports SQLite3 (default), MySQL, and PostgreSQL. GORM with generated query code in `dal/`. Context timeouts are enforced: 5s default, 10s+ for complex queries. See `CONTEXT_TIMEOUT_GUIDE.md` for patterns.

### Dependency Injection

All services, handlers, and repositories are wired via Uber FX in `injection/`. Avoid using `fx.Populate` (global mutable state anti-pattern — was removed). Add new components as FX `fx.Provide` entries.

### Security Patterns

- CSRF: `X-CSRF-Token` header required for all state-changing operations
- Auth: JWT middleware in `handler/middleware/`
- Rate limiting: login endpoint limited to 5 req/min
- Passwords: bcrypt-hashed (including category passwords)
- Path traversal: validate all file paths from user input before file operations

### Storage

Object storage abstraction in `service/storage/`. Supports MinIO, AWS S3, Aliyun OSS, Google Cloud. Local filesystem is the default.

### Configuration

Config files in `conf/config.yaml` (prod) and `conf/config.dev.yaml` (dev). Key settings: server port (default 8080), work directory (logs/DB/uploads/templates), database DSN.

AI features are configured at runtime via the admin API (`GET/POST /api/admin/ai/config`) and stored in the DB property system. The four properties are: `ai_provider` (anthropic/openai/ollama, default: anthropic), `ai_api_key`, `ai_model`, and `ai_base_url` (for OpenAI-compatible endpoints). The YAML `ai:` block in config files sets startup defaults for `api_key` and `model`; leave `api_key` empty to disable AI features entirely.

## gstack

Use the `/browse` skill from gstack for all web browsing. Never use `mcp__claude-in-chrome__*` tools.

## Skill routing

When the user's request matches an available skill, ALWAYS invoke it using the Skill
tool as your FIRST action. Do NOT answer directly, do NOT use other tools first.
The skill has specialized workflows that produce better results than ad-hoc answers.

Key routing rules:
- Product ideas, "is this worth building", brainstorming → invoke office-hours
- Bugs, errors, "why is this broken", 500 errors → invoke investigate
- Ship, deploy, push, create PR → invoke ship
- QA, test the site, find bugs → invoke qa
- Code review, check my diff → invoke review
- Update docs after shipping → invoke document-release
- Weekly retro → invoke retro
- Design system, brand → invoke design-consultation
- Visual audit, design polish → invoke design-review
- Architecture review → invoke plan-eng-review
- Save progress, checkpoint, resume → invoke checkpoint
- Code quality, health check → invoke health

Available skills: `/office-hours`, `/plan-ceo-review`, `/plan-eng-review`, `/plan-design-review`, `/design-consultation`, `/design-shotgun`, `/design-html`, `/review`, `/ship`, `/land-and-deploy`, `/canary`, `/benchmark`, `/browse`, `/connect-chrome`, `/qa`, `/qa-only`, `/design-review`, `/setup-browser-cookies`, `/setup-deploy`, `/retro`, `/investigate`, `/document-release`, `/codex`, `/cso`, `/autoplan`, `/plan-devex-review`, `/devex-review`, `/careful`, `/freeze`, `/guard`, `/unfreeze`, `/gstack-upgrade`, `/learn`
