package competency

import (
	"context"
	"time"

	dto "github.com/sed-evaluacion-desempeno/api/internal/dto/competency"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	repo "github.com/sed-evaluacion-desempeno/api/internal/repository/competency"
)

// scaleService implements ScaleService.
type scaleService struct {
	competencyRepo repo.CompetencyRepo
	scaleRepo      repo.ScaleRepo
}

// NewScaleService creates a new ScaleService.
func NewScaleService(competencyRepo repo.CompetencyRepo, scaleRepo repo.ScaleRepo) ScaleService {
	return &scaleService{
		competencyRepo: competencyRepo,
		scaleRepo:      scaleRepo,
	}
}

func (s *scaleService) GetByCompetency(ctx context.Context, competencyID string) (*dto.ScaleCriteriaResponse, error) {
	comp, err := s.competencyRepo.Get(ctx, competencyID)
	if err != nil {
		return nil, err
	}

	criteria, err := s.scaleRepo.GetByCompetency(ctx, competencyID)
	if err != nil {
		return nil, err
	}

	grouped := make(map[int][]string)
	for _, c := range criteria {
		grouped[c.Level] = append(grouped[c.Level], c.Description)
	}

	return &dto.ScaleCriteriaResponse{
		CompetencyID: competencyID,
		Criteria:     grouped,
		UpdatedAt:    comp.UpdatedAt,
	}, nil
}

func (s *scaleService) Upsert(ctx context.Context, competencyID string, req dto.ScaleCriteriaBulkRequest) (*dto.ScaleCriteriaResponse, error) {
	if len(req.Criteria) == 0 {
		return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"at least one criterion is required", nil)
	}

	// Validate levels and detect duplicates
	seen := make(map[int]struct{})
	for _, item := range req.Criteria {
		if item.Level < 1 || item.Level > 5 {
			return nil, pkgerrors.NewDomainError("INVALID_LEVEL",
				"level must be between 1 and 5", nil)
		}
		if _, ok := seen[item.Level]; ok {
			return nil, pkgerrors.NewDomainError("DUPLICATE_LEVEL",
				"duplicate level in request; each level can appear at most once", nil)
		}
		seen[item.Level] = struct{}{}
	}

	// Verify competency exists and get pillar ID
	comp, err := s.competencyRepo.Get(ctx, competencyID)
	if err != nil {
		return nil, err
	}

	// Build inputs and call ReplaceAll (which manages its own transaction)
	scaleInput := make([]repo.ScaleCriterionInput, len(req.Criteria))
	for i, item := range req.Criteria {
		scaleInput[i] = repo.ScaleCriterionInput{
			Level:       item.Level,
			Description: item.Description,
		}
	}

	if err := s.scaleRepo.ReplaceAll(ctx, competencyID, comp.PillarID.String(), scaleInput); err != nil {
		return nil, err
	}

	// Fetch updated criteria
	criteria, err := s.scaleRepo.GetByCompetency(ctx, competencyID)
	if err != nil {
		return nil, err
	}

	grouped := make(map[int][]string)
	for _, c := range criteria {
		grouped[c.Level] = append(grouped[c.Level], c.Description)
	}

	return &dto.ScaleCriteriaResponse{
		CompetencyID: competencyID,
		Criteria:     grouped,
		UpdatedAt:    time.Now(),
	}, nil
}
