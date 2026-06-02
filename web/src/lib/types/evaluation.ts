export type EvaluationProfile =
	| 'colaborador'
	| 'jefe'
	| 'vendedor'
	| 'gerente-tienda'
	| 'divisional'
	| 'regional'
	| 'director'
	| 'rh';

export type CyclePhase = 'inicio-anio' | 'medio-anio' | 'fin-anio';

export const EVALUATION_PROFILES: EvaluationProfile[] = [
	'colaborador',
	'jefe',
	'vendedor',
	'gerente-tienda',
	'divisional',
	'regional',
	'director',
	'rh'
];

export const CYCLE_PHASES: CyclePhase[] = ['inicio-anio', 'medio-anio', 'fin-anio'];

export const PROFILE_LABELS: Record<EvaluationProfile, string> = {
	colaborador: 'Colaborador',
	jefe: 'Jefe',
	vendedor: 'Vendedor',
	'gerente-tienda': 'Gerente de tienda',
	divisional: 'Divisional',
	regional: 'Regional',
	director: 'Director',
	rh: 'Recursos Humanos'
};

export const PHASE_LABELS: Record<CyclePhase, string> = {
	'inicio-anio': 'Inicio de año',
	'medio-anio': 'Medio año',
	'fin-anio': 'Fin de año'
};