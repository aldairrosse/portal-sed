package evaluation

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	dto "github.com/sed-evaluacion-desempeno/api/internal/dto/evaluation"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/quadrant"
	repo "github.com/sed-evaluacion-desempeno/api/internal/repository/evaluation"
)

// NineBoxService handles 9×9 matrix operations and quadrant computation.
type NineBoxService struct {
	nineBoxRepo NineBoxRepo
	catalogRepo CatalogRepo
	db          DB
}

// NewNineBoxService creates a new NineBoxService.
func NewNineBoxService(nineBoxRepo NineBoxRepo, catalogRepo CatalogRepo, db DB) *NineBoxService {
	return &NineBoxService{
		nineBoxRepo: nineBoxRepo,
		catalogRepo: catalogRepo,
		db:          db,
	}
}

// CreateMatrix creates a new 9×9 matrix for an evaluator in a cycle.
func (s *NineBoxService) CreateMatrix(ctx context.Context, cycleID, evaluatorID uuid.UUID) (*dto.NineBoxMatrixResponse, error) {
	m, err := s.nineBoxRepo.CreateMatrix(ctx, cycleID, evaluatorID)
	if err != nil {
		return nil, err
	}
	return &dto.NineBoxMatrixResponse{
		ID: m.ID, CycleID: m.CycleID, EvaluatorID: m.EvaluatorID,
		Entries: []dto.NineBoxEntryDTO{}, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}, nil
}

// GetMatrix retrieves a matrix with entries.
func (s *NineBoxService) GetMatrix(ctx context.Context, matrixID uuid.UUID) (*dto.NineBoxMatrixResponse, error) {
	m, err := s.nineBoxRepo.GetMatrixByID(ctx, matrixID)
	if err != nil {
		return nil, err
	}
	resp := &dto.NineBoxMatrixResponse{
		ID: m.ID, CycleID: m.CycleID, EvaluatorID: m.EvaluatorID,
		CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
	entries := m.Edges.Entries
	if entries == nil {
		entries = []*internal.NineBoxEntry{}
	}
	resp.Entries = make([]dto.NineBoxEntryDTO, len(entries))
	for i, e := range entries {
		resp.Entries[i] = s.toEntryDTO(ctx, e)
	}
	return resp, nil
}

// ListMatrices returns matrices filtered by cycle and/or evaluator.
func (s *NineBoxService) ListMatrices(ctx context.Context, cycleID, evaluatorID uuid.UUID) ([]dto.NineBoxMatrixResponse, error) {
	matrices, err := s.nineBoxRepo.ListMatrices(ctx, cycleID, evaluatorID)
	if err != nil {
		return nil, err
	}
	resp := make([]dto.NineBoxMatrixResponse, len(matrices))
	for i, m := range matrices {
		resp[i] = dto.NineBoxMatrixResponse{
			ID: m.ID, CycleID: m.CycleID, EvaluatorID: m.EvaluatorID,
			Entries: []dto.NineBoxEntryDTO{}, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
		}
	}
	return resp, nil
}

// UpsertEntry creates or updates a single matrix entry with quadrant computation.
func (s *NineBoxService) UpsertEntry(ctx context.Context, matrixID uuid.UUID, req dto.NineBoxEntryInput) (*dto.NineBoxEntryDTO, error) {
	if req.PerformanceScore < 1 || req.PerformanceScore > 9 || req.PotentialScore < 1 || req.PotentialScore > 9 {
		return nil, repo.ErrQuadrantOutOfRange
	}

	q := quadrant.ComputeQuadrant(req.PerformanceScore, req.PotentialScore)
	if q == 0 {
		return nil, repo.ErrQuadrantOutOfRange
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	if err := s.nineBoxRepo.LockEntryForSelect(ctx, tx, matrixID, req.EvaluateeID); err != nil {
		return nil, err
	}

	entry, err := s.nineBoxRepo.UpsertEntry(ctx, tx, matrixID, req.EvaluateeID,
		req.PerformanceScore, req.PotentialScore, q, req.Comments)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	tx = nil

	dto := s.toEntryDTO(ctx, entry)
	return &dto, nil
}

// UpdateEntry updates an existing entry with optimistic lock.
func (s *NineBoxService) UpdateEntry(ctx context.Context, entryID uuid.UUID, req dto.NineBoxEntryInput, ifMatch int) (*dto.NineBoxEntryDTO, error) {
	if req.PerformanceScore < 1 || req.PerformanceScore > 9 || req.PotentialScore < 1 || req.PotentialScore > 9 {
		return nil, repo.ErrQuadrantOutOfRange
	}
	q := quadrant.ComputeQuadrant(req.PerformanceScore, req.PotentialScore)
	if q == 0 {
		return nil, repo.ErrQuadrantOutOfRange
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	entry, err := s.nineBoxRepo.UpdateEntry(ctx, tx, entryID,
		req.PerformanceScore, req.PotentialScore, q, req.Comments, ifMatch)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	tx = nil

	dto := s.toEntryDTO(ctx, entry)
	return &dto, nil
}

// BatchSubmitEntries atomically submits multiple entries in a single transaction.
func (s *NineBoxService) BatchSubmitEntries(ctx context.Context, matrixID uuid.UUID, req dto.NineBoxBatchRequest) ([]dto.NineBoxEntryDTO, error) {
	if len(req.Entries) > 20 {
		return nil, errors.NewDomainError(errors.BatchSizeExceeded,
			"Batch size exceeds the maximum allowed (20).", nil)
	}

	seen := make(map[uuid.UUID]bool)
	for _, e := range req.Entries {
		if seen[e.EvaluateeID] {
			return nil, errors.NewDomainError(errors.InvalidRequest,
				"Duplicate evaluateeId in batch request.", nil,
			).WithDetails("evaluatee_id: " + e.EvaluateeID.String())
		}
		seen[e.EvaluateeID] = true
		if e.PerformanceScore < 1 || e.PerformanceScore > 9 || e.PotentialScore < 1 || e.PotentialScore > 9 {
			return nil, repo.ErrQuadrantOutOfRange
		}
	}

	items := make([]repo.EntryUpsert, len(req.Entries))
	for i, e := range req.Entries {
		q := quadrant.ComputeQuadrant(e.PerformanceScore, e.PotentialScore)
		items[i] = repo.EntryUpsert{
			EvaluateeID: e.EvaluateeID, PerformanceScore: e.PerformanceScore,
			PotentialScore: e.PotentialScore, Quadrant: q, Comments: e.Comments,
		}
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	entries, err := s.nineBoxRepo.BatchUpsertEntries(ctx, tx, matrixID, items)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	tx = nil

	dtos := make([]dto.NineBoxEntryDTO, 0, len(entries))
	for _, e := range entries {
		dtos = append(dtos, s.toEntryDTO(ctx, e))
	}
	return dtos, nil
}

// GetScales returns all 9×9 scale definitions.
func (s *NineBoxService) GetScales(ctx context.Context) ([]dto.NineBoxScaleDTO, error) {
	scales, err := s.catalogRepo.GetScales(ctx)
	if err != nil {
		return nil, err
	}
	dtos := make([]dto.NineBoxScaleDTO, len(scales))
	for i, sc := range scales {
		dtos[i] = dto.NineBoxScaleDTO{
			Axis: string(sc.Axis), Level: sc.Level,
			Label: sc.Label, Description: sc.Description,
		}
	}
	return dtos, nil
}

// GetQuadrants returns all 9 quadrant definitions.
func (s *NineBoxService) GetQuadrants(ctx context.Context) ([]dto.NineBoxQuadrantDTO, error) {
	quadrants, err := s.catalogRepo.GetQuadrants(ctx)
	if err != nil {
		return nil, err
	}
	dtos := make([]dto.NineBoxQuadrantDTO, len(quadrants))
	for i, q := range quadrants {
		dtos[i] = dto.NineBoxQuadrantDTO{
			Quadrant: q.Quadrant, Label: q.Label, Description: q.Description,
			Color: q.Color, ActionRecommendation: q.ActionRecommendation,
		}
	}
	return dtos, nil
}

func (s *NineBoxService) toEntryDTO(ctx context.Context, e *internal.NineBoxEntry) dto.NineBoxEntryDTO {
	dto := dto.NineBoxEntryDTO{
		ID: e.ID, EvaluateeID: e.EvaluateeID,
		PerformanceScore: e.PerformanceScore, PotentialScore: e.PotentialScore,
		Quadrant: e.Quadrant, Comments: e.Comments,
	}
	quad, err := s.catalogRepo.GetQuadrantByNumber(ctx, e.Quadrant)
	if err == nil && quad != nil {
		dto.QuadrantLabel = quad.Label
		dto.QuadrantColor = quad.Color
	}
	version, err := s.nineBoxRepo.FetchEntryVersion(ctx, e.ID)
	if err == nil {
		dto.Version = version
	}
	return dto
}
