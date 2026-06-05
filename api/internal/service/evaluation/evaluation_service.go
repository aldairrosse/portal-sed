// Package evaluation provides business logic for evaluation self-evaluation,
// RH evaluation, and finalization during the year-end "cierre" phase.
package evaluation

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	dto "github.com/sed-evaluacion-desempeno/api/internal/dto/evaluation"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/state"
	repo "github.com/sed-evaluacion-desempeno/api/internal/repository/evaluation"
)

// CyclePhaseChecker is an interface for checking the current phase of a cycle.
type CyclePhaseChecker interface {
	GetPhase(ctx context.Context, cycleID uuid.UUID) (string, error)
	GetSelfEvalDeadline(ctx context.Context, cycleID uuid.UUID) (*time.Time, error)
}

// IdempotencyCache is an interface for idempotency key storage.
type IdempotencyCache interface {
	Get(ctx context.Context, key string) (*IdempotencyCacheEntry, error)
	Set(ctx context.Context, key string, entry *IdempotencyCacheEntry, ttl time.Duration) error
}

// IdempotencyCacheEntry holds cached response data.
type IdempotencyCacheEntry struct {
	Body        []byte `json:"body"`
	PayloadHash string `json:"payload_hash"`
}

// EvaluationService orchestrates evaluation lifecycle operations.
type EvaluationService struct {
	evalRepo   EvaluationRepo
	compRepo   CompetencyRatingRepo
	goalRepo   GoalRatingRepo
	cycleCheck CyclePhaseChecker
	idemCache  IdempotencyCache
}

// NewEvaluationService creates a new EvaluationService.
func NewEvaluationService(
	evalRepo EvaluationRepo,
	compRepo CompetencyRatingRepo,
	goalRepo GoalRatingRepo,
	cycleCheck CyclePhaseChecker,
	idemCache IdempotencyCache,
) *EvaluationService {
	return &EvaluationService{
		evalRepo:   evalRepo,
		compRepo:   compRepo,
		goalRepo:   goalRepo,
		cycleCheck: cycleCheck,
		idemCache:  idemCache,
	}
}

// SubmitSelfEvaluation handles employee self-evaluation submission.
func (s *EvaluationService) SubmitSelfEvaluation(ctx context.Context, evaluationID uuid.UUID, req dto.SelfEvaluationRequest, idempotencyKey string) (*dto.EvaluationDetailResponse, error) {
	row, err := s.evalRepo.GetByID(ctx, evaluationID)
	if err != nil {
		return nil, err
	}

	if err := s.validatePhase(ctx, row.CycleID); err != nil {
		return nil, err
	}
	if err := s.checkSelfEvalDeadline(ctx, row.CycleID); err != nil {
		return nil, err
	}

	tx, err := s.evalRepo.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	lockedRow, err := s.evalRepo.LockEvalForUpdate(ctx, tx, evaluationID)
	if err != nil {
		return nil, err
	}
	if lockedRow.State == state.StateCompleted.String() {
		return nil, repo.ErrEvaluationFinalized
	}

	newState := lockedRow.State
	if lockedRow.State == state.StatePendingEvalFinal.String() {
		newState = state.StateInProgress.String()
	}

	comps := make([]repo.CompetencyUpsert, len(req.Competencies))
	for i, c := range req.Competencies {
		comps[i] = repo.CompetencyUpsert{
			CompetencyID: c.CompetencyID,
			Rating:       c.Rating,
			Comments:     c.Comments,
		}
	}
	goals := make([]repo.GoalCommentUpsert, len(req.GoalComments))
	for i, g := range req.GoalComments {
		goals[i] = repo.GoalCommentUpsert{
			GoalID:  g.GoalID,
			Comment: g.Comment,
		}
	}

	if err := s.evalRepo.SubmitEval(ctx, tx, evaluationID, comps, goals, newState, true, false); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	tx = nil

	if idempotencyKey != "" && s.idemCache != nil {
		_ = s.idemCache.Set(ctx, "idempotency:"+idempotencyKey, &IdempotencyCacheEntry{
			PayloadHash: hashSelfEvalPayload(req),
		}, 24*time.Hour)
	}

	return s.GetEvaluation(ctx, evaluationID)
}

// UpdateSelfEvaluation updates a previously submitted self-evaluation.
func (s *EvaluationService) UpdateSelfEvaluation(ctx context.Context, evaluationID uuid.UUID, req dto.SelfEvaluationRequest, ifMatch int) (*dto.EvaluationDetailResponse, error) {
	row, err := s.evalRepo.GetByID(ctx, evaluationID)
	if err != nil {
		return nil, err
	}
	if err := s.validatePhase(ctx, row.CycleID); err != nil {
		return nil, err
	}
	if row.State == state.StateCompleted.String() {
		return nil, repo.ErrEvaluationFinalized
	}

	tx, err := s.evalRepo.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	lockedRow, err := s.evalRepo.LockEvalForUpdate(ctx, tx, evaluationID)
	if err != nil {
		return nil, err
	}
	if lockedRow.Version != ifMatch {
		return nil, errors.ErrConcurrentUpdate.WithDetails(
			fmt.Sprintf("expected_version: %d", ifMatch),
			fmt.Sprintf("actual_version: %d", lockedRow.Version),
		)
	}

	comps := make([]repo.CompetencyUpsert, len(req.Competencies))
	for i, c := range req.Competencies {
		comps[i] = repo.CompetencyUpsert{
			CompetencyID: c.CompetencyID,
			Rating:       c.Rating,
			Comments:     c.Comments,
		}
	}
	goals := make([]repo.GoalCommentUpsert, len(req.GoalComments))
	for i, g := range req.GoalComments {
		goals[i] = repo.GoalCommentUpsert{
			GoalID:  g.GoalID,
			Comment: g.Comment,
		}
	}

	if err := s.evalRepo.SubmitEval(ctx, tx, evaluationID, comps, goals, lockedRow.State, true, false); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	tx = nil

	return s.GetEvaluation(ctx, evaluationID)
}

// SubmitRHEvaluation handles RH evaluation submission.
func (s *EvaluationService) SubmitRHEvaluation(ctx context.Context, evaluationID uuid.UUID, req dto.RHEvaluationRequest, idempotencyKey string) (*dto.EvaluationDetailResponse, error) {
	row, err := s.evalRepo.GetByID(ctx, evaluationID)
	if err != nil {
		return nil, err
	}
	if err := s.validatePhase(ctx, row.CycleID); err != nil {
		return nil, err
	}

	tx, err := s.evalRepo.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	lockedRow, err := s.evalRepo.LockEvalForUpdate(ctx, tx, evaluationID)
	if err != nil {
		return nil, err
	}
	if lockedRow.State == state.StateCompleted.String() {
		return nil, repo.ErrEvaluationFinalized
	}

	newState := lockedRow.State
	if lockedRow.State == state.StatePendingEvalFinal.String() {
		newState = state.StateInProgress.String()
	}

	comps := make([]repo.CompetencyUpsert, len(req.Competencies))
	for i, c := range req.Competencies {
		comps[i] = repo.CompetencyUpsert{
			CompetencyID: c.CompetencyID,
			Rating:       c.Rating,
			Comments:     c.Comments,
		}
	}

	if err := s.evalRepo.SubmitEval(ctx, tx, evaluationID, comps, nil, newState, false, true); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	tx = nil

	if idempotencyKey != "" && s.idemCache != nil {
		_ = s.idemCache.Set(ctx, "idempotency:"+idempotencyKey, &IdempotencyCacheEntry{
			PayloadHash: hashRHEvalPayload(req),
		}, 24*time.Hour)
	}

	return s.GetEvaluation(ctx, evaluationID)
}

// UpdateRHEvaluation updates a previously submitted RH evaluation.
func (s *EvaluationService) UpdateRHEvaluation(ctx context.Context, evaluationID uuid.UUID, req dto.RHEvaluationRequest, ifMatch int) (*dto.EvaluationDetailResponse, error) {
	row, err := s.evalRepo.GetByID(ctx, evaluationID)
	if err != nil {
		return nil, err
	}
	if err := s.validatePhase(ctx, row.CycleID); err != nil {
		return nil, err
	}
	if row.State == state.StateCompleted.String() {
		return nil, repo.ErrEvaluationFinalized
	}

	tx, err := s.evalRepo.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	lockedRow, err := s.evalRepo.LockEvalForUpdate(ctx, tx, evaluationID)
	if err != nil {
		return nil, err
	}
	if lockedRow.Version != ifMatch {
		return nil, errors.ErrConcurrentUpdate.WithDetails(
			fmt.Sprintf("expected_version: %d", ifMatch),
			fmt.Sprintf("actual_version: %d", lockedRow.Version),
		)
	}

	comps := make([]repo.CompetencyUpsert, len(req.Competencies))
	for i, c := range req.Competencies {
		comps[i] = repo.CompetencyUpsert{
			CompetencyID: c.CompetencyID,
			Rating:       c.Rating,
			Comments:     c.Comments,
		}
	}

	if err := s.evalRepo.SubmitEval(ctx, tx, evaluationID, comps, nil, lockedRow.State, false, true); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	tx = nil

	return s.GetEvaluation(ctx, evaluationID)
}

// FinalizeEvaluation performs the one-way finalization with advisory lock.
func (s *EvaluationService) FinalizeEvaluation(ctx context.Context, evaluationID uuid.UUID, req dto.FinalizeEvaluationRequest) (*dto.EvaluationDetailResponse, error) {
	// Advisory lock connection
	lockKey := "eval:finalize:" + evaluationID.String()
	lockConn, err := s.evalRepo.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	_, err = lockConn.ExecContext(ctx, `SELECT pg_advisory_lock(hashtext($1))`, lockKey)
	if err != nil {
		_ = lockConn.Rollback()
		return nil, fmt.Errorf("failed to acquire advisory lock: %w", err)
	}
	defer func() {
		_, _ = lockConn.ExecContext(ctx, `SELECT pg_advisory_unlock(hashtext($1))`, lockKey)
		_ = lockConn.Rollback()
	}()

	row, err := s.evalRepo.GetByID(ctx, evaluationID)
	if err != nil {
		return nil, err
	}
	if row.State == state.StateCompleted.String() {
		return nil, repo.ErrEvaluationFinalized
	}
	if err := s.validatePhase(ctx, row.CycleID); err != nil {
		return nil, err
	}
	if row.SelfEvaluationCompletedAt == nil {
		return nil, errors.NewDomainError(errors.InvalidTransition,
			"cannot finalize evaluation: self-evaluation has not been submitted", nil)
	}
	if row.RhEvaluationCompletedAt == nil {
		return nil, errors.NewDomainError(errors.InvalidTransition,
			"cannot finalize evaluation: RH evaluation has not been submitted", nil)
	}

	tx, err := s.evalRepo.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err := s.evalRepo.LockEvalForUpdate(ctx, tx, evaluationID); err != nil {
		return nil, err
	}

	if err := s.evalRepo.FinalizeEval(ctx, tx, evaluationID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	tx = nil

	_ = s.evalRepo.RefreshSummaryView(ctx)

	return s.GetEvaluation(ctx, evaluationID)
}

// GetEvaluation retrieves the full evaluation detail with competencies and goals.
func (s *EvaluationService) GetEvaluation(ctx context.Context, id uuid.UUID) (*dto.EvaluationDetailResponse, error) {
	row, comps, goals, err := s.evalRepo.GetDetail(ctx, id)
	if err != nil {
		return nil, err
	}

	resp := &dto.EvaluationDetailResponse{
		ID:                  row.ID,
		EmployeeID:          row.EmployeeID,
		CycleID:             row.CycleID,
		State:               row.State,
		SelfEvalCompletedAt: row.SelfEvaluationCompletedAt,
		RHEvalCompletedAt:   row.RhEvaluationCompletedAt,
		Version:             row.Version,
		CreatedAt:           row.CreatedAt,
		UpdatedAt:           row.UpdatedAt,
	}

	resp.CompetencyRatings = make([]dto.CompetencyRatingDTO, len(comps))
	for i, c := range comps {
		resp.CompetencyRatings[i] = dto.CompetencyRatingDTO{
			CompetencyID: c.CompetencyID,
			Rating:       c.Rating,
			Comments:     c.Comments,
		}
	}
	resp.GoalRatings = make([]dto.GoalRatingDTO, len(goals))
	for i, g := range goals {
		resp.GoalRatings[i] = dto.GoalRatingDTO{
			GoalID:        g.GoalID,
			FinalRating:   g.FinalRating,
			FinalComments: g.FinalComments,
		}
	}
	if resp.CompetencyRatings == nil {
		resp.CompetencyRatings = []dto.CompetencyRatingDTO{}
	}
	if resp.GoalRatings == nil {
		resp.GoalRatings = []dto.GoalRatingDTO{}
	}

	return resp, nil
}

// ListEvaluations returns a cursor-paginated list of evaluations for a cycle.
func (s *EvaluationService) ListEvaluations(ctx context.Context, cycleID uuid.UUID, stateFilter string, cursor string, limit int) (*dto.EvaluationListResponse, error) {
	rows, nextCursor, err := s.evalRepo.ListByCycle(ctx, cycleID, stateFilter, cursor, limit)
	if err != nil {
		return nil, err
	}
	items := make([]dto.EvaluationListItem, len(rows))
	for i, r := range rows {
		items[i] = dto.EvaluationListItem{
			ID: r.ID, EmployeeID: r.EmployeeID, CycleID: r.CycleID,
			State: r.State, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		}
	}
	return &dto.EvaluationListResponse{Data: items, NextCursor: nextCursor}, nil
}

func (s *EvaluationService) validatePhase(ctx context.Context, cycleID uuid.UUID) error {
	phase, err := s.cycleCheck.GetPhase(ctx, cycleID)
	if err != nil {
		return fmt.Errorf("failed to get cycle phase: %w", err)
	}
	return state.RequiresPhase(phase)
}

func (s *EvaluationService) checkSelfEvalDeadline(ctx context.Context, cycleID uuid.UUID) error {
	deadline, err := s.cycleCheck.GetSelfEvalDeadline(ctx, cycleID)
	if err != nil {
		return nil
	}
	if deadline != nil && time.Now().After(*deadline) {
		return repo.ErrSelfEvalDeadlinePassed
	}
	return nil
}

func hashSelfEvalPayload(req dto.SelfEvaluationRequest) string {
	h := sha256.New()
	for _, c := range req.Competencies {
		h.Write([]byte(c.CompetencyID.String()))
		h.Write([]byte{0})
		h.Write([]byte(fmt.Sprintf("%d", c.Rating)))
		h.Write([]byte{0})
		h.Write([]byte(c.Comments))
		h.Write([]byte{0})
	}
	for _, g := range req.GoalComments {
		h.Write([]byte(g.GoalID.String()))
		h.Write([]byte{0})
		h.Write([]byte(g.Comment))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

func hashRHEvalPayload(req dto.RHEvaluationRequest) string {
	h := sha256.New()
	for _, c := range req.Competencies {
		h.Write([]byte(c.CompetencyID.String()))
		h.Write([]byte{0})
		h.Write([]byte(fmt.Sprintf("%d", c.Rating)))
		h.Write([]byte{0})
		h.Write([]byte(c.Comments))
		h.Write([]byte{0})
	}
	h.Write([]byte(req.FinalComments))
	return hex.EncodeToString(h.Sum(nil))
}
