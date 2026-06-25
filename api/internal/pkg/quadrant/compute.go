// Package quadrant provides pure, deterministic computations for the 9-box
// matrix: performance tiers from goal progress, potential tiers from
// competency ratings, and quadrant mapping from tier pairs.
//
// Quadrant numbering (potential tier as rows, performance tier as columns):
//
//	                  Performance
//	                Low(1) Med(2) High(3)
//	Pot High(3)   |   7  |   8  |   9   |
//	Pot Med(2)    |   4  |   5  |   6   |
//	Pot Low(1)    |   1  |   2  |   3   |
//
// This numbering MUST match the seed data in the NineBoxQuadrant catalog table.
package quadrant

// ComputePerformanceTier maps average goal progress (0–100) to a tier 1–3.
//
//	< 34%   → 1 (low)
//	34–66%  → 2 (medium)
//	≥ 67%   → 3 (high)
//
// This uses < 34 as the tier-1 boundary so that e.g. 33.9% is still tier 1,
// matching the spec's "≤ 33%" integer intent with continuous float inputs.
func ComputePerformanceTier(avgProgress float64) int {
	if avgProgress < 34 {
		return 1
	}
	if avgProgress < 67 {
		return 2
	}
	return 3
}

// ComputePotentialTier maps competency ratings (scale 1–5) to a tier 1–3.
//
//	1.00–2.33 → 1 (low)
//	2.34–3.66 → 2 (medium)
//	3.67–5.00 → 3 (high)
//
// If both selfRating and hrRating are non-nil, their average is used.
// If only one is provided, that value is used.
// If neither is provided, returns 2 (default / unknown).
func ComputePotentialTier(selfRating, hrRating *float64) int {
	if selfRating == nil && hrRating == nil {
		return 2
	}

	var sum, count float64
	if selfRating != nil {
		sum += *selfRating
		count++
	}
	if hrRating != nil {
		sum += *hrRating
		count++
	}

	avg := sum / count

	switch {
	case avg <= 2.33:
		return 1
	case avg <= 3.66:
		return 2
	default:
		return 3
	}
}

// ComputeQuadrant maps performance (1–9) and potential (1–9) scores to a
// quadrant (1–9). It is a pure, deterministic function with no side effects.
//
// Returns 0 if either score is out of range.
func ComputeQuadrant(performance, potential int) int {
	if performance < 1 || performance > 9 || potential < 1 || potential > 9 {
		return 0
	}
	perfTier := tier(performance) // 1=low, 2=med, 3=high
	potTier := tier(potential)    // 1=low, 2=med, 3=high
	return (potTier-1)*3 + perfTier
}

// ComputeQuadrantFromTiers maps a (performanceTier, potentialTier) pair (1–3)
// to a quadrant (1–9). Uses the same formula as ComputeQuadrant:
//
//	quadrant = (potTier-1)*3 + perfTier
//
// Returns 0 if either tier is out of range.
func ComputeQuadrantFromTiers(perfTier, potTier int) int {
	if perfTier < 1 || perfTier > 3 || potTier < 1 || potTier > 3 {
		return 0
	}
	return (potTier-1)*3 + perfTier
}

// tier maps a score (1–9) to a tier:
//   - 1–3 → 1 (low)
//   - 4–6 → 2 (medium)
//   - 7–9 → 3 (high)
func tier(score int) int {
	switch {
	case score <= 3:
		return 1
	case score <= 6:
		return 2
	default:
		return 3
	}
}
