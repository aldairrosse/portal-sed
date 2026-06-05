package competency

import (
	"context"
	"fmt"
	"time"

	dto "github.com/sed-evaluacion-desempeno/api/internal/dto/competency"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	repo "github.com/sed-evaluacion-desempeno/api/internal/repository/competency"
)

// pillarService implements PillarService.
type pillarService struct {
	repo repo.PillarRepo
}

// NewPillarService creates a new PillarService.
func NewPillarService(repo repo.PillarRepo) PillarService {
	return &pillarService{repo: repo}
}

func (s *pillarService) List(ctx context.Context, opts ListOptions) (*ListResult[dto.PillarListItem], error) {
	limit := opts.Limit
	if limit < 1 {
		limit = 1
	}
	if limit > 100 {
		limit = 100
	}

	includeCompetencies := false
	for _, inc := range opts.Include {
		if inc == "competencies" {
			includeCompetencies = true
			break
		}
	}

	pillars, nextCursor, err := s.repo.List(ctx, opts.Cursor, limit, includeCompetencies)
	if err != nil {
		return nil, err
	}

	items := make([]dto.PillarListItem, len(pillars))
	for i, p := range pillars {
		items[i] = dto.PillarListItem{
			ID:          p.ID.String(),
			Name:        p.Name,
			Description: p.Description,
		}
		count := len(p.Edges.Competencies)
		if count == 0 {
			c, err := s.repo.CountCompetencies(ctx, p.ID.String())
			if err == nil {
				count = c
			}
		}
		items[i].CompetencyCount = count
	}

	hasMore := nextCursor != ""
	return &ListResult[dto.PillarListItem]{
		Data:       items,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (s *pillarService) Get(ctx context.Context, id string, includeCompetencies bool) (*dto.PillarDetail, error) {
	p, err := s.repo.Get(ctx, id, includeCompetencies)
	if err != nil {
		return nil, err
	}

	detail := &dto.PillarDetail{
		ID:          p.ID.String(),
		Name:        p.Name,
		Description: p.Description,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}

	if includeCompetencies && p.Edges.Competencies != nil {
		detail.Competencies = make([]dto.CompetencyLite, len(p.Edges.Competencies))
		for i, c := range p.Edges.Competencies {
			detail.Competencies[i] = dto.CompetencyLite{
				ID:          c.ID.String(),
				Name:        c.Name,
				Description: c.Description,
			}
		}
	}

	return detail, nil
}

func (s *pillarService) Create(ctx context.Context, req dto.CreatePillarRequest) (*dto.PillarDetail, error) {
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

	p, err := s.repo.Create(ctx, req.Name, req.Description)
	if err != nil {
		return nil, err
	}

	return &dto.PillarDetail{
		ID:          p.ID.String(),
		Name:        p.Name,
		Description: p.Description,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}, nil
}

func (s *pillarService) Update(ctx context.Context, id string, req dto.UpdatePillarRequest, ifMatch time.Time) (*dto.PillarDetail, error) {
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

	p, err := s.repo.Update(ctx, id, req.Name, req.Description, ifMatch)
	if err != nil {
		return nil, err
	}

	return &dto.PillarDetail{
		ID:          p.ID.String(),
		Name:        p.Name,
		Description: p.Description,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}, nil
}

func (s *pillarService) Delete(ctx context.Context, id string, force bool) error {
	count, err := s.repo.CountCompetencies(ctx, id)
	if err != nil {
		return err
	}

	if count > 0 && !force {
		return pkgerrors.NewDomainError("PILLAR_HAS_COMPETENCIES",
			"cannot delete pillar with existing competencies; use force=true to cascade",
			nil).WithDetails(
			fmt.Sprintf("pillar_id: %s", id),
			fmt.Sprintf("competencies_count: %d", count),
			"action: use force=true to cascade delete",
		)
	}

	return s.repo.Delete(ctx, id)
}
