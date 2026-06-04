import type { EvaluationProfile } from './evaluation';

// ─── Units ────────────────────────────────────────────────────────────────────

export type GoalUnit = 'porcentaje' | 'moneda' | 'numero' | 'binario';
export type KpiUnit = GoalUnit;

// ─── Cycle Phase ──────────────────────────────────────────────────────────────

export type CyclePhase = 'inicio-anio' | 'medio-anio' | 'fin-anio';

// ─── KPI ───────────────────────────────────────────────────────────────────────

export interface KPI {
	id: string;
	name: string;
	description: string;
	unit: KpiUnit;
	direction: 'ascendente' | 'descendente';
	targetValue?: number;
	minValue?: number;
	maxValue?: number;
}

// ─── GoalCategory ──────────────────────────────────────────────────────────────

export interface GoalCategory {
	id: string;
	name: string;
	description: string;
	weight: number;
}

// ─── Goal ──────────────────────────────────────────────────────────────────────

export interface Goal {
	id: string;
	name: string;
	description: string;
	categoryId: string;
	weight: number;
	unit: GoalUnit;
	targetValue: number;
	progress?: number;
	progressUpdatedAt?: string;
	comments?: GoalComment[];
}

// ─── GoalComment ──────────────────────────────────────────────────────────────

export interface GoalComment {
	id: string;
	authorId: string;
	authorName: string;
	content: string;
	createdAt: string;
}

// ─── GoalKpiLink (N:M) ─────────────────────────────────────────────────────────

export interface GoalKpiLink {
	goalId: string;
	kpiId: string;
	weight?: number;
}

// ─── EmployeeAssignment ────────────────────────────────────────────────────────

export interface EmployeeAssignment {
	id: string;
	employeeId: string;
	employeeName: string;
	profileId: EvaluationProfile;
	managerId: string | null;
	goalIds: string[];
	createdAt: string;
	updatedAt: string;
}

// ─── ChangeRequest ─────────────────────────────────────────────────────────────

export interface ChangeRequest {
	id: string;
	entityType: 'goal' | 'category' | 'kpi' | 'link' | 'assignment';
	entityId: string;
	action: 'create' | 'update' | 'delete';
	changes: Record<string, unknown>;
	reason: string;
	requestedBy: string;
	requestedAt: string;
	status: 'pending' | 'approved' | 'rejected';
	approvedBy?: string;
	approvedAt?: string;
}

// ─── MANAGER_MAP ───────────────────────────────────────────────────────────────

/**
 * Mock hierarchy for editor/reader mode detection.
 * Maps a profile to the profile of their direct manager.
 */
export const MANAGER_MAP: Partial<Record<EvaluationProfile, EvaluationProfile>> = {
	colaborador: 'jefe',
	vendedor: 'jefe',
	jefe: 'director',
	'gerente-tienda': 'director',
	divisional: 'director',
	regional: 'director',
	rh: 'director'
};
