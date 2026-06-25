package scoring

// GoalScore represents a single goal's contribution within a category.
type GoalScore struct {
	// Weight is the goal's weight within its category (0-100).
	Weight float64
	// ProgressPercent is the progress percentage for this goal (0-100).
	ProgressPercent float64
}

// CategoryScore represents a category of goals with an overall weight.
type CategoryScore struct {
	// Weight is the category's weight in the overall evaluation (0-100).
	Weight float64
	// Goals is the list of goal scores within this category.
	Goals []GoalScore
}

// EmployeeScore calculates the overall weighted score for an employee
// based on category and goal weights.
//
// Formula:
//
//	score = Σ(cat_weight / 100 × Σ(goal_weight / 100 × progress%))
//
// Returns a value in the range [0, 100].
func EmployeeScore(categories []CategoryScore) float64 {
	var total float64

	for _, cat := range categories {
		var catScore float64
		for _, g := range cat.Goals {
			catScore += (g.Weight / 100) * g.ProgressPercent
		}
		total += (cat.Weight / 100) * catScore
	}

	return total
}
