// Package activity provides business logic for activity log management,
// including ownership verification: an employee can only see their own logs.
package activity

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/auth"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	repo "github.com/sed-evaluacion-desempeno/api/internal/repository/activity"
)

// ActivityResponse is the API response for a single activity log entry.
type ActivityResponse struct {
	ID          string                 `json:"id"`
	Action      string                 `json:"action"`
	Description string                 `json:"description"`
	Module      string                 `json:"module"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   string                 `json:"created_at"`
}

// Service defines the interface for activity log business operations.
type Service interface {
	// LogActivity persists an activity log entry.
	LogActivity(ctx context.Context, employeeID uuid.UUID, action, description, module string, metadata map[string]interface{}) error

	// ListByEmployee returns activity logs for an employee.
	// Verifies ownership: the caller's employeeID must match the target employeeID.
	ListByEmployee(ctx context.Context, employeeID uuid.UUID, limit int) ([]*ActivityResponse, error)
}

// service implements Service.
type service struct {
	repo repo.Repository
}

// NewService creates a new activity service.
func NewService(r repo.Repository) Service {
	return &service{repo: r}
}

// LogActivity persists an activity log entry.
func (s *service) LogActivity(ctx context.Context, employeeID uuid.UUID, action, description, module string, metadata map[string]interface{}) error {
	_, err := s.repo.Create(ctx, employeeID, action, description, module, metadata)
	return err
}

// ListByEmployee returns activity logs for an employee with ownership check.
func (s *service) ListByEmployee(ctx context.Context, employeeID uuid.UUID, limit int) ([]*ActivityResponse, error) {
	// Ownership check: the authenticated user must match the target employeeID.
	callerID, ok := auth.GetEmployeeID(ctx)
	if !ok {
		return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"no authenticated session", nil)
	}
	if callerID != employeeID {
		return nil, pkgerrors.ErrForbidden
	}

	rows, err := s.repo.ListByEmployee(ctx, employeeID, limit)
	if err != nil {
		return nil, err
	}

	responses := make([]*ActivityResponse, len(rows))
	for i, row := range rows {
		responses[i] = modelToResponse(row)
	}
	return responses, nil
}

// modelToResponse converts an Ent ActivityLog to the API response format.
func modelToResponse(m *internal.ActivityLog) *ActivityResponse {
	resp := &ActivityResponse{
		ID:          m.ID.String(),
		Action:      m.Action,
		Description: m.Description,
		Module:      m.Module,
		CreatedAt:   m.CreatedAt.Format(time.RFC3339),
	}
	if len(m.Metadata) > 0 {
		resp.Metadata = m.Metadata
	}
	return resp
}
