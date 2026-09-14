// Package evaluation provides business logic for evaluation self-evaluation,
// RH evaluation, and finalization during the year-end "cierre" phase.
package evaluation

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal/auth"
	dto "github.com/sed-evaluacion-desempeno/api/internal/dto/evaluation"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	"github.com/sed-evaluacion-desempeno/api/internal/pkg/state"
	repo "github.com/sed-evaluacion-desempeno/api/internal/repository/evaluation"
	orgrepo "github.com/sed-evaluacion-desempeno/api/internal/repository/org"
)

// CyclePhaseChecker is an interface for checking the current phase of a cycle.
type CyclePhaseChecker interface {
	GetPhase(ctx context.Context, cycleID uuid.UUID) (string, error)
	GetSelfEvalDeadline(ctx context.Context, cycleID uuid.UUID) (*time.Time, error)
	GetActiveCycleID(ctx context.Context, orgID uuid.UUID) (uuid.UUID, error)
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
	empRepo    *orgrepo.EmployeeRepo
	nodeRepo   *orgrepo.OrgNodeRepo
}

// NewEvaluationService creates a new EvaluationService.
// empRepo and nodeRepo are required for scope=team resolution and may be nil otherwise.
func NewEvaluationService(
	evalRepo EvaluationRepo,
	compRepo CompetencyRatingRepo,
	goalRepo GoalRatingRepo,
	cycleCheck CyclePhaseChecker,
	idemCache IdempotencyCache,
	empRepo *orgrepo.EmployeeRepo,
	nodeRepo *orgrepo.OrgNodeRepo,
) *EvaluationService {
	return &EvaluationService{
		evalRepo:   evalRepo,
		compRepo:   compRepo,
		goalRepo:   goalRepo,
		cycleCheck: cycleCheck,
		idemCache:  idemCache,
		empRepo:    empRepo,
		nodeRepo:   nodeRepo,
	}
}

// resolveCompetencyProfileID returns the employee's evaluation profile
// (employees.profile_id, FK target of evaluation_competencies.profile_id).
// A missing employee, missing repo, or uuid.Nil profile yields a clear
// domain error instead of an FK-violating insert.
func (s *EvaluationService) resolveCompetencyProfileID(ctx context.Context, employeeID uuid.UUID) (uuid.UUID, error) {
	if s.empRepo == nil {
		return uuid.Nil, repo.ErrEvaluationProfileMissing
	}
	emp, err := s.empRepo.GetByID(ctx, employeeID)
	if err != nil {
		return uuid.Nil, err
	}
	if emp.ProfileID == uuid.Nil {
		return uuid.Nil, repo.ErrEvaluationProfileMissing
	}
	return emp.ProfileID, nil
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

	profileID, err := s.resolveCompetencyProfileID(ctx, lockedRow.EmployeeID)
	if err != nil {
		return nil, err
	}

	if err := s.evalRepo.SubmitEval(ctx, tx, evaluationID, profileID, comps, goals, newState, true, false); err != nil {
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
	currentPhase, err := s.cycleCheck.GetPhase(ctx, row.CycleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cycle phase: %w", err)
	}
	if err := validateRowPhase(row.Phase, currentPhase); err != nil {
		return nil, err
	}
	if err := s.validateWritePhase(ctx, row.CycleID, ""); err != nil {
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

	profileID, err := s.resolveCompetencyProfileID(ctx, lockedRow.EmployeeID)
	if err != nil {
		return nil, err
	}

	if err := s.evalRepo.SubmitEval(ctx, tx, evaluationID, profileID, comps, goals, lockedRow.State, true, false); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	tx = nil

	return s.GetEvaluation(ctx, evaluationID)
}

// buildRHCompetencyUpserts maps request competencies to repo upserts, splitting
// the write path by caller role: RH (PermEvalRH) writes comments,
// jefe (allowed via AuthorizeRHEvaluationWrite, no PermEvalRH) writes
// manager_comment without touching comments. Rating is always shared.
// Missing role defaults to the RH path to preserve existing behavior.
func buildRHCompetencyUpserts(ctx context.Context, in []dto.CompetencyRatingInput) []repo.CompetencyUpsert {
	isManager := isManagerCompetencyWrite(ctx)
	comps := make([]repo.CompetencyUpsert, len(in))
	for i, c := range in {
		comps[i] = repo.CompetencyUpsert{
			CompetencyID: c.CompetencyID,
			Rating:       c.Rating,
			IsManager:    isManager,
		}
		if isManager {
			comps[i].ManagerComment = c.ManagerComment
		} else {
			comps[i].Comments = c.Comments
		}
	}
	return comps
}

// isManagerCompetencyWrite reports whether the caller writes competencies via
// the jefe path: any authenticated role without PermEvalRH.
func isManagerCompetencyWrite(ctx context.Context) bool {
	role, ok := auth.GetRole(ctx)
	if !ok {
		return false
	}
	return !auth.HasPermission(role, auth.PermEvalRH)
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

	comps := buildRHCompetencyUpserts(ctx, req.Competencies)

	profileID, err := s.resolveCompetencyProfileID(ctx, lockedRow.EmployeeID)
	if err != nil {
		return nil, err
	}

	if err := s.evalRepo.SubmitEval(ctx, tx, evaluationID, profileID, comps, nil, newState, false, true); err != nil {
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
	currentPhase, err := s.cycleCheck.GetPhase(ctx, row.CycleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cycle phase: %w", err)
	}
	if err := validateRowPhase(row.Phase, currentPhase); err != nil {
		return nil, err
	}
	if err := s.validateWritePhase(ctx, row.CycleID, ""); err != nil {
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

	comps := buildRHCompetencyUpserts(ctx, req.Competencies)

	profileID, err := s.resolveCompetencyProfileID(ctx, lockedRow.EmployeeID)
	if err != nil {
		return nil, err
	}

	if err := s.evalRepo.SubmitEval(ctx, tx, evaluationID, profileID, comps, nil, lockedRow.State, false, true); err != nil {
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
	// Jefe comments live in the manager_comment sidecar (ent model has no such
	// field); overlay them without touching the interface so test fakes
	// without the sidecar keep compiling (assertion fails → skip).
	if r, ok := s.evalRepo.(interface {
		GetCompetencyManagerComments(context.Context, uuid.UUID) (map[string]string, error)
	}); ok {
		if mc, err := r.GetCompetencyManagerComments(ctx, id); err == nil && len(mc) > 0 {
			for i := range resp.CompetencyRatings {
				if v, ok := mc[resp.CompetencyRatings[i].CompetencyID.String()]; ok {
					resp.CompetencyRatings[i].ManagerComment = v
				}
			}
		}
	}
	resp.GoalRatings = make([]dto.GoalRatingDTO, len(goals))
	for i, g := range goals {
		// Direct closing value: cierre ?? avance (never a derived 1-5 rating).
		var finalProgress *float64
		if g.CierreProgress != nil {
			finalProgress = g.CierreProgress
		} else if g.AvanceProgress != nil {
			finalProgress = g.AvanceProgress
		}
		resp.GoalRatings[i] = dto.GoalRatingDTO{
			GoalID:         g.GoalID,
			FinalRating:    g.FinalRating,
			FinalProgress:  finalProgress,
			AvanceProgress: g.AvanceProgress,
			CierreProgress: g.CierreProgress,
			FinalComments:  g.FinalComments,
			RhAssessment:   g.RhAssessment,
			ManagerComment: g.ManagerComment,
			// F2: row timestamps cover self/rh/manager comments alike
			// (no per-comment author column; AuthorName stays empty).
			CreatedAt:               &g.CreatedAt,
			UpdatedAt:               &g.UpdatedAt,
			ManagerCommentCreatedAt: &g.UpdatedAt,
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

// UpdateGoalState updates per-goal final_progress, self_assessment and rh_assessment.
// Only fields provided in the input are modified; the rest are left untouched.
func (s *EvaluationService) UpdateGoalState(ctx context.Context, evaluationID uuid.UUID, input dto.GoalStateUpdateInput, ifMatch int) (*dto.EvaluationDetailResponse, error) {
	row, err := s.evalRepo.GetByID(ctx, evaluationID)
	if err != nil {
		return nil, err
	}
	// ponytail: optimistic lock — reject if version doesn't match.
	if ifMatch > 0 && row.Version != ifMatch {
		return nil, errors.ErrConcurrentUpdate
	}
	if err := s.validateWritePhase(ctx, row.CycleID, ""); err != nil {
		return nil, err
	}
	if row.State == state.StateCompleted.String() {
		return nil, repo.ErrEvaluationFinalized
	}

	// P3: direct closing value in the goal's own unit — 0 allowed, negative
	// rejected (400). No unit/direction in input by contract
	// (GoalStateUpdateInput has no unit field); direction scaling lives in
	// P4 ninebox, not here. No final_rating derivation.
	if input.FinalProgress != nil && *input.FinalProgress < 0 {
		return nil, errors.NewDomainError(errors.InvalidRequest,
			"el progreso no puede ser negativo", nil,
		).WithDetails("INVALID_PROGRESS")
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

	// P2: resolve the write phase from the ACTIVE cycle phase
	// (cycles.current_phase via cycleCheck). Explicit request phase must
	// match the active phase (only active editable, prior immutable).
	// Missing phase defaults to the active phase — never to avance.
	activePhase := ""
	if p, err := s.cycleCheck.GetPhase(ctx, row.CycleID); err == nil {
		activePhase = p
	}
	phase := ""
	if input.Phase != nil && *input.Phase != "" {
		if activePhase != "" && !state.SamePhaseForWrite(*input.Phase, activePhase) {
			return nil, errors.NewDomainError(errors.PhaseNotActive,
				fmt.Sprintf("phase '%s' is not active; current phase is '%s'", *input.Phase, activePhase), nil,
			).WithDetails("requested_phase: " + *input.Phase, "current_phase: " + activePhase)
		}
		phase = *input.Phase
	} else if activePhase != "" {
		phase = activePhase
	} else if row.Phase != "" {
		phase = row.Phase
	}

	// F3: nil = not sent (leave unchanged); empty string is also treated as
	// not-sent — there is no intentional-clear path yet, so never blank comments.
	selfAssessment := input.SelfAssessment
	if selfAssessment != nil && *selfAssessment == "" {
		selfAssessment = nil
	}
	rhAssessment := input.RhAssessment
	if rhAssessment != nil && *rhAssessment == "" {
		rhAssessment = nil
	}

	if err := s.goalRepo.UpsertGoalState(ctx, tx, evaluationID, repo.GoalStateUpsert{
		GoalID:         input.GoalID,
		FinalProgress:  input.FinalProgress,
		FinalRating:    input.FinalRating,
		Phase:          phase,
		SelfAssessment: selfAssessment,
		RhAssessment:   rhAssessment,
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	tx = nil

	return s.GetEvaluation(ctx, evaluationID)
}

// UpdateGoalComments updates per-goal manager_comment.
func (s *EvaluationService) UpdateGoalComments(ctx context.Context, evaluationID uuid.UUID, input dto.GoalCommentUpdateInput, ifMatch int) (*dto.EvaluationDetailResponse, error) {
	row, err := s.evalRepo.GetByID(ctx, evaluationID)
	if err != nil {
		return nil, err
	}
	// ponytail: optimistic lock — reject if version doesn't match.
	if ifMatch > 0 && row.Version != ifMatch {
		return nil, errors.ErrConcurrentUpdate
	}
	if err := s.validateWritePhase(ctx, row.CycleID, ""); err != nil {
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

	if _, err := s.evalRepo.LockEvalForUpdate(ctx, tx, evaluationID); err != nil {
		return nil, err
	}

	if err := s.goalRepo.UpsertGoalComment(ctx, tx, evaluationID, repo.GoalCommentUpsert{
		GoalID:  input.GoalID,
		Comment: input.Comment,
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	tx = nil

	return s.GetEvaluation(ctx, evaluationID)
}

// ResolveEvaluationID resolves the evaluation ID for employee+cycle via FindByEmployeeCycle.
func (s *EvaluationService) ResolveEvaluationID(ctx context.Context, employeeID, cycleID uuid.UUID) (uuid.UUID, error) {
	row, err := s.evalRepo.FindByEmployeeCycle(ctx, employeeID, cycleID)
	if err != nil {
		return uuid.Nil, err
	}
	log.Printf("[evalId] resolved employee=%s cycle=%s eval=%s", employeeID, cycleID, row.ID)
	return row.ID, nil
}

// EnsureEvaluation finds or creates the evaluation for employee+cycle.
// Empty phase defaults to the cycle's current phase.
func (s *EvaluationService) EnsureEvaluation(ctx context.Context, employeeID, cycleID uuid.UUID, phase string) (uuid.UUID, error) {
	if phase == "" {
		if p, err := s.cycleCheck.GetPhase(ctx, cycleID); err == nil && p != "" {
			phase = p
		}
	}
	actorID, _ := auth.GetEmployeeID(ctx)
	if actorID == uuid.Nil {
		actorID = employeeID
	}
	row, err := s.evalRepo.EnsureEvaluation(ctx, employeeID, cycleID, phase, actorID)
	if err != nil {
		return uuid.Nil, err
	}
	log.Printf("[evalId] ensured employee=%s cycle=%s eval=%s phase=%s", employeeID, cycleID, row.ID, phase)
	return row.ID, nil
}

// ResolveActiveCycleID resolves the active (latest) cycle for an employee's organization.
func (s *EvaluationService) ResolveActiveCycleID(ctx context.Context, employeeID uuid.UUID) (uuid.UUID, error) {
	emp, err := s.empRepo.GetByID(ctx, employeeID)
	if err != nil {
		return uuid.Nil, err
	}
	orgNode, err := s.nodeRepo.GetByID(ctx, emp.OrgNodeID)
	if err != nil {
		return uuid.Nil, err
	}
	return s.cycleCheck.GetActiveCycleID(ctx, orgNode.OrganizationID)
}

// GetEmployeeCompetencyRatings returns ALL competencies for an employee's profile in a cycle,
// with self/rh ratings if evaluations exist (or null if not yet evaluated).
// phase is optional: empty defaults to the cycle's current phase so reads
// follow the active phase without requiring the frontend to send it.
func (s *EvaluationService) GetEmployeeCompetencyRatings(ctx context.Context, employeeID, cycleID uuid.UUID, phase string) (*dto.EmployeeCompetencyRatingsResponse, error) {
	emp, err := s.empRepo.GetByID(ctx, employeeID)
	if err != nil {
		return nil, err
	}
	if phase == "" {
		if p, err := s.cycleCheck.GetPhase(ctx, cycleID); err == nil && p != "" {
			phase = p
		}
	}

	rows, err := s.evalRepo.GetCompetencyRatingsByEmployee(ctx, employeeID, cycleID, emp.ProfileID, phase)
	if err != nil {
		return nil, err
	}

	ratings := make([]dto.EmployeeCompetencyRatingDTO, len(rows))
	for i, r := range rows {
		ratings[i] = dto.EmployeeCompetencyRatingDTO{
			CompetencyID:    r.CompetencyID,
			SelfRating:      r.SelfRating,
			RhRating:        r.RhRating,
			Comments:        r.Comments,
			SelfComment:     r.SelfComment,
			RhComment:       r.RhComment,
			ManagerComment:  r.ManagerComment,
			AcceptanceLevel: r.AcceptanceLevel,
		}
	}

	return &dto.EmployeeCompetencyRatingsResponse{
		EmployeeID: employeeID,
		CycleID:    cycleID,
		Ratings:    ratings,
	}, nil
}

// GetCyclePhase exposes the current phase of a cycle for privacy gates.
func (s *EvaluationService) GetCyclePhase(ctx context.Context, cycleID uuid.UUID) (string, error) {
	return s.cycleCheck.GetPhase(ctx, cycleID)
}

// ListEvaluations returns a cursor-paginated list of evaluations for a cycle.
// phase filters by evaluation phase ("avance"/"medio-anio" are equivalent);
// empty defaults to the cycle's current_phase.
func (s *EvaluationService) ListEvaluations(ctx context.Context, cycleID uuid.UUID, stateFilter string, phase string, cursor string, limit int) (*dto.EvaluationListResponse, error) {
	if phase == "" {
		if p, err := s.cycleCheck.GetPhase(ctx, cycleID); err == nil && p != "" {
			phase = p
		}
	}
	rows, nextCursor, err := s.evalRepo.ListByCycle(ctx, cycleID, stateFilter, phase, cursor, limit)
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

// GetCompetencyResults returns paginated competency averages with optional search
// and scope=team filter. phase filters by evaluation phase
// ("avance"/"medio-anio" are equivalent); empty defaults to the cycle's current_phase.
func (s *EvaluationService) GetCompetencyResults(ctx context.Context, cycleID uuid.UUID, phase string, query string, scope string, currentUserID uuid.UUID, offset, limit int) (*dto.CompetencyResultsResponse, error) {
	// Clamp pagination params
	if limit <= 0 {
		limit = 50
	} else if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	if phase == "" {
		if p, err := s.cycleCheck.GetPhase(ctx, cycleID); err == nil && p != "" {
			phase = p
		}
	}

	// Resolve managerID for team scope: filter by direct reports of the current user
	var managerID *uuid.UUID
	if scope == "team" && currentUserID != uuid.Nil {
		managerID = &currentUserID
	}

	total, err := s.evalRepo.CountCompetencyResults(ctx, cycleID, phase, query, managerID)
	if err != nil {
		return nil, err
	}

	rows, err := s.evalRepo.ListCompetencyResults(ctx, cycleID, phase, query, managerID, offset, limit)
	if err != nil {
		return nil, err
	}

	items := make([]dto.CompetencyResultItem, len(rows))
	for i, r := range rows {
		items[i] = dto.CompetencyResultItem{
			ID:            r.ID.String(),
			Name:          r.Name,
			ProfileName:   r.ProfileName,
			SelfRatingAvg: r.SelfRatingAvg,
			RHRatingAvg:   r.RHRatingAvg,
			Status:        r.Status,
		}
	}

	hasMore := offset+len(rows) < total

	return &dto.CompetencyResultsResponse{
		Data: items,
		Meta: dto.PaginationMeta{
			HasMore: hasMore,
			Total:   total,
			Offset:  offset,
			Limit:   limit,
		},
	}, nil
}

func (s *EvaluationService) validatePhase(ctx context.Context, cycleID uuid.UUID) error {
	phase, err := s.cycleCheck.GetPhase(ctx, cycleID)
	if err != nil {
		return fmt.Errorf("failed to get cycle phase: %w", err)
	}
	return state.RequiresPhase(phase)
}

// validateWritePhase verifies the cycle is in the required phase for a
// mid-year write (avance/medio-anio) or cierre. An empty wantPhase allows
// any writable phase; otherwise the cycle's current phase must match
// (with "avance"/"medio-anio" treated as the same phase). Saved rows stay
// editable while the phase is active, including after a revert. Mismatch
// returns PHASE_NOT_ADVANCEABLE (409); missing permission is 403 via RBAC.
func validateRowPhase(rowPhase, currentPhase string) error {
	if rowPhase == "" {
		return nil
	}
	if state.SamePhaseForWrite(rowPhase, currentPhase) {
		return nil
	}
	return errors.NewDomainError(errors.PhaseNotAdvanceable, "phase_mismatch: row is "+rowPhase+" but active is "+currentPhase, nil)
}

func (s *EvaluationService) validateWritePhase(ctx context.Context, cycleID uuid.UUID, wantPhase string) error {
	phase, err := s.cycleCheck.GetPhase(ctx, cycleID)
	if err != nil {
		return fmt.Errorf("failed to get cycle phase: %w", err)
	}
	return state.WritableInPhase(phase, wantPhase)
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
		fmt.Fprintf(h, "%d", c.Rating)
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
		fmt.Fprintf(h, "%d", c.Rating)
		h.Write([]byte{0})
		h.Write([]byte(c.Comments))
		h.Write([]byte{0})
		h.Write([]byte(c.ManagerComment))
		h.Write([]byte{0})
	}
	h.Write([]byte(req.FinalComments))
	return hex.EncodeToString(h.Sum(nil))
}

// --- Colaborador privacy gate (sed-roles-privacidad, punto 5) ---

// selfRestrictedPhase reports whether a colaborador viewer is limited to
// their own evaluations in the given cycle phase (avance / medio-anio).
func selfRestrictedPhase(phase string) bool {
	return state.IsMidYearPhase(phase)
}

// FilterEvaluationsForViewer hides other employees' evaluations from
// colaborador viewers during avance/medio-anio; manager/rh see everything.
func (s *EvaluationService) FilterEvaluationsForViewer(items []dto.EvaluationListItem, viewerID uuid.UUID, viewerRole auth.Role, phase string) []dto.EvaluationListItem {
	if viewerRole != auth.RoleColaborador || !selfRestrictedPhase(phase) {
		return items
	}
	out := make([]dto.EvaluationListItem, 0, len(items))
	for _, it := range items {
		if it.EmployeeID == viewerID {
			out = append(out, it)
		}
	}
	return out
}

// AuthorizeEvaluationAccess returns ErrForbidden (403) when a colaborador
// viewer opens someone else's evaluation during avance/medio-anio.
func (s *EvaluationService) AuthorizeEvaluationAccess(viewerID, employeeID uuid.UUID, viewerRole auth.Role, phase string) error {
	if viewerRole == auth.RoleColaborador && selfRestrictedPhase(phase) && viewerID != employeeID {
		return errors.ErrForbidden
	}
	return nil
}

// canWriteRHEvaluation is the pure ownership rule for rh-evaluation writes:
// RH (PermEvalRH) may write any evaluation; otherwise only the assigned
// manager (ManagerID == actor) may write. No session data is read here so the
// rule stays table-testable; AuthorizeRHEvaluationWrite resolves the inputs.
func canWriteRHEvaluation(role auth.Role, actorID uuid.UUID, managerID *uuid.UUID) bool {
	if auth.HasPermission(role, auth.PermEvalRH) {
		return true
	}
	if managerID == nil || *managerID == uuid.Nil || actorID == uuid.Nil {
		return false
	}
	return *managerID == actorID
}

// AuthorizeRHEvaluationWrite allows rh-evaluation writes for RH holders or the
// assigned manager (jefe directo) of the evaluated employee, validated
// server-side via employees.manager_id. Missing session → 401
// NOT_AUTHENTICATED; anyone else → 403 ErrForbidden. It never grants the
// eval:rh permission itself, so unrelated jefes stay rejected.
func (s *EvaluationService) AuthorizeRHEvaluationWrite(ctx context.Context, evaluationID uuid.UUID) error {
	role, ok := auth.GetRole(ctx)
	if !ok {
		return errors.NewDomainError(errors.NotAuthenticated, "no authenticated session", nil)
	}
	actorID, ok := auth.GetEmployeeID(ctx)
	if !ok || actorID == uuid.Nil {
		return errors.NewDomainError(errors.NotAuthenticated, "no authenticated session", nil)
	}
	if auth.HasPermission(role, auth.PermEvalRH) {
		return nil
	}
	row, err := s.evalRepo.GetByID(ctx, evaluationID)
	if err != nil {
		return err
	}
	if s.empRepo == nil {
		return errors.ErrForbidden
	}
	emp, err := s.empRepo.GetByID(ctx, row.EmployeeID)
	if err != nil {
		return err
	}
	if canWriteRHEvaluation(role, actorID, emp.ManagerID) {
		return nil
	}
	return errors.ErrForbidden
}

// RedactDetailForSelf hides completion timestamps from self (colaborador)
// viewers. F2: goal comments (self/rh/manager) stay visible for every
// viewerMode — only timestamps are redacted, never the comment text.
// Keeps manager/RH views untouched.
func (s *EvaluationService) RedactDetailForSelf(detail *dto.EvaluationDetailResponse, viewerID uuid.UUID, viewerRole auth.Role, phase string) *dto.EvaluationDetailResponse {
	if detail == nil || viewerRole != auth.RoleColaborador || viewerID != detail.EmployeeID {
		return detail
	}
	if selfRestrictedPhase(phase) {
		detail.RHEvalCompletedAt = nil
	}
	// Self viewers never see RH completion timestamp; own progress stays.
	detail.RHEvalCompletedAt = nil
	return detail
}

// SuggestEvaluator returns the employee's manager (jefe recomendado), with
// "rh" fallback when the employee has no manager assigned.
func (s *EvaluationService) SuggestEvaluator(ctx context.Context, employeeID uuid.UUID) (uuid.UUID, string, error) {
	if s.empRepo == nil {
		return uuid.Nil, "rh", nil
	}
	emp, err := s.empRepo.GetByID(ctx, employeeID)
	if err != nil {
		return uuid.Nil, "", err
	}
	if emp.ManagerID != nil && *emp.ManagerID != uuid.Nil {
		return *emp.ManagerID, "jefe", nil
	}
	return uuid.Nil, "rh", nil
}
