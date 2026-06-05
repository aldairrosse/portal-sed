package evaluation_test

import (
	"context"
	"database/sql"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	dto "github.com/sed-evaluacion-desempeno/api/internal/dto/evaluation"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	repo "github.com/sed-evaluacion-desempeno/api/internal/repository/evaluation"
	svc "github.com/sed-evaluacion-desempeno/api/internal/service/evaluation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------- Mock Repositories ----------

type mockEvalRepo struct {
	db        *sql.DB
	sqlmock   sqlmock.Sqlmock
	row       *repo.EvaluationRow
	detailRow *repo.EvaluationRow
	comps     []*internal.EvaluationCompetency
	goals     []*internal.EvaluationGoal
	summary   map[string]int64
	submitErr error
	finalizeErr error
	refreshErr  error
	state     string
	mu        sync.Mutex
}

func (m *mockEvalRepo) GetByID(ctx context.Context, id uuid.UUID) (*repo.EvaluationRow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.row == nil {
		return nil, repo.ErrEvaluationNotFound
	}
	r := *m.row
	r.State = m.state
	return &r, nil
}

func (m *mockEvalRepo) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return m.db.BeginTx(ctx, opts)
}

func (m *mockEvalRepo) LockEvalForUpdate(ctx context.Context, tx *sql.Tx, evalID uuid.UUID) (*repo.EvaluationRow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.row == nil {
		return nil, repo.ErrEvaluationNotFound
	}
	r := *m.row
	r.State = m.state
	return &r, nil
}

func (m *mockEvalRepo) SubmitEval(ctx context.Context, tx *sql.Tx, evalID uuid.UUID, comps []repo.CompetencyUpsert, goals []repo.GoalCommentUpsert, newState string, setSelfCompleted, setRHCompleted bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state = newState
	if setSelfCompleted {
		now := time.Now()
		m.row.SelfEvaluationCompletedAt = &now
	}
	if setRHCompleted {
		now := time.Now()
		m.row.RhEvaluationCompletedAt = &now
	}
	return m.submitErr
}

func (m *mockEvalRepo) GetDetail(ctx context.Context, id uuid.UUID) (*repo.EvaluationRow, []*internal.EvaluationCompetency, []*internal.EvaluationGoal, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.detailRow == nil {
		return nil, nil, nil, repo.ErrEvaluationNotFound
	}
	r := *m.detailRow
	r.State = m.state
	return &r, m.comps, m.goals, nil
}

func (m *mockEvalRepo) ListByCycle(ctx context.Context, cycleID uuid.UUID, state string, cursor string, limit int) ([]*repo.EvaluationRow, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.row == nil {
		return []*repo.EvaluationRow{}, "", nil
	}
	return []*repo.EvaluationRow{m.row}, "", nil
}

func (m *mockEvalRepo) FinalizeEval(ctx context.Context, tx *sql.Tx, evalID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state = "completada"
	now := time.Now()
	m.row.RhEvaluationCompletedAt = &now
	return m.finalizeErr
}

func (m *mockEvalRepo) RefreshSummaryView(ctx context.Context) error {
	return m.refreshErr
}

func (m *mockEvalRepo) GetSummaryByCycle(ctx context.Context, cycleID uuid.UUID) (map[string]int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.summary == nil {
		return map[string]int64{}, nil
	}
	return m.summary, nil
}

// ---------- Mock Cycle Phase Checker ----------

type mockCycleChecker struct {
	phase    string
	deadline *time.Time
	err      error
}

func (m *mockCycleChecker) GetPhase(ctx context.Context, cycleID uuid.UUID) (string, error) {
	return m.phase, m.err
}

func (m *mockCycleChecker) GetSelfEvalDeadline(ctx context.Context, cycleID uuid.UUID) (*time.Time, error) {
	return m.deadline, m.err
}

// ---------- Mock Idempotency Cache ----------

type mockIdemCache struct {
	entries map[string]*svc.IdempotencyCacheEntry
	mu      sync.Mutex
}

func (m *mockIdemCache) Get(ctx context.Context, key string) (*svc.IdempotencyCacheEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if e, ok := m.entries[key]; ok {
		return e, nil
	}
	return nil, nil
}

func (m *mockIdemCache) Set(ctx context.Context, key string, entry *svc.IdempotencyCacheEntry, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries[key] = entry
	return nil
}

// ---------- Mock NineBox Repos ----------

type mockNineBoxRepo struct {
	matrix      *internal.NineBoxMatrix
	matrices    []*internal.NineBoxMatrix
	entry       *internal.NineBoxEntry
	entries     []*internal.NineBoxEntry
	version     int
	lockErr     error
	upsertErr   error
	updateErr   error
	batchErr    error
	fetchVerErr error
	mu          sync.Mutex
}

func (m *mockNineBoxRepo) CreateMatrix(ctx context.Context, cycleID, evaluatorID uuid.UUID) (*internal.NineBoxMatrix, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.matrix, nil
}

func (m *mockNineBoxRepo) GetMatrixByID(ctx context.Context, id uuid.UUID) (*internal.NineBoxMatrix, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.matrix, nil
}

func (m *mockNineBoxRepo) ListMatrices(ctx context.Context, cycleID, evaluatorID uuid.UUID) ([]*internal.NineBoxMatrix, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.matrices, nil
}

func (m *mockNineBoxRepo) GetMatrixEntries(ctx context.Context, matrixID uuid.UUID) ([]*internal.NineBoxEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.entries, nil
}

func (m *mockNineBoxRepo) UpsertEntry(ctx context.Context, tx *sql.Tx, matrixID uuid.UUID, evaluateeID uuid.UUID, perf, pot int, quadrant int, comments string) (*internal.NineBoxEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.upsertErr != nil {
		return nil, m.upsertErr
	}
	return m.entry, nil
}

func (m *mockNineBoxRepo) UpdateEntry(ctx context.Context, tx *sql.Tx, entryID uuid.UUID, perf, pot int, quadrant int, comments string, version int) (*internal.NineBoxEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	return m.entry, nil
}

func (m *mockNineBoxRepo) BatchUpsertEntries(ctx context.Context, tx *sql.Tx, matrixID uuid.UUID, items []repo.EntryUpsert) ([]*internal.NineBoxEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.batchErr != nil {
		return nil, m.batchErr
	}
	return m.entries, nil
}

func (m *mockNineBoxRepo) LockEntryForSelect(ctx context.Context, tx *sql.Tx, matrixID, evaluateeID uuid.UUID) error {
	return m.lockErr
}

func (m *mockNineBoxRepo) FetchEntryVersion(ctx context.Context, entryID uuid.UUID) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.version, m.fetchVerErr
}

// ---------- Mock Catalog Repo ----------

type mockCatalogRepo struct {
	quadrants []*internal.NineBoxQuadrant
	scales    []*internal.NineBoxScale
}

func (m *mockCatalogRepo) GetQuadrants(ctx context.Context) ([]*internal.NineBoxQuadrant, error) {
	return m.quadrants, nil
}

func (m *mockCatalogRepo) GetQuadrantByNumber(ctx context.Context, quadrant int) (*internal.NineBoxQuadrant, error) {
	for _, q := range m.quadrants {
		if q.Quadrant == quadrant {
			return q, nil
		}
	}
	return nil, nil
}

func (m *mockCatalogRepo) GetScales(ctx context.Context) ([]*internal.NineBoxScale, error) {
	return m.scales, nil
}

// ---------- Mock DB ----------

type mockDB struct {
	db *sql.DB
}

func (m *mockDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return m.db.BeginTx(ctx, opts)
}

// ---------- Tests: EvaluationService ----------

func TestEvaluationService_SubmitSelfEvaluation_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	cycleID := uuid.New()
	evalID := uuid.New()
	compID := uuid.New()
	now := time.Now()

	row := &repo.EvaluationRow{
		ID:        evalID,
		CycleID:   cycleID,
		State:     "pendiente_evaluacion_final",
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockRepo := &mockEvalRepo{
		db:        db,
		sqlmock:   mock,
		row:       row,
		detailRow: row,
		comps: []*internal.EvaluationCompetency{
			{CompetencyID: compID, Rating: 4, Comments: "Good"},
		},
		goals: []*internal.EvaluationGoal{},
		state: "pendiente_evaluacion_final",
	}

	mock.ExpectBegin()
	mock.ExpectCommit()

	checker := &mockCycleChecker{phase: "cierre"}
	idem := &mockIdemCache{entries: make(map[string]*svc.IdempotencyCacheEntry)}
	service := svc.NewEvaluationService(mockRepo, nil, nil, checker, idem)

	req := dto.SelfEvaluationRequest{
		Competencies: []dto.CompetencyRatingInput{
			{CompetencyID: compID, Rating: 4, Comments: "Good"},
		},
	}

	resp, err := service.SubmitSelfEvaluation(context.Background(), evalID, req, "idem-key-1")
	require.NoError(t, err)
	assert.Equal(t, evalID, resp.ID)
	assert.Equal(t, "en_progreso", resp.State)
}

func TestEvaluationService_SubmitSelfEvaluation_WrongPhase(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	cycleID := uuid.New()
	evalID := uuid.New()
	now := time.Now()

	row := &repo.EvaluationRow{
		ID:        evalID,
		CycleID:   cycleID,
		State:     "pendiente_evaluacion_final",
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockRepo := &mockEvalRepo{db: db, row: row, state: "pendiente_evaluacion_final"}
	checker := &mockCycleChecker{phase: "avance"}
	service := svc.NewEvaluationService(mockRepo, nil, nil, checker, nil)

	req := dto.SelfEvaluationRequest{
		Competencies: []dto.CompetencyRatingInput{
			{CompetencyID: uuid.New(), Rating: 4},
		},
	}

	_, err = service.SubmitSelfEvaluation(context.Background(), evalID, req, "")
	require.Error(t, err)
	var de *pkgerrors.DomainError
	require.True(t, pkgerrors.AsDomainError(err, &de))
	assert.Equal(t, pkgerrors.PhaseNotAdvanceable, de.Code)
}

func TestEvaluationService_SubmitRHEvaluation_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	cycleID := uuid.New()
	evalID := uuid.New()
	compID := uuid.New()
	now := time.Now()

	row := &repo.EvaluationRow{
		ID:        evalID,
		CycleID:   cycleID,
		State:     "en_progreso",
		Version:   2,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockRepo := &mockEvalRepo{
		db:        db,
		sqlmock:   mock,
		row:       row,
		detailRow: row,
		comps: []*internal.EvaluationCompetency{
			{CompetencyID: compID, Rating: 5},
		},
		goals: []*internal.EvaluationGoal{},
		state: "en_progreso",
	}

	mock.ExpectBegin()
	mock.ExpectCommit()

	checker := &mockCycleChecker{phase: "cierre"}
	idem := &mockIdemCache{entries: make(map[string]*svc.IdempotencyCacheEntry)}
	service := svc.NewEvaluationService(mockRepo, nil, nil, checker, idem)

	req := dto.RHEvaluationRequest{
		Competencies: []dto.CompetencyRatingInput{
			{CompetencyID: compID, Rating: 5},
		},
		FinalComments: "Excellent performance",
	}

	resp, err := service.SubmitRHEvaluation(context.Background(), evalID, req, "idem-rh-1")
	require.NoError(t, err)
	assert.Equal(t, evalID, resp.ID)
}

func TestEvaluationService_FinalizeEvaluation_AllComplete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	evalID := uuid.New()
	cycleID := uuid.New()
	now := time.Now()
	selfComp := now.Add(-time.Hour)
	rhComp := now.Add(-30 * time.Minute)

	row := &repo.EvaluationRow{
		ID:                        evalID,
		CycleID:                   cycleID,
		State:                     "en_progreso",
		SelfEvaluationCompletedAt: &selfComp,
		RhEvaluationCompletedAt:   &rhComp,
		Version:                   2,
	}

	mockRepo := &mockEvalRepo{
		db:        db,
		sqlmock:   mock,
		row:       row,
		detailRow: row,
		state:     "en_progreso",
	}

	// Advisory lock tx + main tx (actual order: Begin lock, Exec lock, Begin main, Commit main, defer: Exec unlock, Rollback lock)
	mock.ExpectBegin()
	mock.ExpectExec("SELECT pg_advisory_lock").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectBegin()
	mock.ExpectCommit()
	mock.ExpectExec("SELECT pg_advisory_unlock").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectRollback()

	checker := &mockCycleChecker{phase: "cierre"}
	service := svc.NewEvaluationService(mockRepo, nil, nil, checker, nil)

	req := dto.FinalizeEvaluationRequest{Reason: "annual closing"}
	resp, err := service.FinalizeEvaluation(context.Background(), evalID, req)
	require.NoError(t, err)
	assert.Equal(t, "completada", resp.State)
}

func TestEvaluationService_FinalizeEvaluation_MissingSelfEval(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	evalID := uuid.New()
	cycleID := uuid.New()
	rhComp := time.Now()

	row := &repo.EvaluationRow{
		ID:                      evalID,
		CycleID:                 cycleID,
		State:                   "en_progreso",
		RhEvaluationCompletedAt: &rhComp,
		Version:                 2,
	}

	mockRepo := &mockEvalRepo{
		db:      db,
		sqlmock: mock,
		row:     row,
		state:   "en_progreso",
	}

	mock.ExpectBegin()
	mock.ExpectExec("SELECT pg_advisory_lock").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("SELECT pg_advisory_unlock").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectRollback()

	checker := &mockCycleChecker{phase: "cierre"}
	service := svc.NewEvaluationService(mockRepo, nil, nil, checker, nil)

	_, err = service.FinalizeEvaluation(context.Background(), evalID, dto.FinalizeEvaluationRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "self-evaluation has not been submitted")
	var de *pkgerrors.DomainError
	require.True(t, pkgerrors.AsDomainError(err, &de))
	assert.Equal(t, pkgerrors.InvalidTransition, de.Code)
}

func TestEvaluationService_ConcurrentSubmission(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	evalID := uuid.New()
	cycleID := uuid.New()
	compID := uuid.New()
	now := time.Now()

	row := &repo.EvaluationRow{
		ID:        evalID,
		CycleID:   cycleID,
		State:     "pendiente_evaluacion_final",
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockRepo := &mockEvalRepo{
		db:        db,
		sqlmock:   mock,
		row:       row,
		detailRow: row,
		comps: []*internal.EvaluationCompetency{
			{CompetencyID: compID, Rating: 4},
		},
		goals: []*internal.EvaluationGoal{},
		state: "pendiente_evaluacion_final",
	}

	checker := &mockCycleChecker{phase: "cierre"}
	idem := &mockIdemCache{entries: make(map[string]*svc.IdempotencyCacheEntry)}
	service := svc.NewEvaluationService(mockRepo, nil, nil, checker, idem)

	const goroutines = 100
	mock.MatchExpectationsInOrder(false)
	for i := 0; i < goroutines; i++ {
		mock.ExpectBegin()
		mock.ExpectCommit()
	}

	var wg sync.WaitGroup
	wg.Add(goroutines)
	var successCount int64

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			req := dto.SelfEvaluationRequest{
				Competencies: []dto.CompetencyRatingInput{
					{CompetencyID: compID, Rating: 4},
				},
			}
			_, err := service.SubmitSelfEvaluation(context.Background(), evalID, req, "")
			if err == nil {
				atomic.AddInt64(&successCount, 1)
			}
		}()
	}

	wg.Wait()
	assert.GreaterOrEqual(t, successCount, int64(1), "at least one submission should succeed")
}

// ---------- Tests: NineBoxService ----------

func TestNineBoxService_UpsertEntry_QuadrantComputed(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	matrixID := uuid.New()
	evaluateeID := uuid.New()
	entryID := uuid.New()

	mockNineBox := &mockNineBoxRepo{
		entry: &internal.NineBoxEntry{
			ID:               entryID,
			EvaluateeID:      evaluateeID,
			PerformanceScore: 5,
			PotentialScore:   5,
			Quadrant:         5,
		},
		version: 1,
	}
	mockCatalog := &mockCatalogRepo{
		quadrants: []*internal.NineBoxQuadrant{
			{Quadrant: 5, Label: "Star", Color: "#00FF00"},
		},
	}

	mock.ExpectBegin()
	mock.ExpectCommit()

	nineBoxSvc := svc.NewNineBoxService(mockNineBox, mockCatalog, &mockDB{db: db})

	req := dto.NineBoxEntryInput{
		EvaluateeID:      evaluateeID,
		PerformanceScore: 5,
		PotentialScore:   5,
		Comments:         "Solid performer",
	}

	resp, err := nineBoxSvc.UpsertEntry(context.Background(), matrixID, req)
	require.NoError(t, err)
	assert.Equal(t, entryID, resp.ID)
	assert.Equal(t, 5, resp.Quadrant)
	assert.Equal(t, "Star", resp.QuadrantLabel)
	assert.Equal(t, "#00FF00", resp.QuadrantColor)
}

func TestNineBoxService_BatchSubmit_Atomic(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	matrixID := uuid.New()
	evaluatee1 := uuid.New()
	evaluatee2 := uuid.New()

	mockNineBox := &mockNineBoxRepo{
		entries: []*internal.NineBoxEntry{
			{ID: uuid.New(), EvaluateeID: evaluatee1, PerformanceScore: 7, PotentialScore: 8, Quadrant: 9},
			{ID: uuid.New(), EvaluateeID: evaluatee2, PerformanceScore: 4, PotentialScore: 5, Quadrant: 5},
		},
	}
	mockCatalog := &mockCatalogRepo{}

	mock.ExpectBegin()
	mock.ExpectCommit()

	nineBoxSvc := svc.NewNineBoxService(mockNineBox, mockCatalog, &mockDB{db: db})

	req := dto.NineBoxBatchRequest{
		Entries: []dto.NineBoxEntryInput{
			{EvaluateeID: evaluatee1, PerformanceScore: 7, PotentialScore: 8},
			{EvaluateeID: evaluatee2, PerformanceScore: 4, PotentialScore: 5},
		},
	}

	resp, err := nineBoxSvc.BatchSubmitEntries(context.Background(), matrixID, req)
	require.NoError(t, err)
	assert.Len(t, resp, 2)
	assert.Equal(t, 9, resp[0].Quadrant)
	assert.Equal(t, 5, resp[1].Quadrant)
}

// ---------- Tests: DashboardService ----------

func TestDashboardService_GetSummary_CountsByState(t *testing.T) {
	cycleID := uuid.New()
	summary := map[string]int64{
		"pendiente_evaluacion_final": 5,
		"en_progreso":                3,
		"completada":                 12,
	}

	mockRepo := &mockEvalRepo{summary: summary}
	dashSvc := svc.NewDashboardService(mockRepo)

	resp, err := dashSvc.GetSummary(context.Background(), cycleID)
	require.NoError(t, err)
	assert.Equal(t, cycleID, resp.CycleID)
	assert.Equal(t, int64(5), resp.Counts["pendiente_evaluacion_final"])
	assert.Equal(t, int64(3), resp.Counts["en_progreso"])
	assert.Equal(t, int64(12), resp.Counts["completada"])
	assert.Equal(t, int64(0), resp.Counts["pendiente_asignacion"])
}
