package evaluation

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	dto "github.com/sed-evaluacion-desempeno/api/internal/dto/evaluation"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
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
		PhaseID: m.PhaseID, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
	if m.Edges.Phase != nil {
		resp.PhaseLabel = m.Edges.Phase.Label
	}
	entries := m.Edges.Entries
	if entries == nil {
		entries = []*internal.NineBoxEntry{}
	}
	infoMap := s.loadEmployeeInfoMap(ctx, entries)
	resp.Entries = make([]dto.NineBoxEntryDTO, len(entries))
	for i, e := range entries {
		resp.Entries[i] = s.toEntryDTO(ctx, e, infoMap)
	}
	return resp, nil
}

// GetMatrixByPhase retrieves a matrix by cycle + evaluator + phase, with entries populated.
func (s *NineBoxService) GetMatrixByPhase(ctx context.Context, cycleID, evaluatorID, phaseID uuid.UUID) (*dto.NineBoxMatrixResponse, error) {
	m, err := s.nineBoxRepo.GetMatrixByPhase(ctx, cycleID, evaluatorID, phaseID)
	if err != nil {
		return nil, err
	}
	resp := &dto.NineBoxMatrixResponse{
		ID: m.ID, CycleID: m.CycleID, EvaluatorID: m.EvaluatorID,
		PhaseID: m.PhaseID, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
	if m.Edges.Phase != nil {
		resp.PhaseLabel = m.Edges.Phase.Label
	}
	entries := m.Edges.Entries
	if entries == nil {
		entries = []*internal.NineBoxEntry{}
	}
	infoMap := s.loadEmployeeInfoMap(ctx, entries)
	resp.Entries = make([]dto.NineBoxEntryDTO, len(entries))
	for i, e := range entries {
		resp.Entries[i] = s.toEntryDTO(ctx, e, infoMap)
	}
	return resp, nil
}

// ListMatrices returns matrices filtered by cycle, evaluator, and/or phase.
func (s *NineBoxService) ListMatrices(ctx context.Context, cycleID, evaluatorID, phaseID uuid.UUID) ([]dto.NineBoxMatrixResponse, error) {
	matrices, err := s.nineBoxRepo.ListMatrices(ctx, cycleID, evaluatorID, phaseID)
	if err != nil {
		return nil, err
	}
	resp := make([]dto.NineBoxMatrixResponse, len(matrices))
	for i, m := range matrices {
		resp[i] = dto.NineBoxMatrixResponse{
			ID: m.ID, CycleID: m.CycleID, EvaluatorID: m.EvaluatorID,
			PhaseID: m.PhaseID, Entries: []dto.NineBoxEntryDTO{},
			CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
		}
	}
	return resp, nil
}

// RecomputeMatrix recalculates all nine-box placements for a given cycle and phase.
//
// The method:
//  1. Retrieves all employees with GoalAssignments in the cycle.
//  2. Resolves the evaluator (manager) for each employee.
//  3. For each employee, computes performance tier from goal progress and potential tier
//     from competency ratings (self + HR).
//  4. Computes quadrant from tiers.
//  5. Creates/get the NineBoxMatrix for each evaluator (using the cycle's evaluator structure).
//  6. Upserts NineBoxEntry for each employee.
func (s *NineBoxService) RecomputeMatrix(ctx context.Context, cycleID, phaseID uuid.UUID) error {
	// 1. Get all employees with goal assignments in this cycle
	employeeIDs, err := s.nineBoxRepo.GetGoalAssigneesByCycle(ctx, cycleID)
	if err != nil {
		return err
	}

	if len(employeeIDs) == 0 {
		return nil
	}

	// 2. Resolve real evaluators from manager mapping. Employees without a manager are skipped.
	managerMapping, err := s.nineBoxRepo.GetManagerMapping(ctx, employeeIDs)
	if err != nil {
		return err
	}

	evaluatorGroups := make(map[uuid.UUID][]uuid.UUID)
	for _, empID := range employeeIDs {
		managerID, ok := managerMapping[empID]
		if !ok {
			continue
		}
		evaluatorGroups[managerID] = append(evaluatorGroups[managerID], empID)
	}

	if len(evaluatorGroups) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	for evaluatorID, evaluatees := range evaluatorGroups {
		// 2. Get or create matrix for (cycleID, evaluatorID, phaseID)
		matrix, err := s.nineBoxRepo.GetMatrixByPhase(ctx, cycleID, evaluatorID, phaseID)
		if err != nil {
			if err == repo.ErrMatrixNotFound {
				matrix, err = s.nineBoxRepo.CreateMatrixWithPhase(ctx, cycleID, evaluatorID, phaseID)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		for _, evaluateeID := range evaluatees {
			// 3a. Compute performance tier from goal progress
			avgProgress, err := s.nineBoxRepo.GetGoalProgressByEmployee(ctx, evaluateeID, cycleID)
			if err != nil {
				return err
			}
			perfTier := quadrant.ComputePerformanceTier(avgProgress)

			// 3b. Compute potential tier from competency ratings
			selfRating, hrRating, err := s.nineBoxRepo.GetCompetencyRatingsByEmployee(ctx, evaluateeID, cycleID)
			if err != nil {
				return err
			}
			potTier := quadrant.ComputePotentialTier(selfRating, hrRating)

			// 3c. Compute quadrant
			q := quadrant.ComputeQuadrantFromTiers(perfTier, potTier)

			// 4. Upsert entry
			_, err = s.nineBoxRepo.UpsertEntryByTiers(ctx, tx, matrix.ID, evaluateeID, perfTier, potTier, q, "")
			if err != nil {
				return err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	tx = nil

	return nil
}

// GetMatrixEntriesFiltered returns matrix entries, optionally filtered by quadrant.
func (s *NineBoxService) GetMatrixEntriesFiltered(ctx context.Context, matrixID uuid.UUID, quadrant *int) ([]dto.NineBoxEntryDTO, error) {
	var entries []*internal.NineBoxEntry
	var err error
	if quadrant != nil {
		entries, err = s.nineBoxRepo.GetMatrixEntriesByQuadrant(ctx, matrixID, *quadrant)
	} else {
		entries, err = s.nineBoxRepo.GetMatrixEntries(ctx, matrixID)
	}
	if err != nil {
		return nil, err
	}

	infoMap := s.loadEmployeeInfoMap(ctx, entries)
	dtos := make([]dto.NineBoxEntryDTO, len(entries))
	for i, e := range entries {
		dtos[i] = s.toEntryDTO(ctx, e, infoMap)
	}
	return dtos, nil
}

// UpsertEntry creates or updates a single matrix entry with quadrant computation.
// Deprecated: Use RecomputeMatrix for automatic tier computation.
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

	comments := ""
	if req.Comments != nil {
		comments = *req.Comments
	}
	entry, err := s.nineBoxRepo.UpsertEntry(ctx, tx, matrixID, req.EvaluateeID,
		req.PerformanceScore, req.PotentialScore, q, comments)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	tx = nil

	dto := s.toEntryDTO(ctx, entry, nil)
	return &dto, nil
}

// UpdateEntry updates an existing entry with optimistic lock.
// Deprecated: Use RecomputeMatrix for automatic tier computation.
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

	comments := ""
	if req.Comments != nil {
		comments = *req.Comments
	}
	entry, err := s.nineBoxRepo.UpdateEntry(ctx, tx, entryID,
		req.PerformanceScore, req.PotentialScore, q, comments, ifMatch)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	tx = nil

	dto := s.toEntryDTO(ctx, entry, nil)
	return &dto, nil
}

// BatchSubmitEntries atomically submits multiple entries in a single transaction.
// Deprecated: Use RecomputeMatrix for automatic tier computation.
func (s *NineBoxService) BatchSubmitEntries(ctx context.Context, matrixID uuid.UUID, req dto.NineBoxBatchRequest) ([]dto.NineBoxEntryDTO, error) {
	if len(req.Entries) > 20 {
		return nil, pkgerrors.NewDomainError(pkgerrors.BatchSizeExceeded,
			"Batch size exceeds the maximum allowed (20).", nil)
	}

	seen := make(map[uuid.UUID]bool)
	for _, e := range req.Entries {
		if seen[e.EvaluateeID] {
			return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
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
		c := ""
		if e.Comments != nil {
			c = *e.Comments
		}
		items[i] = repo.EntryUpsert{
			EvaluateeID: e.EvaluateeID, PerformanceTier: e.PerformanceScore,
			PotentialTier: e.PotentialScore, Quadrant: q, Comments: c,
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
		dtos = append(dtos, s.toEntryDTO(ctx, e, nil))
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
			Color: q.Color, ColorHex: q.ColorHex, Title: q.Title,
			ActionRecommendation: q.ActionRecommendation,
		}
	}
	return dtos, nil
}

// UpdateQuadrantByNumber updates quadrant title, description, and colorHex by quadrant number (1-9).
func (s *NineBoxService) UpdateQuadrantByNumber(ctx context.Context, quadrantNumber int, input dto.NineBoxQuadrantUpdateInput) (*dto.NineBoxQuadrantDTO, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.ExecContext(ctx,
		`UPDATE nine_box_quadrants SET title = $1, description = $2, color_hex = $3 WHERE quadrant = $4`,
		input.Title, input.Description, input.ColorHex, quadrantNumber,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	tx = nil

	quad, err := s.catalogRepo.GetQuadrantByNumber(ctx, quadrantNumber)
	if err != nil {
		return nil, err
	}

	dtoResp := dto.NineBoxQuadrantDTO{
		Quadrant: quad.Quadrant, Label: quad.Label, Description: quad.Description,
		Color: quad.Color, ColorHex: quad.ColorHex, Title: quad.Title,
		ActionRecommendation: quad.ActionRecommendation,
	}
	return &dtoResp, nil
}

func (s *NineBoxService) loadEmployeeInfoMap(ctx context.Context, entries []*internal.NineBoxEntry) map[uuid.UUID]*repo.EmployeeInfo {
	if len(entries) == 0 {
		return map[uuid.UUID]*repo.EmployeeInfo{}
	}
	ids := make([]uuid.UUID, 0, len(entries))
	for _, e := range entries {
		if e == nil {
			continue
		}
		ids = append(ids, e.EvaluateeID)
	}
	infoMap, err := s.nineBoxRepo.GetEmployeesByIDs(ctx, ids)
	if err != nil {
		return map[uuid.UUID]*repo.EmployeeInfo{}
	}
	return infoMap
}

func (s *NineBoxService) toEntryDTO(ctx context.Context, e *internal.NineBoxEntry, infoMap map[uuid.UUID]*repo.EmployeeInfo) dto.NineBoxEntryDTO {
	dtoOut := dto.NineBoxEntryDTO{
		ID: e.ID, EvaluateeID: e.EvaluateeID,
		PerformanceTier: e.PerformanceTier, PotentialTier: e.PotentialTier,
		Quadrant: e.Quadrant, Comments: e.Comments,
	}
	if infoMap != nil {
		if info, ok := infoMap[e.EvaluateeID]; ok && info != nil {
			dtoOut.EmployeeName = info.FirstName + " " + info.LastName
			dtoOut.ProfileID = info.ProfileID
		}
	}
	quad, err := s.catalogRepo.GetQuadrantByNumber(ctx, e.Quadrant)
	if err == nil && quad != nil {
		dtoOut.QuadrantLabel = quad.Label
		if quad.ColorHex != "" {
			dtoOut.QuadrantColor = quad.ColorHex
		} else {
			dtoOut.QuadrantColor = quad.Color
		}
	}
	version, err := s.nineBoxRepo.FetchEntryVersion(ctx, e.ID)
	if err == nil {
		dtoOut.Version = version
	}
	return dtoOut
}
