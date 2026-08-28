/**
 * Scoring utilities for calculating goal/KPI progress and delta indicators.
 *
 * Mirrors the backend scoring logic (see api/internal/scoring/).
 *
 * @module scoring
 */

/**
 * Calculate the progress percentage for a goal or KPI.
 *
 * - `ascendente`: `min(current / target * 100, 100)` — higher is better.
 * - `descendente`: `min((baseline - current) / (baseline - target) * 100, 100)` — lower is better.
 *
 * Result is always clamped between 0 and 100.
 *
 * @param current  - Current measured value.
 * @param target   - Target value to reach.
 * @param baseline - Baseline/reference value (required for `descendente`, ignored for `ascendente`).
 * @param direction - Direction of improvement.
 * @returns Progress percentage clamped to 0–100.
 */
export function progressPercent(
	current: number,
	target: number,
	baseline: number | undefined,
	direction: 'ascendente' | 'descendente'
): number {
	let pct: number;

	if (direction === 'ascendente') {
		// Guard: division by zero
		if (target === 0) return 0;
		pct = (current / target) * 100;
	} else {
		// descendente requires a baseline
		if (baseline === undefined) return 0;
		const range = baseline - target;
		if (range === 0) return 0;
		pct = ((baseline - current) / range) * 100;
	}

	return Math.min(Math.max(pct, 0), 100);
}

/**
 * Compute the absolute delta and its sign type for display.
 *
 * - `ascendente`: `delta = current - target` (positive = exceeding target).
 * - `descendente`: `delta = baseline - current` (positive = improvement from baseline).
 *
 * @param current  - Current measured value.
 * @param target   - Target value to reach.
 * @param baseline - Baseline/reference value (used for `descendente`, optional for `ascendente`).
 * @param direction - Direction of improvement.
 * @returns Object with numeric `value` and display `type`.
 */
export function deltaIndicator(
	current: number,
	target: number,
	baseline: number | undefined,
	direction: 'ascendente' | 'descendente'
): { value: number; type: 'positive' | 'negative' | 'neutral' } {
	const delta =
		direction === 'ascendente'
			? current - target
			: (baseline ?? 0) - current;

	let type: 'positive' | 'negative' | 'neutral';
	if (delta > 0) type = 'positive';
	else if (delta < 0) type = 'negative';
	else type = 'neutral';

	return { value: delta, type };
}

/**
 * Format a delta value as a human-readable string with sign prefix.
 *
 * @param delta - Numeric delta value.
 * @param type  - Sign type from `deltaIndicator`.
 * @returns Formatted string: `"+15"`, `"-8"`, or `""`.
 */
export function formatDelta(
	delta: number,
	type: 'positive' | 'negative' | 'neutral'
): string {
	if (type === 'neutral') return '';
	if (type === 'positive') return `+${delta}`;
	return `${delta}`;
}

/**
 * Hierarchical weighted score: personal * P/100 * PJ/100 with fallback 100.
 * Mirrors backend scoring.HierarchicalScore.
 */
export function hierarchicalScore(personalScore: number, pWeight = 100, pjWeight = 100): number {
	if (pWeight === 0) pWeight = 100;
	if (pjWeight === 0) pjWeight = 100;
	const r = personalScore * (pWeight / 100) * (pjWeight / 100);
	return Math.min(Math.max(r, 0), 100);
}

export function effectiveWeightPersonal(w: number, pWeight = 100, pjWeight = 100): number {
	if (!pWeight) pWeight = 100;
	if (!pjWeight) pjWeight = 100;
	return w * (pWeight / 100) * (pjWeight / 100);
}
export function effectiveWeightGlobal(w: number, pWeight = 100): number {
	if (!pWeight) pWeight = 100;
	const g = 100 - pWeight;
	return w * (Math.max(0, g) / 100);
}
export function effectiveWeightShared(w: number, pWeight = 100, pjWeight = 100): number {
	if (!pWeight) pWeight = 100;
	if (!pjWeight) pjWeight = 100;
	const j = 100 - pjWeight;
	return w * (Math.max(0, j) / 100) * (pWeight / 100);
}
