# Tasks: Backend Server Bootstrap

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 750–900 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 → PR 2 → PR 3 |
| Delivery strategy | auto-chain |
| Chain strategy | stacked-to-main |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Docker infra + config files | PR 1 | Independent; merges to main |
| 2 | Seeder package (all domains) | PR 2 | Depends on Ent schemas only; merges to main |
| 3 | Server entry point + DI wiring | PR 3 | Depends on PR 1 + PR 2 for full integration test |

## PR 1: Docker Infrastructure & Config

- [ ] T1.1 Create `api/Dockerfile` — multi-stage: builder (golang:1.26-alpine, CGO_ENABLED=0) → runtime (alpine:3.19, non-root user `app`), static binary at `/opt/api/api`
- [ ] T1.2 Create `docker-compose.yml` — `postgres:16-alpine` (healthcheck `pg_isready -U sed`, volume `pgdata`) + `api` service (build `./api`, depends_on postgres healthy, env_file `.env`)
- [ ] T1.3 Create `.env.example` — `DATABASE_URL`, `API_PORT=8080`, `CORS_ORIGINS=http://localhost:5173`, `SEED_ON_START=true`
- [x] T1.4 Add Go deps to `api/go.mod`: `go get github.com/go-chi/cors`, `github.com/joho/godotenv`, `github.com/jackc/pgx/v5/stdlib`

**Acceptance**: `docker-compose up` starts postgres (healthy), `docker build -t api ./api` succeeds.

## PR 2: Seeder Package

- [x] T2.1 Create `api/internal/seed/seed.go` — `Run(ctx, client)` orchestrator with: `--seed` flag parsing, `SEED_ON_START` env check, empty-DB check via `Employee.Query().Count()`, FK-ordered invocation of domain seeders
- [x] T2.2 Create `api/internal/seed/org.go` — Organization + OrgNode (31 nodes hierarchical, matching org-tree.json), EvaluationProfile (9 profiles)
- [x] T2.3 Create `api/internal/seed/employees.go` — 31 employees with `org_node_id`, `manager_id`, `profile_id` FK chain; EvaluatorScope records; "Frankil Aldair Perez" replaces "María López García"
- [x] T2.4 Create `api/internal/seed/cycle.go` — Cycle 2026, PhaseDefinition records (3), PhaseTransition rows (2)
- [x] T2.5 Create `api/internal/seed/competency.go` — 3 Pillars → 8 Competencies → ScaleCriteria (40) → LevelDefinitions (5) → CompetencyAcceptanceLevel (64)
- [x] T2.6 Create `api/internal/seed/goals.go` — GoalCategory (4) → Goal (10) → KPI (6) → GoalKpiLink (11) → GoalAssignment (10)
- [x] T2.7 Create `api/internal/seed/evaluation.go` — Evaluation records (3) with self + RH competency ratings (25), evaluation goals (6 closures)
- [x] T2.8 Create `api/internal/seed/ninebox.go` — NineBoxScale (18, 9×2 axes) → NineBoxQuadrant (7) → NineBoxMatrix (1) → NineBoxEntry (6)

**Acceptance**: `seed.Run(ctx, client)` populates all tables with no FK violations. Each domain seeder runs in its own `client.Tx(ctx)` with rollback-log on error.

## PR 3: Server Entry Point & DI Wiring

- [x] T3.1 Create `api/cmd/server/main.go` — env loading via godotenv; sql.Open("pgx") → entsql.OpenDB → internal.NewClient; auto-migrate with client.Schema.Create(ctx)
- [x] T3.2 Wire dependency chain in main.go:
  - Extract `*sql.DB` via sql.Open("pgx", ...); create ent driver with entsql.OpenDB + internal.NewClient(internal.Driver(drv))
  - Build ALL repos (goal: CategoryRepo, GoalRepo, KpiRepo, LinkKpiRepo, AssignmentRepo, WeightQueries; cycle: CycleRepo, PhaseRepo; competency: PillarRepo, CompetencyRepo, ScaleRepo, CatalogRepo, AcceptanceRepo; evaluation: EvaluationRepo, CompetencyRatingRepo, GoalRatingRepo, NineBoxRepo, CatalogRepo; org: OrgTreeRepo, OrgNodeRepo, EmployeeRepo, EvaluatorScopeRepo)
  - Build ALL services (goal: CategoryService, GoalService, ProgressService, KPIService, WeightValidationService, BatchService, PhaseCheck with nopPhaseChecker adapter; cycle: Service, PhaseService; competency: PillarService, CompetencyService, ScaleService, CatalogService, AcceptanceService; evaluation: EvaluationService, NineBoxService, DashboardService; org: OrgTreeService, OrgNodeService, EmployeeService, EvaluateeService, EvaluatorService; auth: SessionStore, EmployeeReader, AuthService)
  - Build ALL handlers (auth.NewAuthHandler, goal.NewGoalHandler, cycle.NewCycleHandler, competency.NewHandler, evaluation.NewEvaluationHandler, org.NewOrgHandler)
  - Stubs: nopPhaseChecker (always "asignacion"), evalCyclePhaseCheck (cycle-repo-backed), inMemoryIdempotencyCache
- [x] T3.3 Mount routes on Chi router with middleware stack: CORS (allow `CORS_ORIGINS`) → RequestID → RealIP → Logger → Recoverer
  - `r.Mount("/api/v1/auth", authhandler.AuthRoutes(authH))`
  - `r.Mount("/", goalhandler.NewRouter(goalH))` — full paths
  - `r.Mount("/api/v1", cyclehandler.NewRouter(cycleH))`
  - `comphandler.RegisterRoutes(r, &comphandler.Dependencies{Handler: compH})`
  - `r.Mount("/api/v1", evalhandler.NewRouter(evalH))`
  - `r.Mount("/api/v1", orghandler.NewRouter(orgH))`
- [x] T3.4 Add `GET /health` returning `{"status":"ok"}` on root router
- [x] T3.5 Add graceful shutdown — `signal.NotifyContext(ctx, SIGINT, SIGTERM)`, `http.Server.Shutdown()` 10s timeout
- [x] T3.6 Add seeder trigger at startup (after auto-migrate, before ListenAndServe): seed.Run(ctx, client) handles --seed flag, SEED_ON_START env, or empty-DB check

**Acceptance**: `go build ./cmd/server` succeeds. `docker-compose up` starts API on `:8080`, `GET /health` returns 200, `GET /api/v1/employees` returns seeded data.

## Phase Organization

| Phase | Tasks | Focus |
|-------|-------|-------|
| Phase 1 | T1.1–T1.4 | Docker & config foundation |
| Phase 2 | T2.1–T2.8 | Seeder with all fixture data |
| Phase 3 | T3.1–T3.6 | Server entry point + full DI + route mounting |

### Implementation Order

PR 1 first (independent infra). PR 2 second (no server runtime needed to write seed data). PR 3 last (depends on repos existing for import, but main.go compiles with or without seed package). Each PR merges to main sequentially.

### Next Step

Ready for implementation (`sdd-apply`). Chain strategy confirmed as `stacked-to-main` — no user decision needed per `auto-chain` delivery strategy.
