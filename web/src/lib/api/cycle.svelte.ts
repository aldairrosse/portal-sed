import { client } from './client';
import { getSession } from './session.svelte';
import type { CyclePhase } from '$lib/types/evaluation';

export interface CycleState {
	activePhase: CyclePhase | null;
	loading: boolean;
	error: string | null;
}

const API_PHASE_MAP: Record<string, CyclePhase> = {
	asignacion: 'inicio-anio',
	avance: 'medio-anio',
	cierre: 'fin-anio'
};

function mapApiPhase(apiPhase: string): CyclePhase {
	if (API_PHASE_MAP[apiPhase]) return API_PHASE_MAP[apiPhase];
	const lower = apiPhase.toLowerCase();
	if (lower.includes('formul') || lower.includes('inicio') || lower.includes('planea') || lower.includes('asignacion')) {
		return 'inicio-anio';
	}
	return API_PHASE_MAP[apiPhase] ?? 'inicio-anio';
}

let activePhase = $state<CyclePhase | null>(null);
let activeCycleYear = $state<number | null>(null)
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
		const raw = data as { data?: Array<{ current_phase?: string; year?: number }> };
		const cycles = raw?.data ?? [];
		if (cycles.length > 0) {
			activePhase = mapApiPhase(cycles[0].current_phase ?? '');
			const y = (cycles[0] as { year?: number }).year
			activeCycleYear = typeof y === 'number' && Number.isFinite(y) ? y : null
		} else {
			activeCycleYear = null
		}
	} catch (e) {
		error = e instanceof Error ? e.message : 'Error al cargar ciclo';
		activePhase = null;
		activeCycleYear = null
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

export function getCycleState(): CycleState {
	return { activePhase, loading, error };
}
