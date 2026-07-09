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

func ptr[T any](v T) *T { return &v }

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

func (m *mockEvalRepo) GetCompetencyRatingsByEmployee(ctx context.Context, employeeID, cycleID, profileID uuid.UUID) ([]repo.EmployeeCompetencyRatingRow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// ponytail: return empty slice by default; tests that need specific data must set up via sqlmock
	return []repo.EmployeeCompetencyRatingRow{}, nil
}

func (m *mockEvalRepo) GetSummaryByCycle(ctx context.Context, cycleID uuid.UUID) (map[string]int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.summary == nil {
		return map[string]int64{}, nil
	}
	return m.summary, nil
}

func (m *mockEvalRepo) ListCompetencyResults(ctx context.Context, cycleID uuid.UUID, query string, managerID *uuid.UUID, offset, limit int) ([]*repo.CompetencyResultRow, error) {
	// ponytail: simplified mock — tests use hasMore directly
	return []*repo.CompetencyResultRow{}, nil
}

func (m *mockEvalRepo) CountCompetencyResults(ctx context.Context, cycleID uuid.UUID, query string, managerID *uuid.UUID) (int, error) {
	return 0, nil
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
	matrix        *internal.NineBoxMatrix
	matrices      []*internal.NineBoxMatrix
	entry         *internal.NineBoxEntry
	entries       []*internal.NineBoxEntry
	version       int
	employeeInfo  map[uuid.UUID]*repo.EmployeeInfo
	managerMap    map[uuid.UUID]uuid.UUID
	goalAssignees []uuid.UUID
	lockErr       error
	upsertErr     error
	updateErr     error
	batchErr      error
	fetchVerErr   error
	mu            sync.Mutex
	callCount     map[string]int
}

func (m *mockNineBoxRepo) recordCall(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.callCount == nil {
		m.callCount = make(map[string]int)
	}
	m.callCount[name]++
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

func (m *mockNineBoxRepo) ListMatrices(ctx context.Context, cycleID, evaluatorID, phaseID uuid.UUID) ([]*internal.NineBoxMatrix, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.matrices, nil
}

func (m *mockNineBoxRepo) GetMatrixEntries(ctx context.Context, matrixID uuid.UUID) ([]*internal.NineBoxEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.entries, nil
}

func (m *mockNineBoxRepo) GetMatrixEntriesByQuadrant(ctx context.Context, matrixID uuid.UUID, quadrant int) ([]*internal.NineBoxEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	filtered := make([]*internal.NineBoxEntry, 0, len(m.entries))
	for _, e := range m.entries {
		if e != nil && e.Quadrant == quadrant {
			filtered = append(filtered, e)
		}
	}
	return filtered, nil
}

func (m *mockNineBoxRepo) GetEmployeesByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*repo.EmployeeInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.employeeInfo == nil {
		return map[uuid.UUID]*repo.EmployeeInfo{}, nil
	}
	result := make(map[uuid.UUID]*repo.EmployeeInfo, len(ids))
	for _, id := range ids {
		if info, ok := m.employeeInfo[id]; ok {
			result[id] = info
		}
	}
	return result, nil
}

func (m *mockNineBoxRepo) GetManagerMapping(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	m.recordCall("GetManagerMapping")
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.managerMap == nil {
		return map[uuid.UUID]uuid.UUID{}, nil
	}
	result := make(map[uuid.UUID]uuid.UUID, len(ids))
	for _, id := range ids {
		if managerID, ok := m.managerMap[id]; ok {
			result[id] = managerID
		}
	}
	return result, nil
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

// --- New interface methods for phase-based matrix ---

func (m *mockNineBoxRepo) CreateMatrixWithPhase(ctx context.Context, cycleID, evaluatorID, phaseID uuid.UUID) (*internal.NineBoxMatrix, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.matrix, nil
}

func (m *mockNineBoxRepo) GetMatrixByPhase(ctx context.Context, cycleID, evaluatorID, phaseID uuid.UUID) (*internal.NineBoxMatrix, error) {
	m.recordCall("GetMatrixByPhase")
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.matrix == nil {
		return nil, repo.ErrMatrixNotFound
	}
	return m.matrix, nil
}

func (m *mockNineBoxRepo) UpsertEntryByTiers(ctx context.Context, tx *sql.Tx, matrixID uuid.UUID, evaluateeID uuid.UUID, perfTier, potTier, quadrant int, comments string) (*internal.NineBoxEntry, error) {
	m.recordCall("UpsertEntryByTiers")
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.upsertErr != nil {
		return nil, m.upsertErr
	}
	return m.entry, nil
}

func (m *mockNineBoxRepo) GetGoalAssigneesByCycle(ctx context.Context, cycleID uuid.UUID) ([]uuid.UUID, error) {
	m.recordCall("GetGoalAssigneesByCycle")
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.goalAssignees == nil {
		return []uuid.UUID{}, nil
	}
	return m.goalAssignees, nil
}

func (m *mockNineBoxRepo) GetGoalProgressByEmployee(ctx context.Context, employeeID, cycleID uuid.UUID) (float64, error) {
	return 50, nil
}

func (m *mockNineBoxRepo) GetCompetencyRatingsByEmployee(ctx context.Context, employeeID, cycleID uuid.UUID) (selfRating, hrRating *float64, err error) {
	return nil, nil, nil
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

func (m *mockCatalogRepo) InvalidateQuadrantCache() {}

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
	service := svc.NewEvaluationService(mockRepo, nil, nil, checker, idem, nil, nil)

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
	service := svc.NewEvaluationService(mockRepo, nil, nil, checker, nil, nil, nil)

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
	service := svc.NewEvaluationService(mockRepo, nil, nil, checker, idem, nil, nil)

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
	service := svc.NewEvaluationService(mockRepo, nil, nil, checker, nil, nil, nil)

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
	service := svc.NewEvaluationService(mockRepo, nil, nil, checker, nil, nil, nil)

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
	service := svc.NewEvaluationService(mockRepo, nil, nil, checker, idem, nil, nil)

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
			ID:              entryID,
			EvaluateeID:     evaluateeID,
			PerformanceTier: 2,
			PotentialTier:   2,
			Quadrant:        5,
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
		Comments:         ptr("Solid performer"),
	}

	resp, err := nineBoxSvc.UpsertEntry(context.Background(), matrixID, req)
	require.NoError(t, err)
	assert.Equal(t, entryID, resp.ID)
	assert.Equal(t, 2, resp.PerformanceTier)
	assert.Equal(t, 2, resp.PotentialTier)
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
			{ID: uuid.New(), EvaluateeID: evaluatee1, PerformanceTier: 3, PotentialTier: 3, Quadrant: 9},
			{ID: uuid.New(), EvaluateeID: evaluatee2, PerformanceTier: 2, PotentialTier: 2, Quadrant: 5},
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

// ---------- Tests: NineBoxService DTO enrichment ----------

func TestNineBoxService_GetMatrix_EnrichesEmployeeInfo(t *testing.T) {
	matrixID := uuid.New()
	evaluateeID := uuid.New()
	profileID := uuid.New()

	mockNineBox := &mockNineBoxRepo{
		matrix: &internal.NineBoxMatrix{
			ID:          matrixID,
			CycleID:     uuid.New(),
			EvaluatorID: uuid.New(),
			Edges: internal.NineBoxMatrixEdges{
				Entries: []*internal.NineBoxEntry{
					{ID: uuid.New(), EvaluateeID: evaluateeID, PerformanceTier: 2, PotentialTier: 2, Quadrant: 5},
				},
			},
		},
		version: 1,
		employeeInfo: map[uuid.UUID]*repo.EmployeeInfo{
			evaluateeID: {ID: evaluateeID, FirstName: "María", LastName: "García", ProfileID: profileID},
		},
	}
	mockCatalog := &mockCatalogRepo{
		quadrants: []*internal.NineBoxQuadrant{
			{Quadrant: 5, Label: "Star", Color: "#00FF00", ColorHex: "#22C55E"},
		},
	}

	nineBoxSvc := svc.NewNineBoxService(mockNineBox, mockCatalog, nil)

	resp, err := nineBoxSvc.GetMatrix(context.Background(), matrixID)
	require.NoError(t, err)
	require.Len(t, resp.Entries, 1)
	assert.Equal(t, "María García", resp.Entries[0].EmployeeName)
	assert.Equal(t, profileID, resp.Entries[0].ProfileID)
	assert.Equal(t, "Star", resp.Entries[0].QuadrantLabel)
	assert.Equal(t, "#22C55E", resp.Entries[0].QuadrantColor)
}

func TestNineBoxService_GetMatrix_OrpantEntry(t *testing.T) {
	matrixID := uuid.New()
	evaluateeID := uuid.New()

	mockNineBox := &mockNineBoxRepo{
		matrix: &internal.NineBoxMatrix{
			ID: matrixID,
			Edges: internal.NineBoxMatrixEdges{
				Entries: []*internal.NineBoxEntry{
					{ID: uuid.New(), EvaluateeID: evaluateeID, PerformanceTier: 2, PotentialTier: 2, Quadrant: 5},
				},
			},
		},
		version:      1,
		employeeInfo: map[uuid.UUID]*repo.EmployeeInfo{},
	}
	mockCatalog := &mockCatalogRepo{
		quadrants: []*internal.NineBoxQuadrant{{Quadrant: 5, Label: "Star", Color: "#00FF00"}},
	}

	nineBoxSvc := svc.NewNineBoxService(mockNineBox, mockCatalog, nil)

	resp, err := nineBoxSvc.GetMatrix(context.Background(), matrixID)
	require.NoError(t, err)
	require.Len(t, resp.Entries, 1)
	assert.Equal(t, "", resp.Entries[0].EmployeeName)
	assert.Equal(t, uuid.Nil, resp.Entries[0].ProfileID)
}

func TestNineBoxService_GetMatrixEntriesFiltered_ByQuadrant(t *testing.T) {
	matrixID := uuid.New()
	eval1 := uuid.New()
	eval2 := uuid.New()

	mockNineBox := &mockNineBoxRepo{
		entries: []*internal.NineBoxEntry{
			{ID: uuid.New(), EvaluateeID: eval1, PerformanceTier: 2, PotentialTier: 2, Quadrant: 5},
			{ID: uuid.New(), EvaluateeID: eval2, PerformanceTier: 3, PotentialTier: 3, Quadrant: 9},
		},
		version: 1,
	}
	mockCatalog := &mockCatalogRepo{}

	nineBoxSvc := svc.NewNineBoxService(mockNineBox, mockCatalog, nil)

	q := 5
	resp, err := nineBoxSvc.GetMatrixEntriesFiltered(context.Background(), matrixID, &q)
	require.NoError(t, err)
	require.Len(t, resp, 1)
	assert.Equal(t, 5, resp[0].Quadrant)
}

func TestNineBoxService_GetMatrixEntriesFiltered_NoFilter(t *testing.T) {
	matrixID := uuid.New()
	eval1 := uuid.New()
	eval2 := uuid.New()

	mockNineBox := &mockNineBoxRepo{
		entries: []*internal.NineBoxEntry{
			{ID: uuid.New(), EvaluateeID: eval1, PerformanceTier: 2, PotentialTier: 2, Quadrant: 5},
			{ID: uuid.New(), EvaluateeID: eval2, PerformanceTier: 3, PotentialTier: 3, Quadrant: 9},
		},
		version: 1,
	}
	mockCatalog := &mockCatalogRepo{}

	nineBoxSvc := svc.NewNineBoxService(mockNineBox, mockCatalog, nil)

	resp, err := nineBoxSvc.GetMatrixEntriesFiltered(context.Background(), matrixID, nil)
	require.NoError(t, err)
	require.Len(t, resp, 2)
}

func TestNineBoxService_RecomputeMatrix_GroupsByManager(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	cycleID := uuid.New()
	phaseID := uuid.New()
	managerID := uuid.New()
	employeeID := uuid.New()

	mockNineBox := &mockNineBoxRepo{
		matrix: &internal.NineBoxMatrix{
			ID:          uuid.New(),
			CycleID:     cycleID,
			EvaluatorID: managerID,
			PhaseID:     phaseID,
		},
		managerMap: map[uuid.UUID]uuid.UUID{
			employeeID: managerID,
		},
		goalAssignees: []uuid.UUID{employeeID},
		entry: &internal.NineBoxEntry{
			ID:          uuid.New(),
			EvaluateeID: employeeID,
			Quadrant:    5,
		},
	}
	mockCatalog := &mockCatalogRepo{}

	mock.ExpectBegin()
	mock.ExpectCommit()

	nineBoxSvc := svc.NewNineBoxService(mockNineBox, mockCatalog, &mockDB{db: db})

	err = nineBoxSvc.RecomputeMatrix(context.Background(), cycleID, phaseID)
	require.NoError(t, err)
	assert.Equal(t, 1, mockNineBox.callCount["GetManagerMapping"])
	assert.Equal(t, 1, mockNineBox.callCount["GetGoalAssigneesByCycle"])
	assert.Equal(t, 1, mockNineBox.callCount["GetMatrixByPhase"])
	assert.Equal(t, 1, mockNineBox.callCount["UpsertEntryByTiers"])
}

func TestNineBoxService_RecomputeMatrix_SkipsRootEmployee(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	cycleID := uuid.New()
	phaseID := uuid.New()
	rootID := uuid.New()

	mockNineBox := &mockNineBoxRepo{
		managerMap:    map[uuid.UUID]uuid.UUID{},
		goalAssignees: []uuid.UUID{rootID},
		entry:         &internal.NineBoxEntry{ID: uuid.New(), EvaluateeID: rootID, Quadrant: 5},
	}
	mockCatalog := &mockCatalogRepo{}

	mock.ExpectBegin()

	nineBoxSvc := svc.NewNineBoxService(mockNineBox, mockCatalog, &mockDB{db: db})

	err = nineBoxSvc.RecomputeMatrix(context.Background(), cycleID, phaseID)
	require.NoError(t, err)
	assert.Equal(t, 1, mockNineBox.callCount["GetManagerMapping"])
	assert.Equal(t, 1, mockNineBox.callCount["GetGoalAssigneesByCycle"])
	assert.Equal(t, 0, mockNineBox.callCount["GetMatrixByPhase"])
	assert.Equal(t, 0, mockNineBox.callCount["UpsertEntryByTiers"])
}

// ---------- Tests: GetCompetencyResults hasMore ----------

func TestGetCompetencyResults_HasMore_ExactPage(t *testing.T) {
	cycleID := uuid.New()
	evalID := uuid.New()
	now := time.Now()

	row := &repo.EvaluationRow{
		ID: evalID, CycleID: cycleID, State: "en_progreso", Version: 1,
		CreatedAt: now, UpdatedAt: now,
	}

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Total matches = limit → no hasMore
	mockEval := &mockHasMoreEvalRepo{
		row:        row,
		totalCount: 50,
		rows:       newFakeCompetencyRows(50),
	}
	checker := &mockCycleChecker{phase: "cierre"}
	service := svc.NewEvaluationService(mockEval, nil, nil, checker, nil, nil, nil)

	resp, err := service.GetCompetencyResults(context.Background(), cycleID, "", "all", uuid.Nil, 0, 50)
	require.NoError(t, err)
	assert.False(t, resp.Meta.HasMore, "hasMore should be false when offset+len == total")
	assert.Equal(t, 50, resp.Meta.Total)
	assert.Equal(t, 50, resp.Meta.Limit)
	assert.Equal(t, 0, resp.Meta.Offset)
}

func TestGetCompetencyResults_HasMore_OneExtraPage(t *testing.T) {
	cycleID := uuid.New()
	evalID := uuid.New()
	now := time.Now()

	row := &repo.EvaluationRow{
		ID: evalID, CycleID: cycleID, State: "en_progreso", Version: 1,
		CreatedAt: now, UpdatedAt: now,
	}

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Total = 51, limit = 50, offset = 0 → hasMore = true
	mockEval := &mockHasMoreEvalRepo{
		row:        row,
		totalCount: 51,
		rows:       newFakeCompetencyRows(50),
	}
	checker := &mockCycleChecker{phase: "cierre"}
	service := svc.NewEvaluationService(mockEval, nil, nil, checker, nil, nil, nil)

	resp, err := service.GetCompetencyResults(context.Background(), cycleID, "", "all", uuid.Nil, 0, 50)
	require.NoError(t, err)
	assert.True(t, resp.Meta.HasMore, "hasMore should be true when offset+len < total")
	assert.Equal(t, 51, resp.Meta.Total)
}

func TestGetCompetencyResults_HasMore_LastPage(t *testing.T) {
	cycleID := uuid.New()
	evalID := uuid.New()
	now := time.Now()

	row := &repo.EvaluationRow{
		ID: evalID, CycleID: cycleID, State: "en_progreso", Version: 1,
		CreatedAt: now, UpdatedAt: now,
	}

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Total = 51, offset = 50, limit = 50 → last page with 1 row → no hasMore
	mockEval := &mockHasMoreEvalRepo{
		row:        row,
		totalCount: 51,
		rows:       newFakeCompetencyRows(1),
	}
	checker := &mockCycleChecker{phase: "cierre"}
	service := svc.NewEvaluationService(mockEval, nil, nil, checker, nil, nil, nil)

	resp, err := service.GetCompetencyResults(context.Background(), cycleID, "", "all", uuid.Nil, 50, 50)
	require.NoError(t, err)
	assert.False(t, resp.Meta.HasMore, "hasMore should be false on last page")
	assert.Equal(t, 51, resp.Meta.Total)
	assert.Equal(t, 50, resp.Meta.Offset)
}

// newFakeCompetencyRows creates n non-nil CompetencyResultRow pointers for testing.
func newFakeCompetencyRows(n int) []*repo.CompetencyResultRow {
	rows := make([]*repo.CompetencyResultRow, n)
	for i := 0; i < n; i++ {
		rows[i] = &repo.CompetencyResultRow{ID: uuid.New(), Name: "Test Employee", ProfileName: "Test"}
	}
	return rows
}

// mockHasMoreEvalRepo implements EvaluationRepo for hasMore testing.
type mockHasMoreEvalRepo struct {
	row        *repo.EvaluationRow
	totalCount int
	rows       []*repo.CompetencyResultRow
}

func (m *mockHasMoreEvalRepo) GetByID(ctx context.Context, id uuid.UUID) (*repo.EvaluationRow, error) {
	return m.row, nil
}
func (m *mockHasMoreEvalRepo) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return nil, nil
}
func (m *mockHasMoreEvalRepo) LockEvalForUpdate(ctx context.Context, tx *sql.Tx, evalID uuid.UUID) (*repo.EvaluationRow, error) {
	return m.row, nil
}
func (m *mockHasMoreEvalRepo) SubmitEval(ctx context.Context, tx *sql.Tx, evalID uuid.UUID, comps []repo.CompetencyUpsert, goals []repo.GoalCommentUpsert, newState string, setSelfCompleted, setRHCompleted bool) error {
	return nil
}
func (m *mockHasMoreEvalRepo) GetDetail(ctx context.Context, id uuid.UUID) (*repo.EvaluationRow, []*internal.EvaluationCompetency, []*internal.EvaluationGoal, error) {
	return m.row, nil, nil, nil
}
func (m *mockHasMoreEvalRepo) ListByCycle(ctx context.Context, cycleID uuid.UUID, state string, cursor string, limit int) ([]*repo.EvaluationRow, string, error) {
	return nil, "", nil
}
func (m *mockHasMoreEvalRepo) GetCompetencyRatingsByEmployee(ctx context.Context, employeeID, cycleID, profileID uuid.UUID) ([]repo.EmployeeCompetencyRatingRow, error) {
	return nil, nil
}
func (m *mockHasMoreEvalRepo) FinalizeEval(ctx context.Context, tx *sql.Tx, evalID uuid.UUID) error {
	return nil
}
func (m *mockHasMoreEvalRepo) RefreshSummaryView(ctx context.Context) error {
	return nil
}
func (m *mockHasMoreEvalRepo) GetSummaryByCycle(ctx context.Context, cycleID uuid.UUID) (map[string]int64, error) {
	return nil, nil
}
func (m *mockHasMoreEvalRepo) ListCompetencyResults(ctx context.Context, cycleID uuid.UUID, query string, managerID *uuid.UUID, offset, limit int) ([]*repo.CompetencyResultRow, error) {
	return m.rows, nil
}
func (m *mockHasMoreEvalRepo) CountCompetencyResults(ctx context.Context, cycleID uuid.UUID, query string, managerID *uuid.UUID) (int, error) {
	return m.totalCount, nil
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
