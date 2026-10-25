# Implementation Plan: Netcheck Web UI (Separate Binary)

## Goal
Create a separate web frontend/executable for DiagnosticToolkit (netcheck). The web UI should run as a local server on localhost, allow starting/stopping monitoring, viewing live logs/results, inspecting the environment, and editing configuration. The existing CLI (`cmd/netcheck`) must remain untouched.

## Rationale
- **Separation of concerns**: Keep CLI for scripting/CI, web UI for interactive inspection/configuration.
- **Reuse**: Leverage existing `internal/app`, `internal/runner`, `internal/domain`, `internal/platform` logic.
- **Safer UX**: Localhost-only server, explicit user action to start monitoring.
- **Type-safe UI**: Use Templ with reusable Templ components. Pair with HTMX for interactivity and SSE for live streaming.

## Architecture Overview

### Binaries
- `cmd/netcheck/` - Existing CLI (unchanged)
- `cmd/netcheck-web/` - New dedicated web binary

### Core Packages
- `internal/web/` - HTTP server, routing, handlers, lifecycle/state management, SSE hub, middleware (localhost enforcement)
- `internal/app/` - Shared app logic (unchanged, reused)
- `web/templates/` - Templ templates and components (`.templ` files, compiled to `*_templ.go`)
- `web/static/` - Static assets (CSS/JS). Embed via `embed.FS`

### State Management (in-memory)
- Single `*app.App` instance
- Monitor lifecycle controlled via `context.Context` + `cancel()` 
- `running` flag with mutex
- Ring buffer of recent `domain.CheckResult` for history/log view
- SSE subscriber set for broadcasting live results

## Feature Set
| Feature | Notes |
| --- | --- |
| Dashboard | Status (running/stopped), environment snapshot, last run time, quick actions (Start/Stop/Run Once) |
| Live Results/Logs | SSE streaming + HTMX partials, filterable by type/status, recent history from ring buffer |
| Run Once | Trigger `app.RunOnce()` without starting continuous monitoring |
| Settings/Config | View/edit full config, validate via `app.Load()`, save atomically with backup, reload app on save |
| Environment | Display discovered gateway/resolvers from `app.Environment()` |
| Graceful Shutdown | Stop monitor on SIGINT/SIGTERM, close server cleanly |

## Implementation Phases

### Phase 1: Scaffolding
- Create `cmd/netcheck-web/main.go` with flags for address/port (must accept a port argument; configurable now, specifics can be defined later)
- Add Templ dependency (`github.com/a-h/templ`) and plan for `templ generate` workflow (justfile/go:generate)
- Create `internal/web/` skeleton, `web/templates/` and `web/static/` directories
- Basic HTTP router setup (stdlib is fine)

### Phase 2: Core Server & Lifecycle
- Implement in-memory state for running/stopped, monitor goroutine, ring buffer, and SSE hub
- Wire `app.App` (load config, create app) 
- Handlers: status, start, stop, run-once, SSE stream
- Enforce localhost binding by default (e.g. default to `127.0.0.1`, allow override via flag/arg for port/addr)
- Start idle by default (no auto-start of monitoring)

### Phase 3: UI with Templ Components
- Base layout component
- Reusable Templ components (status badges, tables, buttons, forms)
- Pages: Dashboard, Logs, Settings, Environment
- HTMX for actions/partial updates, SSE for live logs
- Embed templates and static assets

### Phase 4: Config Editing & Polish
- Full YAML editor/viewer on Settings page
- Validate on save using existing `app.Load()` 
- Atomic write with backup (`netcheck.yaml.bak`)
- Reload app/runner after successful config change
- Basic styling, error/flash messages, empty states

## Decisions & Answers

1. **Port binding**: Must support being called with an argument to specify a port (and/or address). Default binding should remain localhost. Exact default values can be finalized later.
2. **Config format (YAML vs JSON)**: Stick with **YAML** as-is. Converting to JSON is not necessary - the existing `app.Load()` and config structures are YAML-based, and full YAML editing is simpler to implement without introducing a new format. No migration needed.
3. **Auto-start**: **No** - server starts idle. Monitoring must be explicitly started by the user via the UI.
4. **Binary naming/location**: Treat as a separate binary under `cmd/netcheck-web/`. Exact installation/distribution location is flexible for now (can be built in-repo or installed separately). The key point is it's a distinct executable from `cmd/netcheck`.

## Notes & Constraints
- Keep it localhost-only by default
- No changes to existing CLI behavior
- Reuse existing business logic wherever possible
- Use Templ with components as the UI foundation
- In-memory only (no DB)
- Graceful shutdown required
