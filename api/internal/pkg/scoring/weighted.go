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

// HierarchicalScore applies hierarchical weighting P/100 * PJ/100 to personalScore.
// If pWeight or pjWeight is 0, it falls back to 100 (no reduction).
// Result is clamped to [0, 100].
func HierarchicalScore(personalScore float64, pWeight, pjWeight float64) float64 {
	if pWeight == 0 {
		pWeight = 100
	}
	if pjWeight == 0 {
		pjWeight = 100
	}
	result := personalScore * (pWeight / 100) * (pjWeight / 100)
	if result < 0 {
		return 0
	}
	if result > 100 {
		return 100
	}
	return result
}

// EffectiveWeightPersonal returns peso ponderado personal: w * (P/100) * (PJ/100) fallback 100.
func EffectiveWeightPersonal(w, pWeight, pjWeight float64) float64 {
	if pWeight == 0 {
		pWeight = 100
	}
	if pjWeight == 0 {
		pjWeight = 100
	}
	return w * (pWeight / 100) * (pjWeight / 100)
}

// EffectiveWeightGlobal returns peso ponderado global: w * (G/100) where G=100-P fallback 100.
func EffectiveWeightGlobal(w, pWeight float64) float64 {
	if pWeight == 0 {
		pWeight = 100
	}
	g := 100 - pWeight
	if g < 0 {
		g = 0
	}
	return w * (g / 100)
}

// EffectiveWeightShared returns peso ponderado compartida: w * (J/100) * (P/100) where J=100-PJ fallback 100.
func EffectiveWeightShared(w, pWeight, pjWeight float64) float64 {
	if pWeight == 0 {
		pWeight = 100
	}
	if pjWeight == 0 {
		pjWeight = 100
	}
	j := 100 - pjWeight
	if j < 0 {
		j = 0
	}
	return w * (j / 100) * (pWeight / 100)
}
