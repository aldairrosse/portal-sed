import { client } from '$lib/api/client';
import { getSession } from '$lib/api/session.svelte';
import { loadCycle, activateCycle as apiActivateCycle } from '$lib/api/cycle.svelte';
import * as notifications from '$lib/stores/notifications.svelte';
import type { Cycle, PhaseTransition, ApiCyclePhase } from '$lib/types/cycle';
import { normalizePhase } from '$lib/types/cycle';
import type { components } from '$lib/api/schemas/cycle';

// ─── Module state ───────────────────────────────────────────────────────────────

let cycles = $state<Cycle[]>([]);
let activeCycle = $state<Cycle | null>(null);
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

/** Returns the cycle resolved via GET /cycles/current (no list fallback). */
export function getActiveCycle(): Cycle | undefined {
	return activeCycle ?? undefined;
}

/** Returns activeCycle or toasts an error when there is no active cycle. */
export function requireActiveCycle(): Cycle | null {
	const current = getActiveCycle();
	if (!current || !current.is_active) {
		notifications.error('No hay ciclo activo');
		return null;
	}
	return current;
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
			params: { query: { organization_id: orgId } },
		});

		if (apiError) {
			throw new Error(
				typeof apiError === 'string' ? apiError : 'Error al cargar ciclos',
			);
		}

		const raw = data as { data?: Array<components['schemas']['CycleLight']> };
		cycles = (raw?.data ?? []).map((c) => ({
			id: c.id,
			organization_id: c.organization_id,
			year: c.year,
			current_phase: normalizePhase(c.current_phase),
			version: c.version,
			is_active:
				(c as { is_active?: boolean }).is_active ??
				// ponytail: CycleLight aún no trae is_active en schemas — legacy: inactivo por defecto, no enmascarar sin activo
				false,
		// ponytail: CycleLight has no started_at/finished_at — keep null; cierre is detected via current_phase, not finished_at
		started_at: null,
		finished_at: null,
			created_at: c.created_at,
			updated_at: c.updated_at,
		}));
	} catch (e) {
		error =
			e instanceof Error ? e.message : 'Error desconocido al cargar ciclos';
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

/** Resolves the real active cycle via GET /cycles/current (source of truth for tabs). No year fallback. */
export async function loadCurrent(year?: number): Promise<Cycle | null> {
	const orgId = getSession().user?.organizationId;
	if (!orgId) {
		error = 'No hay organización en la sesión';
		return null;
	}
	try {
		const { data, error: apiError } = await client.GET('/cycles/current', {
			params: { query: { organization_id: orgId, ...(year !== undefined ? { year } : {}) } },
		});
		if (apiError || !data) return null;
		const raw = data as components['schemas']['Cycle'];
		activeCycle = {
			id: raw.id,
			organization_id: raw.organization_id,
			year: raw.year,
			current_phase: normalizePhase(raw.current_phase),
			version: raw.version,
			is_active: (raw as { is_active?: boolean }).is_active ?? false,
			started_at: raw.started_at ?? null,
			finished_at: raw.finished_at ?? null,
			created_at: raw.created_at,
			updated_at: raw.updated_at,
		};
		return activeCycle;
	} catch {
		return null;
	}
}

// ─── Mutations ──────────────────────────────────────────────────────────────────

/** Activates a cycle via POST /cycles/{id}/activate and refreshes the store. */
export async function activate(cycleId: string): Promise<boolean> {
	const ok = await apiActivateCycle(cycleId);
	if (!ok) {
		error = 'Error al activar ciclo';
		return false;
	}
	cycles = cycles.map((c) => ({ ...c, is_active: c.id === cycleId }));
	const fresh = await getCycle(cycleId);
	if (fresh) {
		cycles = cycles.map((c) => (c.id === cycleId ? fresh : { ...c, is_active: false }));
		activeCycle = fresh;
	} else {
		await loadCurrent();
	}
	loadCycle();
	return true;
}

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
				header: { 'Idempotency-Key': crypto.randomUUID() },
			},
			body: { organization_id: orgId, year },
		});

		if (apiError) {
			throw new Error(
				typeof apiError === 'string' ? apiError : 'Error al crear ciclo',
			);
		}

		const raw = data as components['schemas']['Cycle'];
		const created: Cycle = {
			id: raw.id,
			organization_id: raw.organization_id,
			year: raw.year,
			current_phase: normalizePhase(raw.current_phase),
			version: raw.version,
			is_active: (raw as { is_active?: boolean }).is_active ?? false,
			started_at: raw.started_at ?? null,
			finished_at: raw.finished_at ?? null,
			created_at: raw.created_at,
			updated_at: raw.updated_at,
		};
		cycles = [created, ...cycles];
		activeCycle = created;
		return created;
	} catch (e) {
		error = e instanceof Error ? e.message : 'Error al crear ciclo';
		return null;
	}
}

async function getCycle(cycleId: string): Promise<Cycle | null> {
	const { data, error: apiError } = await client.GET('/cycles/{id}', {
		params: { path: { id: cycleId } },
	});
	if (apiError || !data) {
		error =
			(apiError as { error?: { message?: string } })?.error?.message ??
			'Error al obtener ciclo';
		return null;
	}
	const raw = data as components['schemas']['Cycle'];
	return {
		id: raw.id,
		organization_id: raw.organization_id,
		year: raw.year,
		current_phase: normalizePhase(raw.current_phase),
		version: raw.version,
		is_active: (raw as { is_active?: boolean }).is_active ?? false,
		started_at: raw.started_at ?? null,
		finished_at: raw.finished_at ?? null,
		created_at: raw.created_at,
		updated_at: raw.updated_at,
	};
}

export async function advancePhase(
	cycleId: string,
	toPhase: ApiCyclePhase,
): Promise<boolean> {
	// ponytail: refresh cycle first to get latest version (avoids stale _loaded guard)
	const fresh = await getCycle(cycleId);
	if (!fresh) return false;

	try {
		const { data, error: apiError } = await client.PUT(
			'/cycles/{id}/transition',
			{
				params: {
					path: { id: cycleId },
					header: {
						'If-Match': String(fresh.version),
						'Idempotency-Key': crypto.randomUUID(),
					},
				},
				body: { trigger: 'manual_rh', to_phase: toPhase, reason: '' },
			},
		);

		if (apiError) {
			const msg =
				(apiError as { error?: { message?: string } })?.error?.message ??
				JSON.stringify(apiError);
			console.error(
				'[advancePhase] 409 body:',
				apiError,
				'sent version:',
				fresh.version,
			);
			throw new Error(msg);
		}

		const raw = data as components['schemas']['Cycle'];
		cycles = cycles.map((c) =>
			c.id === cycleId
				? {
						...c,
					current_phase: normalizePhase(raw.current_phase),
					version: raw.version,
					// ponytail: propagate backend finished_at as-is; spec keeps it NULL during cierre until a new annual cycle is created
					finished_at: raw.finished_at ?? null,
					updated_at: raw.updated_at,
				}
			: c,
		);
		if (activeCycle?.id === cycleId) {
			activeCycle = {
				...activeCycle,
				current_phase: normalizePhase(raw.current_phase),
				version: raw.version,
				finished_at: raw.finished_at ?? null,
				updated_at: raw.updated_at,
			};
		}
		// sync getActivePhase() consumers (cycle.svelte.ts)
		loadCycle();
		return true;
	} catch (e) {
		error = e instanceof Error ? e.message : 'Error al avanzar fase';
		return false;
	}
}

/** Extracts { message, code, trace } from a cycles API error body. */
function revertErrorDetails(apiError: unknown): {
	message: string;
	code: string | null;
	trace: string | null;
} {
	const body = (
		apiError as {
			error?: { code?: unknown; message?: unknown; trace_id?: unknown };
		}
	)?.error;
	const message =
		typeof body?.message === 'string' ? body.message : JSON.stringify(apiError);
	const code = typeof body?.code === 'string' ? body.code : null;
	const trace = typeof body?.trace_id === 'string' ? body.trace_id : null;
	return { message, code, trace };
}

export async function revertPhase(cycleId: string): Promise<boolean> {
	try {
		const { data, error: apiError } = await client.POST('/cycles/{id}/revert', {
			params: {
				path: { id: cycleId },
				header: { 'Idempotency-Key': crypto.randomUUID() },
			},
		});

		if (apiError) {
			const { message, code, trace } = revertErrorDetails(apiError);
			const err = new Error(message);
			(err as Error & { code?: string | null; trace?: string | null }).code =
				code;
			(err as Error & { code?: string | null; trace?: string | null }).trace =
				trace;
			throw err;
		}

		const raw = data as components['schemas']['Cycle'];
		cycles = cycles.map((c) =>
			c.id === cycleId
				? {
						...c,
					current_phase: normalizePhase(raw.current_phase),
					version: raw.version,
					// ponytail: propagate backend finished_at as-is; stale legacy values left untouched (DB cleanup out of scope)
					finished_at: raw.finished_at ?? null,
					updated_at: raw.updated_at,
					}
				: c,
		);
		if (activeCycle?.id === cycleId) {
			activeCycle = {
				...activeCycle,
				current_phase: normalizePhase(raw.current_phase),
				version: raw.version,
				finished_at: raw.finished_at ?? null,
				updated_at: raw.updated_at,
			};
		}
		// sync getActivePhase() consumers (cycle.svelte.ts)
		loadCycle();
		return true;
	} catch (e) {
		const err = e as Error & { code?: string | null; trace?: string | null };
		const base =
			err instanceof Error && err.message
				? err.message
				: 'Error al retroceder fase';
		const withTrace = err.trace ? `${base} (trace: ${err.trace})` : base;
		notifications.errorWithCode(withTrace, err.code ?? null);
		return false;
	}
}

export async function getAvailableTransitions(
	cycleId: string,
): Promise<PhaseTransition[]> {
	try {
		const { data, error: apiError } = await client.GET(
			'/cycles/{id}/transitions',
			{
				params: { path: { id: cycleId } },
			},
		);

		if (apiError) return [];

		const raw = data as {
			data?: Array<components['schemas']['PhaseTransition']>;
		};
		return (raw?.data ?? []).map((t) => ({
			from_phase: t.from_phase,
			to_phase: t.to_phase,
			trigger: t.trigger,
		}));
	} catch {
		return [];
	}
}

export async function assignAll(
	cycleId: string,
): Promise<{ assigned: number; skipped: number; total: number } | null> {
	try {
		const { data, error: apiError } = await client.POST(
			'/cycles/{id}/assign-all',
			{
				params: {
					path: { id: cycleId },
					header: { 'Idempotency-Key': crypto.randomUUID() },
				},
			},
		);

		if (apiError) {
			throw new Error(
				typeof apiError === 'string' ? apiError : 'Error al asignar empleados',
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
// Re-exports: única verdad de fases en $lib/types/cycle.ts.
export {
	normalizePhase,
	getPhaseLabel,
	isMidYearPhase,
	samePhaseForWrite,
} from '$lib/types/cycle';
