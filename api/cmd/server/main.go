package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/auth"
	"github.com/sed-evaluacion-desempeno/api/internal/middleware"
	authsvc "github.com/sed-evaluacion-desempeno/api/internal/service/auth"

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

	// Handlers
	authhandler "github.com/sed-evaluacion-desempeno/api/internal/handler/auth"
	goalhandler "github.com/sed-evaluacion-desempeno/api/internal/handler/goal"
	cyclehandler "github.com/sed-evaluacion-desempeno/api/internal/handler/cycle"
	comphandler "github.com/sed-evaluacion-desempeno/api/internal/handler/competency"
	evalhandler "github.com/sed-evaluacion-desempeno/api/internal/handler/evaluation"
	orghandler "github.com/sed-evaluacion-desempeno/api/internal/handler/org"

	// Seed
	"github.com/sed-evaluacion-desempeno/api/internal/seed"
)

// ---------------------------------------------------------------------------
// Stub implementations for interfaces without production providers at bootstrap.
// ---------------------------------------------------------------------------

// nopPhaseChecker implements goalsvc.PhaseChecker for goal services.
// Always returns "asignacion" as the current phase.
type nopPhaseChecker struct{}

func (nopPhaseChecker) GetCurrentPhase(_ context.Context, _ string) (goalsvc.CyclePhase, error) {
	return goalsvc.PhaseAsignacion, nil
}

// evalCyclePhaseCheck implements evalsvc.CyclePhaseChecker backed by the cycle repo.
type evalCyclePhaseCheck struct {
	cycleRepo *repocycle.CycleRepo
}

func (c *evalCyclePhaseCheck) GetPhase(ctx context.Context, cycleID uuid.UUID) (string, error) {
	row, err := c.cycleRepo.GetCycle(ctx, cycleID)
	if err != nil {
		return "", err
	}
	return string(row.CurrentPhase), nil
}

func (c *evalCyclePhaseCheck) GetSelfEvalDeadline(_ context.Context, _ uuid.UUID) (*time.Time, error) {
	// CycleRow does not expose SelfEvalEndsAt yet; return nil (no deadline).
	return nil, nil
}

// inMemoryIdempotencyCache implements evalsvc.IdempotencyCache for the evaluation service.
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

func main() {
	// -----------------------------------------------------------------------
	// Env & flags
	// -----------------------------------------------------------------------
	flag.Parse()

	_ = godotenv.Load() // ignore error if .env does not exist

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("[server] DATABASE_URL is required")
	}
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}
	corsOrigins := os.Getenv("CORS_ORIGINS")
	if corsOrigins == "" {
		corsOrigins = "http://localhost:5173"
	}

	// -----------------------------------------------------------------------
	// Database & Ent client
	// -----------------------------------------------------------------------
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("[server] failed to open database: %v", err)
	}
	defer db.Close()

	drv := entsql.OpenDB(dialect.Postgres, db)
	client := internal.NewClient(internal.Driver(drv))

	// Auto-migrate
	bgCtx := context.Background()
	if err := client.Schema.Create(bgCtx); err != nil {
		log.Fatalf("[server] failed to auto-migrate: %v", err)
	}
	log.Println("[server] schema migrated")

	// Seeder — seed.Run handles its own guards (flag --seed, SEED_ON_START env, empty-DB check)
	if err := seed.Run(bgCtx, client); err != nil {
		log.Printf("[seed] error: %v", err)
	}

	// -----------------------------------------------------------------------
	// Dependency Injection — Repositories
	// -----------------------------------------------------------------------

	// Goal
	catRepo := repogoal.NewCategoryRepo(client, db)
	goalRepo := repogoal.NewGoalRepo(client, db)
	kpiRepo := repogoal.NewKpiRepo(client, db)
	linkRepo := repogoal.NewLinkKpiRepo(client, db)
	assignRepo := repogoal.NewAssignmentRepo(client, db)
	weightQ := repogoal.NewWeightQueries(db)

	// Cycle
	cycleRepo := repocycle.NewCycleRepo(client, db)
	phaseRepo := repocycle.NewPhaseRepo(client, db)

	// Competency
	pillarRepo := repocompetency.NewPillarRepo(client)
	compRepo := repocompetency.NewCompetencyRepo(client)
	scaleRepo := repocompetency.NewScaleRepo(client)
	catalogCompRepo := repocompetency.NewCatalogRepo(client)
	acceptanceRepo := repocompetency.NewAcceptanceRepo(client)

	// Evaluation
	evalRepo := repoeval.NewEvaluationRepo(client, db)
	compRatingRepo := repoeval.NewCompetencyRatingRepo(client)
	goalRatingRepo := repoeval.NewGoalRatingRepo(client)
	nineBoxRepo := repoeval.NewNineBoxRepo(client, db)
	catalogEvalRepo := repoeval.NewCatalogRepo(client)

	// Org
	orgTreeRepo := repoorganization.NewOrgTreeRepo(client, db)
	orgNodeRepo := repoorganization.NewOrgNodeRepo(client, db)
	employeeRepo := repoorganization.NewEmployeeRepo(client, db)
	scopeRepo := repoorganization.NewEvaluatorScopeRepo(client, db)

	// Auth infrastructure
	sessionStore := auth.NewSessionStore(db)
	employeeReader := authsvc.NewEmployeeReader(db)

	// -----------------------------------------------------------------------
	// Dependency Injection — Services
	// -----------------------------------------------------------------------

	// Auth
	authSvc := authsvc.NewAuthService(sessionStore, employeeReader, db)

	// Goal services
	phaseChecker := nopPhaseChecker{}
	phaseCheck := goalsvc.NewPhaseCheck(phaseChecker)

	catSvc := goalsvc.NewCategoryService(catRepo, phaseCheck)
	goalSvc := goalsvc.NewGoalService(goalRepo, catRepo, kpiRepo, linkRepo, weightQ, phaseCheck)
	progressSvc := goalsvc.NewProgressService(goalRepo, catRepo, phaseCheck)
	kpiSvc := goalsvc.NewKPIService(kpiRepo, linkRepo, goalRepo, catRepo, phaseCheck)
	weightSvc := goalsvc.NewWeightValidationService(catRepo, goalRepo)
	batchSvc := goalsvc.NewBatchService(goalRepo, catRepo, kpiRepo, linkRepo, weightQ, phaseCheck)

	// Cycle services
	cycleSvc := cyclesvc.NewService(cycleRepo, phaseRepo, client)
	phaseSvc := cyclesvc.NewPhaseService(cycleRepo, phaseRepo)

	// Competency services
	pillarSvc := compsvc.NewPillarService(pillarRepo)
	competencySvc := compsvc.NewCompetencyService(pillarRepo, compRepo)
	scaleSvc := compsvc.NewScaleService(compRepo, scaleRepo)
	catalogSvc := compsvc.NewCatalogService(catalogCompRepo)
	acceptanceSvc := compsvc.NewAcceptanceService(compRepo, catalogCompRepo, acceptanceRepo)

	// Evaluation services
	cycleCheck := &evalCyclePhaseCheck{cycleRepo: cycleRepo}
	idemCache := newInMemoryIdempotencyCache()

	evalSvc := evalsvc.NewEvaluationService(evalRepo, compRatingRepo, goalRatingRepo, cycleCheck, idemCache)
	nineBoxSvc := evalsvc.NewNineBoxService(nineBoxRepo, catalogEvalRepo, db)
	dashboardSvc := evalsvc.NewDashboardService(evalRepo)

	// Org services
	orgTreeSvc := orgsvc.NewOrgTreeService(orgTreeRepo, orgNodeRepo, client)
	orgNodeSvc := orgsvc.NewOrgNodeService(orgNodeRepo, client)
	employeeSvc := orgsvc.NewEmployeeService(employeeRepo, client)
	evaluateeSvc := orgsvc.NewEvaluateeService(employeeRepo, orgNodeRepo, scopeRepo, client)
	evaluatorSvc := orgsvc.NewEvaluatorService(employeeRepo, orgNodeRepo, scopeRepo, client)

	// -----------------------------------------------------------------------
	// Dependency Injection — Handlers
	// -----------------------------------------------------------------------

	authH := authhandler.NewAuthHandler(authSvc)
	goalH := goalhandler.NewGoalHandler(
		catSvc, goalSvc, progressSvc, kpiSvc, weightSvc, batchSvc,
		catRepo, goalRepo, kpiRepo, linkRepo, assignRepo,
	)
	cycleH := cyclehandler.NewCycleHandler(cycleSvc, phaseSvc)
	compH := comphandler.NewHandler(pillarSvc, competencySvc, scaleSvc, catalogSvc, acceptanceSvc)
	evalH := evalhandler.NewEvaluationHandler(evalSvc, nineBoxSvc, dashboardSvc)
	orgH := orghandler.NewOrgHandler(orgTreeSvc, orgNodeSvc, employeeSvc, evaluateeSvc, evaluatorSvc)

	// -----------------------------------------------------------------------
	// Router
	// -----------------------------------------------------------------------
	r := chi.NewRouter()

	// Global middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   strings.Split(corsOrigins, ","),
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Idempotency-Key"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)

	// Mount handler routes
	// competency.RegisterRoutes calls r.Use() internally — wrap in Group for a clean subrouter.
	r.Group(func(r chi.Router) {
		comphandler.RegisterRoutes(r, &comphandler.Dependencies{Handler: compH})
	})
	r.Mount("/api/v1/auth", authhandler.AuthRoutes(authH))
	r.Mount("/", goalhandler.NewRouter(goalH))

	// Cycle, evaluation, and org: register all on a single apiV1 subrouter.
	// Apply shared AuthPlaceholder middleware once for all three.
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

	// -----------------------------------------------------------------------
	// Server with graceful shutdown
	// -----------------------------------------------------------------------
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("[server] listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[server] error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("[server] shutting down...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("[server] shutdown error: %v", err)
	}
}
