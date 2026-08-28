package goal

import (
	"context"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/scoring"
)

// WeightResolver resolves hierarchical P/PJ weights for an employee.
type WeightResolver interface {
	GetEmployeeHierarchicalWeights(ctx context.Context, empID uuid.UUID) (float64, float64)
}

// ScoringService handles employee score calculation.
type ScoringService struct {
	catRepo     CategoryRepository
	goalRepo    GoalRepository
	weightSvc   WeightResolver
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
