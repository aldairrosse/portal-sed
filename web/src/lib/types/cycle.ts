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
	'fin-anio': 'cierre'
};

export const CANONICAL_TO_ALIAS: Record<ApiCyclePhase, CyclePhaseAlias> = {
	asignacion: 'inicio-anio',
	avance: 'medio-anio',
	cierre: 'fin-anio'
};

/** Normaliza alias legacy a fase canónica. */
export function normalizePhase(phase: string): ApiCyclePhase {
	if (phase === 'asignacion' || phase === 'avance' || phase === 'cierre') return phase;
	const mapped = (ALIAS_TO_CANONICAL as Record<string, ApiCyclePhase>)[phase];
	return mapped ?? 'asignacion';
}

/** True si la fase es medio-año (avance o alias medio-anio solo-lectura). */
export function isMidYearPhase(phase: string): boolean {
	const n = phase.toLowerCase().trim().replace(/_/g, '-');
	return n === 'avance' || n === 'medio-anio';
}

/** Alias PascalCase para compatibilidad con spec. */
export const IsMidYearPhase = isMidYearPhase;

/** True si dos fases coinciden para escritura (avance ≡ medio-anio). */
export function samePhaseForWrite(a: string, b: string): boolean {
	const na = a.toLowerCase().trim().replace(/_/g, '-');
	const nb = b.toLowerCase().trim().replace(/_/g, '-');
	if (na === nb) return true;
	return isMidYearPhase(na) && isMidYearPhase(nb);
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
	cierre: 'Fin de año'
};
