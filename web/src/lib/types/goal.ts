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
	currentValue?: number;
	progressPercent?: number;
	minValue?: number;
	maxValue?: number;
}

// ─── GoalCategory ──────────────────────────────────────────────────────────────

export interface GoalCategory {
	id: string;
	name: string;
	description: string;
	weight: number;
	pillarId?: string;
	comments?: GoalComment[];
}

// ─── Goal ──────────────────────────────────────────────────────────────────────

export interface Goal {
	id: string;
	name: string;
	description: string;
	categoryId: string;
	weight: number;
	unit: GoalUnit;
	direction: 'ascendente' | 'descendente';
	targetValue: number;
	baselineValue?: number;
	progressPercent?: number;
	progress?: number;
	progressUpdatedAt?: string;
	comments?: GoalComment[];
	pendingProposal?: GoalProposal;
	version: number;
}

// ─── GoalComment ──────────────────────────────────────────────────────────────

export interface GoalComment {
	id: string;
	authorId: string;
	authorName: string;
	content: string;
	createdAt: string;
	goalId?: string;
	categoryId?: string;
	assignmentId?: string;
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
	employeeNumber?: string;
	profileId: EvaluationProfile;
	managerId: string | null;
	goalIds: string[];
	createdAt: string;
	updatedAt: string;
}

// ─── GoalProposal ──────────────────────────────────────────────────────────────

export interface GoalProposal {
	id: string;
	goalId: string;
	requestedBy: string;
	name: string;
	description: string;
	unit: GoalUnit;
	direction: 'ascendente' | 'descendente';
	weight: number;
	targetValue: number;
	baselineValue?: number;
	kpiIds: string[];
	status: 'pending' | 'accepted' | 'rejected';
	reviewedBy?: string;
	reviewedAt?: string;
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
