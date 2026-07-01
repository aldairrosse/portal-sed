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

// UpsertGoalState updates final_progress, self_assessment and rh_assessment for a goal.
// Only non-nil fields are touched. final_progress maps to final_rating (int 1-5 or NULL);
// self_assessment maps to final_comments. rh_assessment uses the new column.
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
		// Map progress 0.0-1.0 to final_rating 1-5; store NULL if zero.
		if *input.FinalProgress == 0 {
			setClauses = append(setClauses, "final_rating = NULL")
		} else {
			rating := int(*input.FinalProgress * 5)
			if rating < 1 {
				rating = 1
			}
			if rating > 5 {
				rating = 5
			}
			setClauses = append(setClauses, "final_rating = $"+strconv.Itoa(idx))
			args = append(args, rating)
			idx++
		}
	}
	if input.SelfAssessment != nil {
		setClauses = append(setClauses, "final_comments = $"+strconv.Itoa(idx))
		args = append(args, *input.SelfAssessment)
		idx++
	}
	if input.RhAssessment != nil {
		setClauses = append(setClauses, "rh_assessment = $"+strconv.Itoa(idx))
		args = append(args, *input.RhAssessment)
		idx++
	}

	args = append(args, evalID, input.GoalID)
	query := "UPDATE evaluation_goals SET " + strings.Join(setClauses, ", ") +
		" WHERE evaluation_id = $" + strconv.Itoa(idx) + " AND goal_id = $" + strconv.Itoa(idx+1)
	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return err
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
