# Design: Backend Server Bootstrap

## Technical Approach

Wire all existing handler/service/repository layers into a single `cmd/server/main.go` entry point. Add a transactional seeder mirroring frontend fixtures. Containerize with multi-stage Docker + docker-compose. Frontend stays on fixtures (`VITE_USE_API=false`).

## Architecture Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Env loading | `os.Getenv` with defaults | No external dep needed; 4 vars only |
| DB driver | `entgo.io/ent/dialect/sql` + `pgx` via stdlib | Already in go.mod; Ent `Open("postgres", dsn)` |
| Auto-migrate | `client.Schema.Create(ctx)` | Ent-generated; idempotent; matches existing `migrate.go` |
| CORS | `github.com/go-chi/cors` | Standard Chi middleware; one import |
| Seeder trigger | `--seed` flag OR `SEED_ON_START` env OR empty DB | Proposal spec; empty-DB check via `Employee.Query().Count()` |
| Seeder tx | One tx per domain group (not global) | Limits lock scope; partial seed visible on failure |
| Route mounting | Mixed: full-path routers on root, relative-path under prefix | Matches existing `routes.go` patterns (goal/competency use full paths; auth/cycle/evaluation/org use relative) |

## Data Flow

```
Startup
  │
  ├─ 1. Load env (DATABASE_URL, API_PORT, CORS_ORIGINS, SEED_ON_START)
  ├─ 2. Ent Open("postgres", DATABASE_URL) → *internal.Client
  ├─ 3. client.Schema.Create(ctx)  [auto-migrate]
  ├─ 4. Should seed? (--seed flag || SEED_ON_START || DB empty)
  │     └─ seed.Run(ctx, client)  [tx per domain group]
  ├─ 5. Build repos → services → handlers  [DI chain]
  ├─ 6. Chi router + middleware stack
  ├─ 7. Mount routes
  └─ 8. http.ListenAndServe + graceful shutdown (SIGINT/SIGTERM)

Request
  CORS → RequestID → Logger → Recovery → [handler-specific middleware] → Handler
```

## Middleware Stack (global, in order)

| # | Middleware | Source | Purpose |
|---|-----------|--------|---------|
| 1 | CORS | `chi/cors` | Allow `localhost:5173` |
| 2 | RequestID | `chi/middleware` | Trace correlation |
| 3 | Logger | `chi/middleware` | Request logging |
| 4 | Recoverer | `chi/middleware` | Panic recovery |

Handler-level middleware (AuthPlaceholder, RateLimit, Idempotency) stays inside each `routes.go` — no change.

## Route Mounting Strategy

| Handler | Router func | Paths | Mount |
|---------|------------|-------|-------|
| auth | `AuthRoutes(h)` | relative (`/login`) | `r.Mount("/api/v1/auth", ...)` |
| goal | `goal.NewRouter(h)` | full (`/api/v1/...`) | `r.Mount("/", ...)` |
| cycle | `cycle.NewRouter(h)` | relative (`/cycles`) | `r.Mount("/api/v1", ...)` |
| competency | `competency.RegisterRoutes(r, deps)` | full (`/api/v1/...`) | `RegisterRoutes(r, deps)` on root |
| evaluation | `evaluation.NewRouter(h)` | relative (`/evaluations`) | `r.Mount("/api/v1", ...)` |
| org | `org.NewRouter(h)` | relative (`/org-trees`) | `r.Mount("/api/v1", ...)` |

**Conflict note**: cycle, evaluation, and org all mount under `/api/v1` with non-overlapping path prefixes. Chi handles this via separate `Mount` calls on the same parent.

## DI Wiring (main.go)

```
*sql.DB ← from Ent client.Driver().(*sql.Driver).DB()
         OR open separately via database/sql for raw queries

Repositories (all take *internal.Client, some take *sql.DB):
  goal:      CategoryRepo, GoalRepo, KpiRepo, LinkKpiRepo, AssignmentRepo, WeightQueries
  cycle:     CycleRepo, PhaseRepo
  competency: PillarRepo, CompetencyRepo, ScaleRepo, CatalogRepo, AcceptanceRepo
  evaluation: EvaluationRepo, NineBoxRepo, CatalogRepo, GoalRatingRepo, CompetencyRatingRepo
  org:       OrgTreeRepo, OrgNodeRepo, EmployeeRepo, EvaluatorScopeRepo
  auth:      EmployeeReader (raw SQL via *sql.DB)

Services (take repos):
  goal:      CategoryService, GoalService, ProgressService, KPIService, WeightValidationService, BatchService
  cycle:     Service, PhaseService
  competency: PillarService, CompetencyService, ScaleService, CatalogService, AcceptanceService
  evaluation: EvaluationService, NineBoxService, DashboardService
  org:       OrgTreeService, OrgNodeService, EmployeeService, EvaluateeService, EvaluatorService
  auth:      SessionStore(*sql.DB), AuthService(SessionStore, EmployeeReader, *sql.DB)

Handlers (take services):
  auth.NewAuthHandler(authSvc)
  goal.NewGoalHandler(6 services + 5 repos)
  cycle.NewCycleHandler(svc, phaseSvc)
  competency.NewHandler(5 services) → competency.Dependencies{Handler}
  evaluation.NewEvaluationHandler(evalSvc, nineBoxSvc, dashSvc)
  org.NewOrgHandler(5 services)
```

## Seeder Package (`api/internal/seed/`)

### File Structure

| File | Responsibility |
|------|---------------|
| `seed.go` | `Run(ctx, client)` orchestrator; empty-DB check; flag parsing |
| `org.go` | Organization + OrgNode tree (25 nodes, hierarchical) |
| `employees.go` | 25 employees with manager FKs, profiles, evaluator scopes |
| `cycle.go` | Cycle 2026, PhaseDefinition, PhaseTransition |
| `competency.go` | 3 Pillars → 8 Competencies → ScaleCriteria → AcceptanceLevels |
| `goals.go` | GoalCategory → Goal → KPI → GoalKpiLink → GoalAssignment |
| `evaluation.go` | Evaluation → EvaluationCompetency (self + RH ratings) → EvaluationGoal (closures) |
| `ninebox.go` | NineBoxScale → NineBoxQuadrant → NineBoxMatrix → NineBoxEntry (9x9 = 12 entries) |

### FK Insertion Order

```
1. Organization
2. OrgNodes (tree with parent_id self-ref)
3. EvaluationProfile (9 profiles)
4. Employee (25, with org_node_id, manager_id, profile_id)
5. EvaluatorScope
6. Cycle + PhaseDefinition + PhaseTransition
7. Pillar → Competency → ScaleCriterion → CompetencyAcceptanceLevel
8. GoalCategory → Goal → KPI → GoalKpiLink → GoalAssignment
9. Evaluation → EvaluationCompetency → EvaluationGoal
10. NineBoxScale → NineBoxQuadrant → NineBoxMatrix → NineBoxEntry
```

Each step runs in its own `client.Tx(ctx)`. On error: rollback that tx, log, continue to next group (partial seed acceptable for dev).

### Seed Trigger Logic

```go
shouldSeed := flagSeed || os.Getenv("SEED_ON_START") == "true"
if !shouldSeed {
    count, _ := client.Employee.Query().Count(ctx)
    shouldSeed = count == 0
}
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `api/cmd/server/main.go` | Create | Server entry point: env, Ent, router, DI, shutdown |
| `api/internal/seed/seed.go` | Create | Seeder orchestrator with trigger logic |
| `api/internal/seed/org.go` | Create | Organization + OrgNode seed data |
| `api/internal/seed/employees.go` | Create | 25 employees + profiles + scopes |
| `api/internal/seed/cycle.go` | Create | Cycle 2026 + phases + transitions |
| `api/internal/seed/competency.go` | Create | Pillars, competencies, scale, acceptance |
| `api/internal/seed/goals.go` | Create | Categories, goals, KPIs, assignments |
| `api/internal/seed/evaluation.go` | Create | Evaluations, competency ratings, goal closures |
| `api/internal/seed/ninebox.go` | Create | Scales, quadrants, matrices, entries |
| `api/Dockerfile` | Create | Multi-stage Go build, non-root |
| `docker-compose.yml` | Create | postgres:16-alpine + api services |
| `.env.example` | Create | DATABASE_URL, API_PORT, CORS_ORIGINS, SEED_ON_START |
| `api/go.mod` | Modify | Add `github.com/go-chi/cors`, `github.com/joho/godotenv`, `github.com/jackc/pgx/v5` |

## Docker Infrastructure

### docker-compose.yml

```yaml
services:
  postgres:
    image: postgres:16-alpine
    ports: ["5432:5432"]
    environment:
      POSTGRES_USER: sed
      POSTGRES_PASSWORD: sed
      POSTGRES_DB: sed
    volumes: [pgdata:/var/lib/postgresql/data]
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U sed"]
      interval: 5s
      retries: 5

  api:
    build: ./api
    ports: ["8080:8080"]
    depends_on:
      postgres: { condition: service_healthy }
    env_file: .env
    environment:
      DATABASE_URL: postgres://sed:sed@postgres:5432/sed?sslmode=disable

volumes:
  pgdata:
```

### api/Dockerfile

```dockerfile
FROM golang:1.26-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/server

FROM alpine:3.19
RUN adduser -D -u 1000 app
USER app
COPY --from=builder /api /opt/api/api
EXPOSE 8080
ENTRYPOINT ["/opt/api/api"]
```

Networking: `api` reaches `postgres` via Docker DNS hostname `postgres` on port 5432. Host machine reaches API on `localhost:8080` and Postgres on `localhost:5432`.

## Config

### .env.example

```
DATABASE_URL=postgres://sed:sed@localhost:5432/sed?sslmode=disable
API_PORT=8080
CORS_ORIGINS=http://localhost:5173
SEED_ON_START=true
```

`.gitignore` already includes `.env` and `.env.*` (with `!.env.example` exception) — no changes needed.

## Testing Strategy

| Layer | What | Approach |
|-------|------|----------|
| Unit | Seeder FK ordering | Test each `seed/*.go` func with `enttest` (existing helper) on SQLite |
| Integration | Full startup + seed | `docker-compose up` + `GET /health` + count seeded rows |
| Manual | Route smoke test | `curl localhost:8080/api/v1/employees` → 200 with seeded data |

## Migration / Rollout

No migration required. Auto-migrate creates tables from Ent schemas. Seeder is idempotent (checks empty DB). Rollback: `docker-compose down -v` drops all data.

## Open Questions

- [ ] None — all dependencies verified against existing code.

## Next Step

Ready for tasks (sdd-tasks).
