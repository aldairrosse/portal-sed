import { client } from './client';
import type { CyclePhase } from '$lib/types/evaluation';

export interface CycleState {
	activePhase: CyclePhase | null;
	loading: boolean;
	error: string | null;
}

const FIXTURE_PHASE: CyclePhase = 'inicio-anio';

const API_PHASE_MAP: Record<string, CyclePhase> = {
	asignacion: 'inicio-anio',
	avance: 'medio-anio',
	cierre: 'fin-anio'
};

function mapApiPhase(apiPhase: string): CyclePhase {
	return API_PHASE_MAP[apiPhase] ?? 'inicio-anio';
}

let activePhase = $state<CyclePhase | null>(null);
let loading = $state(true);
let error = $state<string | null>(null);

export async function loadCycle(): Promise<void> {
	loading = true;
	error = null;

	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		activePhase = FIXTURE_PHASE;
		loading = false;
		return;
	}

	try {
		const { data, error: apiError } = await client.GET('/cycle/current' as never);
		if (apiError) {
			throw new Error(typeof apiError === 'string' ? apiError : 'Error al cargar ciclo');
		}
		const raw = data as { current_phase?: string };
		if (raw?.current_phase) {
			activePhase = mapApiPhase(raw.current_phase);
		}
	} catch (e) {
		error = e instanceof Error ? e.message : 'Error al cargar ciclo';
		activePhase = null;
	} finally {
		loading = false;
	}
}

export function getActivePhase(): CyclePhase | null {
	return activePhase;
}

export function getCycleState(): CycleState {
	return { activePhase, loading, error };
}
