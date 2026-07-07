import { client } from '$lib/api/client';
import { getSession } from '$lib/api/session.svelte';
import type { Cycle, PhaseTransition, ApiCyclePhase } from '$lib/types/cycle';
import type { components } from '$lib/api/schemas/cycle';

// ─── Module state ───────────────────────────────────────────────────────────────

let cycles = $state<Cycle[]>([]);
// ponytail: plain vars, NOT $state — prevents $effect in consumers from tracking
// loading/loaded changes and re-firing in an infinite loop (same pattern as phaseStore).
let _loading = false;
let _loaded = false;
let loading = $state(false);
let error = $state<string | null>(null);

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
	if (_loaded || _loading) return;
	_loading = true;
	loading = true;
	error = null;

	const orgId = getSession().user?.organizationId;
	if (!orgId) {
		error = 'No hay organización en la sesión';
		_loaded = true;
		_loading = false;
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
			version: c.version,
			started_at: null,
			finished_at: null,
			created_at: c.created_at,
			updated_at: c.updated_at
		}));
	} catch (e) {
		error = e instanceof Error ? e.message : 'Error desconocido al cargar ciclos';
	} finally {
		_loaded = true;
		_loading = false;
		loading = false;
	}
}

/** Alias for loadCycles() — bypasses the dedup guard. */
export function reload(): Promise<void> {
	_loaded = false;
	_loading = false;
	cycles = [];
	return loadCycles();
}

// ─── Mutations ──────────────────────────────────────────────────────────────────

export async function createCycle(year: number): Promise<Cycle | null> {
	if (hasCycleForYear(year)) {
		error = `Ya existe un ciclo para el año ${year}`;
		return null;
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
	// ponytail: refresh cycle first to get latest version (avoids stale _loaded guard)
	const fresh = await getCycle(cycleId);
	if (!fresh) return false;

	try {
		const { data, error: apiError } = await client.PUT('/cycles/{id}/transition', {
			params: {
				path: { id: cycleId },
				header: {
					'If-Match': String(fresh.version),
					'Idempotency-Key': crypto.randomUUID()
				}
			},
			body: { trigger: 'manual_rh', reason: '' }
		});

		if (apiError) {
			const msg = typeof apiError === 'object' && apiError !== null
				? ((apiError as Record<string, unknown>).error as Record<string, unknown> ?? {})?.message ?? JSON.stringify(apiError)
				: String(apiError);
			console.error('[advancePhase] 409 body:', apiError, 'sent version:', fresh.version);
			throw new Error(msg);
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

export async function assignAll(cycleId: string): Promise<{ assigned: number; skipped: number; total: number } | null> {
	try {
		const { data, error: apiError } = await client.POST('/cycles/{id}/assign-all', {
			params: {
				path: { id: cycleId },
				header: { 'Idempotency-Key': crypto.randomUUID() }
			}
		});

		if (apiError) {
			throw new Error(
				typeof apiError === 'string' ? apiError : 'Error al asignar empleados'
			);
		}

		const raw = data as { assigned: number; skipped: number; total: number };
		return raw;
	} catch (e) {
		error = e instanceof Error ? e.message : 'Error al asignar empleados';
		return null;
	}
}

// ─── Helpers ────────────────────────────────────────────────────────────────────

function getNextPhase(current: ApiCyclePhase): ApiCyclePhase | null {
	if (current === 'asignacion') return 'avance';
	if (current === 'avance') return 'cierre';
	return null;
}
