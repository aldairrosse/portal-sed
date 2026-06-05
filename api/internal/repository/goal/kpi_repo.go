package goal

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/kpi"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
)

// KpiRow is the full representation of a KPI.
type KpiRow struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Name        string    `json:"name"`
	Unit        string    `json:"unit"`
	Description string    `json:"description"`
}

// kpiToRow converts an ent KPI to a KpiRow.
func kpiToRow(k *internal.KPI) *KpiRow {
	if k == nil {
		return nil
	}
	return &KpiRow{
		ID:          k.ID,
		CreatedAt:   k.CreatedAt,
		UpdatedAt:   k.UpdatedAt,
		Name:        k.Name,
		Unit:        string(k.Unit),
		Description: k.Description,
	}
}

// KpiRepo provides Ent-backed CRUD operations for KPIs.
type KpiRepo struct {
	client *internal.Client
	db     *sql.DB
}

// NewKpiRepo creates a new KpiRepo.
func NewKpiRepo(client *internal.Client, db *sql.DB) *KpiRepo {
	return &KpiRepo{client: client, db: db}
}

// ListKPIs returns all KPIs with optional cursor-based pagination.
func (r *KpiRepo) ListKPIs(ctx context.Context) ([]*KpiRow, error) {
	kpis, err := r.client.KPI.Query().
		Order(internal.Asc(kpi.FieldName)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	rows := make([]*KpiRow, len(kpis))
	for i, k := range kpis {
		rows[i] = kpiToRow(k)
	}
	return rows, nil
}

// GetKPI retrieves a single KPI by ID.
func (r *KpiRepo) GetKPI(ctx context.Context, kpiID uuid.UUID) (*KpiRow, error) {
	k, err := r.client.KPI.Query().
		Where(kpi.ID(kpiID)).
		Only(ctx)
	if err != nil {
		if internal.IsNotFound(err) {
			return nil, pkgerrors.ErrKpiNotFound
		}
		return nil, err
	}
	return kpiToRow(k), nil
}

// CreateKPI inserts a new KPI.
func (r *KpiRepo) CreateKPI(ctx context.Context, name, unit, description string) (*KpiRow, error) {
	k, err := r.client.KPI.Create().
		SetName(name).
		SetUnit(kpi.Unit(unit)).
		SetDescription(description).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return kpiToRow(k), nil
}

// UpdateKPI updates an existing KPI.
func (r *KpiRepo) UpdateKPI(ctx context.Context, kpiID uuid.UUID, name, unit, description string) (*KpiRow, error) {
	k, err := r.client.KPI.UpdateOneID(kpiID).
		SetName(name).
		SetUnit(kpi.Unit(unit)).
		SetDescription(description).
		Save(ctx)
	if err != nil {
		if internal.IsNotFound(err) {
			return nil, pkgerrors.ErrKpiNotFound
		}
		return nil, err
	}
	return kpiToRow(k), nil
}

// DeleteKPI removes a KPI by ID. Returns ErrKpiLinkedCannotDelete if linked to goals.
func (r *KpiRepo) DeleteKPI(ctx context.Context, kpiID uuid.UUID) error {
	// Check for existing links first
	count, err := r.CountGoalLinksByKPI(ctx, kpiID)
	if err != nil {
		return err
	}
	if count > 0 {
		return pkgerrors.ErrKpiLinkedCannotDelete
	}

	err = r.client.KPI.DeleteOneID(kpiID).Exec(ctx)
	if err != nil {
		if internal.IsNotFound(err) {
			return pkgerrors.ErrKpiNotFound
		}
		return err
	}
	return nil
}

// CountGoalLinksByKPI returns the number of goals linked to a given KPI.
func (r *KpiRepo) CountGoalLinksByKPI(ctx context.Context, kpiID uuid.UUID) (int, error) {
	return r.client.KPI.Query().
		Where(kpi.ID(kpiID)).
		QueryGoalLinks().
		Count(ctx)
}
