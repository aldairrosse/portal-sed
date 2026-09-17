package goal

import (
	"context"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/scoring"
	repogoal "github.com/sed-evaluacion-desempeno/api/internal/repository/goal"
)

// WeightResolver resolves hierarchical P/PJ weights for an employee.
type WeightResolver interface {
	GetEmployeeHierarchicalWeights(ctx context.Context, empID uuid.UUID) (float64, float64)
}

// ScoringService handles employee score calculation.
type ScoringService struct {
	catRepo    CategoryRepository
	goalRepo   GoalRepository
	weightSvc  WeightResolver
	assignRepo AssignmentRepository
}

// NewScoringService creates a new ScoringService.
func NewScoringService(
	catRepo CategoryRepository,
	goalRepo GoalRepository,
) *ScoringService {
	return &ScoringService{
		catRepo:  catRepo,
		goalRepo: goalRepo,
	}
}

// WithAssignmentRepo injects the assignment repo for global/shared contributions.
func (s *ScoringService) WithAssignmentRepo(a AssignmentRepository) *ScoringService {
	s.assignRepo = a
	return s
}

// WithWeightResolver injects the hierarchical weight resolver.
func (s *ScoringService) WithWeightResolver(w WeightResolver) *ScoringService {
	s.weightSvc = w
	return s
}

// GetEmployeeScore calculates the overall weighted score for an employee.
func (s *ScoringService) GetEmployeeScore(ctx context.Context, empID uuid.UUID) (float64, error) {
	// Load all categories for the employee
	cats, err := s.catRepo.ListCategoriesByEmployee(ctx, empID)
	if err != nil {
		return 0, err
	}

	// Build category scores
	catScores := make([]scoring.CategoryScore, 0, len(cats))
	for _, cat := range cats {
		goals, err := s.goalRepo.ListGoalsByCategory(ctx, cat.ID)
		if err != nil {
			return 0, err
		}

		goalScores := make([]scoring.GoalScore, 0, len(goals))
		for _, g := range goals {
			baselineVal := 0.0
			if g.BaselineValue != nil {
				baselineVal = *g.BaselineValue
			}
			goalScores = append(goalScores, scoring.GoalScore{
				Weight:          g.Weight,
				ProgressPercent: scoring.ProgressPercent(g.CurrentValue, g.TargetValue, baselineVal, g.Direction),
			})
		}

		catScores = append(catScores, scoring.CategoryScore{
			Weight: cat.Weight,
			Goals:  goalScores,
		})
	}

	personalScore := scoring.EmployeeScore(catScores)
	pWeight, pjWeight := 100.0, 100.0
	if s.weightSvc != nil {
		pWeight, pjWeight = s.weightSvc.GetEmployeeHierarchicalWeights(ctx, empID)
	}
	return scoring.HierarchicalScore(personalScore, pWeight, pjWeight), nil
}

// GetEmployeeHierarchicalScore returns the final 0-100 score (same scorer as 9-box,
// also exposed for home/avance via GoalScorer):
// personalHJ (catW/w_i/P/PJ) + global (w_i*G) + shared (w_i*J*P).
// Each goal uses ProgressPercent(current/target/baseline, direction).
// cycleID is kept for API compatibility and future cycle scoping; avance reads
// live current values (not cycle snapshots), so goals are intentionally not
// filtered by cycleID here. Weights resolve via WeightResolver (active cycle)
// with fallback 100. NULL snapshots are excluded via ProgressPercent (target<=0 → 0).
// hasNoInstitutionalGoals centralizes the "solo metas personales" condition:
// true when the employee has no global goals assigned nor shared goals linked
// (rows without assignments/members don't count as assigned).
func hasNoInstitutionalGoals(globals []*repogoal.GlobalGoalRow, shared []*repogoal.SharedGoalRow) bool {
	for _, gg := range globals {
		if gg != nil && len(gg.Assignments) > 0 {
			return false
		}
	}
	for _, sg := range shared {
		if sg != nil && len(sg.Members) > 0 {
			return false
		}
	}
	return true
}

func (s *ScoringService) GetEmployeeHierarchicalScore(ctx context.Context, empID, cycleID uuid.UUID) (float64, error) {
	_ = cycleID // avance: live values, no filtrar por ciclo; se conserva para scoping futuro/home
	// List institutional goals first: centralized check needs them before weights.
	var globals []*repogoal.GlobalGoalRow
	var shared []*repogoal.SharedGoalRow
	if s.assignRepo != nil {
		if g, err := s.assignRepo.ListGlobalGoalsByEmployee(ctx, empID); err == nil {
			globals = g
		}
		if sh, err := s.assignRepo.ListSharedGoalsAsMember(ctx, empID); err == nil {
			shared = sh
		}
	}
	pWeight, pjWeight := 100.0, 100.0
	if s.weightSvc != nil {
		pWeight, pjWeight = s.weightSvc.GetEmployeeHierarchicalWeights(ctx, empID)
	}
	if hasNoInstitutionalGoals(globals, shared) {
		// Solo personales: ignorar config ciclo/equipo, personal 100%.
		pWeight, pjWeight = 100, 100
	}
	if pWeight == 0 {
		pWeight = 100
	}
	if pjWeight == 0 {
		pjWeight = 100
	}
	gWeight := 100 - pWeight
	if gWeight < 0 {
		gWeight = 0
	}
	jWeight := 100 - pjWeight
	if jWeight < 0 {
		jWeight = 0
	}

	cats, err := s.catRepo.ListCategoriesByEmployee(ctx, empID)
	if err != nil {
		return 0, err
	}
	catScores := make([]scoring.CategoryScore, 0, len(cats))
	for _, cat := range cats {
		goals, err := s.goalRepo.ListGoalsByCategory(ctx, cat.ID)
		if err != nil {
			return 0, err
		}
		goalScores := make([]scoring.GoalScore, 0, len(goals))
		for _, g := range goals {
			baselineVal := 0.0
			if g.BaselineValue != nil {
				baselineVal = *g.BaselineValue
			}
			goalScores = append(goalScores, scoring.GoalScore{
				Weight:          g.Weight,
				ProgressPercent: scoring.ProgressPercent(g.CurrentValue, g.TargetValue, baselineVal, g.Direction),
			})
		}
		catScores = append(catScores, scoring.CategoryScore{
			Weight: cat.Weight,
			Goals:  goalScores,
		})
	}
	personalHJ := scoring.HierarchicalScore(scoring.EmployeeScore(catScores), pWeight, pjWeight)

	var globalSum, sharedSum float64
	for _, gg := range globals {
		if gg == nil || len(gg.Assignments) == 0 {
			continue
		}
		a := gg.Assignments[0]
		target := a.TargetValue
		if target == 0 {
			target = gg.TargetValue
		}
		baseline := 0.0
		if a.BaselineValue != nil {
			baseline = *a.BaselineValue
		} else if gg.BaselineValue != nil {
			baseline = *gg.BaselineValue
		}
		pct := scoring.ProgressPercent(gg.CurrentValue, target, baseline, gg.Direction)
		globalSum += pct * a.Weight / 100 * gWeight / 100
	}
	for _, sg := range shared {
		if sg == nil || len(sg.Members) == 0 {
			continue
		}
		m := sg.Members[0]
		target := m.TargetValue
		if target == 0 {
			target = sg.TargetValue
		}
		baseline := 0.0
		if m.BaselineValue != nil {
			baseline = *m.BaselineValue
		} else if sg.BaselineValue != nil {
			baseline = *sg.BaselineValue
		}
		pct := scoring.ProgressPercent(sg.CurrentValue, target, baseline, sg.Direction)
		sharedSum += pct * m.Weight / 100 * jWeight / 100 * pWeight / 100
	}
	return scoring.FinalScore(personalHJ, globalSum, sharedSum), nil
}
