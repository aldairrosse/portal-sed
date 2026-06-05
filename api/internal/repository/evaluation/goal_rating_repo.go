package evaluation

import (
	"context"
	"database/sql"
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
