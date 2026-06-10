package integration

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
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
	"github.com/sed-evaluacion-desempeno/api/internal/middleware"
	"github.com/sed-evaluacion-desempeno/api/internal/seed"

	// Repositories
	repogoal "github.com/sed-evaluacion-desempeno/api/internal/repository/goal"
	repocycle "github.com/sed-evaluacion-desempeno/api/internal/repository/cycle"
	repocompetency "github.com/sed-evaluacion-desempeno/api/internal/repository/competency"
	repoeval "github.com/sed-evaluacion-desempeno/api/internal/repository/evaluation"
	repoorganization "github.com/sed-evaluacion-desempeno/api/internal/repository/org"

	// Services
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
)

// nopPhaseChecker always returns "asignacion" as the current phase.
type nopPhaseChecker struct{}

func (nopPhaseChecker) GetCurrentPhase(_ context.Context, _ string) (goalsvc.CyclePhase, error) {
	return goalsvc.PhaseAsignacion, nil
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
}

// setupTestServer builds the full router with real DB and returns a testServer.
// It connects to the PostgreSQL instance specified by DATABASE_URL, runs
// auto-migration, seeds data, and wires all handlers.
func setupTestServer(t *testing.T) *testServer {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://sed:sed@localhost:5432/sed?sslmode=disable"
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		t.Fatalf("failed to ping database: %v", err)
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
	if err := seed.Run(context.Background(), client); err != nil {
		log.Printf("[seed] warning: %v", err)
	}

	// --- DI wiring (mirrors main.go) ---

	// Repos
	catRepo := repogoal.NewCategoryRepo(client, db)
	goalRepo := repogoal.NewGoalRepo(client, db)
	kpiRepo := repogoal.NewKpiRepo(client, db)
	linkRepo := repogoal.NewLinkKpiRepo(client, db)
	assignRepo := repogoal.NewAssignmentRepo(client, db)
	weightQ := repogoal.NewWeightQueries(db)

	cycleRepo := repocycle.NewCycleRepo(client, db)
	phaseRepo := repocycle.NewPhaseRepo(client, db)

	pillarRepo := repocompetency.NewPillarRepo(client)
	compRepo := repocompetency.NewCompetencyRepo(client)
	scaleRepo := repocompetency.NewScaleRepo(client)
	catalogCompRepo := repocompetency.NewCatalogRepo(client)
	acceptanceRepo := repocompetency.NewAcceptanceRepo(client)

	evalRepo := repoeval.NewEvaluationRepo(client, db)
	compRatingRepo := repoeval.NewCompetencyRatingRepo(client)
	goalRatingRepo := repoeval.NewGoalRatingRepo(client)
	nineBoxRepo := repoeval.NewNineBoxRepo(client, db)
	catalogEvalRepo := repoeval.NewCatalogRepo(client)

	orgTreeRepo := repoorganization.NewOrgTreeRepo(client, db)
	orgNodeRepo := repoorganization.NewOrgNodeRepo(client, db)
	employeeRepo := repoorganization.NewEmployeeRepo(client, db)
	scopeRepo := repoorganization.NewEvaluatorScopeRepo(client, db)

	sessionStore := auth.NewSessionStore(db)
	employeeReader := authsvc.NewEmployeeReader(db)

	// Services
	authSvc := authsvc.NewAuthService(sessionStore, employeeReader, db)

	phaseChecker := nopPhaseChecker{}
	phaseCheck := goalsvc.NewPhaseCheck(phaseChecker)
	catSvc := goalsvc.NewCategoryService(catRepo, phaseCheck)
	goalSvc := goalsvc.NewGoalService(goalRepo, catRepo, kpiRepo, linkRepo, weightQ, phaseCheck)
	progressSvc := goalsvc.NewProgressService(goalRepo, catRepo, phaseCheck)
	kpiSvc := goalsvc.NewKPIService(kpiRepo, linkRepo, goalRepo, catRepo, phaseCheck)
	weightSvc := goalsvc.NewWeightValidationService(catRepo, goalRepo)
	batchSvc := goalsvc.NewBatchService(goalRepo, catRepo, kpiRepo, linkRepo, weightQ, phaseCheck)

	cycleSvc := cyclesvc.NewService(cycleRepo, phaseRepo, client)
	phaseSvc := cyclesvc.NewPhaseService(cycleRepo, phaseRepo)

	pillarSvc := compsvc.NewPillarService(pillarRepo)
	competencySvc := compsvc.NewCompetencyService(pillarRepo, compRepo)
	scaleSvc := compsvc.NewScaleService(compRepo, scaleRepo)
	catalogSvc := compsvc.NewCatalogService(catalogCompRepo)
	acceptanceSvc := compsvc.NewAcceptanceService(compRepo, catalogCompRepo, acceptanceRepo)

	cycleCheck := &evalCyclePhaseCheck{cycleRepo: cycleRepo}
	idemCache := newInMemoryIdempotencyCache()
	evalSvc := evalsvc.NewEvaluationService(evalRepo, compRatingRepo, goalRatingRepo, cycleCheck, idemCache)
	nineBoxSvc := evalsvc.NewNineBoxService(nineBoxRepo, catalogEvalRepo, db)
	dashboardSvc := evalsvc.NewDashboardService(evalRepo)

	orgTreeSvc := orgsvc.NewOrgTreeService(orgTreeRepo, orgNodeRepo, client)
	orgNodeSvc := orgsvc.NewOrgNodeService(orgNodeRepo, client)
	employeeSvc := orgsvc.NewEmployeeService(employeeRepo, client)
	evaluateeSvc := orgsvc.NewEvaluateeService(employeeRepo, orgNodeRepo, scopeRepo, client)
	evaluatorSvc := orgsvc.NewEvaluatorService(employeeRepo, orgNodeRepo, scopeRepo, client)

	// Handlers
	authH := authhandler.NewAuthHandler(authSvc)
	goalH := goalhandler.NewGoalHandler(
		catSvc, goalSvc, progressSvc, kpiSvc, weightSvc, batchSvc,
		catRepo, goalRepo, kpiRepo, linkRepo, assignRepo,
	)
	cycleH := cyclehandler.NewCycleHandler(cycleSvc, phaseSvc)
	compH := comphandler.NewHandler(pillarSvc, competencySvc, scaleSvc, catalogSvc, acceptanceSvc)
	evalH := evalhandler.NewEvaluationHandler(evalSvc, nineBoxSvc, dashboardSvc)
	orgH := orghandler.NewOrgHandler(orgTreeSvc, orgNodeSvc, employeeSvc, evaluateeSvc, evaluatorSvc)

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
		comphandler.RegisterRoutes(r, &comphandler.Dependencies{Handler: compH})
	})
	r.Mount("/api/v1/auth", authhandler.AuthRoutes(authH))
	r.Mount("/", goalhandler.NewRouter(goalH))

	apiV1 := chi.NewRouter()
	apiV1.Use(middleware.AuthPlaceholder)
	cyclehandler.RegisterRoutes(apiV1, cycleH)
	evalhandler.RegisterRoutes(apiV1, evalH)
	orghandler.RegisterRoutes(apiV1, orgH)
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

	return &testServer{
		Client: client,
		DB:     db,
		Router: r,
		Clean:  clean,
	}
}
