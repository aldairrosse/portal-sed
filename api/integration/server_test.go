package integration

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/seed"

	// Ent entity/enum packages
	"github.com/sed-evaluacion-desempeno/api/internal/competency"
	"github.com/sed-evaluacion-desempeno/api/internal/cycle"
	"github.com/sed-evaluacion-desempeno/api/internal/employee"
	"github.com/sed-evaluacion-desempeno/api/internal/evaluation"
	"github.com/sed-evaluacion-desempeno/api/internal/evaluationcompetency"
	"github.com/sed-evaluacion-desempeno/api/internal/evaluationprofile"
	"github.com/sed-evaluacion-desempeno/api/internal/goal"
	"github.com/sed-evaluacion-desempeno/api/internal/goalassignment"
	"github.com/sed-evaluacion-desempeno/api/internal/goalcategory"
	"github.com/sed-evaluacion-desempeno/api/internal/goalkpilink"
	"github.com/sed-evaluacion-desempeno/api/internal/kpi"
	"github.com/sed-evaluacion-desempeno/api/internal/nineboxscale"
	"github.com/sed-evaluacion-desempeno/api/internal/organization"
	"github.com/sed-evaluacion-desempeno/api/internal/orgnode"
	"github.com/sed-evaluacion-desempeno/api/internal/phasedefinition"
	"github.com/sed-evaluacion-desempeno/api/internal/pillar"
	"github.com/sed-evaluacion-desempeno/api/internal/scalecriterion"

	// Repositories
	repoactivity "github.com/sed-evaluacion-desempeno/api/internal/repository/activity"
	repogoal "github.com/sed-evaluacion-desempeno/api/internal/repository/goal"
	repocycle "github.com/sed-evaluacion-desempeno/api/internal/repository/cycle"
	repocompetency "github.com/sed-evaluacion-desempeno/api/internal/repository/competency"
	repoeval "github.com/sed-evaluacion-desempeno/api/internal/repository/evaluation"
	repoorganization "github.com/sed-evaluacion-desempeno/api/internal/repository/org"

	// Services
	activitysvc "github.com/sed-evaluacion-desempeno/api/internal/service/activity"
	goalsvc "github.com/sed-evaluacion-desempeno/api/internal/service/goal"
	cyclesvc "github.com/sed-evaluacion-desempeno/api/internal/service/cycle"
	compsvc "github.com/sed-evaluacion-desempeno/api/internal/service/competency"
	evalsvc "github.com/sed-evaluacion-desempeno/api/internal/service/evaluation"
	orgsvc "github.com/sed-evaluacion-desempeno/api/internal/service/org"
	authsvc "github.com/sed-evaluacion-desempeno/api/internal/service/auth"

	// Handlers
	authhandler "github.com/sed-evaluacion-desempeno/api/internal/handler/auth"
	goalhandler "github.com/sed-evaluacion-desempeno/api/internal/handler/goal"
	cyclehandler "github.com/sed-evaluacion-desempeno/api/internal/handler/cycle"
	comphandler "github.com/sed-evaluacion-desempeno/api/internal/handler/competency"
	evalhandler "github.com/sed-evaluacion-desempeno/api/internal/handler/evaluation"
	orghandler "github.com/sed-evaluacion-desempeno/api/internal/handler/org"

	"github.com/sed-evaluacion-desempeno/api/internal/auth"
	"github.com/sed-evaluacion-desempeno/api/internal/auth/sso"
)

// nopPhaseChecker always returns "asignacion" as the current phase.
type nopPhaseChecker struct{}

func (nopPhaseChecker) GetCurrentPhase(_ context.Context, _ string) (goalsvc.CyclePhase, error) {
	return goalsvc.PhaseAsignacion, nil
}

// fixedPhaseChecker returns the given phase for all queries.
type fixedPhaseChecker struct {
	phase goalsvc.CyclePhase
}

func (f fixedPhaseChecker) GetCurrentPhase(_ context.Context, _ string) (goalsvc.CyclePhase, error) {
	return f.phase, nil
}

// evalCyclePhaseCheck implements evalsvc.CyclePhaseChecker backed by the cycle repo.
type evalCyclePhaseCheck struct {
	cycleRepo *repocycle.CycleRepo
}

func (c *evalCyclePhaseCheck) GetPhase(_ context.Context, _ uuid.UUID) (string, error) {
	return "asignacion", nil
}

func (c *evalCyclePhaseCheck) GetSelfEvalDeadline(_ context.Context, _ uuid.UUID) (*time.Time, error) {
	return nil, nil
}

func (c *evalCyclePhaseCheck) GetActiveCycleID(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, nil
}

// inMemoryIdempotencyCache is a simple in-memory cache for tests.
type inMemoryIdempotencyCache struct {
	mu    sync.RWMutex
	items map[string]*evalsvc.IdempotencyCacheEntry
}

func newInMemoryIdempotencyCache() *inMemoryIdempotencyCache {
	return &inMemoryIdempotencyCache{items: make(map[string]*evalsvc.IdempotencyCacheEntry)}
}

func (c *inMemoryIdempotencyCache) Get(_ context.Context, key string) (*evalsvc.IdempotencyCacheEntry, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.items[key]
	if !ok {
		return nil, nil
	}
	return entry, nil
}

func (c *inMemoryIdempotencyCache) Set(_ context.Context, key string, entry *evalsvc.IdempotencyCacheEntry, ttl time.Duration) error {
	c.mu.Lock()
	c.items[key] = entry
	c.mu.Unlock()
	time.AfterFunc(ttl, func() {
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
	})
	return nil
}

// testServer holds the HTTP test server and its cleanup function.
type testServer struct {
	Server *http.Server
	Client *internal.Client
	DB     *sql.DB
	Router chi.Router
	Clean  func()
	Token  string
	// TokenRH es la sesión del empleado emp-jefe (perfil rh → RoleRH, que tiene
	// eval:9x9). Los writes nine-box requieren ese permiso; emp-dg-01 (colaborador) no lo tiene.
	TokenRH string
}

// setupTestServer builds the full router with real DB (phase=asignacion).
func setupTestServer(t *testing.T) *testServer {
	return setupTestServerWithPhaseChecker(t, nopPhaseChecker{})
}

// setupTestServerWithPhaseChecker builds the full router with the given phase checker.
// It connects to the PostgreSQL instance specified by DATABASE_URL, runs
// auto-migration, seeds data, and wires all handlers.
func setupTestServerWithPhaseChecker(t *testing.T, phaseChecker goalsvc.PhaseChecker) *testServer {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://sed_test:sed_test@localhost:5433/sed_test?sslmode=disable"
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		t.Fatalf("failed to ping database: %v", err)
	}

	// SQL migrations (goose .up.sql files) — same mechanism as cmd/server.
	// Creates enum types and tables Ent does not manage (phase, sessions, ...).
	if err := runSQLMigrations(ctx, db); err != nil {
		db.Close()
		t.Fatalf("failed to run SQL migrations: %v", err)
	}

	drv := entsql.OpenDB(dialect.Postgres, db)
	client := internal.NewClient(internal.Driver(drv))

	// Auto-migrate
	if err := client.Schema.Create(context.Background()); err != nil {
		db.Close()
		t.Fatalf("failed to auto-migrate: %v", err)
	}

	// Create ltree extension and set up materialized views (not managed by Ent)
	for _, stmt := range []string{
		`CREATE EXTENSION IF NOT EXISTS ltree`,
		`CREATE MATERIALIZED VIEW IF NOT EXISTS evaluation_summary AS
		 SELECT cycle_id, state, COUNT(1) as count
		 FROM evaluations GROUP BY cycle_id, state WITH DATA`,
	} {
		if _, err := db.ExecContext(context.Background(), stmt); err != nil {
			db.Close()
			t.Fatalf("failed to run post-migration setup %q: %v", stmt[:60], err)
		}
	}

	// Seed data
	if err := seedIntegrationData(ctx, client, db); err != nil {
		log.Printf("[seed] warning: %v", err)
	}

	// --- DI wiring (mirrors main.go) ---

	// Repos
	catRepo := repogoal.NewCategoryRepo(client, db)
	goalRepo := repogoal.NewGoalRepo(client, db)
	kpiRepo := repogoal.NewKpiRepo(client, db)
	linkRepo := repogoal.NewLinkKpiRepo(client, db)
	assignRepo := repogoal.NewAssignmentRepo(client, db)
	proposalRepo := repogoal.NewGoalProposalRepo(db)
	weightQ := repogoal.NewWeightQueries(db)

	cycleRepo := repocycle.NewCycleRepo(client, db)
	phaseRepo := repocycle.NewPhaseRepo(client, db)

	pillarRepo := repocompetency.NewPillarRepo(client)
	compRepo := repocompetency.NewCompetencyRepo(client)
	scaleRepo := repocompetency.NewScaleRepo(client)
	catalogCompRepo := repocompetency.NewCatalogRepo(client, db)
	acceptanceRepo := repocompetency.NewAcceptanceRepo(client)

	evalRepo := repoeval.NewEvaluationRepo(client, db)
	compRatingRepo := repoeval.NewCompetencyRatingRepo(client)
	goalRatingRepo := repoeval.NewGoalRatingRepo(client)
	nineBoxRepo := repoeval.NewNineBoxRepo(client, db)
	catalogEvalRepo := repoeval.NewCatalogRepo(client)

	orgTreeRepo := repoorganization.NewOrgTreeRepo(client, db)
	orgNodeRepo := repoorganization.NewOrgNodeRepo(client, db)
	employeeRepo := repoorganization.NewEmployeeRepo(client, db)
	metricsRepo := repoorganization.NewMetricsRepo(client, db)

	sessionStore := auth.NewSessionStore(db)
	employeeReader := authsvc.NewEmployeeReader(db)

	// Services
	authSvc := authsvc.NewAuthService(sessionStore, employeeReader, db)

	phaseCheck := goalsvc.NewPhaseCheck(phaseChecker)
	catSvc := goalsvc.NewCategoryService(catRepo, pillarRepo, phaseCheck, assignRepo, cycleRepo)
	goalSvc := goalsvc.NewGoalService(goalRepo, catRepo, kpiRepo, linkRepo, weightQ, phaseCheck)
	progressSvc := goalsvc.NewProgressService(goalRepo, catRepo, phaseCheck)
 	kpiSvc := goalsvc.NewKPIService(kpiRepo, linkRepo, goalRepo, catRepo, phaseCheck, orgNodeRepo, orgTreeRepo, employeeRepo)
	scoringSvc := goalsvc.NewScoringService(catRepo, goalRepo)
	weightSvc := goalsvc.NewWeightValidationService(catRepo, goalRepo)
	batchSvc := goalsvc.NewBatchService(goalRepo, catRepo, kpiRepo, linkRepo, weightQ, phaseCheck)
	proposalSvc := goalsvc.NewGoalProposalService(proposalRepo, goalRepo, catRepo, linkRepo, weightQ, phaseCheck, db)

	cycleSvc := cyclesvc.NewService(cycleRepo, phaseRepo, client)
	phaseSvc := cyclesvc.NewPhaseService(cycleRepo, phaseRepo)

	pillarSvc := compsvc.NewPillarService(pillarRepo)
	competencySvc := compsvc.NewCompetencyService(pillarRepo, compRepo)
	scaleSvc := compsvc.NewScaleService(compRepo, scaleRepo)
	catalogSvc := compsvc.NewCatalogService(catalogCompRepo)
	acceptanceSvc := compsvc.NewAcceptanceService(compRepo, catalogCompRepo, acceptanceRepo)

	cycleCheck := &evalCyclePhaseCheck{cycleRepo: cycleRepo}
	idemCache := newInMemoryIdempotencyCache()
	evalSvc := evalsvc.NewEvaluationService(evalRepo, compRatingRepo, goalRatingRepo, cycleCheck, idemCache, employeeRepo, orgNodeRepo)
	nineBoxSvc := evalsvc.NewNineBoxService(nineBoxRepo, catalogEvalRepo, db, cycleRepo, orgNodeRepo, employeeRepo)
	dashboardSvc := evalsvc.NewDashboardService(evalRepo)

	orgTreeSvc := orgsvc.NewOrgTreeService(orgTreeRepo, orgNodeRepo, employeeRepo, client)
	orgNodeSvc := orgsvc.NewOrgNodeService(orgNodeRepo, client)
	employeeSvc := orgsvc.NewEmployeeService(employeeRepo, client)
	evaluateeSvc := orgsvc.NewEvaluateeService(employeeRepo, orgNodeRepo, client, nil, nil, nil)
	metricsSvc := orgsvc.NewMetricsService(metricsRepo, orgNodeRepo, client)

	// Activity
	activityRepo := repoactivity.NewRepository(client)
	activitySvc := activitysvc.NewService(activityRepo)

	// Handlers
	authH := authhandler.NewAuthHandler(authSvc, sso.NewNoopAdapter())
	goalH := goalhandler.NewGoalHandler(
		catSvc, goalSvc, progressSvc, kpiSvc, scoringSvc, weightSvc, batchSvc, proposalSvc,
		catRepo, goalRepo, kpiRepo, linkRepo, assignRepo, proposalRepo, activitySvc, evalSvc,
	)
	cycleH := cyclehandler.NewCycleHandler(cycleSvc, phaseSvc, activitySvc, assignRepo, employeeRepo)
	compH := comphandler.NewHandler(pillarSvc, competencySvc, scaleSvc, catalogSvc, acceptanceSvc, activitySvc)
	evalH := evalhandler.NewEvaluationHandler(evalSvc, nineBoxSvc, dashboardSvc, activitySvc)
	orgH := orghandler.NewOrgHandler(orgTreeSvc, orgNodeSvc, employeeSvc, evaluateeSvc, metricsSvc)

	// Router
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Idempotency-Key"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)

	// Mount routes (same order as main.go)
	r.Group(func(r chi.Router) {
		comphandler.RegisterRoutes(r, &comphandler.Dependencies{Handler: compH, AuthSvc: authSvc})
	})
	r.Mount("/api/v1/auth", authhandler.AuthRoutes(authH, authSvc))

	apiV1 := chi.NewRouter()
	// Each RegisterRoutes calls r.Use() internally — wrap in Group so each
	// gets a clean inline sub-router (same pattern as comphandler, line 254).
	apiV1.Group(func(r chi.Router) {
		cyclehandler.RegisterRoutes(r, cycleH, authSvc)
	})
	apiV1.Group(func(r chi.Router) {
		evalhandler.RegisterRoutes(r, evalH, authSvc)
	})
	apiV1.Group(func(r chi.Router) {
		orghandler.RegisterRoutes(r, orgH, authSvc)
	})
	apiV1.Group(func(r chi.Router) {
		goalhandler.RegisterRoutes(r, goalH, authSvc)
	})
	r.Mount("/api/v1", apiV1)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	clean := func() {
		client.Close()
		db.Close()
	}

	ts := &testServer{
		Client: client,
		DB:     db,
		Router: r,
		Clean:  clean,
	}

	// Session token for authenticated integration tests (employee emp-dg-01).
	// Seed arrives later, so a login failure is tolerated.
	res, err := authSvc.Login(ctx, "emp-dg-01@example.com", "127.0.0.1", "integration-test")
	if err != nil {
		log.Printf("[test] login warning: %v", err)
	} else {
		ts.Token = res.Token
	}

	// RH session token for nine-box writes (employee emp-jefe, perfil rh → RoleRH).
	rhRes, rhErr := authSvc.Login(ctx, "jefe@example.com", "127.0.0.1", "integration-test")
	if rhErr != nil {
		log.Printf("[test] rh login warning: %v", rhErr)
	} else {
		ts.TokenRH = rhRes.Token
	}

	return ts
}

// runSQLMigrations applies the goose .up.sql migrations from cmd/server
// using the same mechanism as runSQLMigrations in api/cmd/server/migrations.go
// (custom schema_migrations table + "goose Up" section extraction). The embed
// in cmd/server is package-private, so we read the .sql files from disk; go
// test runs with the package dir as CWD, so "../cmd/server/migrations" resolves.
// Idempotent: already-applied versions are skipped.
func runSQLMigrations(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := os.ReadDir("../cmd/server/migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".up.sql") {
			continue
		}
		files = append(files, e.Name())
	}
	sort.Strings(files)

	for _, name := range files {
		version := strings.TrimSuffix(name, ".up.sql")

		var exists bool
		if err := db.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`, version,
		).Scan(&exists); err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}
		if exists {
			continue
		}

		data, err := os.ReadFile("../cmd/server/migrations/" + name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		sql := extractUpSQL(string(data))
		if sql == "" {
			log.Printf("[migrate] skipping empty %s", name)
			continue
		}

		log.Printf("[migrate] applying %s", name)

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin tx for %s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx, sql); err != nil {
			tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO schema_migrations (version) VALUES ($1)`, version,
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %s: %w", name, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", name, err)
		}
	}

	return nil
}

func extractUpSQL(content string) string {
	var lines []string
	inBlock := false
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "-- +goose Up" {
			inBlock = true
			continue
		}
		if trimmed == "-- +goose Down" {
			break
		}
		if trimmed == "-- +goose StatementBegin" || trimmed == "-- +goose StatementEnd" {
			continue
		}
		if inBlock {
			lines = append(lines, line)
		}
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// seedIntegrationData crea datos deterministas para los tests de integración.
// Reemplaza al antiguo seed.Run() que ahora es no-op.
func seedIntegrationData(ctx context.Context, client *internal.Client, db *sql.DB) error {
	seedID := seed.SeedID

	orgID := seedID("org-sed")
	rootID := seedID("node-root")
	child1ID := seedID("node-child-1")
	child2ID := seedID("node-child-2")
	profileRHID := seedID("profile-rh")
	profileColID := seedID("profile-colaborador")
	jefeID := seedID("emp-jefe")
	dgID := seedID("emp-dg-01")
	colID := seedID("emp-colaborador-01")
	cycleID := seedID("cycle-2026")
	avancePhaseID := seedID("phase-def-avance")
	cierrePhaseID := seedID("phase-def-cierre")
	pillarID := seedID("pillar-competencias")
	compID := seedID("comp-orientacion-resultados")
	catID := seedID("cat-ventas-finanzas")
	goalID := seedID("goal-incrementar-ventas")
	kpiID := seedID("kpi-ventas-mensuales")
	evalID := seedID("eval-dg01-2026")

	// Cleanup idempotente: borrar lo que vamos a crear (hijos antes que padres).
	// Un fallo de DELETE no aborta el seed: se loguea y se continúa para no dejar
	// la BD corrupta; solo los INSERT devuelven error.
	if _, err := client.NineBoxEntry.Delete().Exec(ctx); err != nil {
		log.Printf("cleanup warning: %v", err)
	}
	if _, err := client.NineBoxMatrix.Delete().Exec(ctx); err != nil {
		log.Printf("cleanup warning: %v", err)
	}
	if _, err := client.NineBoxQuadrant.Delete().Exec(ctx); err != nil {
		log.Printf("cleanup warning: %v", err)
	}
	if _, err := client.NineBoxScale.Delete().Exec(ctx); err != nil {
		log.Printf("cleanup warning: %v", err)
	}
	if _, err := client.GoalKpiLink.Delete().Where(goalkpilink.GoalID(goalID), goalkpilink.KpiID(kpiID)).Exec(ctx); err != nil {
		log.Printf("cleanup warning: %v", err)
	}
	if _, err := client.GoalAssignment.Delete().Where(goalassignment.IDIn(seedID("ga-dg01-2026"))).Exec(ctx); err != nil {
		log.Printf("cleanup warning: %v", err)
	}
	if _, err := client.EvaluationCompetency.Delete().Where(evaluationcompetency.IDIn(seedID("evalcomp-dg01-2026"))).Exec(ctx); err != nil {
		log.Printf("cleanup warning: %v", err)
	}
	if _, err := client.Evaluation.Delete().Where(evaluation.IDIn(evalID)).Exec(ctx); err != nil {
		log.Printf("cleanup warning: %v", err)
	}
	if _, err := client.Goal.Delete().Where(goal.IDIn(goalID)).Exec(ctx); err != nil {
		log.Printf("cleanup warning: %v", err)
	}
	if _, err := client.KPI.Delete().Where(kpi.IDIn(kpiID)).Exec(ctx); err != nil {
		log.Printf("cleanup warning: %v", err)
	}
	if _, err := client.GoalCategory.Delete().Where(goalcategory.IDIn(catID)).Exec(ctx); err != nil {
		log.Printf("cleanup warning: %v", err)
	}
	// phase_transitions referencian cycle y phase_definitions sin cascade.
	if _, err := client.PhaseTransition.Delete().Exec(ctx); err != nil {
		log.Printf("cleanup warning: %v", err)
	}
	if _, err := client.PhaseDefinition.Delete().Where(phasedefinition.IDIn(avancePhaseID, cierrePhaseID)).Exec(ctx); err != nil {
		log.Printf("cleanup warning: %v", err)
	}
	if _, err := client.Cycle.Delete().Where(cycle.IDIn(cycleID)).Exec(ctx); err != nil {
		log.Printf("cleanup warning: %v", err)
	}
	// activity_logs y sessions referencian employees sin ON DELETE CASCADE
	// (los tests y los logins los crean): borrarlas antes para que
	// Employee.Delete nunca falle por FK.
	if _, err := client.ActivityLog.Delete().Exec(ctx); err != nil {
		log.Printf("cleanup warning: %v", err)
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM sessions`); err != nil {
		log.Printf("cleanup warning: %v", err)
	}
	if _, err := client.Employee.Delete().Where(employee.IDIn(jefeID, dgID, colID)).Exec(ctx); err != nil {
		log.Printf("cleanup warning: %v", err)
	}
	if _, err := client.ScaleCriterion.Delete().Where(scalecriterion.CompetencyID(compID)).Exec(ctx); err != nil {
		log.Printf("cleanup warning: %v", err)
	}
	if _, err := client.Competency.Delete().Where(competency.IDIn(compID)).Exec(ctx); err != nil {
		log.Printf("cleanup warning: %v", err)
	}
	if _, err := client.Pillar.Delete().Where(pillar.IDIn(pillarID)).Exec(ctx); err != nil {
		log.Printf("cleanup warning: %v", err)
	}
	if _, err := client.OrgNode.Delete().Where(orgnode.IDIn(rootID, child1ID, child2ID)).Exec(ctx); err != nil {
		log.Printf("cleanup warning: %v", err)
	}
	if _, err := client.Organization.Delete().Where(organization.IDIn(orgID)).Exec(ctx); err != nil {
		log.Printf("cleanup warning: %v", err)
	}
	if _, err := client.EvaluationProfile.Delete().Where(evaluationprofile.IDIn(profileRHID, profileColID)).Exec(ctx); err != nil {
		log.Printf("cleanup warning: %v", err)
	}

	// Organización y árbol.
	if _, err := client.Organization.Create().
		SetID(orgID).
		SetName("SED Test").
		SetSlug("sed-test").
		Save(ctx); err != nil {
		return err
	}
	if _, err := client.OrgNode.Create().
		SetID(rootID).
		SetType(orgnode.TypeCorporate).
		SetCode("ROOT").
		SetName("Raiz").
		SetOrganizationID(orgID).
		SetPath("1").
		Save(ctx); err != nil {
		return err
	}
	if _, err := client.OrgNode.Create().
		SetID(child1ID).
		SetType(orgnode.TypeRetail).
		SetCode("CHILD1").
		SetName("Sucursal 1").
		SetOrganizationID(orgID).
		SetParentID(rootID).
		SetPath("1.1").
		Save(ctx); err != nil {
		return err
	}
	if _, err := client.OrgNode.Create().
		SetID(child2ID).
		SetType(orgnode.TypeCorporate).
		SetCode("CHILD2").
		SetName("Sucursal 2").
		SetOrganizationID(orgID).
		SetParentID(rootID).
		SetPath("1.2").
		Save(ctx); err != nil {
		return err
	}

	// Perfiles de evaluación.
	if _, err := client.EvaluationProfile.Create().
		SetID(profileRHID).
		SetName("rh").
		Save(ctx); err != nil {
		return err
	}
	if _, err := client.EvaluationProfile.Create().
		SetID(profileColID).
		SetName("colaborador").
		Save(ctx); err != nil {
		return err
	}

	// Empleados: jefe (sin manager), dg-01 (manager=jefe), colaborador-01 (manager=dg-01).
	if _, err := client.Employee.Create().
		SetID(jefeID).
		SetFirstName("Jefe").
		SetLastName("General").
		SetEmployeeNumber("JEFE-001").
		SetEmail("jefe@example.com").
		SetIsActive(true).
		SetOrgNodeID(rootID).
		SetProfileID(profileRHID).
		SetCreatedBy(jefeID).
		SetUpdatedBy(jefeID).
		Save(ctx); err != nil {
		return err
	}
	if _, err := client.Employee.Create().
		SetID(dgID).
		SetFirstName("Director").
		SetLastName("General").
		SetEmployeeNumber("DG-001").
		SetEmail("emp-dg-01@example.com").
		SetIsActive(true).
		SetOrgNodeID(child1ID).
		SetManagerID(jefeID).
		SetProfileID(profileColID).
		SetCreatedBy(jefeID).
		SetUpdatedBy(jefeID).
		Save(ctx); err != nil {
		return err
	}
	if _, err := client.Employee.Create().
		SetID(colID).
		SetFirstName("Colaborador").
		SetLastName("Uno").
		SetEmployeeNumber("COL-001").
		SetEmail("emp-colaborador-01@example.com").
		SetIsActive(true).
		SetOrgNodeID(child2ID).
		SetManagerID(dgID).
		SetProfileID(profileColID).
		SetCreatedBy(jefeID).
		SetUpdatedBy(jefeID).
		Save(ctx); err != nil {
		return err
	}

	// Ciclo 2026 con fases avance y cierre.
	if _, err := client.Cycle.Create().
		SetID(cycleID).
		SetYear(2026).
		SetCurrentPhase(cycle.CurrentPhaseAvance).
		SetOrganizationID(orgID).
		Save(ctx); err != nil {
		return err
	}
	if _, err := client.PhaseDefinition.Create().
		SetID(avancePhaseID).
		SetPhase(phasedefinition.PhaseAvance).
		SetLabel("Avance").
		SetOrder(1).
		SetCycleID(cycleID).
		Save(ctx); err != nil {
		return err
	}
	if _, err := client.PhaseDefinition.Create().
		SetID(cierrePhaseID).
		SetPhase(phasedefinition.PhaseCierre).
		SetLabel("Cierre").
		SetOrder(2).
		SetCycleID(cycleID).
		Save(ctx); err != nil {
		return err
	}

	// Marco de competencias: pillar + competency + scale_criterion 1..5.
	if _, err := client.Pillar.Create().
		SetID(pillarID).
		SetName("Competencias Corporativas").
		SetType(pillar.TypeCompetencias).
		Save(ctx); err != nil {
		return err
	}
	if _, err := client.Competency.Create().
		SetID(compID).
		SetName("Orientación a Resultados").
		SetPillarID(pillarID).
		Save(ctx); err != nil {
		return err
	}
	for i := 1; i <= 5; i++ {
		if _, err := client.ScaleCriterion.Create().
			SetID(seedID(fmt.Sprintf("criteria-comp-%d", i))).
			SetLevel(i).
			SetDescription(fmt.Sprintf("Nivel %d de desempeño", i)).
			SetCompetencyID(compID).
			SetPillarID(pillarID).
			Save(ctx); err != nil {
			return err
		}
	}

	// Categoría, goal y KPI para emp-dg-01.
	if _, err := client.GoalCategory.Create().
		SetID(catID).
		SetName("Ventas y Finanzas").
		SetWeight(40).
		SetEmployeeID(dgID).
		SetPillarID(pillarID).
		SetCreatedBy(jefeID).
		SetUpdatedBy(jefeID).
		Save(ctx); err != nil {
		return err
	}
	if _, err := client.Goal.Create().
		SetID(goalID).
		SetName("Incrementar ventas").
		SetUnit(goal.UnitNumero).
		SetWeight(40).
		SetTargetValue(100).
		SetCurrentValue(20).
		SetDirection(goal.DirectionAscendente).
		SetState(goal.StateFijada).
		SetType(goal.TypePersonal).
		SetCategoryID(catID).
		SetCreatedBy(jefeID).
		SetUpdatedBy(jefeID).
		Save(ctx); err != nil {
		return err
	}
	if _, err := client.KPI.Create().
		SetID(kpiID).
		SetName("Ventas mensuales").
		SetUnit(kpi.UnitNumero).
		SetDirection(kpi.DirectionAscendente).
		SetCurrentValue(20).
		SetTargetValue(100).
		Save(ctx); err != nil {
		return err
	}
	if _, err := client.GoalKpiLink.Create().
		SetGoalID(goalID).
		SetKpiID(kpiID).
		Save(ctx); err != nil {
		return err
	}
	if _, err := client.GoalAssignment.Create().
		SetID(seedID("ga-dg01-2026")).
		SetEmployeeID(dgID).
		SetCycleID(cycleID).
		Save(ctx); err != nil {
		return err
	}

	// Evaluación de emp-dg-01 en el ciclo con rating de competencia.
	if _, err := client.Evaluation.Create().
		SetID(evalID).
		SetPhase(evaluation.PhaseAvance).
		SetState(evaluation.StatePendienteAvance).
		SetEmployeeID(dgID).
		SetCycleID(cycleID).
		SetCreatedBy(jefeID).
		SetUpdatedBy(jefeID).
		Save(ctx); err != nil {
		return err
	}
	if _, err := client.EvaluationCompetency.Create().
		SetID(seedID("evalcomp-dg01-2026")).
		SetRating(4).
		SetEvaluationID(evalID).
		SetCompetencyID(compID).
		SetProfileID(profileColID).
		Save(ctx); err != nil {
		return err
	}

	// Catálogo nine-box: quadrants 1..9, escalas, 2 matrices (avance/cierre) y entry.
	for i := 1; i <= 9; i++ {
		if _, err := client.NineBoxQuadrant.Create().
			SetID(seedID(fmt.Sprintf("quadrant-%d", i))).
			SetQuadrant(i).
			SetLabel(fmt.Sprintf("Cuadrante %d", i)).
			SetTitle(fmt.Sprintf("Título cuadrante %d", i)).
			SetDescription(fmt.Sprintf("Descripción del cuadrante %d", i)).
			SetColorHex(fmt.Sprintf("#%06X", i*0x111111)).
			Save(ctx); err != nil {
			return err
		}
		if _, err := client.NineBoxScale.Create().
			SetID(seedID(fmt.Sprintf("scale-perf-%d", i))).
			SetAxis(nineboxscale.AxisPerformance).
			SetLevel(i).
			SetLabel(fmt.Sprintf("Desempeño %d", i)).
			Save(ctx); err != nil {
			return err
		}
		if _, err := client.NineBoxScale.Create().
			SetID(seedID(fmt.Sprintf("scale-pot-%d", i))).
			SetAxis(nineboxscale.AxisPotential).
			SetLevel(i).
			SetLabel(fmt.Sprintf("Potencial %d", i)).
			Save(ctx); err != nil {
			return err
		}
	}
	if _, err := client.NineBoxMatrix.Create().
		SetID(seedID("matrix-dg01-2026-avance")).
		SetCycleID(cycleID).
		SetEvaluatorID(dgID).
		SetPhaseID(avancePhaseID).
		Save(ctx); err != nil {
		return err
	}
	if _, err := client.NineBoxMatrix.Create().
		SetID(seedID("matrix-dg01-2026-cierre")).
		SetCycleID(cycleID).
		SetEvaluatorID(dgID).
		SetPhaseID(cierrePhaseID).
		Save(ctx); err != nil {
		return err
	}
	if _, err := client.NineBoxEntry.Create().
		SetID(seedID("entry-dg01-2026-avance")).
		SetPerformanceTier(2).
		SetPotentialTier(2).
		SetQuadrant(5).
		SetMatrixID(seedID("matrix-dg01-2026-avance")).
		SetEvaluateeID(dgID).
		SetCreatedBy(jefeID).
		SetUpdatedBy(jefeID).
		Save(ctx); err != nil {
		return err
	}

	return nil
}
