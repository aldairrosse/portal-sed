# Proposal: Backend Server Bootstrap

## Intent

Create a runnable backend server with Docker infrastructure, seeder matching frontend fixtures, and all handler groups wired under `/api/v1`. Enables local development with `docker-compose up` and provides the foundation for future API development. Frontend remains on fixtures (`VITE_USE_API=false`) throughout.

## Scope

### In Scope
- `docker-compose.yml` with PostgreSQL 16 + API service
- `api/Dockerfile` multi-stage Go build (non-root user)
- `api/cmd/server/main.go` with env config, Ent auto-migrate, Chi router mounting all handler groups, CORS for localhost:5173, health check
- `api/internal/seed/` package mirroring ALL frontend fixture data (profiles, org tree, goals, competencies, evaluations, ninebox)
- Transactional seeder with FK ordering and run-once flag
- `.env.example` with DATABASE_URL, API_PORT, CORS_ORIGINS

### Out of Scope
- Frontend loading states, error toasts, API client wiring
- Real auth (JWT/sessions) — AuthPlaceholder only
- Production hardening (Redis, rate limit persistence, SSL)
- CI/CD pipelines
- Load/skeleton components, global error/confirmation UI

## Capabilities

### New Capabilities
None — this change introduces infrastructure only, no new business capabilities.

### Modified Capabilities
None — existing spec behavior unchanged.

## Approach

1. **Docker Compose**: PostgreSQL 16 container with port 5432 exposed; API service depends_on postgres, runs migrations on startup via entrypoint script.

2. **Dockerfile**: Multi-stage (builder → runtime), non-root user `app`, static binary in `/opt/api/`.

3. **Main entrypoint** (`cmd/server/main.go`):
   - Load env: `DATABASE_URL`, `API_PORT` (default 8080), `CORS_ORIGINS` (comma-separated)
   - Ent client with auto-migrate (no manual migrations needed for new tables)
   - Chi router with middleware stack: CORS → request ID → logger
   - Mount all handler groups: `/api/v1/auth`, `/api/v1/goals`, `/api/v1/cycles`, `/api/v1/competencies`, `/api/v1/evaluations`, `/api/v1/org`
   - Health: `GET /health` returns `{"status":"ok"}`
   - Graceful shutdown on SIGINT/SIGTERM

4. **Seeder** (`internal/seed/`):
   - Organized by domain: `org.go`, `goals.go`, `competency.go`, `evaluations.go`, `ninebox.go`, `seed.go` (main)
   - If `--seed` flag or DB empty, run all seeders in transaction with FK-ordered inserts
   - Names match frontend fixtures; "María López García" → "Frankil Aldair Perez" in seeds only

5. **Config**: `.env.example` committed; real `.env` gitignored.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `api/Dockerfile` | New | Multi-stage Go build |
| `api/cmd/server/main.go` | New | Server entry point, router wiring |
| `api/internal/seed/` | New | Fixture data seeder package |
| `docker-compose.yml` | New | PostgreSQL + API services |
| `.env.example` | New | Environment variable template |
| `api/migrations/` | Modified | Auto-applied on startup (no manual run) |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Ent auto-migrate conflicts with existing migrations | Low | Use `migrate.DropDisabled` mode; reviewed migration order |
| Seeder FK ordering bugs | Medium | Transaction rollback; unit test with fresh DB |
| CORS misconfiguration blocks frontend | Low | Default localhost:5173; documented override via CORS_ORIGINS |

## Rollback Plan

1. Revert docker-compose.yml + Dockerfile changes
2. Remove `cmd/server/main.go` and `internal/seed/` directories
3. No database schema changes — auto-migrate is additive only for new tables
4. If seeder corrupts data: `docker-compose down -v` (drops volumes) + restart

## Dependencies

- Go 1.22+
- Docker + Docker Compose v2
- PostgreSQL 16 (provided via docker-compose)
- Ent generated code at `api/internal/ent/`

## Success Criteria

- [ ] `docker-compose up` starts PostgreSQL + API with no errors
- [ ] `GET /health` returns `{"status":"ok"}` with HTTP 200
- [ ] `GET /api/v1/*` routes respond (even if 401 AuthPlaceholder)
- [ ] Seeder runs on `--seed` flag and populates all domains
- [ ] Frontend fixtures remain untouched (`VITE_USE_API=false` works as before)
- [ ] Dev bar profile/phase switcher functional during transition