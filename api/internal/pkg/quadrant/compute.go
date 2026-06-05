// Package quadrant provides a pure, deterministic computation of the 9×9
// matrix quadrant from performance and potential scores.
//
// Quadrant numbering (potential tier as rows, performance tier as columns):
//
//	                  Performance
//	                Low(1-3) Med(4-6) High(7-9)
//	Pot High(7-9)  |   7   |   8   |    9   |
//	Pot Med(4-6)   |   4   |   5   |    6   |
//	Pot Low(1-3)   |   1   |   2   |    3   |
//
// This numbering MUST match the seed data in the NineBoxQuadrant catalog table.
package quadrant

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
