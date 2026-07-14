package goal

import (
	"context"

	"github.com/google/uuid"
	dtogoal "github.com/sed-evaluacion-desempeno/api/internal/dto/goal"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	repogoal "github.com/sed-evaluacion-desempeno/api/internal/repository/goal"
)

// CategoryService handles business logic for goal categories.
type CategoryService struct {
	catRepo    CategoryRepository
	pillarRepo PillarRepository
	phaseCheck *PhaseCheck
}

// NewCategoryService creates a new CategoryService.
func NewCategoryService(catRepo CategoryRepository, pillarRepo PillarRepository, phaseCheck *PhaseCheck) *CategoryService {
	return &CategoryService{
		catRepo:    catRepo,
		pillarRepo: pillarRepo,
		phaseCheck: phaseCheck,
	}
}

// ListCategories retrieves all categories for an employee.
func (s *CategoryService) ListCategories(ctx context.Context, empID uuid.UUID) ([]*repogoal.CategoryRow, error) {
	return s.catRepo.ListCategoriesByEmployee(ctx, empID)
}

// CreateCategory creates a new category for the employee.
func (s *CategoryService) CreateCategory(ctx context.Context, empID uuid.UUID, req dtogoal.CreateCategoryRequest) (*repogoal.CategoryRow, error) {
	if err := s.phaseCheck.CanCreateCategory(ctx, empID.String()); err != nil {
		return nil, err
	}

	if req.Weight <= 0 || req.Weight > 100 {
		return nil, pkgerrors.ErrInvalidWeightRange
	}
	if req.Name == "" {
		return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "category name is required", nil)
	}

	pillarID, err := s.resolvePillarID(ctx, req.PillarID)
	if err != nil {
		return nil, err
	}

	return s.catRepo.CreateCategory(ctx, empID, req.Name, req.Description, req.Weight, pillarID)
}

// UpdateCategory updates a category.
func (s *CategoryService) UpdateCategory(ctx context.Context, empID, catID uuid.UUID, req dtogoal.UpdateCategoryRequest) (*repogoal.CategoryRow, error) {
	if err := s.phaseCheck.CanUpdateCategory(ctx, empID.String()); err != nil {
		return nil, err
	}

	if req.Weight <= 0 || req.Weight > 100 {
		return nil, pkgerrors.ErrInvalidWeightRange
	}
	if req.Name == "" {
		return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "category name is required", nil)
	}

	pillarID, err := s.resolvePillarID(ctx, req.PillarID)
	if err != nil {
		return nil, err
	}

	// In avance phase, only weight changes are allowed
	// Fetch existing to compare
	existing, err := s.catRepo.GetCategory(ctx, catID)
	if err != nil {
		return nil, err
	}
	if existing.EmployeeID != empID {
		return nil, pkgerrors.ErrCategoryNotFound
	}

	if existing.Name != req.Name || existing.Description != req.Description {
		if err := s.phaseCheck.Enforce(ctx, empID.String(), PhaseAsignacion); err != nil {
			return nil, err
		}
	}

	return s.catRepo.UpdateCategory(ctx, catID, req.Name, req.Description, req.Weight, empID, pillarID)
}

// DeleteCategory deletes a category.
func (s *CategoryService) DeleteCategory(ctx context.Context, empID, catID uuid.UUID) error {
	if err := s.phaseCheck.CanDeleteCategory(ctx, empID.String()); err != nil {
		return err
	}

	// Verify ownership
	cat, err := s.catRepo.GetCategory(ctx, catID)
	if err != nil {
		return err
	}
	if cat.EmployeeID != empID {
		return pkgerrors.ErrCategoryNotFound
	}

	return s.catRepo.DeleteCategory(ctx, catID)
}

// resolvePillarID validates that the given pillar ID exists and has type 'metas'.
// Returns nil when no pillar is provided (optional field).
func (s *CategoryService) resolvePillarID(ctx context.Context, pillarIDStr *string) (*uuid.UUID, error) {
	if pillarIDStr == nil || *pillarIDStr == "" {
		return nil, nil
	}
	pillarID, err := uuid.Parse(*pillarIDStr)
	if err != nil {
		return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid pillar_id", err)
	}
	pillar, err := s.pillarRepo.Get(ctx, pillarID.String(), false)
	if err != nil {
		return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "pillar not found", err)
	}
	if pillar.Type != "metas" {
		return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "pillar must be of type 'metas'", nil)
	}
	return &pillarID, nil
}
