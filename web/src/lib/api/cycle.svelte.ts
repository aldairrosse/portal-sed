import { client } from './client';
import { getSession } from './session.svelte';
import { normalizePhase, CANONICAL_TO_ALIAS } from '$lib/types/cycle';
import type { AnyCyclePhase } from '$lib/types/cycle';

// Re-export: única verdad de fases en $lib/types/cycle.ts.
export type { ApiCyclePhase, AnyCyclePhase, CyclePhaseAlias } from '$lib/types/cycle';
export type CyclePhase = AnyCyclePhase;

export interface CycleState {
	activePhase: CyclePhase | null;
	loading: boolean;
	error: string | null;
}

function mapApiPhase(apiPhase: string): CyclePhase {
	return CANONICAL_TO_ALIAS[normalizePhase(apiPhase)];
}

let activePhase = $state<CyclePhase | null>(null);
let activeCycleYear = $state<number | null>(null)
let activeCycleId = $state<string | null>(null);
let loading = $state(true);
let error = $state<string | null>(null);

export async function loadCycle(): Promise<void> {
	loading = true;
	error = null;

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
			throw new Error(typeof apiError === 'string' ? apiError : 'Error al cargar ciclo');
		}
		const raw = data as { data?: Array<{ id?: string; current_phase?: string; year?: number }> };
		const cycles = raw?.data ?? [];
		if (cycles.length > 0) {
			activePhase = mapApiPhase(cycles[0].current_phase ?? '');
			const y = (cycles[0] as { year?: number }).year
			activeCycleYear = typeof y === 'number' && Number.isFinite(y) ? y : null
			activeCycleId = (cycles[0] as { id?: string }).id ?? null
		} else {
			activeCycleYear = null
			activeCycleId = null
		}
	} catch (e) {
		error = e instanceof Error ? e.message : 'Error al cargar ciclo';
		activePhase = null;
		activeCycleYear = null
		activeCycleId = null
	} finally {
		loading = false;
	}
}

export function getActivePhase(): CyclePhase | null {
	return activePhase;
}

export function getActiveCycleYear(): number | null {
	return activeCycleYear
}

export function getActiveCycleId(): string | null {
	return activeCycleId
}

export function getCycleState(): CycleState {
	return { activePhase, loading, error };
}
