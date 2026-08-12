package goal

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/employee"
	"github.com/sed-evaluacion-desempeno/api/internal/globalgoalassignment"
	"github.com/sed-evaluacion-desempeno/api/internal/globalgoalrule"
	"github.com/sed-evaluacion-desempeno/api/internal/goal"
	pkgerrors "github.com/sed-evaluacion-desempeno/api/internal/pkg/errors"
)

// GlobalGoalRow is the full representation of a global goal with its assignments.
type GlobalGoalRow struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Unit        string                 `json:"unit"`
	Direction   string                 `json:"direction"`
	Weight      float64                `json:"weight"`
	TargetValue float64                `json:"target_value"`
	GoalKind    string                 `json:"goal_kind"`
	State       string                 `json:"state"`
	CreatedBy   uuid.UUID              `json:"created_by"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Assignments []*GlobalAssignmentRow `json:"assignments"`
	Rules       []*GlobalRuleRow       `json:"rules"`
}

// GlobalAssignmentRow represents an assignment of a global goal to an employee.
type GlobalAssignmentRow struct {
	ID            uuid.UUID   `json:"id"`
	GoalID        uuid.UUID   `json:"goal_id"`
	EmployeeID    uuid.UUID   `json:"employee_id"`
	Weight        float64     `json:"weight"`
	TargetValue   float64     `json:"target_value"`
	BaselineValue *float64    `json:"baseline_value,omitempty"`
}

// GlobalRuleRow represents a mass assignment rule.
type GlobalRuleRow struct {
	ID               uuid.UUID  `json:"id"`
	GoalID           uuid.UUID  `json:"goal_id"`
	RuleType         string     `json:"rule_type"`
	DepartmentID     *uuid.UUID `json:"department_id,omitempty"`
	MinDirectReports *int       `json:"min_direct_reports,omitempty"`
	DefaultWeight    float64    `json:"default_weight"`
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
			SetDefaultWeight(rule.DefaultWeight)

		if rule.DepartmentID != nil {
			create = create.SetDepartmentID(*rule.DepartmentID)
		}
		if rule.MinDirectReports != nil {
			create = create.SetMinDirectReports(*rule.MinDirectReports)
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
		GoalKind:    string(*g.GoalKind),
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
			ID:           rule.ID,
			GoalID:       rule.GoalID,
			RuleType:     string(rule.RuleType),
			DefaultWeight: rule.DefaultWeight,
		}
		if rule.DepartmentID != nil {
			ruleRow.DepartmentID = rule.DepartmentID
		}
		if rule.MinDirectReports != nil {
			ruleRow.MinDirectReports = rule.MinDirectReports
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
			GoalKind:    string(*g.GoalKind),
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
	rules, err := r.client.GlobalGoalRule.Query().
		Where(globalgoalrule.GoalID(goalID)).
		All(ctx)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, rule := range rules {
		var employees []*internal.Employee

		switch rule.RuleType {
		case globalgoalrule.RuleTypeDepartment:
			if rule.DepartmentID == nil {
				continue
			}
			// Query employees in the department
			employees, err = r.client.Employee.Query().
				Where(employee.OrgNodeID(*rule.DepartmentID)).
				All(ctx)
		case globalgoalrule.RuleTypeMinDirectReports:
			if rule.MinDirectReports == nil {
				continue
			}
			// Query all active employees (simplified - in real implementation, 
			// you'd need a more complex query to count direct reports)
			employees, err = r.client.Employee.Query().
				Where(employee.IsActive(true)).
				All(ctx)
		}

		if err != nil {
			return count, err
		}

		for _, emp := range employees {
			// Check if assignment already exists
			exists, _ := r.client.GlobalGoalAssignment.Query().
				Where(
					globalgoalassignment.GoalID(goalID),
					globalgoalassignment.EmployeeID(emp.ID),
				).Exist(ctx)

			if !exists {
				_, err = r.client.GlobalGoalAssignment.Create().
					SetGoalID(goalID).
					SetEmployeeID(emp.ID).
					SetWeight(rule.DefaultWeight).
					SetTargetValue(100). // Default target
					Save(ctx)
				if err != nil {
					return count, err
				}
				count++
			}
		}
	}

	return count, nil
}
