package org

import (
	"context"
	"math"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	dto "github.com/sed-evaluacion-desempeno/api/internal/dto/org"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/state"
	repo "github.com/sed-evaluacion-desempeno/api/internal/repository/org"
)

// GoalScorer computes the weighted metas score for a single employee (0-100).
// It mirrors the 9-box hierarchical scorer (personal cat/goal weights + P/PJ + global/shared).
type GoalScorer interface {
	GetEmployeeScore(ctx context.Context, empID uuid.UUID) (float64, error)
	GetEmployeeHierarchicalScore(ctx context.Context, empID, cycleID uuid.UUID) (float64, error)
}

// MetricsService defines the interface for area metrics operations.
type MetricsService interface {
	GetAreaMetrics(ctx context.Context, nodeID, cycleID, phase string) (*dto.AreaMetricsResponse, error)
}

type metricsService struct {
	metricsRepo *repo.MetricsRepo
	nodeRepo    *repo.OrgNodeRepo
	client      *internal.Client
	scorer      GoalScorer
}

// NewMetricsService creates a new MetricsService.
func NewMetricsService(metricsRepo *repo.MetricsRepo, nodeRepo *repo.OrgNodeRepo, client *internal.Client) *metricsService {
	return &metricsService{
		metricsRepo: metricsRepo,
		nodeRepo:    nodeRepo,
		client:      client,
	}
}

// WithScorer injects the hierarchical goal scorer used for weighted department averages.
// When set, AvgProgress and AvgRating are computed as the mean of per-employee hierarchical scores (0 por defecto si no tiene metas), solo metas.
func (s *metricsService) WithScorer(scorer GoalScorer) *metricsService {
	s.scorer = scorer
	return s
}

// GetAreaMetrics computes aggregated metrics for the team of a given org node
// (direct employees plus heads of direct child nodes). When cycleID is set and
// phase is empty, phase defaults to the cycle's current_phase.
func (s *metricsService) GetAreaMetrics(ctx context.Context, nodeID, cycleID, phase string) (*dto.AreaMetricsResponse, error) {
	// Parse and validate nodeId
	nodeUUID, err := uuid.Parse(nodeID)
	if err != nil {
		return nil, errors.NewDomainError(errors.InvalidRequest, "nodeId must be a valid UUID", err)
	}

	// Verify node exists
	_, err = s.nodeRepo.GetByID(ctx, nodeUUID)
	if err != nil {
		if err == repo.ErrNodeNotFound {
			return nil, errors.NewDomainError(errors.NodeNotFound, "Org node not found", err)
		}
		return nil, err
	}

	// Get team employees (node staff plus child-node heads)
	employees, err := s.metricsRepo.GetTeamEmployees(ctx, nodeUUID)
	if err != nil {
		return nil, err
	}

	// Build response
	resp := &dto.AreaMetricsResponse{
		NodeID:        nodeID,
		EmployeeCount: len(employees),
	}

	if len(employees) == 0 {
		resp.Employees = []dto.AreaMetricsEmployee{}
		return resp, nil
	}

	// Extract employee IDs
	empIDs := make([]uuid.UUID, len(employees))
	for i, emp := range employees {
		empIDs[i] = emp.ID
	}

	// Build employees list for response
	resp.Employees = make([]dto.AreaMetricsEmployee, len(employees))
	for i, emp := range employees {
		resp.Employees[i] = dto.AreaMetricsEmployee{
			ID:                 emp.ID.String(),
			FirstName:          emp.FirstName,
			LastName:           emp.LastName,
			ProfileID:          emp.ProfileID.String(),
			ProfileName:        emp.ProfileName,
			ProfileDescription: emp.ProfileDescription,
			JobTitle:           emp.JobTitle,
		}
	}

	// Get goals for legacy counters (completed/pending/employeesWithGoals).
	goals, err := s.metricsRepo.GetGoalsByEmployees(ctx, empIDs)
	if err != nil {
		return nil, err
	}
	resp.CompletedGoals = countCompleted(goals)
	resp.PendingGoals = countPending(goals)
	resp.EmployeesWithGoals = countEmployeesWithGoals(goals)

	// Parse cycle ID and phase (para scorer cycle-scoped y para RatingsCount).
	var cycleUUID uuid.UUID
	if cycleID != "" {
		cycleUUID, err = uuid.Parse(cycleID)
		if err != nil {
			return nil, errors.NewDomainError(errors.InvalidRequest, "cycleId must be a valid UUID", err)
		}
		if phase == "" {
			phase, err = s.metricsRepo.GetCyclePhase(ctx, cycleUUID)
			if err != nil {
				return nil, err
			}
		}
	}
	if phase != "" {
		phase = state.NormalizePhase(phase)
		switch phase {
		case state.PhaseAsignacion, state.PhaseAvance, state.PhaseCierre:
		default:
			return nil, errors.NewDomainError(errors.InvalidRequest, "phase must be one of asignacion, avance, medio-anio, cierre", nil)
		}
	}

	// ── Weighted department average: mean of per-employee hierarchical goal scores ──
	// Solo metas, 0 por defecto para empleados sin metas. Usa el mismo scorer que 9-box
	// (categorías/pesos + P/PJ + global/compartidas). Si no hay scorer inyectado, fallback
	// a promedio simple por empleado con 0 para los sin metas.
	var deptAvg *float64
	if len(employees) > 0 {
		var sum float64
		if s.scorer != nil {
			for _, emp := range employees {
				score, err := s.scorer.GetEmployeeHierarchicalScore(ctx, emp.ID, cycleUUID)
				if err != nil {
					// Fallback to simple score if hierarchical fails
					if s2, err2 := s.scorer.GetEmployeeScore(ctx, emp.ID); err2 == nil {
						score = s2
					} else {
						score = 0
					}
				}
				// GetEmployeeHierarchicalScore ya incluye clamp 0-100 y 0 si no tiene metas.
				sum += score
			}
		} else {
			// Fallback sin scorer: promedio por empleado (current/target) con 0 para sin metas.
			perEmp := groupProgressByEmployee(goals)
			for _, emp := range employees {
				if v, ok := perEmp[emp.ID]; ok {
					sum += v
				} else {
					sum += 0
				}
			}
		}
		avg := sum / float64(len(employees))
		avg = math.Round(avg*10) / 10
		deptAvg = &avg
	}

	// Asignar a ambas métricas para que Avance promedio (medio-año) y Promedio final (cierre)
	// compartan el mismo cálculo ponderado de metas. Se mantiene compat con por-fase.
	if deptAvg != nil {
		resp.AvgProgress = deptAvg
		resp.AvgRating = deptAvg
	} else if len(employees) > 0 {
		zero := 0.0
		resp.AvgProgress = &zero
		resp.AvgRating = &zero
	}

	// RatingsCount sigue siendo conteo de evaluaciones RH con rating para compatibilidad del
	// card Evaluaciones en cierre (1/10). El promedio ya es de metas, pero el conteo se usa
	// solo como numerador del total de colaboradores.
	if cycleUUID != uuid.Nil {
		ratings, err := s.metricsRepo.GetRHEvaluationsByEmployees(ctx, empIDs, cycleUUID, phase)
		if err != nil {
			return nil, err
		}
		resp.RatingsCount = len(ratings)
	}

	return resp, nil
}

// computeAvgProgress calculates the mean progress percentage across all goals.
// Progress for each goal is (current_value / target_value) * 100.
// Returns nil if no goals are provided.
func computeAvgProgress(goals []*repo.GoalRow) *float64 {
	if len(goals) == 0 {
		return nil
	}
	var sum float64
	count := 0
	for _, g := range goals {
		if g.TargetValue > 0 {
			sum += (g.CurrentValue / g.TargetValue) * 100
			count++
		}
	}
	if count == 0 {
		return nil
	}
	avg := sum / float64(count)
	// Round to one decimal place
	avg = math.Round(avg*10) / 10
	return &avg
}

// computeAvgRating calculates the mean of all RH ratings.
// Returns nil if no ratings are provided.
func computeAvgRating(ratings []*repo.RHRatingRow) *float64 {
	if len(ratings) == 0 {
		return nil
	}
	var sum float64
	for _, r := range ratings {
		sum += r.RHRating
	}
	avg := sum / float64(len(ratings))
	// Round to one decimal place
	avg = math.Round(avg*10) / 10
	return &avg
}

// countCompleted counts goals where current_value >= target_value.
func countCompleted(goals []*repo.GoalRow) int {
	count := 0
	for _, g := range goals {
		if g.CurrentValue >= g.TargetValue {
			count++
		}
	}
	return count
}

// countPending counts goals where current_value < target_value.
func countPending(goals []*repo.GoalRow) int {
	count := 0
	for _, g := range goals {
		if g.CurrentValue < g.TargetValue {
			count++
		}
	}
	return count
}

// countEmployeesWithGoals returns the count of distinct employees that have
// at least one goal.
func countEmployeesWithGoals(goals []*repo.GoalRow) int {
	seen := make(map[uuid.UUID]struct{})
	for _, g := range goals {
		seen[g.EmployeeID] = struct{}{}
	}
	return len(seen)
}

// groupProgressByEmployee promedia el progreso por empleado (current/target) para el fallback sin scorer.
// Retorna map employeeID -> avg progress 0-100 de ese empleado. Empleados sin metas no aparecen (caller pone 0).
func groupProgressByEmployee(goals []*repo.GoalRow) map[uuid.UUID]float64 {
	byEmp := make(map[uuid.UUID][]float64)
	for _, g := range goals {
		if g.TargetValue <= 0 {
			continue
		}
		pct := g.CurrentValue / g.TargetValue * 100
		if pct < 0 {
			pct = 0
		}
		if pct > 100 {
			pct = 100
		}
		byEmp[g.EmployeeID] = append(byEmp[g.EmployeeID], pct)
	}
	avg := make(map[uuid.UUID]float64, len(byEmp))
	for emp, vals := range byEmp {
		var sum float64
		for _, v := range vals {
			sum += v
		}
		avg[emp] = sum / float64(len(vals))
	}
	return avg
}
