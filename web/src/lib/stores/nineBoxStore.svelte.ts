import type {
	NineBoxEntry,
	NineBoxScale,
	NineBoxQuadrant,
	NineBoxQuadrantDef
} from '$lib/types/nine-box';
import type { components } from '$lib/api/schemas/evaluations';

import matrixEntriesData from '$lib/fixtures/nine-box/matrix-entries.json';
import quadrantDefsData from '$lib/fixtures/nine-box/quadrant-definitions.json';
import { client } from '$lib/api/client';

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

// ─── Internal data shape ──────────────────────────────────────────────────────

interface StoreData {
	entries: NineBoxEntry[];
	quadrantDefs: NineBoxQuadrantDef[];
}

// ─── Triplete state ───────────────────────────────────────────────────────────

let data = $state<StoreData | null>(null);
let loading = $state(true);
let error = $state<string | null>(null);

// ─── Helpers ──────────────────────────────────────────────────────────────────

function loadFixtures(): StoreData {
	return {
		entries: structuredClone(matrixEntriesData as NineBoxEntry[]),
		quadrantDefs: structuredClone(quadrantDefsData as NineBoxQuadrantDef[])
	};
}

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
		employeeName: '',
		profileId: '',
		performance: (dto.performanceScore ?? 5) as NineBoxScale,
		potential: (dto.potentialScore ?? 5) as NineBoxScale,
		quadrant: computeQuadrant(dto.performanceScore ?? 5, dto.potentialScore ?? 5)
	}));

	const quadrantDefs: NineBoxQuadrantDef[] = apiQuadrants.map((dto) => ({
		id: (dto.label?.toLowerCase().replace(/\s+/g, '-') ?? 'core-player') as NineBoxQuadrant,
		label: dto.label ?? '',
		description: dto.description ?? '',
		colorClass: '',
		perfRange: [1, 9],
		potRange: [1, 9]
	}));

	return { entries, quadrantDefs };
}

/**
 * Load nine-box data.
 *
 * In DEV without VITE_USE_API: loads from fixture files (structured clone).
 * In production / VITE_USE_API=true: fetches from the real API endpoints.
 */
export async function load(): Promise<void> {
	loading = true;
	error = null;

	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		data = loadFixtures();
		loading = false;
		return;
	}

	try {
		const [matricesRes, quadrantsRes] = await Promise.all([
			client.GET('/nine-box/matrices', {}),
			client.GET('/nine-box/quadrants', {})
		]);

		if (matricesRes.error) {
			throw new Error(
				(matricesRes.error as { error?: { message?: string } })?.error?.message ??
					'Error al cargar matrices 9×9'
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
		error = e instanceof Error ? e.message : 'Error desconocido al cargar datos de matriz 9×9';
	} finally {
		loading = false;
	}
}

/** Alias for load(). */
export function reload(): Promise<void> {
	return load();
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
	return data?.quadrantDefs ?? [];
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
	const scoped = scopeIds.length > 0 ? getMatrixEntries(scopeIds) : (data?.entries ?? []);
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

export async function setEntryScores(
	employeeId: string,
	performance: NineBoxScale,
	potential: NineBoxScale
): Promise<void> {
	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		data = {
			...data!,
			entries: (data?.entries ?? []).map((e) =>
				e.employeeId === employeeId
					? { ...e, performance, potential, quadrant: computeQuadrant(performance, potential) }
					: e
			)
		};
		return;
	}

	const entry = (data?.entries ?? []).find((e) => e.employeeId === employeeId);

	if (entry?.id) {
		const { error: apiError } = await client.PUT('/nine-box/entries/{entryId}', {
			params: { path: { entryId: entry.id } },
			body: {
				evaluateeId: employeeId,
				performanceScore: performance,
				potentialScore: potential
			},
			headers: { 'If-Match': 1 } as Record<string, number>
		});
		if (apiError) {
			throw new Error(
				(apiError as { error?: { message?: string } })?.error?.message ??
					'Error al actualizar entrada 9×9'
			);
		}
	} else {
		// New entry — local-only until we have matrix context
		data = {
			...data!,
			entries: [
				...(data?.entries ?? []),
				{
					id: crypto.randomUUID(),
					employeeId,
					employeeName: '',
					profileId: '',
					performance,
					potential,
					quadrant: computeQuadrant(performance, potential)
				}
			]
		};
		return;
	}

	await reload();
}

export async function bulkSetEntries(newEntries: NineBoxEntry[]): Promise<void> {
	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		data = { ...data!, entries: structuredClone(newEntries) };
		return;
	}

	// API mode: local-only for now — batch endpoint needs matrix context
	data = { ...data!, entries: structuredClone(newEntries) };
}
