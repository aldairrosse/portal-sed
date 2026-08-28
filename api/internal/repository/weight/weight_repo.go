package weight

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/cycleconfig"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	"github.com/sed-evaluacion-desempeno/api/internal/teamweightconfig"
)

// Repo provides Ent-backed CRUD for hierarchical weights with fallback 100.
type Repo struct {
	client *internal.Client
	db     *sql.DB
}

// NewRepo creates a Repo.
func NewRepo(client *internal.Client, db *sql.DB) *Repo {
	return &Repo{client: client, db: db}
}

// CycleWeights holds G/P pair.
type CycleWeights struct {
	G float64 `json:"g_weight"`
	P float64 `json:"p_weight"`
}

// TeamWeights holds J/PJ pair.
type TeamWeights struct {
	J  float64 `json:"j_weight"`
	PJ float64 `json:"pj_weight"`
}

// GetOrFallbackCycleConfig returns CycleWeights for cycleID, fallback P=100 when no row.
func (r *Repo) GetOrFallbackCycleConfig(ctx context.Context, cycleID uuid.UUID) (CycleWeights, error) {
	cfg, err := r.client.CycleConfig.Query().
		Where(cycleconfig.CycleID(cycleID)).
		Only(ctx)
	if err != nil {
		if internal.IsNotFound(err) || err == sql.ErrNoRows {
			return CycleWeights{G: 0, P: 100}, nil
		}
		return CycleWeights{}, err
	}
	return CycleWeights{G: cfg.GWeight, P: cfg.PWeight}, nil
}

// UpsertCycleConfig validates 0-100 and sets P=100-g via Ent upsert.
func (r *Repo) UpsertCycleConfig(ctx context.Context, cycleID uuid.UUID, g float64) (CycleWeights, error) {
	if g < 0 || g > 100 {
		return CycleWeights{}, errors.NewDomainError(errors.InvalidRequest, "g_weight debe estar entre 0 y 100", nil)
	}
	p := 100 - g
	// Try update, else create
	existing, err := r.client.CycleConfig.Query().Where(cycleconfig.CycleID(cycleID)).Only(ctx)
	if err != nil && !internal.IsNotFound(err) {
		return CycleWeights{}, err
	}
	if existing != nil {
		updated, err := existing.Update().SetGWeight(g).SetPWeight(p).Save(ctx)
		if err != nil {
			return CycleWeights{}, err
		}
		return CycleWeights{G: updated.GWeight, P: updated.PWeight}, nil
	}
	created, err := r.client.CycleConfig.Create().
		SetCycleID(cycleID).
		SetGWeight(g).
		SetPWeight(p).
		Save(ctx)
	if err != nil {
		return CycleWeights{}, err
	}
	return CycleWeights{G: created.GWeight, P: created.PWeight}, nil
}

// GetOrFallbackTeamConfig returns TeamWeights for cycleID+teamID, fallback PJ=100 when no row.
func (r *Repo) GetOrFallbackTeamConfig(ctx context.Context, cycleID, teamID uuid.UUID) (TeamWeights, error) {
	cfg, err := r.client.TeamWeightConfig.Query().
		Where(teamweightconfig.CycleID(cycleID), teamweightconfig.TeamID(teamID)).
		Only(ctx)
	if err != nil {
		if internal.IsNotFound(err) || err == sql.ErrNoRows {
			return TeamWeights{J: 0, PJ: 100}, nil
		}
		return TeamWeights{}, err
	}
	return TeamWeights{J: cfg.JWeight, PJ: cfg.PjWeight}, nil
}

// UpsertTeamConfig validates 0-100 and sets PJ=100-j.
func (r *Repo) UpsertTeamConfig(ctx context.Context, cycleID, teamID uuid.UUID, j float64) (TeamWeights, error) {
	if j < 0 || j > 100 {
		return TeamWeights{}, errors.NewDomainError(errors.InvalidRequest, "j_weight debe estar entre 0 y 100", nil)
	}
	pj := 100 - j
	existing, err := r.client.TeamWeightConfig.Query().
		Where(teamweightconfig.CycleID(cycleID), teamweightconfig.TeamID(teamID)).
		Only(ctx)
	if err != nil && !internal.IsNotFound(err) {
		return TeamWeights{}, err
	}
	if existing != nil {
		updated, err := existing.Update().SetJWeight(j).SetPjWeight(pj).Save(ctx)
		if err != nil {
			return TeamWeights{}, err
		}
		return TeamWeights{J: updated.JWeight, PJ: updated.PjWeight}, nil
	}
	created, err := r.client.TeamWeightConfig.Create().
		SetCycleID(cycleID).
		SetTeamID(teamID).
		SetJWeight(j).
		SetPjWeight(pj).
		Save(ctx)
	if err != nil {
		return TeamWeights{}, err
	}
	return TeamWeights{J: created.JWeight, PJ: created.PjWeight}, nil
}
