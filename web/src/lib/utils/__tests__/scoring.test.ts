import { describe, it, expect } from 'vitest';
import { progressPercent, deltaIndicator, formatDelta } from '../scoring';

// ─── progressPercent ─────────────────────────────────────────────────────────────

describe('progressPercent', () => {
	describe('ascendente', () => {
		it('returns 50 when current is halfway to target', () => {
			expect(progressPercent(50, 100, undefined, 'ascendente')).toBe(50);
		});

		it('returns 100 when target is met', () => {
			expect(progressPercent(100, 100, undefined, 'ascendente')).toBe(100);
		});

		it('caps at 100 when exceeding target', () => {
			expect(progressPercent(150, 100, undefined, 'ascendente')).toBe(100);
		});

		it('returns 0 when target is 0 (division guard)', () => {
			expect(progressPercent(50, 0, undefined, 'ascendente')).toBe(0);
		});

		it('clamps negative percentage to 0', () => {
			expect(progressPercent(-10, 100, undefined, 'ascendente')).toBe(0);
		});

		it('ignores baseline value', () => {
			expect(progressPercent(75, 100, 50, 'ascendente')).toBe(75);
		});

		it('returns 0 when current is 0', () => {
			expect(progressPercent(0, 100, undefined, 'ascendente')).toBe(0);
		});
	});

	describe('descendente', () => {
		it('returns 50 when halfway from baseline to target', () => {
			// baseline=100, target=0, current=50 → (100-50)/(100-0)*100 = 50
			expect(progressPercent(50, 0, 100, 'descendente')).toBe(50);
		});

		it('returns 100 when target is reached (target=0)', () => {
			// baseline=100, target=0, current=0 → (100-0)/(100-0)*100 = 100
			expect(progressPercent(0, 0, 100, 'descendente')).toBe(100);
		});

		it('returns 100 when at target with non-zero target', () => {
			// baseline=100, target=50, current=50 → (100-50)/(100-50)*100 = 100
			expect(progressPercent(50, 50, 100, 'descendente')).toBe(100);
		});

		it('caps at 100 when current is below target', () => {
			// baseline=100, target=50, current=30 → (100-30)/(100-50)*100 = 140 → capped to 100
			expect(progressPercent(30, 50, 100, 'descendente')).toBe(100);
		});

		it('clamps to 0 when current is above baseline (worsening)', () => {
			// baseline=100, target=50, current=120 → (100-120)/(100-50)*100 = -40 → clamped to 0
			expect(progressPercent(120, 50, 100, 'descendente')).toBe(0);
		});

		it('returns 0 when baseline is undefined', () => {
			expect(progressPercent(50, 0, undefined, 'descendente')).toBe(0);
		});

		it('returns 0 when baseline equals target', () => {
			expect(progressPercent(50, 100, 100, 'descendente')).toBe(0);
		});


	});
});

// ─── deltaIndicator ──────────────────────────────────────────────────────────────

describe('deltaIndicator', () => {
	describe('ascendente', () => {
		it('returns positive delta when exceeding target', () => {
			const result = deltaIndicator(110, 100, undefined, 'ascendente');
			expect(result).toEqual({ value: 10, type: 'positive' });
		});

		it('returns negative delta when below target', () => {
			const result = deltaIndicator(90, 100, undefined, 'ascendente');
			expect(result).toEqual({ value: -10, type: 'negative' });
		});

		it('returns neutral when exactly at target', () => {
			const result = deltaIndicator(100, 100, undefined, 'ascendente');
			expect(result).toEqual({ value: 0, type: 'neutral' });
		});
	});

	describe('descendente', () => {
		it('returns positive delta when current is below baseline (improving)', () => {
			// baseline=100, current=30 → 100-30 = 70
			const result = deltaIndicator(30, 50, 100, 'descendente');
			expect(result).toEqual({ value: 70, type: 'positive' });
		});

		it('returns negative delta when current is above baseline (worsening)', () => {
			// baseline=100, current=120 → 100-120 = -20
			const result = deltaIndicator(120, 50, 100, 'descendente');
			expect(result).toEqual({ value: -20, type: 'negative' });
		});

		it('returns neutral when current equals baseline', () => {
			const result = deltaIndicator(100, 50, 100, 'descendente');
			expect(result).toEqual({ value: 0, type: 'neutral' });
		});

		it('uses 0 as baseline when baseline is undefined', () => {
			const result = deltaIndicator(30, 50, undefined, 'descendente');
			expect(result).toEqual({ value: -30, type: 'negative' });
		});
	});
});

// ─── formatDelta ─────────────────────────────────────────────────────────────────

describe('formatDelta', () => {
	it('formats positive delta with + prefix', () => {
		expect(formatDelta(15, 'positive')).toBe('+15');
	});

	it('formats negative delta with - prefix', () => {
		expect(formatDelta(-8, 'negative')).toBe('-8');
	});

	it('returns empty string for neutral', () => {
		expect(formatDelta(0, 'neutral')).toBe('');
	});
});
