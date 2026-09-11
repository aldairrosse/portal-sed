import type { AnyCyclePhase } from './cycle';

export type EvaluationProfile =
	| 'colaborador'
	| 'jefe'
	| 'gerente'
	| 'coordinador'
	| 'vendedor'
	| 'gerente-tienda'
	| 'divisional'
	| 'regional'
	| 'director'
	| 'director-general'
	| 'rh';

export type CyclePhase = AnyCyclePhase;

export const EVALUATION_PROFILES: EvaluationProfile[] = [
	'colaborador',
	'jefe',
	'gerente',
	'coordinador',
	'vendedor',
	'gerente-tienda',
	'divisional',
	'regional',
	'director',
	'director-general',
	'rh',
];

export const CYCLE_PHASES: CyclePhase[] = [
	'avance',
	'inicio-anio',
	'medio-anio',
	'fin-anio',
];

export const PROFILE_LABELS: Record<EvaluationProfile, string> = {
	colaborador: 'Colaborador',
	jefe: 'Jefe',
	gerente: 'Gerente',
	coordinador: 'Coordinador',
	vendedor: 'Vendedor',
	'gerente-tienda': 'Gerente de tienda',
	divisional: 'Divisional',
	regional: 'Regional',
	director: 'Director',
	'director-general': 'Director General',
	rh: 'Recursos Humanos',
};

export const PHASE_LABELS: Record<CyclePhase, string> = {
	asignacion: 'Inicio de año',
	avance: 'Medio año',
	cierre: 'Fin de año',
	'inicio-anio': 'Inicio de año',
	'medio-anio': 'Medio año',
	'fin-anio': 'Fin de año',
};
