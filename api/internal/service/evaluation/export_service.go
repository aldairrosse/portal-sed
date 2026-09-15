// Package evaluation provides business logic for evaluation self-evaluation,
// RH evaluation, and finalization during the year-end "cierre" phase.
package evaluation

import (
	"context"
	"math"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal/auth"
	dto "github.com/sed-evaluacion-desempeno/api/internal/dto/evaluation"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/quadrant"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/state"
	cyclerepo "github.com/sed-evaluacion-desempeno/api/internal/repository/cycle"
	repoeval "github.com/sed-evaluacion-desempeno/api/internal/repository/evaluation"
	orgrepo "github.com/sed-evaluacion-desempeno/api/internal/repository/org"
)

// ExportService builds the 7-column evaluations export over the full viewer
// scope (all employees for RH, direct reports otherwise). It reuses the
// hierarchical goal scorer, batch competency averages, and the same
// pending/in-progress/completed derivation as the employee lists.
type ExportService struct {
	evalRepo  *repoeval.EvaluationRepo
	empRepo   *orgrepo.EmployeeRepo
	cycleRepo *cyclerepo.CycleRepo
	scorer    HierarchicalScorer
}

// NewExportService creates a new ExportService.
func NewExportService(
	evalRepo *repoeval.EvaluationRepo,
	empRepo *orgrepo.EmployeeRepo,
	cycleRepo *cyclerepo.CycleRepo,
	scorer HierarchicalScorer,
) *ExportService {
	return &ExportService{
		evalRepo:  evalRepo,
		empRepo:   empRepo,
		cycleRepo: cycleRepo,
		scorer:    scorer,
	}
}

// Export returns one row per in-scope employee for the active (or given)
// cycle+phase. Averages default to 0 when there is no phase data.
func (s *ExportService) Export(ctx context.Context, cycleID uuid.UUID, phase, query string, viewerID uuid.UUID, viewerRole auth.Role) (*dto.EvaluationExportResponse, error) {
	if cycleID == uuid.Nil {
		active, err := s.cycleRepo.GetActive(ctx)
		if err != nil {
			return nil, err
		}
		if active == nil {
			return nil, pkgerrors.NewDomainError(pkgerrors.CycleNotFound, "no active cycle", nil)
		}
		cycleID = active.ID
	}
	if phase == "" {
		if c, err := s.cycleRepo.GetCycle(ctx, cycleID); err == nil && c != nil {
			phase = string(c.CurrentPhase)
		}
	}
	phase = state.NormalizePhase(phase)

	activeOnly := true
	ids, err := s.empRepo.ListIDsWithProfiles(ctx, orgrepo.EmployeeFilter{IsActive: &activeOnly, Query: query})
	if err != nil {
		return nil, err
	}
	if !auth.RoleSeesAll(viewerRole) {
		if ids, err = s.scopeToReports(ctx, viewerID, ids); err != nil {
			return nil, err
		}
	}

	// Batch employee rows (GetByIDs caps at 100 per call).
	var empRows []*orgrepo.EmployeeRow
	for i := 0; i < len(ids); i += 100 {
		end := i + 100
		if end > len(ids) {
			end = len(ids)
		}
		chunk, err := s.empRepo.GetByIDsWithProfiles(ctx, ids[i:end])
		if err != nil {
			return nil, err
		}
		empRows = append(empRows, chunk...)
	}
	empIDs := make([]uuid.UUID, len(empRows))
	for i, r := range empRows {
		empIDs[i] = r.ID
	}

	// Same 3-query enrichment as the employee lists (nil-safe: errors => empty).
	avgMap := map[uuid.UUID]repoeval.EmployeeAvg{}
	goalMap := map[uuid.UUID]repoeval.EmployeeGoalStatus{}
	compTotal := 0
	if len(empIDs) > 0 {
		if m, err := s.evalRepo.BatchAvgByEmployeeIDs(ctx, cycleID, phase, empIDs); err == nil {
			avgMap = m
		}
		if m, err := s.evalRepo.BatchGoalStatusByEmployeeIDs(ctx, cycleID, phase, empIDs); err == nil {
			goalMap = m
		}
		if n, err := s.evalRepo.CompetencyTotalCount(ctx); err == nil {
			compTotal = n
		}
	}

	rows := make([]dto.EvaluationExportRow, 0, len(empRows))
	for _, r := range empRows {
		var self, rh float64
		rated := 0
		if a, ok := avgMap[r.ID]; ok {
			if a.SelfAvg != nil {
				self = *a.SelfAvg
			}
			if a.RhAvg != nil {
				rh = *a.RhAvg
			}
			rated = a.SelfCount
		}
		gt, gd := 0, 0
		if g, ok := goalMap[r.ID]; ok {
			gt, gd = g.Total, g.Done
		}
		progress := 0.0
		if s.scorer != nil {
			if p, err := s.scorer.GetEmployeeHierarchicalScore(ctx, r.ID, cycleID); err == nil {
				progress = p
			}
		}
		// Same visible-state derivation as ListEmployees enrichment,
		// labeled in Spanish sentence case for the xlsx export.
		status := "Pendiente"
		switch {
		case rated == 0 && gd == 0:
			status = "Pendiente"
		case compTotal <= 0:
			status = "En progreso"
		case rated < compTotal || gd < gt:
			status = "En progreso"
		default:
			status = "Completada"
		}
		rows = append(rows, dto.EvaluationExportRow{
			EmployeeID:     r.ID,
			EmployeeNumber: r.EmployeeNumber,
			EmployeeName:   r.FirstName + " " + r.LastName,
			GoalProgress:   math.Round(progress*100) / 100,
			SelfAvg:        self,
			RhAvg:          rh,
			Rating:         self*quadrant.DefaultWeightSelf + rh*quadrant.DefaultWeightRH,
			Status:         status,
		})
	}
	if rows == nil {
		rows = []dto.EvaluationExportRow{}
	}
	return &dto.EvaluationExportResponse{
		Data: rows,
		Meta: dto.EvaluationExportMeta{CycleID: cycleID, Phase: phase, Total: len(rows)},
	}, nil
}

// scopeToReports restricts IDs to the viewer's direct reports only (1 level,
// active). Mirrors GetMyEvaluatees/ListByManager; RoleSeesAll bypasses upstream.
func (s *ExportService) scopeToReports(ctx context.Context, viewerID uuid.UUID, employeeIDs []uuid.UUID) ([]uuid.UUID, error) {
	reports, err := s.empRepo.ListByManager(ctx, viewerID, true)
	if err != nil {
		return nil, err
	}
	allowed := make(map[uuid.UUID]struct{}, len(reports))
	for _, r := range reports {
		allowed[r.ID] = struct{}{}
	}
	filtered := make([]uuid.UUID, 0, len(employeeIDs))
	for _, id := range employeeIDs {
		if _, ok := allowed[id]; ok {
			filtered = append(filtered, id)
		}
	}
	return filtered, nil
}
