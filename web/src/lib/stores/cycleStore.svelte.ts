import { client } from '$lib/api/client';
import { getSession } from '$lib/api/session.svelte';
import type { Cycle, PhaseTransition, ApiCyclePhase } from '$lib/types/cycle';
import type { components } from '$lib/api/schemas/cycle';

// ─── Module state ───────────────────────────────────────────────────────────────

let cycles = $state<Cycle[]>([]);
let loading = $state(false);
let error = $state<string | null>(null);

// ─── Fixture data ───────────────────────────────────────────────────────────────

const FIXTURE_CYCLES: Cycle[] = [
	{
		id: 'cyc-001',
		organization_id: '00000000-0000-0000-0000-000000000001',
		year: 2026,
		current_phase: 'asignacion',
		version: 1,
		started_at: '2026-01-15T09:00:00Z',
		finished_at: null,
		created_at: '2026-01-15T09:00:00Z',
		updated_at: '2026-01-15T09:00:00Z'
	},
	{
		id: 'cyc-000',
		organization_id: '00000000-0000-0000-0000-000000000001',
		year: 2025,
		current_phase: 'cierre',
		version: 3,
		started_at: '2025-01-10T09:00:00Z',
		finished_at: '2025-12-20T18:00:00Z',
		created_at: '2025-01-10T09:00:00Z',
		updated_at: '2025-12-20T18:00:00Z'
	}
];

// ─── Getters ────────────────────────────────────────────────────────────────────

export function getCycles(): Cycle[] {
	return cycles;
}

export function isLoading(): boolean {
	return loading;
}

export function getError(): string | null {
	return error;
}

/** Returns the active (not finished) cycle, if any. */
export function getActiveCycle(): Cycle | undefined {
	return cycles.find((c) => !c.finished_at);
}

/** Returns true if there is a cycle for the given year. */
export function hasCycleForYear(year: number): boolean {
	return cycles.some((c) => c.year === year);
}

// ─── Load ───────────────────────────────────────────────────────────────────────

export async function loadCycles(): Promise<void> {
	loading = true;
	error = null;

	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		cycles = structuredClone(FIXTURE_CYCLES);
		loading = false;
		return;
	}

	const orgId = getSession().user?.organizationId;
	if (!orgId) {
		error = 'No hay organización en la sesión';
		loading = false;
		return;
	}

	try {
		const { data, error: apiError } = await client.GET('/cycles', {
			params: { query: { organization_id: orgId } }
		});

		if (apiError) {
			throw new Error(
				typeof apiError === 'string' ? apiError : 'Error al cargar ciclos'
			);
		}

		const raw = data as { data?: Array<components['schemas']['CycleLight']> };
		cycles = (raw?.data ?? []).map((c) => ({
			id: c.id,
			organization_id: c.organization_id,
			year: c.year,
			current_phase: c.current_phase,
			version: 0,
			started_at: null,
			finished_at: null,
			created_at: c.created_at,
			updated_at: c.updated_at
		}));
	} catch (e) {
		error = e instanceof Error ? e.message : 'Error desconocido al cargar ciclos';
	} finally {
		loading = false;
	}
}

/** Alias for loadCycles(). */
export function reload(): Promise<void> {
	return loadCycles();
}

// ─── Mutations ──────────────────────────────────────────────────────────────────

export async function createCycle(year: number): Promise<Cycle | null> {
	if (hasCycleForYear(year)) {
		error = `Ya existe un ciclo para el año ${year}`;
		return null;
	}

	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		const newCycle: Cycle = {
			id: `cyc-${Date.now()}`,
			organization_id: getSession().user?.organizationId ?? '',
			year,
			current_phase: 'asignacion',
			version: 1,
			started_at: new Date().toISOString(),
			finished_at: null,
			created_at: new Date().toISOString(),
			updated_at: new Date().toISOString()
		};
		cycles = [newCycle, ...cycles];
		return newCycle;
	}

	const orgId = getSession().user?.organizationId;
	if (!orgId) {
		error = 'No hay organización en la sesión';
		return null;
	}

	try {
		const { data, error: apiError } = await client.POST('/cycles', {
			params: {
				header: { 'Idempotency-Key': crypto.randomUUID() }
			},
			body: { organization_id: orgId, year }
		});

		if (apiError) {
			throw new Error(
				typeof apiError === 'string' ? apiError : 'Error al crear ciclo'
			);
		}

		const raw = data as components['schemas']['Cycle'];
		const created: Cycle = {
			id: raw.id,
			organization_id: raw.organization_id,
			year: raw.year,
			current_phase: raw.current_phase,
			version: raw.version,
			started_at: raw.started_at ?? null,
			finished_at: raw.finished_at ?? null,
			created_at: raw.created_at,
			updated_at: raw.updated_at
		};
		cycles = [created, ...cycles];
		return created;
	} catch (e) {
		error = e instanceof Error ? e.message : 'Error al crear ciclo';
		return null;
	}
}

export async function advancePhase(cycleId: string): Promise<boolean> {
	const cycle = cycles.find((c) => c.id === cycleId);
	if (!cycle) {
		error = 'Ciclo no encontrado';
		return false;
	}

	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		const nextPhase = getNextPhase(cycle.current_phase);
		if (!nextPhase) {
			error = 'El ciclo ya está en la última fase';
			return false;
		}
		cycles = cycles.map((c) =>
			c.id === cycleId
				? {
						...c,
						current_phase: nextPhase,
						version: c.version + 1,
						finished_at: nextPhase === 'cierre' ? new Date().toISOString() : c.finished_at,
						updated_at: new Date().toISOString()
					}
				: c
		);
		return true;
	}

	try {
		const { data, error: apiError } = await client.PUT('/cycles/{id}/transition', {
			params: {
				path: { id: cycleId },
				header: {
					'If-Match': String(cycle.version),
					'Idempotency-Key': crypto.randomUUID()
				}
			},
			body: { trigger: 'manual_rh', reason: '' }
		});

		if (apiError) {
			throw new Error(
				typeof apiError === 'string' ? apiError : 'Error al avanzar fase'
			);
		}

		const raw = data as components['schemas']['Cycle'];
		cycles = cycles.map((c) =>
			c.id === cycleId
				? {
						...c,
						current_phase: raw.current_phase,
						version: raw.version,
						finished_at: raw.finished_at ?? c.finished_at,
						updated_at: raw.updated_at
					}
				: c
		);
		return true;
	} catch (e) {
		error = e instanceof Error ? e.message : 'Error al avanzar fase';
		return false;
	}
}

export async function getAvailableTransitions(cycleId: string): Promise<PhaseTransition[]> {
	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		const cycle = cycles.find((c) => c.id === cycleId);
		if (!cycle) return [];
		const next = getNextPhase(cycle.current_phase);
		return next ? [{ from_phase: cycle.current_phase, to_phase: next, trigger: 'manual_rh' }] : [];
	}

	try {
		const { data, error: apiError } = await client.GET('/cycles/{id}/transitions', {
			params: { path: { id: cycleId } }
		});

		if (apiError) return [];

		const raw = data as { data?: Array<components['schemas']['PhaseTransition']> };
		return (raw?.data ?? []).map((t) => ({
			from_phase: t.from_phase,
			to_phase: t.to_phase,
			trigger: t.trigger
		}));
	} catch {
		return [];
	}
}

// ─── Helpers ────────────────────────────────────────────────────────────────────

function getNextPhase(current: ApiCyclePhase): ApiCyclePhase | null {
	if (current === 'asignacion') return 'avance';
	if (current === 'avance') return 'cierre';
	return null;
}
