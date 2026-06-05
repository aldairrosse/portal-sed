package goal

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/goalkpilink"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
)

// LinkKpiRepo provides Ent-backed operations for GoalKpiLink.
type LinkKpiRepo struct {
	client *internal.Client
	db     *sql.DB
}

// NewLinkKpiRepo creates a new LinkKpiRepo.
func NewLinkKpiRepo(client *internal.Client, db *sql.DB) *LinkKpiRepo {
	return &LinkKpiRepo{client: client, db: db}
}

// LinkKPI creates a GoalKpiLink. Idempotent — duplicate calls return no error.
// Uses raw SQL because Ent's create builder does not expose OnConflict().
func (r *LinkKpiRepo) LinkKPI(ctx context.Context, goalID, kpiID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO goal_kpi_links (goal_id, kpi_id, created_at)
		 VALUES ($1, $2, NOW())
		 ON CONFLICT (goal_id, kpi_id) DO NOTHING`,
		goalID, kpiID,
	)
	return err
}

// UnlinkKPI removes a specific GoalKpiLink.
func (r *LinkKpiRepo) UnlinkKPI(ctx context.Context, goalID, kpiID uuid.UUID) error {
	n, err := r.client.GoalKpiLink.Delete().
		Where(goalkpilink.GoalID(goalID), goalkpilink.KpiID(kpiID)).
		Exec(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return pkgerrors.ErrKpiNotFound
	}
	return nil
}

// CountGoalKPILinks counts the number of KPIs linked to a goal.
func (r *LinkKpiRepo) CountGoalKPILinks(ctx context.Context, goalID uuid.UUID) (int, error) {
	return r.client.GoalKpiLink.Query().
		Where(goalkpilink.GoalID(goalID)).
		Count(ctx)
}

// ReplaceGoalKpiLinks atomically replaces all KPI links for a goal.
// It deletes existing links and inserts new ones. Validates max 5 KPIs.
func (r *LinkKpiRepo) ReplaceGoalKpiLinks(ctx context.Context, goalID uuid.UUID, kpiIDs []uuid.UUID) error {
	if len(kpiIDs) > 5 {
		return pkgerrors.ErrKpiLinkLimitExceeded
	}

	// Delete existing links
	_, err := r.client.GoalKpiLink.Delete().
		Where(goalkpilink.GoalID(goalID)).
		Exec(ctx)
	if err != nil {
		return err
	}

	// Re-insert new links
	for _, kpiID := range kpiIDs {
		if err := r.LinkKPI(ctx, goalID, kpiID); err != nil {
			return err
		}
	}
	return nil
}

// ListKpiIDsByGoal returns the KPI IDs linked to a goal.
func (r *LinkKpiRepo) ListKpiIDsByGoal(ctx context.Context, goalID uuid.UUID) ([]uuid.UUID, error) {
	links, err := r.client.GoalKpiLink.Query().
		Where(goalkpilink.GoalID(goalID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, len(links))
	for i, l := range links {
		ids[i] = l.KpiID
	}
	return ids, nil
}
