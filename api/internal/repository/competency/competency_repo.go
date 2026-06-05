package competency

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/competency"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
)

// competencyRepo implements CompetencyRepo.
type competencyRepo struct {
	client *internal.Client
}

// NewCompetencyRepo creates a new CompetencyRepo.
func NewCompetencyRepo(client *internal.Client) CompetencyRepo {
	return &competencyRepo{client: client}
}

func (r *competencyRepo) WithTx(ctx context.Context, fn TxFunc) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if v := recover(); v != nil {
			_ = tx.Rollback()
			panic(v)
		}
	}()
	if err := fn(tx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			err = pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
				"transaction rollback failed", rerr)
		}
		return err
	}
	return tx.Commit()
}

func (r *competencyRepo) ListByPillar(ctx context.Context, pillarID string, cursor string, limit int) ([]*internal.Competency, string, error) {
	q := r.client.Competency.Query().
		Where(competency.PillarIDEQ(uuid.MustParse(pillarID))).
		Order(internal.Asc(competency.FieldName)).
		Limit(limit + 1)

	if cursor != "" {
		nc, err := decodeNameCursor(cursor)
		if err != nil {
			return nil, "", err
		}
		q = q.Where(competency.NameGT(nc.Name))
	}

	results, err := q.All(ctx)
	if err != nil {
		return nil, "", err
	}

	var nextCursor string
	hasMore := len(results) > limit
	if hasMore {
		results = results[:limit]
		last := results[len(results)-1]
		nextCursor = encodeNameCursor(last.Name)
	}

	return results, nextCursor, nil
}

func (r *competencyRepo) Get(ctx context.Context, id string) (*internal.Competency, error) {
	c, err := r.client.Competency.Query().
		Where(competency.IDEQ(uuid.MustParse(id))).
		WithScaleCriteria().
		Only(ctx)
	if err != nil {
		if internal.IsNotFound(err) {
			return nil, pkgerrors.NewDomainError("COMPETENCY_NOT_FOUND",
				"competency not found", err)
		}
		return nil, err
	}
	return c, nil
}

func (r *competencyRepo) Create(ctx context.Context, pillarID, name, description string) (*internal.Competency, error) {
	c, err := r.client.Competency.Create().
		SetPillarID(uuid.MustParse(pillarID)).
		SetName(name).
		SetNillableDescription(strPtr(description)).
		Save(ctx)
	if err != nil {
		if internal.IsConstraintError(err) {
			return nil, pkgerrors.NewDomainError("DUPLICATE_NAME",
				"a competency with this name already exists in the pillar", err)
		}
		return nil, err
	}
	return c, nil
}

func (r *competencyRepo) Update(ctx context.Context, id string, name, description, pillarID string, ifMatch time.Time) (*internal.Competency, error) {
	// Fetch current for optimistic lock check
	current, err := r.client.Competency.Query().
		Where(competency.IDEQ(uuid.MustParse(id))).
		Only(ctx)
	if err != nil {
		if internal.IsNotFound(err) {
			return nil, pkgerrors.NewDomainError("COMPETENCY_NOT_FOUND",
				"competency not found", err)
		}
		return nil, err
	}

	// Optimistic lock check
	if !current.UpdatedAt.Truncate(time.Microsecond).Equal(ifMatch.Truncate(time.Microsecond)) {
		return nil, pkgerrors.NewDomainError(pkgerrors.ConcurrentUpdate,
			"the competency was modified by another request; retry with the latest version",
			nil)
	}

	updater := r.client.Competency.UpdateOneID(uuid.MustParse(id)).
		SetName(name).
		SetNillableDescription(strPtr(description))

	if pillarID != "" {
		updater = updater.SetPillarID(uuid.MustParse(pillarID))
	}

	updated, err := updater.Save(ctx)
	if err != nil {
		if internal.IsConstraintError(err) {
			return nil, pkgerrors.NewDomainError("DUPLICATE_NAME",
				"a competency with this name already exists", err)
		}
		return nil, err
	}

	// Re-fetch with eager loading for the response
	updated, err = r.client.Competency.Query().
		Where(competency.IDEQ(uuid.MustParse(id))).
		WithScaleCriteria().
		Only(ctx)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (r *competencyRepo) Delete(ctx context.Context, id string) error {
	_, err := r.client.Competency.Query().
		Where(competency.IDEQ(uuid.MustParse(id))).
		Only(ctx)
	if err != nil {
		if internal.IsNotFound(err) {
			return pkgerrors.NewDomainError("COMPETENCY_NOT_FOUND",
				"competency not found", err)
		}
		return err
	}

	if err := r.client.Competency.DeleteOneID(uuid.MustParse(id)).Exec(ctx); err != nil {
		return err
	}
	return nil
}

func (r *competencyRepo) CountScaleCriteria(ctx context.Context, competencyID string) (int, error) {
	return r.client.Competency.Query().
		Where(competency.IDEQ(uuid.MustParse(competencyID))).
		QueryScaleCriteria().Count(ctx)
}
