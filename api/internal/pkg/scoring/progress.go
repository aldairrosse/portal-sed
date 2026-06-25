// Package scoring provides functions to calculate progress percentages
// and weighted scores for employee evaluations.
package scoring

import "math"

// ProgressPercent calculates how much progress has been made toward a target.
//
// For ascendente (higher is better):
//
//	progress = min(current / target * 100, 100), clamped to [0, 100]
//
// For descendente (lower is better):
//
//	progress = min((baseline - current) / (baseline - target) * 100, 100), clamped to [0, 100]
//
// Edge cases:
//   - target == 0 returns 0 (no target defined = no progress)
//   - descendente with baseline == target returns 0 (division by zero guard)
//   - results are clamped to the [0, 100] range.
func ProgressPercent(currentValue, targetValue, baselineValue float64, direction string) float64 {
	if targetValue == 0 {
		return 0
	}

	var pct float64
	switch direction {
	case "descendente":
		if baselineValue == targetValue {
			return 0
		}
		pct = (baselineValue - currentValue) / (baselineValue - targetValue) * 100
	default: // "ascendente" or any other value
		pct = currentValue / targetValue * 100
	}

	return math.Max(0, math.Min(pct, 100))
}
