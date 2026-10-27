# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Commands

### Backend (Go)

```bash
go build -o http2smtp ./        # Build
go build ./...                  # Compile-check everything
go vet ./...
go test ./...                   # Run all tests
go test ./util -run TestParserCron   # Run a single test
```

The go.mod carries `replace github.com/chuccp/go-web-frame => ../go-web-frame`, so the
sibling checkout is what actually gets compiled — editing it takes effect immediately,
and a version bump in go.mod alone changes nothing.

Run the server:

```bash
./http2smtp                                    # Defaults: 12566 (mgmt) + 12567 (API)
./http2smtp -web_port 12566 -api_port 12567    # Custom ports
./http2smtp -storage_root /path/to/storage     # Custom storage
```

**A port given by flag or environment means Docker** (`core.isdocker`): the container's run
command fixed it and its port mapping depends on it, so the UI may not move it. Both UIs
therefore only *display* the ports there — the wizard always, the settings page when
`isDocker` — and `RestartService` refuses to change them. The flag is written on every
start, true or false: it lives in the config file, and a stale `true` from an earlier
container run would otherwise lock the ports for good.

Port resolution is flag/env → `config.ini` → the 12566/12567 constants, read through
`model.GetConfig` so the INI's strings survive. `config.ini` matters here: the wizard and
the settings page both write it, and ignoring it used to mean a port saved in the UI was
silently dropped on the next start. The wizard echoes the ports back in `/dbInit` even
though it does not edit them, because that handler binds onto `model.DefaultConfig()` and an
omitted port would be persisted as 12566/12567.

Stop a running instance:

```bash
taskkill /F /IM http2smtp.exe                   # Windows
lsof -ti :12566 | xargs -I {} kill -9 {}        # macOS / Linux
```

Cross-compile and Docker:

```bash
make release-linux      # → build_assets/http2smtp
make release-osx        # → build_assets/http2smtp
cd docker && docker compose up -d
docker build -t http2smtp .
```

### Web UI (`webui/`, Go html/template)

The management UI is server-rendered and embedded in the binary (`webui/embed.go`
`//go:embed templates static i18n`) — there is no Node/pnpm in the build or the release.

```bash
go test ./webui    # includes the four-locale key-sync test
```

**Dev mode is the build default.** An ordinary `go build` / `go run` binary reads
templates and static assets from `webui/templates` and `webui/static` on every
request, so edits show up on a browser refresh — no rebuild, no restart. Run it from
the repository root (Init fails loudly otherwise). Release binaries are built with
**`-tags release`** (the release workflow and `make release-*` do) and serve the
embedded copies instead; `WEBUI_DEV=0` / `1` overrides either default. The locale
JSON is embedded in every build.

- Pages are registered in `webui/pages.go` (`Pages.Init` → `registerPageRoutes`); handlers
  live in `webui/handlers_*.go`. Page routes sit at the site root (`/login`, `/smtp`, …)
  and call `requireUser`/`requireAdmin` themselves — they deliberately do **not** carry the
  framework's `WithLogin` meta, whose 403 JSON is wrong for a browser navigation.
- Rendering: `webui/render.go`. Each page is its own template set (layout skeleton +
  page file defining `content`); handlers write HTML via `Pages.render` and return
  `nil, nil` so the framework converter leaves the response alone. Form saves redirect
  (PRG) with `?ok=<i18n key>`; failures re-render the form with the values echoed back.
- `webui/static/ui.js` is the hand-rolled component library: modal dialogs
  (alert/confirm/prompt), toast, loading mask, partial navigation (pjax), form dialogs
  and page dialogs. Menu/tab/links fetch the target page and swap only `.content`, so
  the fixed sidebar, header and tab bar stay put; GET filter forms and POST forms
  marked `data-pjax-form` also submit through fetch, so list searches and settings
  saves do not reload the shell. A response without the app shell (login redirect,
  setup, refusal page) falls back to a normal navigation; unmarked form posts stay
  full navigations (PRG).
- `select[data-picker]` is progressively enhanced into the Vue-style selector: a tag
  trigger with a × per tag, and a nested modal listing the options as radio (single) or
  checkbox (multiple) rows. The native select stays in the form — invisible but
  focusable — so required validation and FormData still work. Used by the token form
  (SMTP, recipients), the schedule form (token) and the SMTP send dialog (recipient).
- Add/edit forms are the existing server-rendered form pages opened through
  `data-form-dialog`: the page head is dropped, Cancel is appended to the form actions,
  a redirect means saved (close + toast + refresh the list), and a 400 re-render is
  shown inside the dialog. Read-only/action pages (token send, SMTP send, log detail)
  open through `data-page-dialog`; their scripts call `UI.dialogDone(btn, msg)` on
  success. The same form URLs still render standalone.
- `app.js` wires the shell — fixed viewport, collapsible sidebar (mobile off-canvas),
  breadcrumb, tab bar (localStorage, keyed by path; closing the active tab lands on a
  neighbour), nav highlight, logout confirmation, password visibility, clipboard
  fallback and the delete/trigger AJAX actions — and re-syncs after each swap through
  `UI.onNavigate`. Dialog button labels come from `data-ui-ok`/`data-ui-cancel` on
  `<body>`, so no copy is hardcoded in the library.
- Deletes and trigger-now hit the existing `/api/*` JSON endpoints, toast the optional
  `data-flash`, and refresh only the content area; log cleanup opens a modal with the
  days input, posts to `/api/log/cleanup` and toasts the deleted count.
- **Translations**: `webui/i18n/{zh-cn,zh-tw,en,ja}.json`, flat dot-keys, embedded. These
  files are the source of truth, using `{name}` placeholders; `webui/i18n_test.go` fails
  the build if the four fall out of key-for-key sync. Templates read them via
  `{{.L.T "key"}}` — never a template func, since funcs bind at parse time and cannot see
  the request's language.
- The wizard/settings persistence is shared with the JSON API through the
  `configapply` package (`ApplyDbInit`, `ApplyAdminInit`, `ApplyReSet`, `ProbeDb`); the
  form handlers and `rest.Set` both call it, so the two cannot drift.

## Architecture

HTTP2SMTP is an HTTP-to-SMTP gateway: configure SMTP servers through the web UI, then send
mail over HTTP from your own code.

### Dual server

- **Management server** (`manage.port`, default 12566, context path `/api`) — web UI plus
  admin APIs. Authenticated by a session cookie.
- **Public API server** (`api.port`, default 12567, no context path) — only `/sendMail`.
  Authenticated by a token from the `t_token` table, passed as a query/body param. No
  session cookie is involved, so the two auth paths never mix.

`rest/API` is the only REST type on the public server; everything else lives on the
management server. Keeping them separate is what lets you expose mail-sending publicly
while the admin surface stays private.

### Web UI hosting

The management UI is server-rendered html/template — no nginx, no reverse proxy, no
`web/` directory. `main.go` registers `webui.NewPages()` on the management group
alongside the JSON REST types; its routes live at the site root (`/login`, `/smtp`, …)
and the static assets under `/static`. Release builds embed everything through
`webui/embed.go`; local builds serve templates/assets from the working tree (dev mode,
see above). The old `manage.webPath` key is no longer read — the config struct still
carries the field so an existing `config.ini` parses, but nothing consumes it.

There is no `ContextPath` on the management group: it is a server-wide prefix in the
framework and would land on the page routes too. The `/api` prefix lives in the JSON
route registrations (`rest/*.go`) instead — the URLs callers see are unchanged, and the
public API group on its own port is untouched.

Auth split: `/api/*` routes carry `WithLogin` meta as before; page routes skip it (its
403 JSON is the wrong answer for a browser navigation) and gate in the handler through
`Pages.requireUser`/`requireAdmin`, redirecting to `/login` or rendering the refusal
page. A restart re-runs every `Rest.Init` on a fresh engine, which is why `webui.Pages`
parses its templates once per process (`sync.Once`) and only re-registers routes.

### Backend layout

| Path | Role |
|---|---|
| `main.go` | Wires both rest groups, models, services and the cron runner |
| `rest/` | JSON API handlers. `Init(context)` registers routes under `/api`; protect with `.WithMeta(auth2.WithLogin())` |
| `webui/` | Server-rendered management UI (templates, handlers, embedded static/i18n) |
| `configapply/` | Configuration-writing bodies shared by the JSON API and the web forms |
| `service/` | Business logic: `TokenService`, `ScheduleService`, `LogService`, `SmtpService`, `UserService`, `RestartService` |
| `model/` | GORM models plus the `Config` struct mapped to `config.ini` |
| `entity/` | Request/response DTOs and the log status constants |
| `auth/` | Session cookie (`authentication.go`) and the AES-CBC helpers (`token.go`) |
| `smtp/` | Delivery via `wneessen/go-mail`; `*2` suffixed functions are the live ones |
| `runner/` | `ScheduleRunner` — reloads scheduled tasks into cron every minute |
| `db/` | Builds the GORM connection from config (sqlite or mysql) |
| `util/` | String, time, template and HTTP helpers |

Tables: `t_user`, `t_SMTP`, `t_mail`, `t_token`, `t_schedule`, `t_log`.

### Sessions and authorisation

The session is an AES-CBC-encrypted `LoginUser` carried in an **HttpOnly cookie named
`user_token`**, validated against `t_user.salt`. `Salt` is a session *epoch*, not part of
password hashing (bcrypt carries its own salt): the token embeds the salt it was minted
with, and `auth.User` rejects it once the stored value no longer matches.

Salt is therefore rotated whenever sessions must end — **sign-out, password change, and
user disable**. That means signing out ends every session for that user, on every device.
`auth.User` also accepts the token via an `Authorization` header, which the browser never
uses but external callers can.

Ownership: non-admin users only see their own rows, filtered by `user_id`. Admin-only
endpoints re-check `IsAdmin` in the handler — `/reSet` included, since it rewrites the
database and port configuration and hot-swaps the live connection under the running
models. `PUT` handlers verify ownership and preserve the original owner, so an admin
editing someone else's row does not take it over. Row-level reads (`GET /smtp/:id`,
`/mail/:id`, …) answer `data: null` for a row the caller does not own rather than
reporting a permission error, so existence does not leak; the list endpoints are what
admins use to reach other users' rows, and those exclude nothing by default
(`adminOnly=true` narrows them to admin-owned rows).

Two references are checked as well, because resolving them by id alone would hand out
another account's credentials: a token may only point at the owner's own SMTP server and
its own contacts. And a token stops sending once its owner is disabled — the public API
has no session, so `TokenService` re-reads the owner (`t_user.is_use`) on every send.
Rows with `user_id = 0` predate token ownership; they are left alone rather than broken.

Recipients work the other way round from what the token form's "Allowed Recipients" label
suggests: the token's list is the *default* and the request's `recipients` are added to it
(`MergeRecipients`), so a request without recipients still sends to the configured list and
an address in both is sent to once. Making the list a hard whitelist would be a deliberate
change of semantics, not a bug fix.

Two routes are open until `core.init` is set and admin-only afterwards, because the setup
wizard needs them before any account exists: `/adminExists` (it returns the administrator's
username) and `/readSet` (the database settings). `/reSet` and `/restart` are admin-only
throughout. Passwords go through `util.ValidatePassword` (8 characters) on every path that
sets one, and any signed-in user can change their own password through `POST /changePassword`
— the salt rotation ends their other sessions, this browser's included.

Mail logs are per-owner: `LogService.Log` records the `user_id` it is given (the token
owner for a token send, the schedule owner for a scheduled call), because the log list
and the attachment download both filter on it. The admin case is the exception — the log
page is the instance's send record, so an admin reads every user's entries (the list
carries the same `adminOnly` filter as the other lists, and fills the owner name) and can
fetch their attachments; everyone else gets `data: null` on a row that is not theirs.

Nothing prunes `t_log` or the attachment cache on its own — cleanup is manual
(`DELETE /log/:id` and `POST /log/cleanup`, which takes `days`: 0 clears everything).
Both remove the attachments a deleted entry recorded, which is the only thing that keeps
the cache from growing forever. The bulk cleanup is scoped to the caller's own entries
even for an admin; deleting a single row is not, since that row was on screen.

### Restart

`POST /api/restart` (login + `IsAdmin`) restarts in place through `WebFrame.ReStart`, which
cancels the run context and lets the framework re-init and serve again in the same process.
The settings page's save-and-restart button saves through `/reSet` first, then calls it.

The catch: a restart re-runs the server groups `main.go` built, and those carry the listen
ports captured at startup — a static `Port()` would freeze the old value forever. So
`main.go` hands each group a `*web.ServerConfig` it keeps, and `service.RestartService`
rewrites `Port` on those two before triggering the restart — except under `core.isdocker`,
where those ports are the container's and stay put. It also refuses a change the app could
not come back up with (port in use, both ports equal), leaving the running instance alone;
the two must differ because groups sharing a port are merged onto one gin engine and the
public API would inherit the management server's auth filter. The response carries the ports
now in effect — the client's own origin may be gone by the time it arrives, which is what
the warning on the settings page reports.

### Configuration

Precedence is defaults → `config.ini` → CLI flags → environment variables.

> **INI values load as strings, and the decoder only weak-converts when the target is a
> struct.** Two consequences that have bitten this codebase:
>
> - Read config through `model.GetConfig()` / `model.CoreState()`. `IConfig.GetBoolOrDefault("core.init")`
>   only accepts a real bool, so against an INI-sourced `init = true` it silently returns the
>   default — which once left the setup guards dead and stopped the database from attaching
>   at startup. Pass the `*Config` itself, never `&cfg`: a `**Config` makes the decoder fall
>   back to `encoding/json`, which cannot coerce `"true"` → bool or `"12566"` → int.
> - A `[]string` field cannot be set from an INI file at all (`cannot convert string to
>   []string`), and a bad value crashes startup rather than reporting a config error.

## Web UI (`webui/`)

Server-rendered management UI — see "Web UI (`webui/`, Go html/template)" under Common
Commands for the layout, the render/auth rules, dev mode and the i18n setup. The four
locales (`zh-cn`, `en`, `zh-tw`, `ja`) must stay in key-for-key sync; `go test ./webui`
enforces it. Release builds (`-tags release`) embed templates, static assets and locale
JSON, so the shipped binary needs no companion files; a plain local build serves the
templates and assets from the working tree by design.

### Release packaging

The release workflow only sets up Go and runs `go build -tags release`; the tarball
carries the binary alone (the UI travels inside it). The Dockerfile unpacks that tarball
and needs no change — it never referenced `web/` directly.

## Known gaps

- **Signing out ends every session for that user**, on all devices — it rotates the session
  salt (see above). Per-device revocation would need a server-side session store.
- **The public `/sendMail` endpoint has no rate limiting.** The framework ships a
  `ratelimit` component; it is not wired up here.
- **INI cannot express a `[]string` field.** A bad `[]string` value makes startup fail
  with `cannot convert string to []string` rather than reporting a config error; the app
  no longer has such a field (the old `web.server.locations` usage went away with the
  static frontend), but the framework limitation remains.
