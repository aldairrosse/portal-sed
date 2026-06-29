// Package activity provides the repository layer for ActivityLog entities.
// It uses Ent-generated queries for CRUD operations on the activity_logs table.
package activity

import (
	"context"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/activitylog"
)

// Repository defines the data access interface for activity logs.
type Repository interface {
	// Create inserts a new activity log entry and returns the created entity.
	Create(ctx context.Context, employeeID uuid.UUID, action, description, module string, metadata map[string]interface{}) (*internal.ActivityLog, error)

	// ListByEmployee returns the latest activity logs for an employee,
	// ordered by created_at DESC, limited to `limit` rows.
	ListByEmployee(ctx context.Context, employeeID uuid.UUID, limit int) ([]*internal.ActivityLog, error)
}

// repo implements Repository backed by the Ent client.
type repo struct {
	client *internal.Client
}

// NewRepository creates a new Repository.
func NewRepository(client *internal.Client) Repository {
	return &repo{client: client}
}

// Create inserts a new activity log entry.
func (r *repo) Create(ctx context.Context, employeeID uuid.UUID, action, description, module string, metadata map[string]interface{}) (*internal.ActivityLog, error) {
	return r.client.ActivityLog.Create().
		SetEmployeeID(employeeID).
		SetAction(action).
		SetDescription(description).
		SetModule(module).
		SetMetadata(metadata).
		Save(ctx)
}

// ListByEmployee returns the latest activity logs for an employee.
func (r *repo) ListByEmployee(ctx context.Context, employeeID uuid.UUID, limit int) ([]*internal.ActivityLog, error) {
	return r.client.ActivityLog.Query().
		Where(activitylog.EmployeeID(employeeID)).
		Order(internal.Desc(activitylog.FieldCreatedAt)).
		Limit(limit).
		All(ctx)
}
