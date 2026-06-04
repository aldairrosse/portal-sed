import type { CompetencyRating, GoalClosure, EvaluationStatus } from '$lib/types/evaluation-result';
import type { EvaluationProfile } from '$lib/types/evaluation';
import { getPhase } from '$lib/stores/devContext.svelte';
import { getGoalsByCategory } from '$lib/stores/goalsStore.svelte';

import selfEvaluationsData from '$lib/fixtures/evaluations/self-evaluations.json';
import goalClosuresData from '$lib/fixtures/evaluations/goal-closures.json';
import rhEvaluationsData from '$lib/fixtures/evaluations/rh-evaluations.json';

// ─── State ────────────────────────────────────────────────────────────────────

let competencyRatings = $state<CompetencyRating[]>(
	structuredClone(selfEvaluationsData as CompetencyRating[])
);
let goalClosures = $state<GoalClosure[]>(structuredClone(goalClosuresData as GoalClosure[]));

// Merge RH evaluations into competencyRatings on init
for (const rh of rhEvaluationsData as CompetencyRating[]) {
	const idx = competencyRatings.findIndex(
		(cr) => cr.employeeId === rh.employeeId && cr.competencyId === rh.competencyId
	);
	if (idx >= 0) {
		competencyRatings[idx] = { ...competencyRatings[idx], rhRating: rh.rhRating, rhComment: rh.rhComment };
	} else {
		competencyRatings = [...competencyRatings, structuredClone(rh)];
	}
}

// ─── Helpers ───────────────────────────────────────────────────────────────────

function isFinAnio(): boolean {
	return getPhase() === 'fin-anio';
}

function isRhProfile(profile: EvaluationProfile): boolean {
	return profile === 'rh';
}

// ─── Getters ───────────────────────────────────────────────────────────────────

export function getCompetencyRatings(employeeId: string): CompetencyRating[] {
	return competencyRatings.filter((cr) => cr.employeeId === employeeId);
}

export function getCompetencyRating(
	employeeId: string,
	competencyId: string
): CompetencyRating | undefined {
	return competencyRatings.find(
		(cr) => cr.employeeId === employeeId && cr.competencyId === competencyId
	);
}

export function getGoalClosures(employeeId: string): GoalClosure[] {
	return goalClosures.filter((gc) => gc.employeeId === employeeId);
}

export function getGoalClosure(employeeId: string, goalId: string): GoalClosure | undefined {
	return goalClosures.find((gc) => gc.employeeId === employeeId && gc.goalId === goalId);
}

export function getEvaluationStatus(
	employeeId: string,
	totalCompetencies: number,
	goalIds: string[]
): EvaluationStatus {
	const empRatings = competencyRatings.filter((cr) => cr.employeeId === employeeId);
	const ratedCompetencies = empRatings.filter((cr) => cr.selfRating !== undefined).length;

	if (ratedCompetencies === 0) return 'pending';

	const empClosures = goalClosures.filter((gc) => gc.employeeId === employeeId);
	const closedGoals = empClosures.filter((gc) => gc.closedAt !== undefined).length;

	if (ratedCompetencies >= totalCompetencies && closedGoals >= goalIds.length) return 'completed';

	return 'in-progress';
}

// ─── Mutations: Employee Self-Evaluation ───────────────────────────────────────

export function rateCompetency(
	employeeId: string,
	competencyId: string,
	level: 1 | 2 | 3 | 4 | 5,
	comment?: string
): void {
	if (!isFinAnio()) return;

	const existing = competencyRatings.find(
		(cr) => cr.employeeId === employeeId && cr.competencyId === competencyId
	);
	if (existing) {
		competencyRatings = competencyRatings.map((cr) =>
			cr.employeeId === employeeId && cr.competencyId === competencyId
				? { ...cr, selfRating: level, selfComment: comment ?? cr.selfComment }
				: cr
		);
	} else {
		competencyRatings = [
			...competencyRatings,
			{
				id: `sr-${employeeId}-${competencyId}-${Date.now()}`,
				employeeId,
				competencyId,
				selfRating: level,
				selfComment: comment
			}
		];
	}
}

export function closeGoal(
	employeeId: string,
	goalId: string,
	finalProgress: number,
	selfAssessment: string
): void {
	if (!isFinAnio()) return;

	const existing = goalClosures.find(
		(gc) => gc.employeeId === employeeId && gc.goalId === goalId
	);
	if (existing) {
		goalClosures = goalClosures.map((gc) =>
			gc.employeeId === employeeId && gc.goalId === goalId
				? {
						...gc,
						finalProgress,
						selfAssessment,
						closedAt: gc.closedAt ?? new Date().toISOString()
					}
				: gc
		);
	} else {
		goalClosures = [
			...goalClosures,
			{
				id: `gc-${employeeId}-${goalId}-${Date.now()}`,
				employeeId,
				goalId,
				finalProgress,
				selfAssessment,
				closedAt: new Date().toISOString()
			}
		];
	}
}

// ─── Mutations: RH Evaluation ─────────────────────────────────────────────────

export function rhRateCompetency(
	employeeId: string,
	competencyId: string,
	level: 1 | 2 | 3 | 4 | 5,
	comment?: string
): void {
	if (!isFinAnio()) return;

	const existing = competencyRatings.find(
		(cr) => cr.employeeId === employeeId && cr.competencyId === competencyId
	);
	if (existing) {
		competencyRatings = competencyRatings.map((cr) =>
			cr.employeeId === employeeId && cr.competencyId === competencyId
				? { ...cr, rhRating: level, rhComment: comment ?? cr.rhComment }
				: cr
		);
	} else {
		competencyRatings = [
			...competencyRatings,
			{
				id: `rh-${employeeId}-${competencyId}-${Date.now()}`,
				employeeId,
				competencyId,
				rhRating: level,
				rhComment: comment
			}
		];
	}
}

export function rhAssessGoal(
	employeeId: string,
	goalId: string,
	rhAssessment: string
): void {
	if (!isFinAnio()) return;

	goalClosures = goalClosures.map((gc) =>
		gc.employeeId === employeeId && gc.goalId === goalId
			? { ...gc, rhAssessment }
			: gc
	);
}

// ─── Mutations: Manager ───────────────────────────────────────────────────────

export function addManagerComment(
	employeeId: string,
	goalId: string,
	comment: string
): void {
	if (!isFinAnio()) return;

	goalClosures = goalClosures.map((gc) =>
		gc.employeeId === employeeId && gc.goalId === goalId
			? { ...gc, managerComment: comment }
			: gc
	);
}
