package competency

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/pillar"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
)

// nameCursor is a simple cursor for name-based pagination.
type nameCursor struct {
	Name string `json:"n"`
}

// encodeNameCursor encodes a name cursor to base64 JSON.
func encodeNameCursor(name string) string {
	nc := nameCursor{Name: name}
	raw, _ := json.Marshal(nc)
	return base64.URLEncoding.EncodeToString(raw)
}

// decodeNameCursor decodes a name cursor from base64 JSON.
func decodeNameCursor(encoded string) (*nameCursor, error) {
	raw, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"cursor: invalid base64 encoding", err)
	}
	var nc nameCursor
	if err := json.Unmarshal(raw, &nc); err != nil {
		return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest,
			"cursor: invalid JSON in cursor payload", err)
	}
	return &nc, nil
}

// pillarRepo implements PillarRepo.
type pillarRepo struct {
	client *internal.Client
}

// NewPillarRepo creates a new PillarRepo.
func NewPillarRepo(client *internal.Client) PillarRepo {
	return &pillarRepo{client: client}
}

func (r *pillarRepo) WithTx(ctx context.Context, fn TxFunc) error {
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

func (r *pillarRepo) List(ctx context.Context, cursor string, limit int, includeCompetencies bool) ([]*internal.Pillar, string, error) {
	q := r.client.Pillar.Query().
		Order(internal.Asc(pillar.FieldName)).
		Limit(limit + 1)

	if includeCompetencies {
		q = q.WithCompetencies(func(cq *internal.CompetencyQuery) {
			cq.Order(internal.Asc("name"))
		})
	}

	if cursor != "" {
		nc, err := decodeNameCursor(cursor)
		if err != nil {
			return nil, "", err
		}
		q = q.Where(pillar.NameGT(nc.Name))
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

func (r *pillarRepo) Get(ctx context.Context, id string, includeCompetencies bool) (*internal.Pillar, error) {
	q := r.client.Pillar.Query().Where(pillar.IDEQ(uuid.MustParse(id)))
	if includeCompetencies {
		q = q.WithCompetencies(func(cq *internal.CompetencyQuery) {
			cq.Order(internal.Asc("name"))
		})
	}
	p, err := q.Only(ctx)
	if err != nil {
		if internal.IsNotFound(err) {
			return nil, pkgerrors.NewDomainError("PILLAR_NOT_FOUND",
				"pillar not found", err)
		}
		return nil, err
	}
	return p, nil
}

func (r *pillarRepo) Create(ctx context.Context, name, description string) (*internal.Pillar, error) {
	p, err := r.client.Pillar.Create().
		SetName(name).
		SetNillableDescription(strPtr(description)).
		Save(ctx)
	if err != nil {
		if internal.IsConstraintError(err) {
			return nil, pkgerrors.NewDomainError("DUPLICATE_NAME",
				"a pillar with this name already exists", err)
		}
		return nil, err
	}
	return p, nil
}

func (r *pillarRepo) Update(ctx context.Context, id string, name, description string, ifMatch time.Time) (*internal.Pillar, error) {
	current, err := r.client.Pillar.Query().
		Where(pillar.IDEQ(uuid.MustParse(id))).
		Only(ctx)
	if err != nil {
		if internal.IsNotFound(err) {
			return nil, pkgerrors.NewDomainError("PILLAR_NOT_FOUND",
				"pillar not found", err)
		}
		return nil, err
	}

	// Optimistic lock check
	if !current.UpdatedAt.Truncate(time.Microsecond).Equal(ifMatch.Truncate(time.Microsecond)) {
		return nil, pkgerrors.NewDomainError(pkgerrors.ConcurrentUpdate,
			"the pillar was modified by another request; retry with the latest version",
			nil)
	}

	updated, err := r.client.Pillar.UpdateOneID(uuid.MustParse(id)).
		SetName(name).
		SetNillableDescription(strPtr(description)).
		Save(ctx)
	if err != nil {
		if internal.IsConstraintError(err) {
			return nil, pkgerrors.NewDomainError("DUPLICATE_NAME",
				"a pillar with this name already exists", err)
		}
		return nil, err
	}
	return updated, nil
}

func (r *pillarRepo) Delete(ctx context.Context, id string) error {
	_, err := r.client.Pillar.Query().
		Where(pillar.IDEQ(uuid.MustParse(id))).
		Only(ctx)
	if err != nil {
		if internal.IsNotFound(err) {
			return pkgerrors.NewDomainError("PILLAR_NOT_FOUND",
				"pillar not found", err)
		}
		return err
	}

	if err := r.client.Pillar.DeleteOneID(uuid.MustParse(id)).Exec(ctx); err != nil {
		return err
	}
	return nil
}

func (r *pillarRepo) CountCompetencies(ctx context.Context, pillarID string) (int, error) {
	return r.client.Pillar.Query().Where(pillar.IDEQ(uuid.MustParse(pillarID))).
		QueryCompetencies().Count(ctx)
}

// strPtr returns a pointer to s, or nil if s is empty.
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
