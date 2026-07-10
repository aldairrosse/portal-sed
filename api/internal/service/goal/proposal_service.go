package goal

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	dtogoal "github.com/sed-evaluacion-desempeno/api/internal/dto/goal"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	repogoal "github.com/sed-evaluacion-desempeno/api/internal/repository/goal"
)

// GoalProposalService handles business logic for goal proposals.
type GoalProposalService struct {
	proposalRepo  GoalProposalRepository
	goalRepo      GoalRepository
	catRepo       CategoryRepository
	linkRepo      LinkKPIRepository
	weightQueries WeightQuerier
	phaseCheck    *PhaseCheck
	db            *sql.DB
}

// NewGoalProposalService creates a new GoalProposalService.
func NewGoalProposalService(
	proposalRepo GoalProposalRepository,
	goalRepo GoalRepository,
	catRepo CategoryRepository,
	linkRepo LinkKPIRepository,
	weightQueries WeightQuerier,
	phaseCheck *PhaseCheck,
	db *sql.DB,
) *GoalProposalService {
	return &GoalProposalService{
		proposalRepo:  proposalRepo,
		goalRepo:      goalRepo,
		catRepo:       catRepo,
		linkRepo:      linkRepo,
		weightQueries: weightQueries,
		phaseCheck:    phaseCheck,
		db:            db,
	}
}

// CreateProposal creates a new goal change proposal.
func (s *GoalProposalService) CreateProposal(ctx context.Context, requestedBy, goalID uuid.UUID, req dtogoal.CreateGoalProposalRequest) (*repogoal.GoalProposalRow, error) {
	if err := s.phaseCheck.CanCreateGoal(ctx, requestedBy.String()); err != nil {
		return nil, err
	}

	if _, err := s.goalRepo.GetGoal(ctx, goalID); err != nil {
		return nil, err
	}

	if err := validateGoalRequest(dtogoal.CreateGoalRequest{
		Name:        req.Name,
		Description: req.Description,
		Unit:        req.Unit,
		Weight:      req.Weight,
		TargetValue: req.TargetValue,
		Direction:   req.Direction,
	}); err != nil {
		return nil, err
	}

	if err := validateDirection(req.Direction, req.BaselineValue, req.TargetValue); err != nil {
		return nil, err
	}

	if len(req.KpiIDs) > 5 {
		return nil, pkgerrors.ErrKpiLinkLimitExceeded
	}

	proposal, err := s.proposalRepo.Create(ctx, goalID, requestedBy,
		req.Name, req.Description, req.Unit, req.Direction,
		req.Weight, req.TargetValue, req.BaselineValue,
	)
	if err != nil {
		return nil, fmt.Errorf("create proposal: %w", err)
	}

	if len(req.KpiIDs) > 0 {
		kpiUUIDs := make([]uuid.UUID, len(req.KpiIDs))
		for i, id := range req.KpiIDs {
			kpiUUIDs[i], err = uuid.Parse(id)
			if err != nil {
				return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "invalid kpi_id format", err)
			}
		}
		if err := s.proposalRepo.SetKpis(ctx, proposal.ID, kpiUUIDs); err != nil {
			return nil, fmt.Errorf("set proposal kpis: %w", err)
		}
	}

	return proposal, nil
}

// AcceptProposal accepts a pending proposal and applies its values to the goal.
func (s *GoalProposalService) AcceptProposal(ctx context.Context, ownerID, proposalID uuid.UUID) (*repogoal.GoalRow, error) {
	proposal, err := s.proposalRepo.GetByID(ctx, proposalID)
	if err != nil {
		return nil, err
	}
	if proposal.Status != "pending" {
		return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "proposal is not pending", nil)
	}

	goal, err := s.goalRepo.GetGoal(ctx, proposal.GoalID)
	if err != nil {
		return nil, err
	}

	cat, err := s.catRepo.GetCategory(ctx, goal.CategoryID)
	if err != nil {
		return nil, err
	}
	if cat.EmployeeID != ownerID {
		return nil, pkgerrors.ErrGoalNotFound
	}

	if err := s.phaseCheck.CanUpdateGoal(ctx, ownerID.String()); err != nil {
		return nil, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var catEmpID uuid.UUID
	err = tx.QueryRowContext(ctx,
		`SELECT employee_id FROM goal_categories WHERE id = $1 FOR UPDATE`,
		goal.CategoryID,
	).Scan(&catEmpID)
	if err != nil {
		return nil, fmt.Errorf("lock category: %w", err)
	}
	if catEmpID != ownerID {
		return nil, pkgerrors.ErrGoalNotFound
	}

	baselineVal := proposal.BaselineValue
	var baselineDesc interface{} = nil
	if baselineVal != nil {
		baselineDesc = *baselineVal
	}
	var desc interface{} = nil
	if proposal.Description != "" {
		desc = proposal.Description
	}

	res, execErr := tx.ExecContext(ctx,
		`UPDATE goals
		 SET name = $1, description = $2, unit = $3, direction = $4,
		     weight = $5, target_value = $6, baseline_value = $7,
		     updated_at = now(), updated_by = $8, version = version + 1
		 WHERE id = $9 AND version = $10`,
		proposal.Name, desc, proposal.Unit, proposal.Direction,
		proposal.Weight, proposal.TargetValue, baselineDesc,
		ownerID, goal.ID, goal.Version,
	)
	if execErr != nil {
		return nil, fmt.Errorf("update goal in tx: %w", execErr)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return nil, pkgerrors.ErrConcurrentModification
	}

	kpiIDs, err := s.proposalRepo.GetKpiIDs(ctx, proposalID)
	if err != nil {
		return nil, fmt.Errorf("get proposal kpis: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM goal_kpi_links WHERE goal_id = $1`,
		goal.ID,
	); err != nil {
		return nil, fmt.Errorf("delete goal kpi links: %w", err)
	}
	for _, kpiID := range kpiIDs {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO goal_kpi_links (goal_id, kpi_id, created_at)
			 VALUES ($1, $2, now())
			 ON CONFLICT (goal_id, kpi_id) DO NOTHING`,
			goal.ID, kpiID,
		); err != nil {
			return nil, fmt.Errorf("insert goal kpi link: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE goal_proposals
		 SET status = 'accepted', reviewed_by = $1, reviewed_at = now(), updated_at = now()
		 WHERE id = $2`,
		ownerID.String(), proposalID,
	); err != nil {
		return nil, fmt.Errorf("update proposal status: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE goal_proposals
		 SET status = 'rejected', reviewed_by = $1, reviewed_at = now(), updated_at = now()
		 WHERE goal_id = $2 AND status = 'pending' AND id != $3`,
		ownerID.String(), goal.ID, proposalID,
	); err != nil {
		return nil, fmt.Errorf("reject other proposals: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	updatedGoal, err := s.goalRepo.GetGoal(ctx, goal.ID)
	if err != nil {
		return nil, err
	}
	return updatedGoal, nil
}

// RejectProposal rejects a pending proposal.
func (s *GoalProposalService) RejectProposal(ctx context.Context, ownerID, proposalID uuid.UUID) (*repogoal.GoalProposalRow, error) {
	proposal, err := s.proposalRepo.GetByID(ctx, proposalID)
	if err != nil {
		return nil, err
	}
	if proposal.Status != "pending" {
		return nil, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "proposal is not pending", nil)
	}

	goal, err := s.goalRepo.GetGoal(ctx, proposal.GoalID)
	if err != nil {
		return nil, err
	}

	cat, err := s.catRepo.GetCategory(ctx, goal.CategoryID)
	if err != nil {
		return nil, err
	}
	if cat.EmployeeID != ownerID {
		return nil, pkgerrors.ErrGoalNotFound
	}

	if err := s.phaseCheck.CanUpdateGoal(ctx, ownerID.String()); err != nil {
		return nil, err
	}

	return s.proposalRepo.UpdateStatus(ctx, proposalID, "rejected", ownerID)
}
