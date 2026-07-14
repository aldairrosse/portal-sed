package goal

import (
	"context"

	"github.com/google/uuid"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
	repoorg "github.com/sed-evaluacion-desempeno/api/internal/repository/org"
	repocycle "github.com/sed-evaluacion-desempeno/api/internal/repository/cycle"
)

// cyclePhaseCheck implements PhaseChecker backed by the cycle repository.
// ponytail: replaces nopPhaseChecker stub. Reads the real active cycle phase
// for the employee's organization instead of always returning "asignacion".
type cyclePhaseCheck struct {
	cycleRepo    *repocycle.CycleRepo
	employeeRepo *repoorg.EmployeeRepo
	orgNodeRepo  *repoorg.OrgNodeRepo
}

// NewCyclePhaseCheck creates a PhaseChecker that reads the real cycle phase from DB.
func NewCyclePhaseCheck(
	cycleRepo *repocycle.CycleRepo,
	employeeRepo *repoorg.EmployeeRepo,
	orgNodeRepo *repoorg.OrgNodeRepo,
) PhaseChecker {
	return &cyclePhaseCheck{
		cycleRepo:    cycleRepo,
		employeeRepo: employeeRepo,
		orgNodeRepo:  orgNodeRepo,
	}
}

func (c *cyclePhaseCheck) GetCurrentPhase(ctx context.Context, empID string) (CyclePhase, error) {
	eid, err := uuid.Parse(empID)
	if err != nil {
		return "", pkgerrors.ErrInvalidRequest
	}

	emp, err := c.employeeRepo.GetByID(ctx, eid)
	if err != nil {
		return "", err
	}

	node, err := c.orgNodeRepo.GetByID(ctx, emp.OrgNodeID)
	if err != nil {
		return "", err
	}

	cycleID, err := c.cycleRepo.GetActiveCycleID(ctx, node.OrganizationID)
	if err != nil {
		return "", err
	}

	row, err := c.cycleRepo.GetCycle(ctx, cycleID)
	if err != nil {
		return "", err
	}

	return CyclePhase(row.CurrentPhase), nil
}
