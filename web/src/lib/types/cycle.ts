export type ApiCyclePhase = 'asignacion' | 'avance' | 'cierre';

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

export const API_PHASE_LABELS: Record<ApiCyclePhase, string> = {
	asignacion: 'Asignación de objetivos',
	avance: 'Seguimiento de avance',
	cierre: 'Cierre y evaluación'
};

export const PHASE_ORDER: ApiCyclePhase[] = ['asignacion', 'avance', 'cierre'];
