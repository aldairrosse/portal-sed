import type {
	NineBoxEntry,
	NineBoxScale,
	NineBoxQuadrant,
	NineBoxQuadrantDef
} from '$lib/types/nine-box';

import matrixEntriesData from '$lib/fixtures/nine-box/matrix-entries.json';
import quadrantDefsData from '$lib/fixtures/nine-box/quadrant-definitions.json';

// ─── Quadrant calculation ─────────────────────────────────────────────────────

const QUADRANT_MAP: Record<string, NineBoxQuadrant> = {
	'high-high': 'star',
	'high-mid': 'growth',
	'mid-high': 'high-potential',
	'mid-mid': 'core-player',
	'low-high': 'risk',
	'low-mid': 'effective',
	'low-low': 'underperformer',
	'mid-low': 'effective',
	'high-low': 'growth'
};

function perfBand(n: number): 'low' | 'mid' | 'high' {
	if (n <= 3) return 'low';
	if (n <= 6) return 'mid';
	return 'high';
}

function potBand(n: number): 'low' | 'mid' | 'high' {
	if (n <= 3) return 'low';
	if (n <= 6) return 'mid';
	return 'high';
}

export function computeQuadrant(perf: number, pot: number): NineBoxQuadrant {
	const key = `${perfBand(perf)}-${potBand(pot)}`;
	return QUADRANT_MAP[key] ?? 'core-player';
}

// ─── State ────────────────────────────────────────────────────────────────────

let entries = $state<NineBoxEntry[]>(structuredClone(matrixEntriesData as NineBoxEntry[]));
const quadrantDefs = $state<NineBoxQuadrantDef[]>(
	structuredClone(quadrantDefsData as NineBoxQuadrantDef[])
);

// ─── Getters ──────────────────────────────────────────────────────────────────

export function getAllEntries(): NineBoxEntry[] {
	return entries;
}

export function getMatrixEntries(scopeIds: string[]): NineBoxEntry[] {
	if (scopeIds.length === 0) return [];
	return entries.filter((e) => scopeIds.includes(e.employeeId));
}

export function getEntryByEmployee(employeeId: string): NineBoxEntry | undefined {
	return entries.find((e) => e.employeeId === employeeId);
}

export function getQuadrantDefs(): NineBoxQuadrantDef[] {
	return quadrantDefs;
}

export function getQuadrantForScores(
	perf: NineBoxScale,
	pot: NineBoxScale
): NineBoxQuadrant {
	return computeQuadrant(perf, pot);
}

export function getEntriesByQuadrant(
	scopeIds: string[],
	quadrant: NineBoxQuadrant
): NineBoxEntry[] {
	const scoped = scopeIds.length > 0 ? getMatrixEntries(scopeIds) : entries;
	return scoped.filter((e) => e.quadrant === quadrant);
}

export function getQuadrantStats(
	scopeIds: string[]
): Record<NineBoxQuadrant, number> {
	const scoped = getMatrixEntries(scopeIds);
	const stats: Record<string, number> = {
		star: 0,
		growth: 0,
		'high-potential': 0,
		'core-player': 0,
		risk: 0,
		effective: 0,
		underperformer: 0
	};
	for (const entry of scoped) {
		stats[entry.quadrant] = (stats[entry.quadrant] ?? 0) + 1;
	}
	return stats as Record<NineBoxQuadrant, number>;
}

// ─── Mutations ────────────────────────────────────────────────────────────────

export function setEntryScores(
	employeeId: string,
	performance: NineBoxScale,
	potential: NineBoxScale
): void {
	entries = entries.map((e) =>
		e.employeeId === employeeId
			? { ...e, performance, potential, quadrant: computeQuadrant(performance, potential) }
			: e
	);
}

export function bulkSetEntries(newEntries: NineBoxEntry[]): void {
	entries = structuredClone(newEntries);
}
