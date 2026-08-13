package goal

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/auth"
	"github.com/sed-evaluacion-desempeno/api/internal/cycle"
	"github.com/sed-evaluacion-desempeno/api/internal/employee"
	"github.com/sed-evaluacion-desempeno/api/internal/globalgoalassignment"
	"github.com/sed-evaluacion-desempeno/api/internal/globalgoalrule"
	"github.com/sed-evaluacion-desempeno/api/internal/goal"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
)

// GlobalGoalRow is the full representation of a global goal with its assignments.
type GlobalGoalRow struct {
	ID           uuid.UUID              `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Unit         string                 `json:"unit"`
	Direction    string                 `json:"direction"`
	Weight       float64                `json:"weight"`
	TargetValue  float64                `json:"target_value"`
	CurrentValue float64                `json:"current_value"`
	GoalKind     string                 `json:"goal_kind"`
	State        string                 `json:"state"`
	CreatedBy    uuid.UUID              `json:"created_by"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	Assignments  []*GlobalAssignmentRow `json:"assignments"`
	Rules        []*GlobalRuleRow       `json:"rules"`
}

func goalKindValue(value *goal.GoalKind) string {
	if value == nil {
		return ""
	}
	return string(*value)
}

// GlobalAssignmentRow represents an assignment of a global goal to an employee.
type GlobalAssignmentRow struct {
	ID            uuid.UUID `json:"id"`
	GoalID        uuid.UUID `json:"goal_id"`
	EmployeeID    uuid.UUID `json:"employee_id"`
	Weight        float64   `json:"weight"`
	TargetValue   float64   `json:"target_value"`
	BaselineValue *float64  `json:"baseline_value,omitempty"`
}

// GlobalRuleRow represents a mass assignment rule.
type GlobalRuleRow struct {
	ID               uuid.UUID  `json:"id"`
	GoalID           uuid.UUID  `json:"goal_id"`
	RuleType         string     `json:"rule_type"`
	DepartmentID     *uuid.UUID `json:"department_id,omitempty"`
	MinDirectReports *int       `json:"min_direct_reports,omitempty"`
	ProfileID        *uuid.UUID `json:"profile_id,omitempty"`
	DefaultWeight    float64    `json:"default_weight"`
	DefaultTarget    float64    `json:"default_target"`
}

// GlobalGoalRepo provides Ent-backed operations for global goals.
type GlobalGoalRepo struct {
	client *internal.Client
	db     *sql.DB
}

// NewGlobalGoalRepo creates a new GlobalGoalRepo.
func NewGlobalGoalRepo(client *internal.Client, db *sql.DB) *GlobalGoalRepo {
	return &GlobalGoalRepo{client: client, db: db}
}

// CreateGlobalGoal creates a new global goal with assignments and rules.
func (r *GlobalGoalRepo) CreateGlobalGoal(ctx context.Context, cycleID, createdBy uuid.UUID, name, description, unit, direction, goalKind string, weight, targetValue float64, assignments []*GlobalAssignmentRow, rules []*GlobalRuleRow) (*GlobalGoalRow, error) {
	// Create the goal
	g, err := r.client.Goal.Create().
		SetName(name).
		SetDescription(description).
		SetUnit(goal.Unit(unit)).
		SetDirection(goal.Direction(direction)).
		SetWeight(weight).
		SetTargetValue(targetValue).
		SetGoalKind(goal.GoalKind(goalKind)).
		SetState(goal.StateBorrador).
		SetCategoryID(uuid.Nil). // Global goals don't belong to a category
		SetCreatedBy(createdBy).
		SetUpdatedBy(createdBy).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	// Create assignments
	for _, a := range assignments {
		create := r.client.GlobalGoalAssignment.Create().
			SetGoalID(g.ID).
			SetEmployeeID(a.EmployeeID).
			SetWeight(a.Weight).
			SetTargetValue(a.TargetValue)

		if a.BaselineValue != nil {
			create = create.SetBaselineValue(*a.BaselineValue)
		}

		_, err := create.Save(ctx)
		if err != nil {
			return nil, err
		}
	}

	// Create rules
	for _, rule := range rules {
		create := r.client.GlobalGoalRule.Create().
			SetGoalID(g.ID).
			SetRuleType(globalgoalrule.RuleType(rule.RuleType)).
			SetDefaultWeight(rule.DefaultWeight).
			SetDefaultTarget(rule.DefaultTarget)

		if rule.DepartmentID != nil {
			create = create.SetDepartmentID(*rule.DepartmentID)
		}
		if rule.MinDirectReports != nil {
			create = create.SetMinDirectReports(*rule.MinDirectReports)
		}
		if rule.ProfileID != nil {
			create = create.SetProfileID(*rule.ProfileID)
		}

		_, err = create.Save(ctx)
		if err != nil {
			return nil, err
		}
	}

	return r.GetGlobalGoal(ctx, g.ID)
}

// GetGlobalGoal retrieves a global goal with its assignments and rules.
func (r *GlobalGoalRepo) GetGlobalGoal(ctx context.Context, goalID uuid.UUID) (*GlobalGoalRow, error) {
	g, err := r.client.Goal.Query().
		Where(goal.ID(goalID), goal.TypeEQ(goal.TypeGlobal)).
		WithGlobalAssignments(func(q *internal.GlobalGoalAssignmentQuery) {
			q.Order(internal.Asc(globalgoalassignment.FieldEmployeeID))
		}).
		WithGlobalRules(func(q *internal.GlobalGoalRuleQuery) {
			q.Order(internal.Asc(globalgoalrule.FieldRuleType))
		}).
		Only(ctx)
	if err != nil {
		if internal.IsNotFound(err) {
			return nil, pkgerrors.ErrGoalNotFound
		}
		return nil, err
	}

	row := &GlobalGoalRow{
		ID:          g.ID,
		Name:        g.Name,
		Description: g.Description,
		Unit:        string(g.Unit),
		Direction:   string(g.Direction),
		Weight:      g.Weight,
		TargetValue: g.TargetValue,
		GoalKind:    goalKindValue(g.GoalKind),
		State:       string(g.State),
		CreatedBy:   g.CreatedBy,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
		Assignments: make([]*GlobalAssignmentRow, 0),
		Rules:       make([]*GlobalRuleRow, 0),
	}

	for _, a := range g.Edges.GlobalAssignments {
		row.Assignments = append(row.Assignments, &GlobalAssignmentRow{
			ID:            a.ID,
			GoalID:        a.GoalID,
			EmployeeID:    a.EmployeeID,
			Weight:        a.Weight,
			TargetValue:   a.TargetValue,
			BaselineValue: a.BaselineValue,
		})
	}

	for _, rule := range g.Edges.GlobalRules {
		ruleRow := &GlobalRuleRow{
			ID:            rule.ID,
			GoalID:        rule.GoalID,
			RuleType:      string(rule.RuleType),
			DefaultWeight: rule.DefaultWeight,
			DefaultTarget: rule.DefaultTarget,
		}
		if rule.DepartmentID != nil {
			ruleRow.DepartmentID = rule.DepartmentID
		}
		if rule.MinDirectReports != nil {
			ruleRow.MinDirectReports = rule.MinDirectReports
		}
		if rule.ProfileID != nil {
			ruleRow.ProfileID = rule.ProfileID
		}
		row.Rules = append(row.Rules, ruleRow)
	}

	return row, nil
}

// ListGlobalGoalsByCycle lists all global goals for a cycle.
func (r *GlobalGoalRepo) ListGlobalGoalsByCycle(ctx context.Context, cycleID uuid.UUID) ([]*GlobalGoalRow, error) {
	goals, err := r.client.Goal.Query().
		Where(goal.TypeEQ(goal.TypeGlobal)).
		WithGlobalAssignments().
		WithGlobalRules().
		Order(internal.Desc(goal.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	rows := make([]*GlobalGoalRow, 0, len(goals))
	for _, g := range goals {
		row := &GlobalGoalRow{
			ID:          g.ID,
			Name:        g.Name,
			Description: g.Description,
			Unit:        string(g.Unit),
			Direction:   string(g.Direction),
			Weight:      g.Weight,
			TargetValue: g.TargetValue,
			GoalKind:    goalKindValue(g.GoalKind),
			State:       string(g.State),
			CreatedBy:   g.CreatedBy,
			CreatedAt:   g.CreatedAt,
			UpdatedAt:   g.UpdatedAt,
			Assignments: make([]*GlobalAssignmentRow, 0),
			Rules:       make([]*GlobalRuleRow, 0),
		}
		rows = append(rows, row)
	}

	return rows, nil
}

// UpdateGlobalGoal updates a global goal.
func (r *GlobalGoalRepo) UpdateGlobalGoal(ctx context.Context, goalID uuid.UUID, name, description, unit, direction, goalKind string, weight, targetValue float64, expectedVersion int) (*GlobalGoalRow, error) {
	g, err := r.client.Goal.UpdateOneID(goalID).
		SetName(name).
		SetDescription(description).
		SetUnit(goal.Unit(unit)).
		SetDirection(goal.Direction(direction)).
		SetGoalKind(goal.GoalKind(goalKind)).
		SetWeight(weight).
		SetTargetValue(targetValue).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return r.GetGlobalGoal(ctx, g.ID)
}

// DeleteGlobalGoal deletes a global goal and its assignments/rules.
func (r *GlobalGoalRepo) DeleteGlobalGoal(ctx context.Context, goalID uuid.UUID) error {
	// Cascading deletes will handle assignments and rules
	return r.client.Goal.DeleteOneID(goalID).Exec(ctx)
}

// ExecuteRules executes mass assignment rules for a global goal.
func (r *GlobalGoalRepo) ExecuteRules(ctx context.Context, goalID uuid.UUID) (int, error) {
	employeeID, ok := auth.GetEmployeeID(ctx)
	if !ok {
		return 0, pkgerrors.NewDomainError(pkgerrors.NotAuthenticated, "no authenticated user", nil)
	}
	authEmployee, err := r.client.Employee.Query().Where(employee.ID(employeeID)).WithOrgNode().Only(ctx)
	if err != nil {
		return 0, err
	}
	if authEmployee.Edges.OrgNode == nil {
		return 0, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "La organización actual no está disponible.", nil)
	}
	orgID := authEmployee.Edges.OrgNode.OrganizationID
	if orgID == uuid.Nil {
		return 0, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "La organización actual no es válida.", nil)
	}
	if !r.client.Cycle.Query().Where(cycle.OrganizationID(orgID), cycle.FinishedAtIsNil()).ExistX(ctx) {
		return 0, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "No existe un ciclo activo para la organización actual.", nil)
	}

	tx, err := r.client.Tx(ctx)
	if err != nil {
		return 0, err
	}
	rollback := func(err error) (int, error) { _ = tx.Rollback(); return 0, err }
	rules, err := tx.GlobalGoalRule.Query().Where(globalgoalrule.GoalID(goalID)).Order(internal.Asc(globalgoalrule.FieldID)).All(ctx)
	if err != nil {
		return rollback(err)
	}
	if len(rules) == 0 {
		return rollback(pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "La meta no tiene reglas de asignación.", nil))
	}
	employees, err := tx.Employee.Query().Where(employee.IsActive(true)).WithOrgNode().All(ctx)
	if err != nil {
		return rollback(err)
	}
	orgEmployees := make([]*internal.Employee, 0, len(employees))
	for _, emp := range employees {
		if emp.Edges.OrgNode != nil && emp.Edges.OrgNode.OrganizationID == orgID {
			orgEmployees = append(orgEmployees, emp)
		}
	}
	if len(orgEmployees) == 0 {
		return rollback(pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "No hay empleados activos en la organización actual.", nil))
	}

	departmentPaths := make(map[uuid.UUID]string)
	sets := make([]map[uuid.UUID]bool, 0, len(rules))
	for _, rule := range rules {
		set := make(map[uuid.UUID]bool)
		var departmentPath string
		if rule.RuleType == globalgoalrule.RuleTypeDepartment && rule.DepartmentID != nil {
			departmentPath, ok = departmentPaths[*rule.DepartmentID]
			if !ok {
				node, e := tx.OrgNode.Get(ctx, *rule.DepartmentID)
				if e != nil {
					return rollback(e)
				}
				departmentPath = node.Path
				departmentPaths[*rule.DepartmentID] = departmentPath
			}
		}
		for _, emp := range orgEmployees {
			if emp.Edges.OrgNode == nil || emp.Edges.OrgNode.OrganizationID != orgID {
				continue
			}
			match := false
			switch rule.RuleType {
			case globalgoalrule.RuleTypeDepartment:
				if rule.DepartmentID != nil {
					employeePath := emp.Edges.OrgNode.Path
					match = employeePath == departmentPath || strings.HasPrefix(employeePath, departmentPath+".")
				}
			case globalgoalrule.RuleTypeRole:
				match = rule.ProfileID != nil && emp.ProfileID == *rule.ProfileID
			case globalgoalrule.RuleTypeMinDirectReports:
				if rule.MinDirectReports != nil {
					count := 0
					for _, report := range orgEmployees {
						if report.Edges.OrgNode == nil || report.Edges.OrgNode.OrganizationID != orgID {
							continue
						}
						managerMatch := report.ManagerID != nil && *report.ManagerID == emp.ID
						departmentMatch := report.Edges.OrgNode.ParentID != nil && *report.Edges.OrgNode.ParentID == emp.OrgNodeID
						if managerMatch || departmentMatch {
							count++
						}
					}
					match = count >= *rule.MinDirectReports
				}
			}
			if match {
				set[emp.ID] = true
			}
		}
		sets = append(sets, set)
	}

	assigned := 0
	for _, emp := range orgEmployees {
		if len(sets) == 0 {
			continue
		}
		match := true
		for _, set := range sets {
			if !set[emp.ID] {
				match = false
				break
			}
		}
		if !match {
			continue
		}
		exists, err := tx.GlobalGoalAssignment.Query().Where(globalgoalassignment.GoalID(goalID), globalgoalassignment.EmployeeID(emp.ID)).Exist(ctx)
		if err != nil {
			return rollback(err)
		}
		if exists {
			continue
		}
		// A single assignment stores one target/weight. Rules are ordered by ID,
		// so the first persisted rule deterministically wins for an intersection.
		weight, target := rules[0].DefaultWeight, rules[0].DefaultTarget
		if _, err = tx.GlobalGoalAssignment.Create().SetGoalID(goalID).SetEmployeeID(emp.ID).SetWeight(weight).SetTargetValue(target).Save(ctx); err != nil {
			return rollback(err)
		}
		assigned++
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return assigned, nil
}
