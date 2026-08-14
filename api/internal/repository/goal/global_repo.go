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
		SetNillableCategoryID(nil). // Global goals don't belong to a category
		SetCreatedBy(createdBy).
		SetUpdatedBy(createdBy).
		SetType(goal.TypeGlobal).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	// Persist rules.
	for _, rule := range rules {
		if err := r.createRule(ctx, g.ID, rule); err != nil {
			return nil, err
		}
	}

	// Materialize assignments (manual + rule-matched).
	if _, err := r.evaluateAndAssign(ctx, g.ID, rules, assignments); err != nil {
		return nil, err
	}

	return r.GetGlobalGoal(ctx, g.ID)
}

// createRule persists a single global goal rule.
func (r *GlobalGoalRepo) createRule(ctx context.Context, goalID uuid.UUID, rule *GlobalRuleRow) error {
	create := r.client.GlobalGoalRule.Create().
		SetGoalID(goalID).
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

	_, err := create.Save(ctx)
	return err
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

	row := goalRowFromEnt(g)
	return row, nil
}

// goalRowFromEnt builds a GlobalGoalRow from an ent Goal entity, populating
// Assignments and Rules from the preloaded edges.
func goalRowFromEnt(g *internal.Goal) *GlobalGoalRow {
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
		row.Rules = append(row.Rules, ruleRowFromEnt(rule))
	}

	return row
}

// ruleRowFromEnt converts an ent GlobalGoalRule entity into a GlobalRuleRow.
func ruleRowFromEnt(rule *internal.GlobalGoalRule) *GlobalRuleRow {
	row := &GlobalRuleRow{
		ID:            rule.ID,
		GoalID:        rule.GoalID,
		RuleType:      string(rule.RuleType),
		DefaultWeight: rule.DefaultWeight,
		DefaultTarget: rule.DefaultTarget,
	}
	if rule.DepartmentID != nil {
		row.DepartmentID = rule.DepartmentID
	}
	if rule.MinDirectReports != nil {
		row.MinDirectReports = rule.MinDirectReports
	}
	if rule.ProfileID != nil {
		row.ProfileID = rule.ProfileID
	}
	return row
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
		rows = append(rows, goalRowFromEnt(g))
	}

	return rows, nil
}

// UpdateGlobalGoal updates a global goal and re-applies its rules and assignments.
func (r *GlobalGoalRepo) UpdateGlobalGoal(ctx context.Context, goalID uuid.UUID, name, description, unit, direction, goalKind string, weight, targetValue float64, assignments []*GlobalAssignmentRow, rules []*GlobalRuleRow) (*GlobalGoalRow, error) {
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

	// Remove the goal from everyone, then re-apply rules and assignments.
	if _, err := r.client.GlobalGoalAssignment.Delete().Where(globalgoalassignment.GoalID(goalID)).Exec(ctx); err != nil {
		return nil, err
	}
	if _, err := r.client.GlobalGoalRule.Delete().Where(globalgoalrule.GoalID(goalID)).Exec(ctx); err != nil {
		return nil, err
	}

	for _, rule := range rules {
		if err := r.createRule(ctx, goalID, rule); err != nil {
			return nil, err
		}
	}

	if _, err := r.evaluateAndAssign(ctx, goalID, rules, assignments); err != nil {
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

	entRules, err := r.client.GlobalGoalRule.Query().Where(globalgoalrule.GoalID(goalID)).Order(internal.Asc(globalgoalrule.FieldID)).All(ctx)
	if err != nil {
		return 0, err
	}
	if len(entRules) == 0 {
		return 0, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "La meta no tiene reglas de asignación.", nil)
	}

	rules := make([]*GlobalRuleRow, 0, len(entRules))
	for _, rule := range entRules {
		rules = append(rules, ruleRowFromEnt(rule))
	}

	return r.evaluateAndAssign(ctx, goalID, rules, nil)
}

// evaluateAndAssign materializes global_goal_assignments rows for a goal by
// evaluating its rules over the organization's active employees and unioning the
// result with the manual assignments. Rule semantics: departments ∩ min_direct_reports ∩ roles,
// with OR within each category and identity for categories without rules.
// Manual assignments always win. It is idempotent: an already-assigned employee is
// never duplicated. It returns the number of newly created assignments.
func (r *GlobalGoalRepo) evaluateAndAssign(ctx context.Context, goalID uuid.UUID, rules []*GlobalRuleRow, manualAssignments []*GlobalAssignmentRow) (int, error) {
	// Resolve the goal's organization from its creator.
	creator, err := r.client.Goal.Query().Where(goal.ID(goalID)).Only(ctx)
	if err != nil {
		return 0, err
	}
	creatorEmp, err := r.client.Employee.Query().Where(employee.ID(creator.CreatedBy)).WithOrgNode().Only(ctx)
	if err != nil {
		return 0, err
	}
	if creatorEmp.Edges.OrgNode == nil {
		return 0, pkgerrors.NewDomainError(pkgerrors.InvalidRequest, "La organización del creador no está disponible.", nil)
	}
	orgID := creatorEmp.Edges.OrgNode.OrganizationID

	employees, err := r.client.Employee.Query().Where(employee.IsActive(true)).WithOrgNode().All(ctx)
	if err != nil {
		return 0, err
	}
	orgEmployees := make([]*internal.Employee, 0, len(employees))
	for _, emp := range employees {
		if emp.Edges.OrgNode != nil && emp.Edges.OrgNode.OrganizationID == orgID {
			orgEmployees = append(orgEmployees, emp)
		}
	}

	assigned := 0
	assignedSet := make(map[uuid.UUID]bool)

	// Manual assignments first: they always win over rule-matched employees.
	for _, a := range manualAssignments {
		if assignedSet[a.EmployeeID] {
			continue
		}
		exists, err := r.client.GlobalGoalAssignment.Query().
			Where(globalgoalassignment.GoalID(goalID), globalgoalassignment.EmployeeID(a.EmployeeID)).
			Exist(ctx)
		if err != nil {
			return 0, err
		}
		if exists {
			assignedSet[a.EmployeeID] = true
			continue
		}
		create := r.client.GlobalGoalAssignment.Create().
			SetGoalID(goalID).
			SetEmployeeID(a.EmployeeID).
			SetWeight(a.Weight).
			SetTargetValue(a.TargetValue)
		if a.BaselineValue != nil {
			create = create.SetBaselineValue(*a.BaselineValue)
		}
		if _, err := create.Save(ctx); err != nil {
			return 0, err
		}
		assignedSet[a.EmployeeID] = true
		assigned++
	}

	if len(rules) == 0 {
		return assigned, nil
	}

	// Categorize rules: OR within a category, AND across categories.
	var deptRules, minReportRules, roleRules []*GlobalRuleRow
	for _, rule := range rules {
		switch rule.RuleType {
		case string(globalgoalrule.RuleTypeDepartment):
			deptRules = append(deptRules, rule)
		case string(globalgoalrule.RuleTypeMinDirectReports):
			minReportRules = append(minReportRules, rule)
		case string(globalgoalrule.RuleTypeRole):
			roleRules = append(roleRules, rule)
		}
	}

	// Precompute department ltree paths.
	departmentPaths := make(map[uuid.UUID]string, len(deptRules))
	for _, rule := range deptRules {
		if rule.DepartmentID == nil {
			continue
		}
		if _, ok := departmentPaths[*rule.DepartmentID]; ok {
			continue
		}
		node, err := r.client.OrgNode.Get(ctx, *rule.DepartmentID)
		if err != nil {
			return 0, err
		}
		departmentPaths[*rule.DepartmentID] = node.Path
	}

	// Max threshold across min_direct_reports rules.
	var minReports *int
	for _, rule := range minReportRules {
		if rule.MinDirectReports == nil {
			continue
		}
		if minReports == nil || *rule.MinDirectReports > *minReports {
			v := *rule.MinDirectReports
			minReports = &v
		}
	}

	// Direct report counts, cached per employee.
	reportCounts := make(map[uuid.UUID]int)
	directReports := func(emp *internal.Employee) int {
		if c, ok := reportCounts[emp.ID]; ok {
			return c
		}
		count := 0
		for _, report := range orgEmployees {
			if report.Edges.OrgNode == nil {
				continue
			}
			managerMatch := report.ManagerID != nil && *report.ManagerID == emp.ID
			departmentMatch := report.Edges.OrgNode.ParentID != nil && *report.Edges.OrgNode.ParentID == emp.OrgNodeID
			if managerMatch || departmentMatch {
				count++
			}
		}
		reportCounts[emp.ID] = count
		return count
	}

	for _, emp := range orgEmployees {
		if emp.Edges.OrgNode == nil || assignedSet[emp.ID] {
			continue
		}

		// Department filter (OR across department rules).
		if len(deptRules) > 0 {
			matched := false
			for _, rule := range deptRules {
				if rule.DepartmentID == nil {
					continue
				}
				deptPath := departmentPaths[*rule.DepartmentID]
				employeePath := emp.Edges.OrgNode.Path
				if employeePath == deptPath || strings.HasPrefix(employeePath, deptPath+".") {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		// Role filter (OR across role rules).
		if len(roleRules) > 0 {
			matched := false
			for _, rule := range roleRules {
				if rule.ProfileID != nil && emp.ProfileID == *rule.ProfileID {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		// Min direct reports filter (single max threshold).
		if len(minReportRules) > 0 && minReports != nil && directReports(emp) < *minReports {
			continue
		}

		// Rule weight/target precedence: department → min_direct_reports → role.
		weight, target := 0.0, 100.0
		switch {
		case len(deptRules) > 0:
			weight, target = deptRules[0].DefaultWeight, deptRules[0].DefaultTarget
		case len(minReportRules) > 0:
			weight, target = minReportRules[0].DefaultWeight, minReportRules[0].DefaultTarget
		case len(roleRules) > 0:
			weight, target = roleRules[0].DefaultWeight, roleRules[0].DefaultTarget
		}

		exists, err := r.client.GlobalGoalAssignment.Query().
			Where(globalgoalassignment.GoalID(goalID), globalgoalassignment.EmployeeID(emp.ID)).
			Exist(ctx)
		if err != nil {
			return 0, err
		}
		if exists {
			assignedSet[emp.ID] = true
			continue
		}
		if _, err := r.client.GlobalGoalAssignment.Create().
			SetGoalID(goalID).
			SetEmployeeID(emp.ID).
			SetWeight(weight).
			SetTargetValue(target).
			Save(ctx); err != nil {
			return 0, err
		}
		assignedSet[emp.ID] = true
		assigned++
	}

	return assigned, nil
}
