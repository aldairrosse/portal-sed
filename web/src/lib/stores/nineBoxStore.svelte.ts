import type {
	NineBoxEntry,
	NineBoxTier,
	NineBoxQuadrantDef
} from '$lib/types/nine-box';
import type { components } from '$lib/api/schemas/evaluations';

import { client } from '$lib/api/client';

// ─── Internal data shape ──────────────────────────────────────────────────────

interface StoreData {
	entries: NineBoxEntry[];
	quadrantDefs: NineBoxQuadrantDef[];
}

// ─── Triplete state ───────────────────────────────────────────────────────────

let data = $state<StoreData | null>(null);
let loading = $state(true);
let error = $state<string | null>(null);

let currentCycleId = $state<string>('');
let currentPhaseId = $state<string>('');

// ─── Helpers ──────────────────────────────────────────────────────────────────

/**
 * Normalize API responses into the flat StoreData format that getters consume.
 */
function normalizeApiData(
	apiEntries: components['schemas']['NineBoxEntryDTO'][],
	apiQuadrants: components['schemas']['NineBoxQuadrantDTO'][]
): StoreData {
	const entries: NineBoxEntry[] = apiEntries.map((dto) => ({
		id: dto.id ?? crypto.randomUUID(),
		employeeId: dto.evaluateeId ?? '',
		employeeName: dto.employeeName ?? '',
		profileId: dto.profileId ?? '',
		performanceTier: (dto.performanceTier ?? 2) as NineBoxTier,
		potentialTier: (dto.potentialTier ?? 2) as NineBoxTier,
		quadrant: dto.quadrant ?? 5
	}));

	const quadrantDefs: NineBoxQuadrantDef[] = apiQuadrants.map((dto) => ({
		quadrant: dto.quadrant ?? 0,
		label: dto.label ?? '',
		title: dto.title ?? '',
		description: dto.description ?? '',
		colorHex: dto.colorHex ?? '#6B7280',
		actionRecommendation: dto.actionRecommendation ?? ''
	}));

	return { entries, quadrantDefs };
}

const DEFAULT_QUADRANT_DEFS: NineBoxQuadrantDef[] = [
	{ quadrant: 1, label: 'Q1', title: 'Rendimiento insuficiente', description: '', colorHex: '#EF4444', actionRecommendation: '' },
	{ quadrant: 2, label: 'Q2', title: 'En riesgo', description: '', colorHex: '#F97316', actionRecommendation: '' },
	{ quadrant: 3, label: 'Q3', title: 'Estrella fugaz', description: '', colorHex: '#EAB308', actionRecommendation: '' },
	{ quadrant: 4, label: 'Q4', title: 'Colaborador sólido', description: '', colorHex: '#A855F7', actionRecommendation: '' },
	{ quadrant: 5, label: 'Q5', title: 'Rendimiento consistente', description: '', colorHex: '#3B82F6', actionRecommendation: '' },
	{ quadrant: 6, label: 'Q6', title: 'Talento emergente', description: '', colorHex: '#06B6D4', actionRecommendation: '' },
	{ quadrant: 7, label: 'Q7', title: 'Potencial sin explotar', description: '', colorHex: '#22C55E', actionRecommendation: '' },
	{ quadrant: 8, label: 'Q8', title: 'Futuro líder', description: '', colorHex: '#14B8A6', actionRecommendation: '' },
	{ quadrant: 9, label: 'Q9', title: 'Estrella', description: '', colorHex: '#6366F1', actionRecommendation: '' }
];

// ─── Loading / error state accessors ──────────────────────────────────────────

export function isLoading(): boolean {
	return loading;
}

export function getError(): string | null {
	return error;
}

export function getCurrentCycleId(): string {
	return currentCycleId;
}

export function getCurrentPhaseId(): string {
	return currentPhaseId;
}

// ─── Load ─────────────────────────────────────────────────────────────────────

/**
 * Load nine-box data for a given cycle and phase.
 *
 * In DEV without VITE_USE_API: loads from fixture files (structured clone).
 * In production / VITE_USE_API=true: fetches from the real API endpoints.
 */
export async function load(cycleId?: string, phaseId?: string): Promise<void> {
	loading = true;
	error = null;

	if (cycleId) currentCycleId = cycleId;
	if (phaseId) currentPhaseId = phaseId;

	try {
		const [matricesRes, quadrantsRes] = await Promise.all([
			client.GET('/nine-box/matrices', {
				params: {
					query: {
						cycle_id: currentCycleId || undefined,
						phase_id: currentPhaseId || undefined
					}
				}
			}),
			client.GET('/nine-box/quadrants', {})
		]);

		if (matricesRes.error) {
			const error = (matricesRes as { error: { error?: { message?: string } }}).error;
			throw new Error(
				error?.error?.message ?? 'Error al cargar matrices'
			);
		}

		const matrixList =
			(matricesRes.data as components['schemas']['NineBoxMatrixResponse'][] | undefined) ?? [];
		const apiEntries: components['schemas']['NineBoxEntryDTO'][] = [];
		for (const matrix of matrixList) {
			if (matrix.entries) {
				apiEntries.push(...matrix.entries);
			}
		}

		const apiQuadrants =
			(quadrantsRes.data as components['schemas']['NineBoxQuadrantDTO'][] | undefined) ?? [];

		data = normalizeApiData(apiEntries, apiQuadrants);
	} catch (e) {
		error = e instanceof Error ? e.message : 'Error desconocido al cargar datos de matriz';
	} finally {
		loading = false;
	}
}

/** Alias for load(). */
export function reload(): Promise<void> {
	return load(currentCycleId, currentPhaseId);
}

/**
 * Force loading to false without fetching data.
 * Use when there's no cycle/phase to load — shows empty grid instead of infinite skeleton.
 */
export function markReady(): void {
	loading = false;
	error = null;
}

// ─── Getters ──────────────────────────────────────────────────────────────────

export function getAllEntries(): NineBoxEntry[] {
	return data?.entries ?? [];
}

export function getMatrixEntries(scopeIds: string[]): NineBoxEntry[] {
	if (scopeIds.length === 0) return [];
	return (data?.entries ?? []).filter((e) => scopeIds.includes(e.employeeId));
}

export function getEntryByEmployee(employeeId: string): NineBoxEntry | undefined {
	return (data?.entries ?? []).find((e) => e.employeeId === employeeId);
}

export function getQuadrantDefs(): NineBoxQuadrantDef[] {
	const defs = data?.quadrantDefs ?? [];
	return [...DEFAULT_QUADRANT_DEFS.filter((_, i) => !defs.some((d) => d.quadrant === i + 1)), ...defs].sort((a, b) => a.quadrant - b.quadrant);
}

export function getQuadrantDef(quadrant: number): NineBoxQuadrantDef | undefined {
	return getQuadrantDefs().find((d) => d.quadrant === quadrant);
}

export function getEntriesByQuadrant(
	scopeIds: string[],
	quadrant: number
): NineBoxEntry[] {
	const scoped = scopeIds.length > 0 ? getMatrixEntries(scopeIds) : (data?.entries ?? []);
	return scoped.filter((e) => e.quadrant === quadrant);
}

export function getQuadrantStats(scopeIds: string[]): Record<number, number> {
	const scoped = getMatrixEntries(scopeIds);
	const stats: Record<number, number> = {};
	for (let i = 1; i <= 9; i++) stats[i] = 0;
	for (const entry of scoped) {
		stats[entry.quadrant] = (stats[entry.quadrant] ?? 0) + 1;
	}
	return stats;
}

// ─── Mutations ────────────────────────────────────────────────────────────────

export async function updateQuadrantDef(
	quadrant: number,
	payload: { title?: string; description?: string; colorHex?: string }
): Promise<void> {
	const { error: apiError } = await client.PUT('/nine-box/quadrants/{quadrant}', {
		params: { path: { quadrant } },
		body: {
			title: payload.title ?? '',
			description: payload.description ?? '',
			colorHex: payload.colorHex ?? '#6B7280'
		}
	});

	if (apiError) {
		throw new Error(
			(apiError as { error?: { message?: string } })?.error?.message ??
				'Error al actualizar cuadrante'
		);
	}

	await reload();
}

export async function recomputeMatrix(cycleId: string, phaseId: string): Promise<void> {
	const { error: apiError } = await client.POST('/nine-box/recompute/{cycleId}/{phaseId}', {
		params: { path: { cycleId, phaseId } }
	});

	if (apiError) {
		throw new Error(
			(apiError as { error?: { message?: string } })?.error?.message ??
				'Error al recalcular matriz'
		);
	}

	await load(cycleId, phaseId);
}
