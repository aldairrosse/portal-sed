package goal

import (
	"context"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/scoring"
)

// ScoringService handles employee score calculation.
type ScoringService struct {
	catRepo  CategoryRepository
	goalRepo GoalRepository
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
	// Hierarchical G/P J/PJ: fetch weights with fallback 100 when no cycle/team config
	// TODO: wire CycleConfig/TeamWeightConfig lookup once cycle/team resolvers available
	pWeight, pjWeight := 100.0, 100.0
	_ = ctx // placeholder for future config fetch
	return scoring.HierarchicalScore(personalScore, pWeight, pjWeight), nil
}
