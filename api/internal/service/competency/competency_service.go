package competency

import (
	"context"
	"fmt"
	"time"

	dto "github.com/sed-evaluacion-desempeno/api/internal/dto/competency"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	repo "github.com/sed-evaluacion-desempeno/api/internal/repository/competency"
)

// competencyService implements CompetencyService.
type competencyService struct {
	pillarRepo repo.PillarRepo
	repo       repo.CompetencyRepo
}

// NewCompetencyService creates a new CompetencyService.
func NewCompetencyService(pillarRepo repo.PillarRepo, competencyRepo repo.CompetencyRepo) CompetencyService {
	return &competencyService{
		pillarRepo: pillarRepo,
		repo:       competencyRepo,
	}
}

func (s *competencyService) ListByPillar(ctx context.Context, pillarID string, opts ListOptions) (*ListResult[dto.CompetencyLite], error) {
	limit := opts.Limit
	if limit < 1 {
		limit = 1
	}
	if limit > 100 {
		limit = 100
	}

	// Verify pillar exists
	_, err := s.pillarRepo.Get(ctx, pillarID, false)
	if err != nil {
		return nil, err
	}

	competencies, nextCursor, err := s.repo.ListByPillar(ctx, pillarID, opts.Cursor, limit)
	if err != nil {
		return nil, err
	}

	items := make([]dto.CompetencyLite, len(competencies))
	for i, c := range competencies {
		items[i] = dto.CompetencyLite{
			ID:          c.ID.String(),
			Name:        c.Name,
			Description: c.Description,
		}
	}

	hasMore := nextCursor != ""
	return &ListResult[dto.CompetencyLite]{
		Data:       items,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (s *competencyService) Get(ctx context.Context, id string) (*dto.CompetencyDetail, error) {
	c, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	detail := &dto.CompetencyDetail{
		ID:          c.ID.String(),
		PillarID:    c.PillarID.String(),
		Name:        c.Name,
		Description: c.Description,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}

	// Group scale criteria by level
	if c.Edges.ScaleCriteria != nil {
		criteria := make(map[int][]string)
		for _, sc := range c.Edges.ScaleCriteria {
			criteria[sc.Level] = append(criteria[sc.Level], sc.Description)
		}
		detail.ScaleCriteria = criteria
	}

	return detail, nil
}

func (s *competencyService) Create(ctx context.Context, pillarID string, req dto.CreateCompetencyRequest) (*dto.CompetencyDetail, error) {
	if req.Name == "" {
		return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"name is required", nil)
	}
	if len(req.Name) > 255 {
		return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"name must be at most 255 characters", nil)
	}
	if len(req.Description) > 2000 {
		return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"description must be at most 2000 characters", nil)
	}

	// Verify pillar exists
	_, err := s.pillarRepo.Get(ctx, pillarID, false)
	if err != nil {
		return nil, err
	}

	c, err := s.repo.Create(ctx, pillarID, req.Name, req.Description)
	if err != nil {
		return nil, err
	}

	return &dto.CompetencyDetail{
		ID:          c.ID.String(),
		PillarID:    c.PillarID.String(),
		Name:        c.Name,
		Description: c.Description,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}, nil
}

func (s *competencyService) Update(ctx context.Context, id string, req dto.UpdateCompetencyRequest, ifMatch time.Time) (*dto.CompetencyDetail, error) {
	if req.Name == "" {
		return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"name is required", nil)
	}
	if len(req.Name) > 255 {
		return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"name must be at most 255 characters", nil)
	}
	if len(req.Description) > 2000 {
		return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"description must be at most 2000 characters", nil)
	}

	// If pillar move requested, verify target pillar exists
	if req.PillarID != "" {
		_, err := s.pillarRepo.Get(ctx, req.PillarID, false)
		if err != nil {
			return nil, err
		}
	}

	c, err := s.repo.Update(ctx, id, req.Name, req.Description, req.PillarID, ifMatch)
	if err != nil {
		return nil, err
	}

	detail := &dto.CompetencyDetail{
		ID:          c.ID.String(),
		PillarID:    c.PillarID.String(),
		Name:        c.Name,
		Description: c.Description,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}

	if c.Edges.ScaleCriteria != nil {
		criteria := make(map[int][]string)
		for _, sc := range c.Edges.ScaleCriteria {
			criteria[sc.Level] = append(criteria[sc.Level], sc.Description)
		}
		detail.ScaleCriteria = criteria
	}

	return detail, nil
}

func (s *competencyService) Delete(ctx context.Context, id string, force bool) error {
	count, err := s.repo.CountScaleCriteria(ctx, id)
	if err != nil {
		return err
	}

	if count > 0 && !force {
		return pkgerrors.NewDomainError("COMPETENCY_HAS_CRITERIA",
			"cannot delete competency with existing scale criteria; use force=true to cascade",
			nil).WithDetails(
			fmt.Sprintf("competency_id: %s", id),
			fmt.Sprintf("criteria_count: %d", count),
			"action: use force=true to cascade delete",
		)
	}

	return s.repo.Delete(ctx, id)
}
