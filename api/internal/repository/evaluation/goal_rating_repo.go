package evaluation

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/evaluationgoal"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
)

// GoalRatingRepo handles operations for EvaluationGoal comments.
type GoalRatingRepo struct {
	client *internal.Client
}

// NewGoalRatingRepo creates a new GoalRatingRepo.
func NewGoalRatingRepo(client *internal.Client) *GoalRatingRepo {
	return &GoalRatingRepo{client: client}
}

// UpdateComments updates the closing comments for goals linked to an evaluation.
func (r *GoalRatingRepo) UpdateComments(ctx context.Context, tx *sql.Tx, evalID uuid.UUID, goals []GoalCommentUpsert) error {
	now := time.Now()
	for _, g := range goals {
		_, err := tx.ExecContext(ctx,
			`UPDATE evaluation_goals SET final_comments = $1, updated_at = $2
			 WHERE evaluation_id = $3 AND goal_id = $4`,
			g.Comment, now, evalID, g.GoalID,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetByEvaluation retrieves all goal ratings for an evaluation.
func (r *GoalRatingRepo) GetByEvaluation(ctx context.Context, evalID uuid.UUID) ([]*internal.EvaluationGoal, error) {
	results, err := r.client.EvaluationGoal.Query().
		Where(evaluationgoal.EvaluationID(evalID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	if results == nil {
		return []*internal.EvaluationGoal{}, nil
	}
	return results, nil
}

// VerifyGoalsExist checks that all goal IDs are linked to the evaluation.
func (r *GoalRatingRepo) VerifyGoalsExist(ctx context.Context, evalID uuid.UUID, goalIDs []uuid.UUID) error {
	for _, gid := range goalIDs {
		count, err := r.client.EvaluationGoal.Query().
			Where(
				evaluationgoal.EvaluationID(evalID),
				evaluationgoal.GoalID(gid),
			).
			Count(ctx)
		if err != nil {
			return err
		}
		if count == 0 {
			return pkgerrors.NewDomainError(pkgerrors.GoalNotFound,
				"Goal is not linked to this evaluation.", nil,
			).WithDetails("goal_id: "+gid.String(), "evaluation_id: "+evalID.String())
		}
	}
	return nil
}

// SnapshotColumnForPhase maps a cycle/evaluation phase to its snapshot column.
// "avance"/"medio-anio"/"medio_anio" -> avance_progress;
// "cierre"/"fin-anio"/"fin_anio" -> cierre_progress.
// Unknown or empty phase defaults to avance_progress (mid-year is the safe
// default: avance writes never clobber the closing snapshot).
func SnapshotColumnForPhase(phase string) string {
	switch strings.ToLower(strings.TrimSpace(phase)) {
	case "avance", "medio-anio", "medio_anio":
		return "avance_progress"
	case "cierre", "fin-anio", "fin_anio":
		return "cierre_progress"
	default:
		return "avance_progress"
	}
}

// UpsertGoalState writes the DIRECT goal value (in the goal's own unit:
// porcentaje/moneda/numero/binario) into the active-phase snapshot column,
// syncs goals.current_value = cierre_progress ?? avance_progress, and leaves
// final_rating NULL unless an explicit rating is provided (no int(FP*5)).
// final_comments/rh_assessment are only overwritten with non-empty values;
// empty strings are treated as not-sent (F3: no destructive ” writes).
func (r *GoalRatingRepo) UpsertGoalState(ctx context.Context, tx *sql.Tx, evalID uuid.UUID, input GoalStateUpsert) error {
	now := time.Now()
	// Ensure the goal row exists for this evaluation (idempotent insert).
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO evaluation_goals (id, created_at, updated_at, evaluation_id, goal_id)
		 VALUES (gen_random_uuid(), $1, $1, $2, $3)
		 ON CONFLICT (evaluation_id, goal_id) DO NOTHING`,
		now, evalID, input.GoalID,
	); err != nil {
		return err
	}

	// Build dynamic SET clause: only include non-nil fields.
	setClauses := []string{"updated_at = $1"}
	args := []interface{}{now}
	idx := 2

	if input.FinalProgress != nil {
		// Direct value into the closing-phase snapshot (never int(FP*5)).
		col := SnapshotColumnForPhase(input.Phase)
		setClauses = append(setClauses, col+" = $"+strconv.Itoa(idx))
		args = append(args, *input.FinalProgress)
		idx++
	}
	if input.SelfAssessment != nil && *input.SelfAssessment != "" {
		setClauses = append(setClauses, "final_comments = $"+strconv.Itoa(idx))
		args = append(args, *input.SelfAssessment)
		idx++
	}
	if input.RhAssessment != nil && *input.RhAssessment != "" {
		setClauses = append(setClauses, "rh_assessment = $"+strconv.Itoa(idx))
		args = append(args, *input.RhAssessment)
		idx++
	}
	// F3: final_rating is only written when explicitly provided (non-nil).
	// Never NULL it implicitly: progress-only saves must preserve prior ratings/comments.
	if input.FinalRating != nil {
		setClauses = append(setClauses, "final_rating = $"+strconv.Itoa(idx))
		args = append(args, *input.FinalRating)
		idx++
	}

	args = append(args, evalID, input.GoalID)
	query := "UPDATE evaluation_goals SET " + strings.Join(setClauses, ", ") +
		" WHERE evaluation_id = $" + strconv.Itoa(idx) + " AND goal_id = $" + strconv.Itoa(idx+1)
	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return err
	}
	if input.FinalProgress != nil {
		// Sync: goals.current_value = cierre_progress ?? avance_progress.
		var avance, cierre sql.NullFloat64
		if err := tx.QueryRowContext(ctx,
			`SELECT avance_progress, cierre_progress FROM evaluation_goals WHERE evaluation_id = $1 AND goal_id = $2`,
			evalID, input.GoalID,
		).Scan(&avance, &cierre); err != nil {
			return err
		}
		current := *input.FinalProgress
		switch {
		case cierre.Valid:
			current = cierre.Float64
		case avance.Valid:
			current = avance.Float64
		}
		if _, err := tx.ExecContext(ctx,
			`UPDATE goals SET current_value = $1, updated_at = $2 WHERE id = $3`,
			current, now, input.GoalID,
		); err != nil {
			return err
		}
	}
	return nil
}

// UpsertGoalComment updates the manager_comment for a goal within an evaluation.
func (r *GoalRatingRepo) UpsertGoalComment(ctx context.Context, tx *sql.Tx, evalID uuid.UUID, input GoalCommentUpsert) error {
	now := time.Now()
	// Ensure the goal row exists (idempotent insert).
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO evaluation_goals (id, created_at, updated_at, evaluation_id, goal_id)
		 VALUES (gen_random_uuid(), $1, $1, $2, $3)
		 ON CONFLICT (evaluation_id, goal_id) DO NOTHING`,
		now, evalID, input.GoalID,
	); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE evaluation_goals SET manager_comment = $1, updated_at = $2
		 WHERE evaluation_id = $3 AND goal_id = $4`,
		input.Comment, now, evalID, input.GoalID,
	); err != nil {
		return err
	}
	return nil
}
