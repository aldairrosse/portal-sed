package competency

import (
	"context"

	dto "github.com/sed-evaluacion-desempeno/api/internal/dto/competency"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	repo "github.com/sed-evaluacion-desempeno/api/internal/repository/competency"
)

// acceptanceService implements AcceptanceService.
type acceptanceService struct {
	competencyRepo repo.CompetencyRepo
	catalogRepo    repo.CatalogRepo
	acceptanceRepo repo.AcceptanceRepo
}

// NewAcceptanceService creates a new AcceptanceService.
func NewAcceptanceService(
	competencyRepo repo.CompetencyRepo,
	catalogRepo repo.CatalogRepo,
	acceptanceRepo repo.AcceptanceRepo,
) AcceptanceService {
	return &acceptanceService{
		competencyRepo: competencyRepo,
		catalogRepo:    catalogRepo,
		acceptanceRepo: acceptanceRepo,
	}
}

func (s *acceptanceService) List(ctx context.Context, filter AcceptanceFilter) ([]dto.AcceptanceLevelItem, error) {
	acceptanceLevels, err := s.acceptanceRepo.List(ctx, filter.CompetencyID, filter.ProfileID)
	if err != nil {
		return nil, err
	}

	items := make([]dto.AcceptanceLevelItem, len(acceptanceLevels))
	for i, a := range acceptanceLevels {
		items[i] = dto.AcceptanceLevelItem{
			ID:           a.ID.String(),
			CompetencyID: a.CompetencyID.String(),
			ProfileID:    a.ProfileID.String(),
			Level:        a.Level,
			CreatedAt:    a.CreatedAt,
			UpdatedAt:    a.UpdatedAt,
		}
	}
	return items, nil
}

func (s *acceptanceService) Upsert(ctx context.Context, req dto.UpsertAcceptanceRequest) (*dto.AcceptanceLevelItem, error) {
	// Validate level range
	if req.Level < 1 || req.Level > 5 {
		return nil, pkgerrors.NewDomainError("INVALID_LEVEL",
			"level must be between 1 and 5", nil)
	}

	// Verify competency exists
	_, err := s.competencyRepo.Get(ctx, req.CompetencyID)
	if err != nil {
		return nil, err
	}

	// Verify profile exists
	profiles, err := s.catalogRepo.ListProfiles(ctx)
	if err != nil {
		return nil, err
	}
	profileFound := false
	for _, p := range profiles {
		if p.ID.String() == req.ProfileID {
			profileFound = true
			break
		}
	}
	if !profileFound {
		return nil, pkgerrors.NewDomainError("RESOURCE_NOT_FOUND",
			"evaluation profile not found", nil)
	}

	// Upsert
	result, err := s.acceptanceRepo.Upsert(ctx, req.CompetencyID, req.ProfileID, req.Level)
	if err != nil {
		return nil, err
	}

	return &dto.AcceptanceLevelItem{
		ID:           result.ID.String(),
		CompetencyID: result.CompetencyID.String(),
		ProfileID:    result.ProfileID.String(),
		Level:        result.Level,
		CreatedAt:    result.CreatedAt,
		UpdatedAt:    result.UpdatedAt,
	}, nil
}
