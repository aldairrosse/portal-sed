import { client } from '$lib/api/client';
import type { components } from '$lib/api/schemas/cycle';

// ─── State ───────────────────────────────────────────────────────────────────

let phaseDefinitions = $state<components['schemas']['PhaseDefinition'][]>([]);

// ponytail: plain vars, NOT $state — prevents $effect in consumers from tracking
// loading/loaded changes and re-firing in an infinite loop.
let _loading = false;
let _loaded = false;

// ─── Load ────────────────────────────────────────────────────────────────────

export async function loadPhases(): Promise<void> {
	if (_loaded || _loading) return;
	_loading = true;
	try {
		const { data, error: apiError } = await client.GET('/phases', {});
		if (apiError) throw new Error('Error al cargar fases');
		const raw = data as { data?: components['schemas']['PhaseDefinition'][] };
		phaseDefinitions = raw?.data ?? [];
	} catch {
		// silently fail — phase UUID lookup will return undefined
	} finally {
		_loaded = true;
		_loading = false;
	}
}

// ─── Getters ─────────────────────────────────────────────────────────────────

/** Get the UUID for a phase enum value ('asignacion', 'avance', 'cierre'). */
export function getPhaseId(phaseEnum: string): string | undefined {
	return phaseDefinitions.find((p) => p.phase === phaseEnum)?.id;
}

export function isPhaseStoreLoading(): boolean {
	return _loading;
}
