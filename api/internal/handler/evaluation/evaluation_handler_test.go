package evaluation_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	dto "github.com/sed-evaluacion-desempeno/api/internal/dto/evaluation"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	repo "github.com/sed-evaluacion-desempeno/api/internal/repository/evaluation"
	handler "github.com/sed-evaluacion-desempeno/api/internal/handler/evaluation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------- Mock Services ----------

type mockEvalService struct {
	listResp       *dto.EvaluationListResponse
	listErr        error
	getResp        *dto.EvaluationDetailResponse
	getErr         error
	submitSelfResp *dto.EvaluationDetailResponse
	submitSelfErr  error
	updateSelfResp *dto.EvaluationDetailResponse
	updateSelfErr  error
	submitRHResp   *dto.EvaluationDetailResponse
	submitRHErr    error
	updateRHResp   *dto.EvaluationDetailResponse
	updateRHErr    error
	finalizeResp   *dto.EvaluationDetailResponse
	finalizeErr    error
	mu             sync.Mutex
	callCount      map[string]int
	delay          time.Duration
}

func (m *mockEvalService) recordCall(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.callCount == nil {
		m.callCount = make(map[string]int)
	}
	m.callCount[name]++
}

func (m *mockEvalService) ListEvaluations(ctx context.Context, cycleID uuid.UUID, stateFilter string, cursor string, limit int) (*dto.EvaluationListResponse, error) {
	m.recordCall("ListEvaluations")
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	return m.listResp, m.listErr
}

func (m *mockEvalService) GetEvaluation(ctx context.Context, id uuid.UUID) (*dto.EvaluationDetailResponse, error) {
	m.recordCall("GetEvaluation")
	return m.getResp, m.getErr
}

func (m *mockEvalService) SubmitSelfEvaluation(ctx context.Context, evaluationID uuid.UUID, req dto.SelfEvaluationRequest, idempotencyKey string) (*dto.EvaluationDetailResponse, error) {
	m.recordCall("SubmitSelfEvaluation")
	return m.submitSelfResp, m.submitSelfErr
}

func (m *mockEvalService) UpdateSelfEvaluation(ctx context.Context, evaluationID uuid.UUID, req dto.SelfEvaluationRequest, ifMatch int) (*dto.EvaluationDetailResponse, error) {
	m.recordCall("UpdateSelfEvaluation")
	return m.updateSelfResp, m.updateSelfErr
}

func (m *mockEvalService) SubmitRHEvaluation(ctx context.Context, evaluationID uuid.UUID, req dto.RHEvaluationRequest, idempotencyKey string) (*dto.EvaluationDetailResponse, error) {
	m.recordCall("SubmitRHEvaluation")
	return m.submitRHResp, m.submitRHErr
}

func (m *mockEvalService) UpdateRHEvaluation(ctx context.Context, evaluationID uuid.UUID, req dto.RHEvaluationRequest, ifMatch int) (*dto.EvaluationDetailResponse, error) {
	m.recordCall("UpdateRHEvaluation")
	return m.updateRHResp, m.updateRHErr
}

func (m *mockEvalService) FinalizeEvaluation(ctx context.Context, evaluationID uuid.UUID, req dto.FinalizeEvaluationRequest) (*dto.EvaluationDetailResponse, error) {
	m.recordCall("FinalizeEvaluation")
	return m.finalizeResp, m.finalizeErr
}

type mockBoxService struct {
	listResp      []dto.NineBoxMatrixResponse
	listErr       error
	createResp    *dto.NineBoxMatrixResponse
	createErr     error
	getResp       *dto.NineBoxMatrixResponse
	getErr        error
	upsertResp    *dto.NineBoxEntryDTO
	upsertErr     error
	updateResp    *dto.NineBoxEntryDTO
	updateErr     error
	batchResp     []dto.NineBoxEntryDTO
	batchErr      error
	scalesResp    []dto.NineBoxScaleDTO
	scalesErr     error
	quadrantsResp []dto.NineBoxQuadrantDTO
	quadrantsErr  error
	mu            sync.Mutex
	callCount     map[string]int
}

func (m *mockBoxService) recordCall(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.callCount == nil {
		m.callCount = make(map[string]int)
	}
	m.callCount[name]++
}

func (m *mockBoxService) ListMatrices(ctx context.Context, cycleID, evaluatorID uuid.UUID) ([]dto.NineBoxMatrixResponse, error) {
	m.recordCall("ListMatrices")
	return m.listResp, m.listErr
}

func (m *mockBoxService) CreateMatrix(ctx context.Context, cycleID, evaluatorID uuid.UUID) (*dto.NineBoxMatrixResponse, error) {
	m.recordCall("CreateMatrix")
	return m.createResp, m.createErr
}

func (m *mockBoxService) GetMatrix(ctx context.Context, matrixID uuid.UUID) (*dto.NineBoxMatrixResponse, error) {
	m.recordCall("GetMatrix")
	return m.getResp, m.getErr
}

func (m *mockBoxService) UpsertEntry(ctx context.Context, matrixID uuid.UUID, req dto.NineBoxEntryInput) (*dto.NineBoxEntryDTO, error) {
	m.recordCall("UpsertEntry")
	return m.upsertResp, m.upsertErr
}

func (m *mockBoxService) UpdateEntry(ctx context.Context, entryID uuid.UUID, req dto.NineBoxEntryInput, ifMatch int) (*dto.NineBoxEntryDTO, error) {
	m.recordCall("UpdateEntry")
	return m.updateResp, m.updateErr
}

func (m *mockBoxService) BatchSubmitEntries(ctx context.Context, matrixID uuid.UUID, req dto.NineBoxBatchRequest) ([]dto.NineBoxEntryDTO, error) {
	m.recordCall("BatchSubmitEntries")
	return m.batchResp, m.batchErr
}

func (m *mockBoxService) GetScales(ctx context.Context) ([]dto.NineBoxScaleDTO, error) {
	m.recordCall("GetScales")
	return m.scalesResp, m.scalesErr
}

func (m *mockBoxService) GetQuadrants(ctx context.Context) ([]dto.NineBoxQuadrantDTO, error) {
	m.recordCall("GetQuadrants")
	return m.quadrantsResp, m.quadrantsErr
}

type mockDashService struct {
	summaryResp *dto.EvaluationSummaryResponse
	summaryErr  error
	mu          sync.Mutex
	callCount   map[string]int
	delay       time.Duration
}

func (m *mockDashService) recordCall(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.callCount == nil {
		m.callCount = make(map[string]int)
	}
	m.callCount[name]++
}

func (m *mockDashService) GetSummary(ctx context.Context, cycleID uuid.UUID) (*dto.EvaluationSummaryResponse, error) {
	m.recordCall("GetSummary")
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	return m.summaryResp, m.summaryErr
}

// ---------- Helpers ----------

func setupHandler(t *testing.T, evalSvc handler.EvalService, boxSvc handler.BoxService, dashSvc handler.DashService) (*handler.EvaluationHandler, chi.Router) {
	h := handler.NewEvaluationHandler(evalSvc, boxSvc, dashSvc)
	r := chi.NewRouter()
	return h, r
}

func doRequest(t *testing.T, r chi.Router, method, path string, body []byte, query string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	if query != "" {
		req.URL.RawQuery = query
	}
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

// ---------- Happy Path Tests ----------

func TestListEvaluations_Success(t *testing.T) {
	cycleID := uuid.New()
	mockEval := &mockEvalService{
		listResp: &dto.EvaluationListResponse{
			Data: []dto.EvaluationListItem{
				{ID: uuid.New(), EmployeeID: uuid.New(), CycleID: cycleID, State: "pendiente_evaluacion_final"},
			},
			NextCursor: "",
		},
	}
	h, r := setupHandler(t, mockEval, nil, nil)
	r.Get("/evaluations", h.ListEvaluations)

	rec := doRequest(t, r, http.MethodGet, "/evaluations", nil, "cycle_id="+cycleID.String())
	assert.Equal(t, http.StatusOK, rec.Code)
	var resp dto.EvaluationListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 1)
}

func TestGetEvaluation_Success(t *testing.T) {
	evalID := uuid.New()
	mockEval := &mockEvalService{
		getResp: &dto.EvaluationDetailResponse{
			ID:    evalID,
			State: "en_progreso",
			CompetencyRatings: []dto.CompetencyRatingDTO{
				{CompetencyID: uuid.New(), Rating: 4},
			},
		},
	}
	h, r := setupHandler(t, mockEval, nil, nil)
	r.Get("/evaluations/{id}", h.GetEvaluation)

	rec := doRequest(t, r, http.MethodGet, "/evaluations/"+evalID.String(), nil, "")
	assert.Equal(t, http.StatusOK, rec.Code)
	var resp dto.EvaluationDetailResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, evalID, resp.ID)
}

func TestSubmitSelfEvaluation_Success(t *testing.T) {
	evalID := uuid.New()
	compID := uuid.New()
	mockEval := &mockEvalService{
		submitSelfResp: &dto.EvaluationDetailResponse{
			ID:    evalID,
			State: "en_progreso",
			CompetencyRatings: []dto.CompetencyRatingDTO{
				{CompetencyID: compID, Rating: 4},
			},
		},
	}
	h, r := setupHandler(t, mockEval, nil, nil)
	r.Post("/evaluations/{id}/self-evaluation", h.SubmitSelfEvaluation)

	reqBody, _ := json.Marshal(dto.SelfEvaluationRequest{
		Competencies: []dto.CompetencyRatingInput{
			{CompetencyID: compID, Rating: 4, Comments: "Good"},
		},
	})
	rec := doRequest(t, r, http.MethodPost, "/evaluations/"+evalID.String()+"/self-evaluation", reqBody, "")
	assert.Equal(t, http.StatusOK, rec.Code)
	var resp dto.EvaluationDetailResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, evalID, resp.ID)
}

func TestSubmitRHEvaluation_Success(t *testing.T) {
	evalID := uuid.New()
	compID := uuid.New()
	mockEval := &mockEvalService{
		submitRHResp: &dto.EvaluationDetailResponse{
			ID:    evalID,
			State: "en_progreso",
			CompetencyRatings: []dto.CompetencyRatingDTO{
				{CompetencyID: compID, Rating: 5},
			},
		},
	}
	h, r := setupHandler(t, mockEval, nil, nil)
	r.Post("/evaluations/{id}/rh-evaluation", h.SubmitRHEvaluation)

	reqBody, _ := json.Marshal(dto.RHEvaluationRequest{
		Competencies: []dto.CompetencyRatingInput{
			{CompetencyID: compID, Rating: 5},
		},
		FinalComments: "Strong performer",
	})
	rec := doRequest(t, r, http.MethodPost, "/evaluations/"+evalID.String()+"/rh-evaluation", reqBody, "")
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestFinalizeEvaluation_Success(t *testing.T) {
	evalID := uuid.New()
	mockEval := &mockEvalService{
		finalizeResp: &dto.EvaluationDetailResponse{
			ID:    evalID,
			State: "completada",
		},
	}
	h, r := setupHandler(t, mockEval, nil, nil)
	r.Post("/evaluations/{id}/finalize", h.FinalizeEvaluation)

	rec := doRequest(t, r, http.MethodPost, "/evaluations/"+evalID.String()+"/finalize", []byte(`{}`), "")
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCreateNineBoxMatrix_Success(t *testing.T) {
	cycleID := uuid.New()
	evaluatorID := uuid.New()
	matrixID := uuid.New()
	mockBox := &mockBoxService{
		createResp: &dto.NineBoxMatrixResponse{
			ID:          matrixID,
			CycleID:     cycleID,
			EvaluatorID: evaluatorID,
			Entries:     []dto.NineBoxEntryDTO{},
		},
	}
	h, r := setupHandler(t, nil, mockBox, nil)
	r.Post("/nine-box/matrices", h.CreateMatrix)

	reqBody, _ := json.Marshal(map[string]interface{}{
		"cycleId":     cycleID,
		"evaluatorId": evaluatorID,
	})
	rec := doRequest(t, r, http.MethodPost, "/nine-box/matrices", reqBody, "")
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestGetNineBoxMatrix_Success(t *testing.T) {
	matrixID := uuid.New()
	mockBox := &mockBoxService{
		getResp: &dto.NineBoxMatrixResponse{
			ID:      matrixID,
			Entries: []dto.NineBoxEntryDTO{},
		},
	}
	h, r := setupHandler(t, nil, mockBox, nil)
	r.Get("/nine-box/matrices/{matrixId}", h.GetMatrix)

	rec := doRequest(t, r, http.MethodGet, "/nine-box/matrices/"+matrixID.String(), nil, "")
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUpsertNineBoxEntry_Success(t *testing.T) {
	matrixID := uuid.New()
	evaluateeID := uuid.New()
	entryID := uuid.New()
	mockBox := &mockBoxService{
		upsertResp: &dto.NineBoxEntryDTO{
			ID:               entryID,
			EvaluateeID:      evaluateeID,
			PerformanceScore: 5,
			PotentialScore:   5,
			Quadrant:         5,
		},
	}
	h, r := setupHandler(t, nil, mockBox, nil)
	r.Post("/nine-box/matrices/{matrixId}/entries", h.UpsertMatrixEntry)

	reqBody, _ := json.Marshal(dto.NineBoxEntryInput{
		EvaluateeID:      evaluateeID,
		PerformanceScore: 5,
		PotentialScore:   5,
	})
	rec := doRequest(t, r, http.MethodPost, "/nine-box/matrices/"+matrixID.String()+"/entries", reqBody, "")
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestBatchNineBoxEntries_Success(t *testing.T) {
	matrixID := uuid.New()
	evaluateeID := uuid.New()
	mockBox := &mockBoxService{
		batchResp: []dto.NineBoxEntryDTO{
			{ID: uuid.New(), EvaluateeID: evaluateeID, PerformanceScore: 7, PotentialScore: 8, Quadrant: 9},
		},
	}
	h, r := setupHandler(t, nil, mockBox, nil)
	r.Post("/nine-box/batch", h.BatchSubmitEntries)

	reqBody, _ := json.Marshal(dto.NineBoxBatchRequest{
		Entries: []dto.NineBoxEntryInput{
			{EvaluateeID: evaluateeID, PerformanceScore: 7, PotentialScore: 8},
		},
	})
	rec := doRequest(t, r, http.MethodPost, "/nine-box/batch?matrixId="+matrixID.String(), reqBody, "")
	assert.Equal(t, http.StatusOK, rec.Code)
	var resp []dto.NineBoxEntryDTO
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp, 1)
}

func TestGetNineBoxScales_Success(t *testing.T) {
	mockBox := &mockBoxService{
		scalesResp: []dto.NineBoxScaleDTO{
			{Axis: "performance", Level: 1, Label: "Low", Description: "Low performance"},
		},
	}
	h, r := setupHandler(t, nil, mockBox, nil)
	r.Get("/nine-box/scales", h.GetScales)

	rec := doRequest(t, r, http.MethodGet, "/nine-box/scales", nil, "")
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetNineBoxQuadrants_Success(t *testing.T) {
	mockBox := &mockBoxService{
		quadrantsResp: []dto.NineBoxQuadrantDTO{
			{Quadrant: 1, Label: "Low/Low", Description: "Low performance and potential", Color: "#FF0000"},
		},
	}
	h, r := setupHandler(t, nil, mockBox, nil)
	r.Get("/nine-box/quadrants", h.GetQuadrants)

	rec := doRequest(t, r, http.MethodGet, "/nine-box/quadrants", nil, "")
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetEvaluationSummary_Success(t *testing.T) {
	cycleID := uuid.New()
	mockDash := &mockDashService{
		summaryResp: &dto.EvaluationSummaryResponse{
			CycleID: cycleID,
			Counts: map[string]int64{
				"pendiente_evaluacion_final": 2,
				"en_progreso":                1,
				"completada":                 5,
			},
		},
	}
	h, r := setupHandler(t, nil, nil, mockDash)
	r.Get("/evaluations/summary", h.GetEvaluationSummary)

	rec := doRequest(t, r, http.MethodGet, "/evaluations/summary", nil, "cycle_id="+cycleID.String())
	assert.Equal(t, http.StatusOK, rec.Code)
	var resp dto.EvaluationSummaryResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, int64(5), resp.Counts["completada"])
}

// ---------- Error Tests ----------

func TestSubmitSelfEvaluation_WrongPhase(t *testing.T) {
	evalID := uuid.New()
	mockEval := &mockEvalService{
		submitSelfErr: pkgerrors.NewDomainError(pkgerrors.PhaseNotAdvanceable,
			"this operation requires the cycle to be in 'cierre' phase; current phase is 'avance'", nil),
	}
	h, r := setupHandler(t, mockEval, nil, nil)
	r.Post("/evaluations/{id}/self-evaluation", h.SubmitSelfEvaluation)

	reqBody, _ := json.Marshal(dto.SelfEvaluationRequest{
		Competencies: []dto.CompetencyRatingInput{{CompetencyID: uuid.New(), Rating: 4}},
	})
	rec := doRequest(t, r, http.MethodPost, "/evaluations/"+evalID.String()+"/self-evaluation", reqBody, "")
	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestFinalizeEvaluation_NotAllComplete(t *testing.T) {
	evalID := uuid.New()
	mockEval := &mockEvalService{
		finalizeErr: pkgerrors.NewDomainError(pkgerrors.InvalidTransition,
			"cannot finalize evaluation: self-evaluation has not been submitted", nil),
	}
	h, r := setupHandler(t, mockEval, nil, nil)
	r.Post("/evaluations/{id}/finalize", h.FinalizeEvaluation)

	rec := doRequest(t, r, http.MethodPost, "/evaluations/"+evalID.String()+"/finalize", []byte(`{}`), "")
	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestUpsertNineBoxEntry_InvalidScore(t *testing.T) {
	matrixID := uuid.New()
	evaluateeID := uuid.New()
	mockBox := &mockBoxService{
		upsertErr: repo.ErrQuadrantOutOfRange,
	}
	h, r := setupHandler(t, nil, mockBox, nil)
	r.Post("/nine-box/matrices/{matrixId}/entries", h.UpsertMatrixEntry)

	reqBody, _ := json.Marshal(dto.NineBoxEntryInput{
		EvaluateeID:      evaluateeID,
		PerformanceScore: 10,
		PotentialScore:   5,
	})
	rec := doRequest(t, r, http.MethodPost, "/nine-box/matrices/"+matrixID.String()+"/entries", reqBody, "")
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetEvaluation_NotFound(t *testing.T) {
	evalID := uuid.New()
	mockEval := &mockEvalService{
		getErr: repo.ErrEvaluationNotFound,
	}
	h, r := setupHandler(t, mockEval, nil, nil)
	r.Get("/evaluations/{id}", h.GetEvaluation)

	rec := doRequest(t, r, http.MethodGet, "/evaluations/"+evalID.String(), nil, "")
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestSubmitSelfEvaluation_AlreadyFinalized(t *testing.T) {
	evalID := uuid.New()
	mockEval := &mockEvalService{
		submitSelfErr: repo.ErrEvaluationFinalized,
	}
	h, r := setupHandler(t, mockEval, nil, nil)
	r.Post("/evaluations/{id}/self-evaluation", h.SubmitSelfEvaluation)

	reqBody, _ := json.Marshal(dto.SelfEvaluationRequest{
		Competencies: []dto.CompetencyRatingInput{{CompetencyID: uuid.New(), Rating: 4}},
	})
	rec := doRequest(t, r, http.MethodPost, "/evaluations/"+evalID.String()+"/self-evaluation", reqBody, "")
	assert.Equal(t, http.StatusConflict, rec.Code)
}

// ---------- Response Time Tests ----------

func TestListEvaluations_ResponseTime(t *testing.T) {
	cycleID := uuid.New()
	mockEval := &mockEvalService{
		listResp: &dto.EvaluationListResponse{
			Data: []dto.EvaluationListItem{
				{ID: uuid.New(), EmployeeID: uuid.New(), CycleID: cycleID, State: "pendiente_evaluacion_final"},
			},
		},
		delay: 10 * time.Millisecond,
	}
	h, r := setupHandler(t, mockEval, nil, nil)
	r.Get("/evaluations", h.ListEvaluations)

	start := time.Now()
	rec := doRequest(t, r, http.MethodGet, "/evaluations", nil, "cycle_id="+cycleID.String())
	elapsed := time.Since(start)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Less(t, elapsed, 200*time.Millisecond, "response time should be under 200ms")
}

func TestGetEvaluationSummary_ResponseTime(t *testing.T) {
	cycleID := uuid.New()
	mockDash := &mockDashService{
		summaryResp: &dto.EvaluationSummaryResponse{
			CycleID: cycleID,
			Counts:  map[string]int64{"completada": 10},
		},
		delay: 5 * time.Millisecond,
	}
	h, r := setupHandler(t, nil, nil, mockDash)
	r.Get("/evaluations/summary", h.GetEvaluationSummary)

	start := time.Now()
	rec := doRequest(t, r, http.MethodGet, "/evaluations/summary", nil, "cycle_id="+cycleID.String())
	elapsed := time.Since(start)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Less(t, elapsed, 150*time.Millisecond, "response time should be under 150ms")
}

// ---------- Concurrency Tests ----------

func TestSubmitSelfEvaluation_Concurrent(t *testing.T) {
	evalID := uuid.New()
	compID := uuid.New()
	mockEval := &mockEvalService{
		submitSelfResp: &dto.EvaluationDetailResponse{
			ID:    evalID,
			State: "en_progreso",
		},
	}
	h, r := setupHandler(t, mockEval, nil, nil)
	r.Post("/evaluations/{id}/self-evaluation", h.SubmitSelfEvaluation)

	reqBody, _ := json.Marshal(dto.SelfEvaluationRequest{
		Competencies: []dto.CompetencyRatingInput{
			{CompetencyID: compID, Rating: 4},
		},
	})

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			rec := doRequest(t, r, http.MethodPost, "/evaluations/"+evalID.String()+"/self-evaluation", reqBody, "")
			assert.Equal(t, http.StatusOK, rec.Code)
		}()
	}

	wg.Wait()
	assert.GreaterOrEqual(t, mockEval.callCount["SubmitSelfEvaluation"], 100, "all 100 goroutines should call the service")
}

func TestUpsertNineBoxEntry_Concurrent(t *testing.T) {
	matrixID := uuid.New()
	evaluateeID := uuid.New()
	entryID := uuid.New()
	mockBox := &mockBoxService{
		upsertResp: &dto.NineBoxEntryDTO{
			ID:          entryID,
			EvaluateeID: evaluateeID,
			Quadrant:    5,
		},
	}
	h, r := setupHandler(t, nil, mockBox, nil)
	r.Post("/nine-box/matrices/{matrixId}/entries", h.UpsertMatrixEntry)

	reqBody, _ := json.Marshal(dto.NineBoxEntryInput{
		EvaluateeID:      evaluateeID,
		PerformanceScore: 5,
		PotentialScore:   5,
	})

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			rec := doRequest(t, r, http.MethodPost, "/nine-box/matrices/"+matrixID.String()+"/entries", reqBody, "")
			assert.Equal(t, http.StatusOK, rec.Code)
		}()
	}

	wg.Wait()
	assert.GreaterOrEqual(t, mockBox.callCount["UpsertEntry"], 50, "all 50 goroutines should call the service")
}

func TestBatchNineBoxEntries_Concurrent(t *testing.T) {
	matrixID := uuid.New()
	evaluateeID := uuid.New()
	mockBox := &mockBoxService{
		batchResp: []dto.NineBoxEntryDTO{
			{ID: uuid.New(), EvaluateeID: evaluateeID, Quadrant: 5},
		},
	}
	h, r := setupHandler(t, nil, mockBox, nil)
	r.Post("/nine-box/batch", h.BatchSubmitEntries)

	reqBody, _ := json.Marshal(dto.NineBoxBatchRequest{
		Entries: []dto.NineBoxEntryInput{
			{EvaluateeID: evaluateeID, PerformanceScore: 5, PotentialScore: 5},
		},
	})

	const goroutines = 30
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			rec := doRequest(t, r, http.MethodPost, "/nine-box/batch?matrixId="+matrixID.String(), reqBody, "")
			assert.Equal(t, http.StatusOK, rec.Code)
		}()
	}

	wg.Wait()
	assert.GreaterOrEqual(t, mockBox.callCount["BatchSubmitEntries"], 30, "all 30 goroutines should call the service")
}
