# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Commands

### Backend (Go)

```bash
go build -o http2smtp ./           # Build
make release-linux                  # Cross-compile for Linux AMD64
make release-osx                    # Cross-compile for macOS AMD64
go test ./...                       # Run all tests
go test ./util -run TestSchedule    # Run specific test
```

Run the server:
```bash
./http2smtp                                    # Default ports 12566 (mgmt) + 12567 (API)
./http2smtp -web_port 12566 -api_port 12567    # Custom ports
./http2smtp -storage_root /path/to/storage     # Custom storage
```

Docker:
```bash
cd docker && docker compose up -d
docker build -t http2smtp .
```

### Frontend (Vue 3 + Vite)

```bash
cd view && npm install       # Install deps
cd view && npm run dev       # Dev server (port 3000, proxies /api → localhost:12566)
cd view && npm run build     # Production build (runs vue-tsc + vite build)
cd view && npm run build:force  # Build skipping type check
```

## Architecture

HTTP2SMTP is an HTTP-to-SMTP gateway: configure SMTP servers via web UI, then send emails through HTTP API calls.

### Dual Server Setup

- **Management server** (port 12566, path prefix `/api`): Web UI + admin APIs, requires cookie-based auth
- **Public API server** (port 12567, no prefix): Only `/sendMail` endpoints, uses token-based auth

This separation lets you expose the email-sending API publicly while keeping admin controls private.

### Backend (Go + go-web-frame)

Built on a custom web framework (`go-web-frame`, local path: `/Users/cao/Documents/GitHub/go-web-frame`). Layered architecture:

- **`rest/`** — REST handlers. Each resource implements `Init(context)` to register routes. Protect routes with `.WithMeta(auth2.WithLogin())`.
- **`service2/`** — Business logic (TokenService, ScheduleService, LogService)
- **`model/`** — GORM models + `Config` struct (maps to `config.ini`). Tables: `t_user`, `t_SMTP`, `t_mail`, `t_token`, `t_schedule`, `t_log`
- **`auth/`** — Two auth mechanisms: `authentication.go` (cookie sessions, AES-CBC encrypted) and `token.go` (API token validation)
- **`smtp/`** — Email delivery via `wneessen/go-mail`
- **`schedule/`** — Cron-based scheduled tasks using `robfig/cron/v3`. `ScheduleRunner` syncs DB tasks every minute. Tasks can fetch external URLs and use templates.
- **`config/`** — Config loading: defaults → `config.ini` → CLI flags → env vars

### Frontend (Vue 3 + Element Plus + Vite)

- **`view/src/api/request.ts`** — Axios instance, baseURL `/api`, withCredentials, response interceptor unwraps `response.data` and handles 401 redirects
- **`view/src/store/auth.ts`** — Pinia store: login/logout, persists token/username/isAdmin to localStorage
- **`view/src/router/index.ts`** — Auth guard on routes with `meta.requiresAuth: true`
- **`view/src/locales/`** — vue-i18n with 4 languages (zh-cn, en, zh-tw, ja)
- **`view/src/components/`** — Reusable form dialogs and selectors (SmtpSelector, TokenSelector, RecipientSelector)
- **`view/vite.config.ts`** — Dev proxy: `/api` → `http://localhost:12566`, Element Plus auto-import, chunk splitting

### Key Patterns

- REST handlers use dependency injection via go-web-frame context helpers
- Auth filter `auth2.NewAuthenticationFilter(&auth.Authentication{})` applied globally to management API group
- API tokens can be passed as URL param (`token=...`) or in JSON body on the public API
- Scheduled tasks support "send on failure only" mode for monitoring use cases
- Frontend uses `crypto-js` for client-side signature calculation during login

## Development Workflow

### Full Stack

1. Start backend: `./http2smtp` (or `go build -o http2smtp ./ && ./http2smtp`)
2. Start frontend: `cd view && npm run dev`
3. Frontend dev server (port 3000) proxies `/api` requests to backend (port 12566)

### Kill backend process

```bash
lsof -ti :12567 | xargs -I {} kill -9 {}
```

## References

- Frontend feature reference: `/Users/cao/Documents/GitHub/d-mail-view`
- API feature reference: https://github.com/chuccp/http2smtp/tree/v0.2.10
- go-web-frame local path: `/Users/cao/Documents/GitHub/go-web-frame`
