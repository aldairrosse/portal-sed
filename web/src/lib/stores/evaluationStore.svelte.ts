import type { CompetencyRating, GoalClosure, EvaluationStatus } from '$lib/types/evaluation-result';
import { getActivePhase } from '$lib/api/cycle.svelte';
import { getSession } from '$lib/api/session.svelte';
import { client } from '$lib/api/client';


// ─── Internal data shape ──────────────────────────────────────────────────────

interface StoreData {
	competencyRatings: CompetencyRating[];
	goalClosures: GoalClosure[];
}

// ─── Triplete state ───────────────────────────────────────────────────────────

let data = $state<StoreData | null>(null);
let loading = $state(true);
let error = $state<string | null>(null);
let loadPromise: Promise<void> | null = null;
let lastLoadTime = 0;
const FRESHNESS_MS = 5000;

/** @returns true while load() is in progress. */
export function isLoading(): boolean {
	return loading;
}

/** @returns the current error message, or null if no error. */
export function getError(): string | null {
	return error;
}

// ─── Normalize API response → StoreData ───────────────────────────────────────

function normalizeApiData(
	detail: {
		employeeId?: string;
		competencies?: Array<{
			competencyId?: string;
			rating?: number;
			selfRating?: number | null;
			rhRating?: number | null;
			comments?: string;
			acceptanceLevel?: number | null;
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
		selfRating: (c.selfRating ?? c.rating) != null ? ((c.selfRating ?? c.rating) as 1 | 2 | 3 | 4 | 5) : undefined,
		selfComment: c.comments,
		rhRating: c.rhRating != null ? (c.rhRating as 1 | 2 | 3 | 4 | 5) : undefined,
		acceptanceLevel: c.acceptanceLevel ?? undefined
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
 * When `employeeId` is provided: fetches ratings for that employee via the
 * employee-specific endpoint (used by competency network view for arbitrary employees).
 * When `employeeId` is absent: fetches the current session user's evaluation (existing path).
 *
 * In DEV without VITE_USE_API: loads from fixture files (structured clone + RH merge).
 * In production / VITE_USE_API=true: fetches from the real API endpoints.
 */
export async function load(employeeId?: string): Promise<void> {
	if (loadPromise) return loadPromise;
	if (data && Date.now() - lastLoadTime < FRESHNESS_MS) return;
	loadPromise = _doLoad(employeeId);
	try {
		await loadPromise;
	} finally {
		loadPromise = null;
		lastLoadTime = Date.now();
	}
}

async function _doLoad(employeeId?: string): Promise<void> {
	loading = true;
	error = null;

	try {
		if (employeeId) {
			// Employee-specific path (competency network for manager/RH view)
			// cycle_id is optional — backend resolves the active cycle automatically
			const { data: apiData, error: apiError } = await client.GET(
				'/evaluations/employee/{employeeId}',
				{ params: { path: { employeeId } } }
			);

			if (apiError) {
				throw new Error(
					typeof apiError === 'string' ? apiError : 'Error al cargar competencias del empleado'
				);
			}

			// Map ratings to competencies for normalizeApiData
			const resp = apiData as { employeeId?: string; cycleId?: string; ratings?: Array<{ competencyId?: string; selfRating?: number | null; rhRating?: number | null; comments?: string; acceptanceLevel?: number | null }> } | undefined;
			const mapped = {
				employeeId: resp?.employeeId ?? employeeId,
				competencies: (resp?.ratings ?? []).map((r) => ({
					competencyId: r.competencyId,
					selfRating: r.selfRating,
					rhRating: r.rhRating,
					comments: r.comments,
					acceptanceLevel: r.acceptanceLevel
				})),
				goals: []
			};
			data = normalizeApiData(mapped, employeeId);
		} else {
			// Existing path: self-evaluation by session user
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
		}
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

	const empId = getSession().user?.employeeId;
	if (!empId) return;

	const { error: apiError } = await client.PUT('/evaluations/{id}/goal-state', {
		params: { path: { id: empId } },
		header: { 'If-Match': 1 } as never,
		body: {
			goalId,
			finalProgress,
			selfAssessment
		}
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

	const empId = getSession().user?.employeeId;
	if (!empId) return;

	const { error: apiError } = await client.PUT('/evaluations/{id}/goal-state', {
		params: { path: { id: empId } },
		header: { 'If-Match': 1 } as never,
		body: {
			goalId,
			rhAssessment
		}
	});
	if (apiError) throw new Error('Error al guardar evaluación RH de la meta');
	await reload();
}

// ─── Mutations: Manager ───────────────────────────────────────────────────────

export async function addManagerComment(
	employeeId: string,
	goalId: string,
	comment: string
): Promise<void> {
	if (!isFinAnio()) return;

	const empId = getSession().user?.employeeId;
	if (!empId) return;

	const { error: apiError } = await client.PUT('/evaluations/{id}/goal-comments', {
		params: { path: { id: empId } },
		header: { 'If-Match': 1 } as never,
		body: {
			goalId,
			role: 'manager',
			comment
		}
	});
	if (apiError) throw new Error('Error al guardar comentario del manager');
	await reload();
}

// ─── Batch Submit / Finalize ───────────────────────────────────────────────────

/**
 * Submit the full self-evaluation (competencies + goal comments) for the
 * current employee. In DEV mode this is a no-op (data is already local).
 * In production it calls POST /evaluations/{id}/self-evaluation.
 */
export async function submitSelfEvaluation(): Promise<void> {
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
