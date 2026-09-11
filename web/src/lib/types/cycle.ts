export type ApiCyclePhase = 'asignacion' | 'avance' | 'cierre';

/** Aliases de compatibilidad UI (inicio-anio/medio-anio/fin-anio). */
export type CyclePhaseAlias = 'inicio-anio' | 'medio-anio' | 'fin-anio';

export type AnyCyclePhase = ApiCyclePhase | CyclePhaseAlias;

export interface Cycle {
	id: string;
	organization_id: string;
	year: number;
	current_phase: ApiCyclePhase;
	version: number;
	started_at: string | null;
	finished_at: string | null;
	created_at: string;
	updated_at: string;
}

export interface PhaseTransition {
	from_phase: ApiCyclePhase;
	to_phase: ApiCyclePhase;
	trigger: 'auto' | 'manual_rh';
}

export const PHASE_ORDER: ApiCyclePhase[] = ['asignacion', 'avance', 'cierre'];

export const ALIAS_TO_CANONICAL: Record<CyclePhaseAlias, ApiCyclePhase> = {
	'inicio-anio': 'asignacion',
	'medio-anio': 'avance',
	'fin-anio': 'cierre',
};

export const CANONICAL_TO_ALIAS: Record<ApiCyclePhase, CyclePhaseAlias> = {
	asignacion: 'inicio-anio',
	avance: 'medio-anio',
	cierre: 'fin-anio',
};

/** Normaliza alias legacy a fase canónica (case-insensitive, acepta _ como -). */
export function normalizePhase(phase: string): ApiCyclePhase {
	const n = phase.toLowerCase().trim().replace(/_/g, '-');
	if (n === 'asignacion' || n === 'avance' || n === 'cierre') return n;
	if (n === 'inicio-anio' || n === 'medio-anio' || n === 'fin-anio') {
		return ALIAS_TO_CANONICAL[n as CyclePhaseAlias];
	}
	const mapped = (ALIAS_TO_CANONICAL as Record<string, ApiCyclePhase>)[phase];
	return mapped ?? 'asignacion';
}

/** True si la fase es medio-año (avance ≡ medio-anio). Acepta canónica o alias. */
export function isMidYearPhase(phase: string): boolean {
	return normalizePhase(phase) === 'avance';
}

/** True si la fase es avance o su alias medio-anio. */
export function isAvance(phase: string): boolean {
	return normalizePhase(phase) === 'avance';
}

/** True si la fase es cierre o su alias fin-anio. */
export function isCierre(phase: string): boolean {
	return normalizePhase(phase) === 'cierre';
}

/** True si la fase es asignación o su alias inicio-anio. */
export function isAsignacion(phase: string): boolean {
	return normalizePhase(phase) === 'asignacion';
}

/** Aliases UI: medio-anio ≡ avance, fin-anio ≡ cierre, inicio-anio ≡ asignacion. */
export const isMedioAnio = isAvance;
export const isFinAnio = isCierre;
export const isInicioAnio = isAsignacion;

/** Alias PascalCase para compatibilidad con spec. */
export const IsMidYearPhase = isMidYearPhase;

/** True si dos fases coinciden para escritura (avance ≡ medio-anio, cierre ≡ fin-anio). */
export function samePhaseForWrite(a: string, b: string): boolean {
	return normalizePhase(a) === normalizePhase(b);
}

/** Aliases PascalCase para compatibilidad con spec. */
export const SamePhaseForWrite = samePhaseForWrite;
export const SamePhase = samePhaseForWrite;

/** Label ES canónico: Inicio de año / Medio año / Fin de año. Acepta alias. */
export function getPhaseLabel(phase: string): string {
	const canonical = normalizePhase(phase);
	if (canonical === 'asignacion') return 'Inicio de año';
	if (canonical === 'avance') return 'Medio año';
	return 'Fin de año';
}

export const API_PHASE_LABELS: Record<ApiCyclePhase, string> = {
	asignacion: 'Inicio de año',
	avance: 'Medio año',
	cierre: 'Fin de año',
};

/** R4: ciclo editable si está activo (sin finished_at) o su año es el actual. */
export function isCycleActive(
	cycle: Pick<Cycle, 'year'> & { finished_at?: string | null },
): boolean {
	if (cycle.finished_at === null || cycle.finished_at === undefined)
		return true;
	return cycle.year === new Date().getFullYear();
}
