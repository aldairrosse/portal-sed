import type { CompetencyRating, GoalClosure, EvaluationStatus } from '$lib/types/evaluation-result';
import { getActivePhase } from '$lib/api/cycle.svelte';
import { getSession } from '$lib/api/session.svelte';
import { client } from '$lib/api/client';

import selfEvaluationsData from '$lib/fixtures/evaluations/self-evaluations.json';
import goalClosuresData from '$lib/fixtures/evaluations/goal-closures.json';
import rhEvaluationsData from '$lib/fixtures/evaluations/rh-evaluations.json';

// ─── Internal data shape ──────────────────────────────────────────────────────

interface StoreData {
	competencyRatings: CompetencyRating[];
	goalClosures: GoalClosure[];
}

// ─── Triplete state ───────────────────────────────────────────────────────────

let data = $state<StoreData | null>(null);
let loading = $state(true);
let error = $state<string | null>(null);

// ─── Fixture loaders ──────────────────────────────────────────────────────────

function mergeRHEvaluations(
	selfRatings: CompetencyRating[],
	rhRatings: CompetencyRating[]
): CompetencyRating[] {
	const merged = [...selfRatings];
	for (const rh of rhRatings) {
		const idx = merged.findIndex(
			(cr) => cr.employeeId === rh.employeeId && cr.competencyId === rh.competencyId
		);
		if (idx >= 0) {
			merged[idx] = { ...merged[idx], rhRating: rh.rhRating, rhComment: rh.rhComment };
		} else {
			merged.push(structuredClone(rh));
		}
	}
	return merged;
}

function loadFixtures(): StoreData {
	const selfRatings = structuredClone(selfEvaluationsData as CompetencyRating[]);
	const rhRatings = structuredClone(rhEvaluationsData as CompetencyRating[]);

	return {
		competencyRatings: mergeRHEvaluations(selfRatings, rhRatings),
		goalClosures: structuredClone(goalClosuresData as GoalClosure[])
	};
}

// ─── Normalize API response → StoreData ───────────────────────────────────────

function normalizeApiData(
	detail: {
		employeeId?: string;
		competencies?: Array<{
			competencyId?: string;
			rating?: number;
			comments?: string;
		}>;
		goals?: Array<{
			goalId?: string;
			finalRating?: number | null;
			finalComments?: string;
		}>;
	} | null,
	empId: string
): StoreData {
	const competencyRatings: CompetencyRating[] = (detail?.competencies ?? []).map((c, i) => ({
		id: `api-cr-${empId}-${c.competencyId ?? i}-${Date.now()}`,
		employeeId: empId,
		competencyId: c.competencyId ?? '',
		selfRating: c.rating != null ? (c.rating as 1 | 2 | 3 | 4 | 5) : undefined,
		selfComment: c.comments
	}));

	const goalClosures: GoalClosure[] = (detail?.goals ?? []).map((g, i) => ({
		id: `api-gc-${empId}-${g.goalId ?? i}-${Date.now()}`,
		employeeId: empId,
		goalId: g.goalId ?? '',
		finalProgress: g.finalRating ?? 0,
		selfAssessment: g.finalComments
	}));

	return { competencyRatings, goalClosures };
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

function isFinAnio(): boolean {
	return getActivePhase() === 'fin-anio';
}

// ─── Load / Reload ────────────────────────────────────────────────────────────

/**
 * Load evaluation data.
 *
 * In DEV without VITE_USE_API: loads from fixture files (structured clone + RH merge).
 * In production / VITE_USE_API=true: fetches from the real API endpoints.
 */
export async function load(): Promise<void> {
	loading = true;
	error = null;

	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		data = loadFixtures();
		loading = false;
		return;
	}

	try {
		const empId = getSession().user?.employeeId;
		if (!empId) throw new Error('No hay sesión activa');

		const { data: apiData, error: apiError } = await client.GET('/evaluations/{id}', {
			params: { path: { id: empId } }
		});

		if (apiError) {
			throw new Error(
				typeof apiError === 'string' ? apiError : 'Error al cargar evaluación'
			);
		}

		data = normalizeApiData(apiData as Parameters<typeof normalizeApiData>[0], empId);
	} catch (e) {
		data = null;
		error = e instanceof Error ? e.message : 'Error desconocido al cargar evaluaciones';
	} finally {
		loading = false;
	}
}

/** Alias for load(). */
export function reload(): Promise<void> {
	return load();
}

// ─── Getters ───────────────────────────────────────────────────────────────────

export function getCompetencyRatings(employeeId: string): CompetencyRating[] {
	return (data?.competencyRatings ?? []).filter((cr) => cr.employeeId === employeeId);
}

export function getCompetencyRating(
	employeeId: string,
	competencyId: string
): CompetencyRating | undefined {
	return (data?.competencyRatings ?? []).find(
		(cr) => cr.employeeId === employeeId && cr.competencyId === competencyId
	);
}

export function getGoalClosures(employeeId: string): GoalClosure[] {
	return (data?.goalClosures ?? []).filter((gc) => gc.employeeId === employeeId);
}

export function getGoalClosure(employeeId: string, goalId: string): GoalClosure | undefined {
	return (data?.goalClosures ?? []).find(
		(gc) => gc.employeeId === employeeId && gc.goalId === goalId
	);
}

export function getEvaluationStatus(
	employeeId: string,
	totalCompetencies: number,
	goalIds: string[]
): EvaluationStatus {
	const empRatings = (data?.competencyRatings ?? []).filter(
		(cr) => cr.employeeId === employeeId
	);
	const ratedCompetencies = empRatings.filter((cr) => cr.selfRating !== undefined).length;

	if (ratedCompetencies === 0) return 'pending';

	const empClosures = (data?.goalClosures ?? []).filter(
		(gc) => gc.employeeId === employeeId
	);
	const closedGoals = empClosures.filter((gc) => gc.closedAt !== undefined).length;

	if (ratedCompetencies >= totalCompetencies && closedGoals >= goalIds.length)
		return 'completed';

	return 'in-progress';
}

// ─── Mutations: Employee Self-Evaluation ───────────────────────────────────────

export async function rateCompetency(
	employeeId: string,
	competencyId: string,
	level: 1 | 2 | 3 | 4 | 5,
	comment?: string
): Promise<void> {
	if (!isFinAnio()) return;

	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		data = {
			...data!,
			competencyRatings: applyRating(data?.competencyRatings ?? [], employeeId, competencyId, level, comment, 'self')
		};
		return;
	}

	const empId = getSession().user?.employeeId;
	if (!empId) return;

	const allRatings = (data?.competencyRatings ?? []).filter(
		(cr) => cr.employeeId === employeeId
	);
	const competencies = allRatings.map((r) => ({
		competencyId: r.competencyId,
		rating:
			r.competencyId === competencyId
				? level
				: (r.selfRating ?? 0),
		comments:
			r.competencyId === competencyId
				? comment
				: r.selfComment
	}));
	// If the competency isn't in the list yet, add it
	const existingIds = new Set(competencies.map((c) => c.competencyId));
	if (!existingIds.has(competencyId)) {
		competencies.push({ competencyId, rating: level, comments: comment });
	}

	const { error: apiError } = await client.PUT('/evaluations/{id}/self-evaluation', {
		params: { path: { id: empId } },
		header: { 'If-Match': 1 } as never,
		body: { competencies }
	});
	if (apiError) throw new Error('Error al guardar autoevaluación');
	await reload();
}

export async function closeGoal(
	employeeId: string,
	goalId: string,
	finalProgress: number,
	selfAssessment: string
): Promise<void> {
	if (!isFinAnio()) return;

	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		const existing = (data?.goalClosures ?? []).find(
			(gc) => gc.employeeId === employeeId && gc.goalId === goalId
		);
		if (existing) {
			data = {
				...data!,
				goalClosures: (data?.goalClosures ?? []).map((gc) =>
					gc.employeeId === employeeId && gc.goalId === goalId
						? {
								...gc,
								finalProgress,
								selfAssessment,
								// eslint-disable-next-line svelte/prefer-svelte-reactivity
								closedAt: gc.closedAt ?? new Date().toISOString()
							}
						: gc
				)
			};
		} else {
			data = {
				...data!,
				goalClosures: [
					...(data?.goalClosures ?? []),
					{
						id: `gc-${employeeId}-${goalId}-${Date.now()}`,
						employeeId,
						goalId,
						finalProgress,
						selfAssessment,
						// eslint-disable-next-line svelte/prefer-svelte-reactivity
						closedAt: new Date().toISOString()
					}
				]
			};
		}
		return;
	}

	const empId = getSession().user?.employeeId;
	if (!empId) return;

	const goalComments = (data?.goalClosures ?? [])
		.filter((gc) => gc.employeeId === employeeId)
		.map((gc) => ({
			goalId: gc.goalId,
			comment:
				gc.goalId === goalId ? selfAssessment : gc.selfAssessment
		}));
	// If not in list yet, add it
	if (!goalComments.some((g) => g.goalId === goalId)) {
		goalComments.push({ goalId, comment: selfAssessment });
	}

	const { error: apiError } = await client.PUT('/evaluations/{id}/self-evaluation', {
		params: { path: { id: empId } },
		header: { 'If-Match': 1 } as never,
		body: { goalComments, competencies: [] }
	});
	if (apiError) throw new Error('Error al cerrar meta');
	await reload();
}

// ─── Mutations: RH Evaluation ─────────────────────────────────────────────────

export async function rhRateCompetency(
	employeeId: string,
	competencyId: string,
	level: 1 | 2 | 3 | 4 | 5,
	comment?: string
): Promise<void> {
	if (!isFinAnio()) return;

	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		data = {
			...data!,
			competencyRatings: applyRating(data?.competencyRatings ?? [], employeeId, competencyId, level, comment, 'rh')
		};
		return;
	}

	const empId = getSession().user?.employeeId;
	if (!empId) return;

	const allRatings = (data?.competencyRatings ?? []).filter(
		(cr) => cr.employeeId === employeeId
	);
	const competencies = allRatings.map((r) => ({
		competencyId: r.competencyId,
		rating:
			r.competencyId === competencyId
				? level
				: (r.rhRating ?? r.selfRating ?? 0),
		comments:
			r.competencyId === competencyId
				? comment
				: (r.rhComment ?? r.selfComment)
	}));
	const existingIds = new Set(competencies.map((c) => c.competencyId));
	if (!existingIds.has(competencyId)) {
		competencies.push({ competencyId, rating: level, comments: comment });
	}

	const { error: apiError } = await client.PUT('/evaluations/{id}/rh-evaluation', {
		params: { path: { id: empId } },
		header: { 'If-Match': 1 } as never,
		body: { competencies }
	});
	if (apiError) throw new Error('Error al guardar evaluación RH');
	await reload();
}

export async function rhAssessGoal(
	employeeId: string,
	goalId: string,
	rhAssessment: string
): Promise<void> {
	if (!isFinAnio()) return;

	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		data = {
			...data!,
			goalClosures: (data?.goalClosures ?? []).map((gc) =>
				gc.employeeId === employeeId && gc.goalId === goalId
					? { ...gc, rhAssessment }
					: gc
			)
		};
		return;
	}

	const empId = getSession().user?.employeeId;
	if (!empId) return;

	// For RH goal assessment there's no dedicated endpoint in the schema;
	// fall back to local-only update + reload
	data = {
		...data!,
		goalClosures: (data?.goalClosures ?? []).map((gc) =>
			gc.employeeId === employeeId && gc.goalId === goalId
				? { ...gc, rhAssessment }
				: gc
		)
	};
	await reload();
}

// ─── Mutations: Manager ───────────────────────────────────────────────────────

export async function addManagerComment(
	employeeId: string,
	goalId: string,
	comment: string
): Promise<void> {
	if (!isFinAnio()) return;

	// No dedicated API endpoint for manager comments; local-only for now.
	data = {
		...data!,
		goalClosures: (data?.goalClosures ?? []).map((gc) =>
			gc.employeeId === employeeId && gc.goalId === goalId
				? { ...gc, managerComment: comment }
				: gc
		)
	};
}

// ─── Batch Submit / Finalize ───────────────────────────────────────────────────

/**
 * Submit the full self-evaluation (competencies + goal comments) for the
 * current employee. In DEV mode this is a no-op (data is already local).
 * In production it calls POST /evaluations/{id}/self-evaluation.
 */
export async function submitSelfEvaluation(): Promise<void> {
	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) return;

	const empId = getSession().user?.employeeId;
	if (!empId) return;

	const empRatings = (data?.competencyRatings ?? []).filter(
		(cr) => cr.employeeId === empId
	);
	const empClosures = (data?.goalClosures ?? []).filter(
		(gc) => gc.employeeId === empId
	);

	const { error: apiError } = await client.POST('/evaluations/{id}/self-evaluation', {
		params: { path: { id: empId } },
		header: { 'Idempotency-Key': `self-eval-${empId}-${Date.now()}` } as never,
		body: {
			competencies: empRatings.map((r) => ({
				competencyId: r.competencyId,
				rating: r.selfRating ?? 0,
				comments: r.selfComment
			})),
			goalComments: empClosures.map((c) => ({
				goalId: c.goalId,
				comment: c.selfAssessment
			}))
		}
	});
	if (apiError) throw new Error('Error al enviar autoevaluación');
	await reload();
}

/**
 * Submit the full RH evaluation for the given employee.
 * In production it calls POST /evaluations/{id}/rh-evaluation.
 */
export async function submitRHEvaluation(employeeId: string): Promise<void> {
	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) return;

	const empId = getSession().user?.employeeId;
	if (!empId) return;

	const empRatings = (data?.competencyRatings ?? []).filter(
		(cr) => cr.employeeId === employeeId
	);

	const { error: apiError } = await client.POST('/evaluations/{id}/rh-evaluation', {
		params: { path: { id: empId } },
		header: { 'Idempotency-Key': `rh-eval-${empId}-${employeeId}-${Date.now()}` } as never,
		body: {
			competencies: empRatings.map((r) => ({
				competencyId: r.competencyId,
				rating: r.rhRating ?? r.selfRating ?? 0,
				comments: r.rhComment ?? r.selfComment
			}))
		}
	});
	if (apiError) throw new Error('Error al enviar evaluación RH');
	await reload();
}

/**
 * Finalize the evaluation for the current employee.
 * In production it calls POST /evaluations/{id}/finalize.
 */
export async function finalizeEvaluation(reason?: string): Promise<void> {
	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) return;

	const empId = getSession().user?.employeeId;
	if (!empId) return;

	const { error: apiError } = await client.POST('/evaluations/{id}/finalize', {
		params: { path: { id: empId } },
		body: reason ? { reason } : undefined
	});
	if (apiError) throw new Error('Error al finalizar evaluación');
	await reload();
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

function applyRating(
	ratings: CompetencyRating[],
	employeeId: string,
	competencyId: string,
	level: 1 | 2 | 3 | 4 | 5,
	comment: string | undefined,
	mode: 'self' | 'rh'
): CompetencyRating[] {
	const existing = ratings.find(
		(cr) => cr.employeeId === employeeId && cr.competencyId === competencyId
	);
	if (existing) {
		return ratings.map((cr) =>
			cr.employeeId === employeeId && cr.competencyId === competencyId
				? {
						...cr,
						...(mode === 'self'
							? { selfRating: level, selfComment: comment ?? cr.selfComment }
							: { rhRating: level, rhComment: comment ?? cr.rhComment })
					}
				: cr
		);
	}
	return [
		...ratings,
		{
			id: `${mode === 'self' ? 'sr' : 'rh'}-${employeeId}-${competencyId}-${Date.now()}`,
			employeeId,
			competencyId,
			...(mode === 'self'
				? { selfRating: level, selfComment: comment }
				: { rhRating: level, rhComment: comment })
		}
	];
}
