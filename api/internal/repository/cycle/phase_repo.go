package cycle

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/phasedefinition"
	"github.com/sed-evaluacion-desempeno/api/internal/phasetransition"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
)

// PhaseDefinitionRow represents a phase definition from the database.
type PhaseDefinitionRow struct {
	ID             uuid.UUID `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Phase          string    `json:"phase"`
	Label          string    `json:"label"`
	Order          int       `json:"order"`
	AllowedActors  []string  `json:"allowed_actors"`
	AllowedActions []string  `json:"allowed_actions"`
	BlockedActions []string  `json:"blocked_actions"`
	CycleID        uuid.UUID `json:"cycle_id"`
}

// ToPhaseDefinition converts to the generated PhaseDefinition model.
func (r *PhaseDefinitionRow) ToPhaseDefinition() *internal.PhaseDefinition {
	return &internal.PhaseDefinition{
		ID:             r.ID,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
		Phase:          phasedefinition.Phase(r.Phase),
		Label:          r.Label,
		Order:          r.Order,
		AllowedActors:  r.AllowedActors,
		AllowedActions: r.AllowedActions,
		BlockedActions: r.BlockedActions,
		CycleID:        r.CycleID,
	}
}

// PhaseTransitionRow represents a phase transition rule from the database.
type PhaseTransitionRow struct {
	ID          uuid.UUID              `json:"id"`
	FromPhase   string                 `json:"from_phase"`
	ToPhase     string                 `json:"to_phase"`
	Trigger     string                 `json:"trigger"`
	Conditions  map[string]interface{} `json:"conditions,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	CycleID     uuid.UUID              `json:"cycle_id"`
	FromPhaseID uuid.UUID              `json:"from_phase_id"`
	ToPhaseID   uuid.UUID              `json:"to_phase_id"`
}

// ToPhaseTransition converts to the generated PhaseTransition model.
func (r *PhaseTransitionRow) ToPhaseTransition() *internal.PhaseTransition {
	return &internal.PhaseTransition{
		ID:          r.ID,
		FromPhase:   phasetransition.FromPhase(r.FromPhase),
		ToPhase:     phasetransition.ToPhase(r.ToPhase),
		Trigger:     phasetransition.Trigger(r.Trigger),
		Conditions:  r.Conditions,
		CreatedAt:   r.CreatedAt,
		CycleID:     r.CycleID,
		FromPhaseID: r.FromPhaseID,
		ToPhaseID:   r.ToPhaseID,
	}
}

// PhaseRepo provides read-only repository operations for PhaseDefinition
// and PhaseTransition. These are static/catalog data.
type PhaseRepo struct {
	client *internal.Client
	db     *sql.DB
}

// NewPhaseRepo creates a new PhaseRepo.
func NewPhaseRepo(client *internal.Client, db *sql.DB) *PhaseRepo {
	return &PhaseRepo{client: client, db: db}
}

// clientFor returns the appropriate client based on context db role hint.
func (r *PhaseRepo) clientFor(ctx context.Context) *internal.Client {
	return r.client
}

// GetPhaseDefinitions returns all phase definitions ordered by "order" ascending.
func (r *PhaseRepo) GetPhaseDefinitions(ctx context.Context) ([]*PhaseDefinitionRow, error) {
	return r.queryPhaseDefinitions(ctx,
		`SELECT id, created_at, updated_at, phase, label, "order", allowed_actors, allowed_actions, blocked_actions, cycle_id
		 FROM phase_definitions ORDER BY "order" ASC`)
}

// GetPhaseDefinitionByPhase returns a single phase definition by its phase enum value.
func (r *PhaseRepo) GetPhaseDefinitionByPhase(ctx context.Context, phase string) (*PhaseDefinitionRow, error) {
	results, err := r.queryPhaseDefinitions(ctx,
		`SELECT id, created_at, updated_at, phase, label, "order", allowed_actors, allowed_actions, blocked_actions, cycle_id
		 FROM phase_definitions WHERE phase = $1 ORDER BY "order" ASC LIMIT 1`, phase)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, sql.ErrNoRows
	}
	return results[0], nil
}

// GetTransitionsByFromPhase returns transitions where from_phase matches, ordered by to_phase.
func (r *PhaseRepo) GetTransitionsByFromPhase(ctx context.Context, fromPhase string) ([]*PhaseTransitionRow, error) {
	return r.queryPhaseTransitions(ctx,
		`SELECT id, from_phase, to_phase, trigger, conditions, created_at, cycle_id, from_phase_id, to_phase_id
		 FROM phase_transitions WHERE from_phase = $1 ORDER BY to_phase ASC`, fromPhase)
}

// ValidateTransition checks if a transition exists for the given (fromPhase, toPhase, trigger) tuple.
// Returns nil if valid, INVALID_TRANSITION error otherwise.
func (r *PhaseRepo) ValidateTransition(ctx context.Context, fromPhase, toPhase, trigger string) error {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(1) FROM phase_transitions WHERE from_phase = $1 AND to_phase = $2 AND trigger = $3`,
		fromPhase, toPhase, trigger,
	).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.ErrInvalidTransition.WithDetails(
			"from_phase: " + fromPhase,
			"to_phase: " + toPhase,
			"trigger: " + trigger,
		)
	}
	return nil
}

// queryPhaseDefinitions runs a raw SQL query and scans into PhaseDefinitionRow.
func (r *PhaseRepo) queryPhaseDefinitions(ctx context.Context, query string, args ...interface{}) ([]*PhaseDefinitionRow, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*PhaseDefinitionRow
	for rows.Next() {
		row := &PhaseDefinitionRow{}
		var actors, actions, blocked []byte
		err := rows.Scan(&row.ID, &row.CreatedAt, &row.UpdatedAt, &row.Phase, &row.Label,
			&row.Order, &actors, &actions, &blocked, &row.CycleID)
		if err != nil {
			return nil, err
		}
		if len(actors) > 0 {
			if err := json.Unmarshal(actors, &row.AllowedActors); err != nil {
				return nil, err
			}
		}
		if len(actions) > 0 {
			if err := json.Unmarshal(actions, &row.AllowedActions); err != nil {
				return nil, err
			}
		}
		if len(blocked) > 0 {
			if err := json.Unmarshal(blocked, &row.BlockedActions); err != nil {
				return nil, err
			}
		}
		results = append(results, row)
	}
	return results, rows.Err()
}

// queryPhaseTransitions runs a raw SQL query and scans into PhaseTransitionRow.
func (r *PhaseRepo) queryPhaseTransitions(ctx context.Context, query string, args ...interface{}) ([]*PhaseTransitionRow, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*PhaseTransitionRow
	for rows.Next() {
		row := &PhaseTransitionRow{}
		err := rows.Scan(&row.ID, &row.FromPhase, &row.ToPhase, &row.Trigger,
			&row.Conditions, &row.CreatedAt, &row.CycleID, &row.FromPhaseID, &row.ToPhaseID)
		if err != nil {
			return nil, err
		}
		results = append(results, row)
	}
	return results, rows.Err()
}
