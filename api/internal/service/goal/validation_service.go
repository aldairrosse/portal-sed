package goal

import (
	"context"
	"math"

	"github.com/google/uuid"
	dtogoal "github.com/sed-evaluacion-desempeno/api/internal/dto/goal"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/validation"
)

// WeightValidationService handles Double 100% weight validation.
type WeightValidationService struct {
	catRepo  CategoryRepository
	goalRepo GoalRepository
}

// NewWeightValidationService creates a new WeightValidationService.
func NewWeightValidationService(
	catRepo CategoryRepository,
	goalRepo GoalRepository,
) *WeightValidationService {
	return &WeightValidationService{
		catRepo:  catRepo,
		goalRepo: goalRepo,
	}
}

// ValidateDoubleWeighting checks that category weights sum to 100% and
// that goals within each category also sum to 100%.
func (s *WeightValidationService) ValidateDoubleWeighting(ctx context.Context, empID uuid.UUID) (*dtogoal.WeightValidationResponse, error) {
	cats, err := s.catRepo.ListCategoriesByEmployee(ctx, empID)
	if err != nil {
		return nil, err
	}

	if len(cats) == 0 {
		return &dtogoal.WeightValidationResponse{
			Valid:       false,
			CategorySum: 0,
			ExpectedSum: 100.0,
			Deficit:     100.0,
			GoalSums:    []dtogoal.CategoryGoalSum{},
		}, nil
	}

	var catSum float64
	var goalSums []dtogoal.CategoryGoalSum
	valid := true

	for _, c := range cats {
		catSum += c.Weight

		goals, err := s.goalRepo.ListGoalsByCategory(ctx, c.ID)
		if err != nil {
			return nil, err
		}

		var gSum float64
		for _, g := range goals {
			gSum += g.Weight
		}

		if math.Abs(gSum-100.0) > validation.Epsilon {
			valid = false
		}

		deficit := 100.0 - gSum
		goalSums = append(goalSums, dtogoal.CategoryGoalSum{
			CategoryID:   c.ID.String(),
			CategoryName: c.Name,
			Sum:          gSum,
			ExpectedSum:  100.0,
			Deficit:      deficit,
		})
	}

	if math.Abs(catSum-100.0) > validation.Epsilon {
		valid = false
	}

	return &dtogoal.WeightValidationResponse{
		Valid:       valid,
		CategorySum: catSum,
		ExpectedSum: 100.0,
		Deficit:     100.0 - catSum,
		GoalSums:    goalSums,
	}, nil
}
