package weight

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	repocycle "github.com/sed-evaluacion-desempeno/api/internal/repository/cycle"
	repoweight "github.com/sed-evaluacion-desempeno/api/internal/repository/weight"
)

// Service resolves active cycle per org and exposes Get/Save for RH and jefe.
type Service struct {
	repo      *repoweight.Repo
	cycleRepo *repocycle.CycleRepo
	db        *sql.DB
}

// NewService creates a Service.
func NewService(repo *repoweight.Repo, cycleRepo *repocycle.CycleRepo, db *sql.DB) *Service {
	return &Service{repo: repo, cycleRepo: cycleRepo, db: db}
}

// GetCycleConfig returns CycleWeights for org's active cycle (fallback P=100).
func (s *Service) GetCycleConfig(ctx context.Context, orgID uuid.UUID) (repoweight.CycleWeights, uuid.UUID, error) {
	cycleID, err := s.cycleRepo.GetActiveCycleID(ctx, orgID)
	if err != nil {
		// fallback when no active cycle
		return repoweight.CycleWeights{G: 0, P: 100}, uuid.Nil, nil
	}
	w, err := s.repo.GetOrFallbackCycleConfig(ctx, cycleID)
	return w, cycleID, err
}

// SaveCycleConfig validates and upserts G% for org's active cycle; P=100-G.
func (s *Service) SaveCycleConfig(ctx context.Context, orgID uuid.UUID, g float64) (repoweight.CycleWeights, error) {
	cycleID, err := s.cycleRepo.GetActiveCycleID(ctx, orgID)
	if err != nil {
		return repoweight.CycleWeights{}, err
	}
	return s.repo.UpsertCycleConfig(ctx, cycleID, g)
}

// GetTeamConfig returns TeamWeights for cycle+team (fallback PJ=100). If teamID is Nil, uses caller's org_node.
func (s *Service) GetTeamConfig(ctx context.Context, orgID, teamID uuid.UUID, callerID uuid.UUID) (repoweight.TeamWeights, uuid.UUID, error) {
	cycleID, err := s.cycleRepo.GetActiveCycleID(ctx, orgID)
	if err != nil {
		return repoweight.TeamWeights{J: 0, PJ: 100}, uuid.Nil, nil
	}
	if teamID == uuid.Nil {
		if resolved, err := s.resolveTeamID(ctx, callerID); err == nil {
			teamID = resolved
		} else {
			return repoweight.TeamWeights{J: 0, PJ: 100}, cycleID, nil
		}
	}
	w, err := s.repo.GetOrFallbackTeamConfig(ctx, cycleID, teamID)
	return w, cycleID, err
}

// SaveTeamConfig upserts J% for cycle+team; PJ=100-J.
func (s *Service) SaveTeamConfig(ctx context.Context, orgID, teamID uuid.UUID, callerID uuid.UUID, j float64) (repoweight.TeamWeights, error) {
	cycleID, err := s.cycleRepo.GetActiveCycleID(ctx, orgID)
	if err != nil {
		return repoweight.TeamWeights{}, err
	}
	if teamID == uuid.Nil {
		if resolved, rerr := s.resolveTeamID(ctx, callerID); rerr == nil {
			teamID = resolved
		} else {
			return repoweight.TeamWeights{}, rerr
		}
	}
	return s.repo.UpsertTeamConfig(ctx, cycleID, teamID, j)
}

// GetEmployeeHierarchicalWeights returns pWeight and pjWeight for empID with fallback 100.
func (s *Service) GetEmployeeHierarchicalWeights(ctx context.Context, empID uuid.UUID) (float64, float64) {
	// resolve employee org_node -> team and org
	var orgNodeID uuid.UUID
	var orgID uuid.UUID
	// lightweight queries: get employee's org_node_id and its organization_id
	err := s.db.QueryRowContext(ctx, `SELECT org_node_id FROM employees WHERE id=$1`, empID).Scan(&orgNodeID)
	if err != nil || orgNodeID == uuid.Nil {
		return 100, 100
	}
	_ = s.db.QueryRowContext(ctx, `SELECT organization_id FROM org_nodes WHERE id=$1`, orgNodeID).Scan(&orgID)
	if orgID == uuid.Nil {
		return 100, 100
	}
	cycleID, err := s.cycleRepo.GetActiveCycleID(ctx, orgID)
	if err != nil {
		return 100, 100
	}
	cw, _ := s.repo.GetOrFallbackCycleConfig(ctx, cycleID)
	tw, _ := s.repo.GetOrFallbackTeamConfig(ctx, cycleID, orgNodeID)
	return cw.P, tw.PJ
}

func (s *Service) resolveTeamID(ctx context.Context, callerID uuid.UUID) (uuid.UUID, error) {
	var teamID uuid.UUID
	err := s.db.QueryRowContext(ctx, `SELECT org_node_id FROM employees WHERE id=$1`, callerID).Scan(&teamID)
	if err != nil {
		return uuid.Nil, err
	}
	return teamID, nil
}

// ResolveOrgIDForEmployee returns organization_id for an employee's org_node.
func (s *Service) ResolveOrgIDForEmployee(ctx context.Context, empID uuid.UUID) (uuid.UUID, error) {
	var orgID uuid.UUID
	err := s.db.QueryRowContext(ctx, `SELECT organization_id FROM org_nodes WHERE id=(SELECT org_node_id FROM employees WHERE id=$1)`, empID).Scan(&orgID)
	return orgID, err
}
