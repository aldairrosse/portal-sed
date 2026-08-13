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
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"

	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/auth"
	"github.com/sed-evaluacion-desempeno/api/internal/auth/sso"
	authsvc "github.com/sed-evaluacion-desempeno/api/internal/service/auth"

	// Repositories
	repoactivity "github.com/sed-evaluacion-desempeno/api/internal/repository/activity"
	repocompetency "github.com/sed-evaluacion-desempeno/api/internal/repository/competency"
	repocycle "github.com/sed-evaluacion-desempeno/api/internal/repository/cycle"
	repoeval "github.com/sed-evaluacion-desempeno/api/internal/repository/evaluation"
	repogoal "github.com/sed-evaluacion-desempeno/api/internal/repository/goal"
	repoorganization "github.com/sed-evaluacion-desempeno/api/internal/repository/org"

	// Services
	activitysvc "github.com/sed-evaluacion-desempeno/api/internal/service/activity"
	compsvc "github.com/sed-evaluacion-desempeno/api/internal/service/competency"
	cyclesvc "github.com/sed-evaluacion-desempeno/api/internal/service/cycle"
	evalsvc "github.com/sed-evaluacion-desempeno/api/internal/service/evaluation"
	goalsvc "github.com/sed-evaluacion-desempeno/api/internal/service/goal"
	notifypkg "github.com/sed-evaluacion-desempeno/api/internal/service/notify"
	orgsvc "github.com/sed-evaluacion-desempeno/api/internal/service/org"

	// Handlers
	activityhandler "github.com/sed-evaluacion-desempeno/api/internal/handler/activity"
	authhandler "github.com/sed-evaluacion-desempeno/api/internal/handler/auth"
	commentchangehandler "github.com/sed-evaluacion-desempeno/api/internal/handler/commentchange"
	comphandler "github.com/sed-evaluacion-desempeno/api/internal/handler/competency"
	cyclehandler "github.com/sed-evaluacion-desempeno/api/internal/handler/cycle"
	evalhandler "github.com/sed-evaluacion-desempeno/api/internal/handler/evaluation"
	goalhandler "github.com/sed-evaluacion-desempeno/api/internal/handler/goal"
	orghandler "github.com/sed-evaluacion-desempeno/api/internal/handler/org"

	// Middleware
	"github.com/sed-evaluacion-desempeno/api/internal/middleware"

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

func (c *evalCyclePhaseCheck) GetActiveCycleID(ctx context.Context, orgID uuid.UUID) (uuid.UUID, error) {
	return c.cycleRepo.GetActiveCycleID(ctx, orgID)
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

	// SQL migrations (goose .up.sql files)
	bgCtx := context.Background()
	if err := runSQLMigrations(bgCtx, db); err != nil {
		log.Fatalf("[server] failed to run SQL migrations: %v", err)
	}

	// ponytail: drop materialized views and indexes that block Ent column type alters
	db.ExecContext(bgCtx, `DROP MATERIALIZED VIEW IF EXISTS evaluation_summary`)
	db.ExecContext(bgCtx, `DROP INDEX IF EXISTS idx_org_nodes_path`)
	// ponytail: Ent can't auto-cast between enum types; ensure custom PG types before migration
	for _, stmt := range []string{
		// phase
		`ALTER TABLE cycles ALTER COLUMN current_phase TYPE phase USING current_phase::text::phase`,
		`ALTER TABLE evaluations ALTER COLUMN phase TYPE phase USING phase::text::phase`,
		`ALTER TABLE phase_definitions ALTER COLUMN phase TYPE phase USING phase::text::phase`,
		`ALTER TABLE phase_transitions ALTER COLUMN from_phase TYPE phase USING from_phase::text::phase`,
		`ALTER TABLE phase_transitions ALTER COLUMN to_phase TYPE phase USING to_phase::text::phase`,
		// evaluation_state
		`ALTER TABLE evaluations ALTER COLUMN state TYPE evaluation_state USING state::text::evaluation_state`,
		// goal_unit
		`ALTER TABLE goals ALTER COLUMN unit TYPE goal_unit USING unit::text::goal_unit`,
		`ALTER TABLE kp_is ALTER COLUMN unit TYPE goal_unit USING unit::text::goal_unit`,
		// goal_state
		`ALTER TABLE goals ALTER COLUMN state TYPE goal_state USING state::text::goal_state`,
		// org_node_type
		`ALTER TABLE org_nodes ALTER COLUMN type TYPE org_node_type USING type::text::org_node_type`,
		// axis
		`ALTER TABLE nine_box_scales ALTER COLUMN axis TYPE axis USING axis::text::axis`,
		// trigger_type
		`ALTER TABLE phase_transitions ALTER COLUMN "trigger" TYPE trigger_type USING "trigger"::text::trigger_type`,
	} {
		db.ExecContext(bgCtx, stmt)
	}

	if err := client.Schema.Create(bgCtx); err != nil {
		log.Fatalf("[server] failed to auto-migrate: %v", err)
	}
	log.Println("[server] schema migrated")

	db.ExecContext(bgCtx, `
		CREATE MATERIALIZED VIEW IF NOT EXISTS evaluation_summary AS
		SELECT cycle_id, state, COUNT(1) as count
		FROM evaluations
		GROUP BY cycle_id, state
		WITH DATA
	`)
	db.ExecContext(bgCtx, `
		CREATE UNIQUE INDEX IF NOT EXISTS idx_evaluation_summary_cycle_state
		ON evaluation_summary (cycle_id, state)
	`)
	// ponytail: Ent changes path from ltree→varchar; restore ltree + gist index
	db.ExecContext(bgCtx, `ALTER TABLE org_nodes ALTER COLUMN path TYPE ltree USING path::ltree`)
	if _, err := db.ExecContext(bgCtx, `
		CREATE INDEX IF NOT EXISTS idx_org_nodes_path
		ON org_nodes USING gist (path)
	`); err != nil {
		log.Printf("[server] warn: failed to recreate idx_org_nodes_path: %v", err)
	}

	// Seeder — seed.Run handles its own guards
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
	proposalRepo := repogoal.NewGoalProposalRepo(db)
	weightQ := repogoal.NewWeightQueries(db)
	globalGoalRepo := repogoal.NewGlobalGoalRepo(client, db)
	sharedGoalRepo := repogoal.NewSharedGoalRepo(client, db)

	// Cycle
	cycleRepo := repocycle.NewCycleRepo(client, db)
	phaseRepo := repocycle.NewPhaseRepo(client, db)

	// Competency
	pillarRepo := repocompetency.NewPillarRepo(client)
	compRepo := repocompetency.NewCompetencyRepo(client)
	scaleRepo := repocompetency.NewScaleRepo(client)
	catalogCompRepo := repocompetency.NewCatalogRepo(client, db)
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

	// Activity
	activityRepo := repoactivity.NewRepository(client)
	activitySvc := activitysvc.NewService(activityRepo)
	activityH := activityhandler.NewActivityHandler(activitySvc)

	// Auth infrastructure
	sessionStore := auth.NewSessionStore(db)
	employeeReader := authsvc.NewEmployeeReader(db)

	// -----------------------------------------------------------------------
	// Dependency Injection — Services
	// -----------------------------------------------------------------------

	// Auth
	authSvc := authsvc.NewAuthService(sessionStore, employeeReader, db)

	// SSO (OIDC) — required for all production logins
	ssoIssuer := os.Getenv("SSO_KC_ISSUER")
	ssoClientID := os.Getenv("SSO_CLIENT_ID")
	ssoClientSecret := os.Getenv("SSO_CLIENT_SECRET")
	ssoRedirectURI := os.Getenv("SSO_REDIRECT_URI")
	ssoPostLogoutURI := os.Getenv("SSO_POST_LOGOUT_URI")
	if ssoIssuer == "" || ssoClientID == "" || ssoClientSecret == "" || ssoRedirectURI == "" {
		log.Fatal("[server] SSO_KC_ISSUER, SSO_CLIENT_ID, SSO_CLIENT_SECRET, SSO_REDIRECT_URI are required")
	}
	ssoAdapter, err := sso.NewOIDCAdapter(bgCtx, ssoIssuer, ssoClientID, ssoClientSecret, ssoRedirectURI, ssoPostLogoutURI)
	if err != nil {
		log.Fatalf("[server] failed to init SSO adapter: %v", err)
	}
	log.Printf("[server] SSO adapter initialized (issuer: %s, client: %s)", ssoIssuer, ssoClientID)
	authSvc.WithSSOValidator(ssoAdapter).WithSSORevalidator(ssoAdapter)

	// Goal services
	phaseChecker := goalsvc.NewCyclePhaseCheck(cycleRepo, employeeRepo, orgNodeRepo)
	phaseCheck := goalsvc.NewPhaseCheck(phaseChecker)

	catSvc := goalsvc.NewCategoryService(catRepo, pillarRepo, phaseCheck)
	goalSvc := goalsvc.NewGoalService(goalRepo, catRepo, kpiRepo, linkRepo, weightQ, phaseCheck)
	progressSvc := goalsvc.NewProgressService(goalRepo, catRepo, phaseCheck)
	kpiSvc := goalsvc.NewKPIService(kpiRepo, linkRepo, goalRepo, catRepo, phaseCheck, orgNodeRepo, orgTreeRepo, employeeRepo)
	scoringSvc := goalsvc.NewScoringService(catRepo, goalRepo)
	weightSvc := goalsvc.NewWeightValidationService(catRepo, goalRepo)
	batchSvc := goalsvc.NewBatchService(goalRepo, catRepo, kpiRepo, linkRepo, weightQ, phaseCheck)
	proposalSvc := goalsvc.NewGoalProposalService(proposalRepo, goalRepo, catRepo, linkRepo, weightQ, phaseCheck, db)
	globalGoalSvc := goalsvc.NewGlobalGoalService(globalGoalRepo)
	sharedGoalSvc := goalsvc.NewSharedGoalService(sharedGoalRepo)

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

	evalSvc := evalsvc.NewEvaluationService(evalRepo, compRatingRepo, goalRatingRepo, cycleCheck, idemCache, employeeRepo, orgNodeRepo)
	nineBoxSvc := evalsvc.NewNineBoxService(nineBoxRepo, catalogEvalRepo, db)
	dashboardSvc := evalsvc.NewDashboardService(evalRepo)

	// Org services
	orgTreeSvc := orgsvc.NewOrgTreeService(orgTreeRepo, orgNodeRepo, employeeRepo, client)
	orgNodeSvc := orgsvc.NewOrgNodeService(orgNodeRepo, client)
	employeeSvc := orgsvc.NewEmployeeService(employeeRepo, client)
	evaluateeSvc := orgsvc.NewEvaluateeService(employeeRepo, orgNodeRepo, client)
	metricsRepo := repoorganization.NewMetricsRepo(client, db)
	metricsSvc := orgsvc.NewMetricsService(metricsRepo, orgNodeRepo, client)

	// -----------------------------------------------------------------------
	// Dependency Injection — Handlers
	// -----------------------------------------------------------------------

	authH := authhandler.NewAuthHandler(authSvc, ssoAdapter)
	goalH := goalhandler.NewGoalHandler(
		catSvc, goalSvc, progressSvc, kpiSvc, scoringSvc, weightSvc, batchSvc, proposalSvc,
		catRepo, goalRepo, kpiRepo, linkRepo, assignRepo, proposalRepo, activitySvc, evalSvc,
	)
	cycleH := cyclehandler.NewCycleHandler(cycleSvc, phaseSvc, activitySvc, assignRepo, employeeRepo)
	compH := comphandler.NewHandler(pillarSvc, competencySvc, scaleSvc, catalogSvc, acceptanceSvc, activitySvc)
	evalH := evalhandler.NewEvaluationHandler(evalSvc, nineBoxSvc, dashboardSvc, activitySvc)
	orgH := orghandler.NewOrgHandler(orgTreeSvc, orgNodeSvc, employeeSvc, evaluateeSvc, metricsSvc)
	commentChangeH := commentchangehandler.NewHandler(db, notifypkg.NoopSender{})
	globalGoalH := goalhandler.NewGlobalGoalHandler(globalGoalSvc)
	sharedGoalH := goalhandler.NewSharedGoalHandler(sharedGoalSvc)

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
		comphandler.RegisterRoutes(r, &comphandler.Dependencies{Handler: compH, AuthSvc: authSvc})
	})
	r.Mount("/api/v1/auth", authhandler.AuthRoutes(authH, authSvc))

	// Cycle, evaluation, org, goals, and activity: register all on a single apiV1 subrouter.
	// Each handler applies its own RequireAuth middleware.
	apiV1 := chi.NewRouter()
	cyclehandler.RegisterRoutes(apiV1, cycleH, authSvc)
	evalhandler.RegisterRoutes(apiV1, evalH, authSvc)
	orghandler.RegisterRoutes(apiV1, orgH, authSvc)
	goalhandler.RegisterRoutes(apiV1, goalH, authSvc)
	apiV1.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(authSvc))
		r.Use(middleware.RequireLoA2())
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(auth.PermGoalGlobal))
			globalGoalH.RegisterRoutes(r)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(auth.PermGoalShared))
			sharedGoalH.RegisterRoutes(r)
		})
	})
	commentchangehandler.RegisterRoutes(apiV1, commentChangeH, authSvc)
	activityhandler.RegisterActivityRoutes(apiV1, activityH, authSvc)
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
