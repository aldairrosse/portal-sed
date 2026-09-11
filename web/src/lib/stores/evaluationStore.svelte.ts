import type { CompetencyRating, GoalClosure, EvaluationStatus } from '$lib/types/evaluation-result';
import { getActivePhase, getActiveCycleId, loadCycle } from '$lib/api/cycle.svelte';
import { isAvance, isCierre } from '$lib/types/cycle';
import { getSession } from '$lib/api/session.svelte';
import { client, HttpNotFoundError } from '$lib/api/client';
import { SvelteMap, SvelteSet } from 'svelte/reactivity';


// ─── Internal data shape ──────────────────────────────────────────────────────

interface StoreData {
	competencyRatings: CompetencyRating[];
	goalClosures: GoalClosure[];
	version: number | null;
}

// ─── Evaluation ID resolution ─────────────────────────────────────────────────
// Backend `{id}` params are evaluationId (uuid), not employeeId. Cache maps
// employeeId → evaluationId from detail/list responses. Never falls back to
// employeeId silently — callers must resolve first (throws EVALUATION_NOT_FOUND).
const evaluationIds = new SvelteMap<string, string>();

function rememberEvaluationId(employeeId: string, evaluationId: string | undefined): void {
	if (evaluationId) evaluationIds.set(employeeId, evaluationId);
}

function resolveEvaluationId(employeeId: string): string {
	const id = evaluationIds.get(employeeId);
	if (!id) {
		const err = new Error(
			`Sin evaluationId para empleado ${employeeId} (ciclo ${getActiveCycleId() ?? 'sin cargar'}): llama load(employeeId) primero para resolverlo vía GET /evaluations?cycle_id; si no existe, créala en fase avance/cierre`
		);
		(err as Error & { code?: string }).code = EVALUATION_NOT_FOUND;
		throw err;
	}
	return id;
}

/** Devuelve el evaluationId en caché sin lanzar (para logs/diagnóstico). */
export function peekEvaluationId(employeeId: string): string | undefined {
	return evaluationIds.get(employeeId);
}

/**
 * Query params para el upsert server-side: si el {id} del path no existe,
 * el backend crea/reutiliza la evaluación para employee_id+cycle_id.
 */
function upsertQuery(employeeId: string): { employee_id: string; cycle_id?: string } {
	const cycleId = getActiveCycleId();
	return cycleId ? { employee_id: employeeId, cycle_id: cycleId } : { employee_id: employeeId };
}

/**
 * Id provisional para el path: evaluationId en caché, o employeeId (uuid
 * válido) para que el request salga y el backend resuelva/cree vía query.
 * Nunca lanza EVALUATION_NOT_FOUND antes del PUT.
 */
function pathIdFor(employeeId: string): string {
	return evaluationIds.get(employeeId) ?? employeeId;
}

interface EvaluationListItem {
	id?: string;
	employeeId?: string;
	cycleId?: string;
}

/** Ensures cycleStore is loaded and returns the active cycle id. */
async function ensureCycleLoaded(): Promise<string> {
	let cycleId = getActiveCycleId();
	if (!cycleId) {
		await loadCycle();
		cycleId = getActiveCycleId();
	}
	if (!cycleId) throw new Error('Sin ciclo activo: no se puede resolver evaluationId');
	return cycleId;
}

/**
 * Finds the evaluationId for an employee via GET /evaluations?cycle_id&limit=100,
 * following cursor pagination. Never sends employee_id as query (backend ignores it).
 */
async function findEvaluationInList(
	employeeId: string,
	cycleId: string
): Promise<string | undefined> {
	let cursor: string | undefined;
	do {
		const { data: listData, error: listError } = await client.GET('/evaluations', {
			params: { query: { cycle_id: cycleId, limit: 100, cursor } }
		});
		if (listError) {
			console.warn('[evalId] NOT_FOUND', { employeeId, cycleId });
			return undefined;
		}
		const page = listData as
			| { data?: EvaluationListItem[]; nextCursor?: string; next_cursor?: string }
			| undefined;
		const items = page?.data ?? [];
		const found = items.find(
			(i) => i.employeeId === employeeId && i.cycleId === cycleId
		);
		if (found?.id) return found.id;
		cursor = page?.nextCursor ?? page?.next_cursor;
	} while (cursor);
	console.warn('[evalId] NOT_FOUND', { employeeId, cycleId });
	return undefined;
}

/** Resolves (and caches) the evaluationId for an employee. Returns undefined if not found. */
async function ensureEvaluationId(employeeId: string): Promise<string | undefined> {
	const cached = evaluationIds.get(employeeId);
	if (cached) {
		return cached;
	}
	const cycleId = await ensureCycleLoaded();
	const found = await findEvaluationInList(employeeId, cycleId);
		if (!found) console.warn('[evalId] NOT_FOUND', { employeeId, cycleId });
		rememberEvaluationId(employeeId, found);
	return found;
}

/** Evicts the cached id and re-resolves via list (heals stale cache after recarga). */
async function refreshEvaluationId(employeeId: string): Promise<string | undefined> {
	evaluationIds.delete(employeeId);
	try {
		return await ensureEvaluationId(employeeId);
	} catch {
		return undefined;
	}
}

/**
 * Best-effort create attempt when no evaluation exists in avance/cierre.
 * No hay endpoint dedicado POST /evaluations (solo GET lista + GET detalle),
 * así que NO se inventa un POST con employeeId como path {id} ni con
 * `competencies: []` (backend responde 400 "at least one competency rating").
 * Solo re-resuelve vía lista; si sigue sin existir, el PUT lanzará error claro.
 * El PUT /goal-state debería crear server-side a futuro.
 */
async function tryAutoCreate(employeeId: string, viewerMode?: ViewerMode): Promise<void> {
	if (!isEditablePhase()) return;
	let cycleId: string | undefined;
	try {
		cycleId = await ensureCycleLoaded();
	} catch (e) {
		console.error('[autoCreate] failed', e);
		return;
	}
	void viewerMode;
	try {
		const found = await findEvaluationInList(employeeId, cycleId);
		if (found) rememberEvaluationId(employeeId, found);
		else console.warn('[evalId] NOT_FOUND', { employeeId, cycleId });
	} catch (e) {
		console.error('[autoCreate] failed', e);
	}
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
		version?: number | null;
		competencies?: Array<{
			competencyId?: string;
			rating?: number;
			selfRating?: number | null;
			rhRating?: number | null;
			comments?: string;
			selfComment?: string | null;
			rhComment?: string | null;
			managerComment?: string | null;
			managerCommentAuthor?: string | null;
			managerCommentCreatedAt?: string | null;
			acceptanceLevel?: number | null;
			authorName?: string | null;
			createdAt?: string | null;
		}>;
		goals?: Array<{
			goalId?: string;
			finalRating?: number | null;
			finalProgress?: number | null;
			avanceProgress?: number | null;
			cierreProgress?: number | null;
			finalComments?: string;
			rhAssessment?: string | null;
			managerComment?: string | null;
			authorName?: string | null;
			createdAt?: string | null;
			updatedAt?: string | null;
			managerCommentAuthor?: string | null;
			managerCommentCreatedAt?: string | null;
		}>;
	} | null,
	empId: string
): StoreData {
	const competencyRatings: CompetencyRating[] = (detail?.competencies ?? []).map((c, i) => ({
		id: `api-cr-${empId}-${c.competencyId ?? i}`,
		employeeId: empId,
		competencyId: c.competencyId ?? '',
		selfRating: c.selfRating != null ? (c.selfRating as 1 | 2 | 3 | 4 | 5) : undefined,
		selfComment: c.selfComment ?? '',
		rhRating: c.rhRating != null ? (c.rhRating as 1 | 2 | 3 | 4 | 5) : undefined,
		rhComment: c.rhComment ?? '',
		managerComment: c.managerComment ?? '',
		managerCommentAuthor: c.managerCommentAuthor ?? undefined,
		managerCommentCreatedAt: c.managerCommentCreatedAt ?? undefined,
		authorName: c.authorName ?? undefined,
		commentCreatedAt: c.createdAt ?? undefined,
		acceptanceLevel: c.acceptanceLevel ?? undefined
	}));

	const goalClosures: GoalClosure[] = (detail?.goals ?? []).map((g, i) => ({
		id: `api-gc-${empId}-${g.goalId ?? i}`,
		employeeId: empId,
		goalId: g.goalId ?? '',
		// Direct snapshot value (cierre ?? avance via backend finalProgress); never derive from 1-5 finalRating.
		finalProgress: g.finalProgress ?? g.cierreProgress ?? g.avanceProgress ?? 0,
		selfAssessment: g.finalComments,
		rhAssessment: g.rhAssessment ?? undefined,
		managerComment: g.managerComment ?? undefined,
		// F2: row-level author/date shared by the three comments.
		selfAssessmentAuthor: g.authorName ?? g.managerCommentAuthor ?? undefined,
		selfAssessmentCreatedAt: g.createdAt ?? g.updatedAt ?? g.managerCommentCreatedAt ?? undefined,
		rhAssessmentAuthor: g.authorName ?? g.managerCommentAuthor ?? undefined,
		rhAssessmentCreatedAt: g.updatedAt ?? g.createdAt ?? g.managerCommentCreatedAt ?? undefined,
		managerCommentAuthor: g.managerCommentAuthor ?? g.authorName ?? undefined,
		managerCommentCreatedAt: g.managerCommentCreatedAt ?? g.updatedAt ?? g.createdAt ?? undefined
	}));

	const version =
		typeof detail?.version === 'number' ? (detail.version as number) : null;
	return { competencyRatings, goalClosures, version };
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

export const EVALUATION_NOT_FOUND = 'EVALUATION_NOT_FOUND';

/** Extracts { message, code } from API/client errors. Code is null when unknown. */
export function errorWithCode(e: unknown): { message: string; code: string | null } {
	if (e instanceof HttpNotFoundError) {
		return { message: 'Evaluación no encontrada, se creará automáticamente', code: EVALUATION_NOT_FOUND };
	}
	if (e instanceof Error) {
		const body = (e as Error & { body?: unknown }).body as
			| { code?: string; error?: { code?: string; message?: string } }
			| undefined;
		const code =
			body?.code ?? body?.error?.code ?? (e as Error & { code?: string }).code ?? null;
		return { message: e.message, code: typeof code === 'string' ? code : null };
	}
	if (typeof e === 'object' && e !== null) {
		const obj = e as { code?: unknown; error?: { code?: unknown; message?: unknown } };
		const code = obj.code ?? obj.error?.code;
		const msg = obj.error?.message;
		return {
			message: typeof msg === 'string' ? msg : 'Error desconocido',
			code: typeof code === 'string' ? code : null
		};
	}
	return { message: 'Error desconocido', code: null };
}

function isNotFoundCode(code: string | null, e: unknown): boolean {
	return code === EVALUATION_NOT_FOUND || e instanceof HttpNotFoundError;
}

/** F4: true when the backend rejected the write for stale version/ETag. Only then a reload is allowed. */
function isVersionConflict(code: string | null, message?: string): boolean {
	const hay = `${code ?? ''} ${message ?? ''}`.toUpperCase();
	return (
		hay.includes('VERSION') ||
		hay.includes('CONFLICT') ||
		hay.includes('409') ||
		hay.includes('412') ||
		hay.includes('IF-MATCH') ||
		hay.includes('IF_MATCH') ||
		hay.includes('PRECONDITION') ||
		hay.includes('STALE')
	);
}

function isEditablePhase(): boolean {
	const phase = getActivePhase() ?? 'inicio-anio';
	return isAvance(phase) || isCierre(phase);
}

function assertValidLevel(level: unknown): asserts level is 1 | 2 | 3 | 4 | 5 {
	if (level !== 1 && level !== 2 && level !== 3 && level !== 4 && level !== 5) {
		throw new Error('Selecciona una calificación de 1 a 5 antes de guardar');
	}
}

/** Current optimistic-lock version (detail.version); 1 when never loaded. */
function currentVersion(): number {
	return data?.version ?? 1;
}

/** Copies version from a detail/PUT response into the store. */
function syncVersion(resp: { version?: number | null } | undefined): void {
	if (resp && typeof resp.version === 'number' && data) {
		data.version = resp.version;
	}
}

/**
 * F4: optimistic patch of a single goal closure without reload.
 * Only the matching goalId entry is replaced; other cards keep identity/state.
 */
function patchGoalClosure(
	employeeId: string,
	goalId: string,
	patch: Partial<GoalClosure>
): void {
	if (!data) return;
	const idx = data.goalClosures.findIndex(
		(gc) => gc.employeeId === employeeId && gc.goalId === goalId
	);
	if (idx >= 0) {
		data.goalClosures[idx] = { ...data.goalClosures[idx], ...patch };
	} else {
		data.goalClosures.push({
			id: `api-gc-${employeeId}-${goalId}`,
			employeeId,
			goalId,
			finalProgress: 0,
			...patch
		});
	}
}

/** F4: optimistic patch of a single competency rating without reload. */
function patchCompetencyRating(
	employeeId: string,
	competencyId: string,
	patch: Partial<CompetencyRating>
): void {
	if (!data) return;
	const idx = data.competencyRatings.findIndex(
		(cr) => cr.employeeId === employeeId && cr.competencyId === competencyId
	);
	if (idx >= 0) {
		data.competencyRatings[idx] = { ...data.competencyRatings[idx], ...patch };
	} else {
		data.competencyRatings.push({
			id: `api-cr-${employeeId}-${competencyId}`,
			employeeId,
			competencyId,
			...patch
		});
	}
}

async function autoCreateSelf(empId: string, competencyId: string, level: number, comment?: string): Promise<string | undefined> {
	if (!isEditablePhase()) return undefined;
	if (!getActiveCycleId()) {
		try {
			await ensureCycleLoaded();
		} catch {
			return undefined;
		}
	}
	const { data } = await client.POST('/evaluations/{id}/self-evaluation', {
		params: { path: { id: empId }, header: { 'Idempotency-Key': crypto.randomUUID() } },
		body: {
			competencies: [{ competencyId, rating: level, comments: comment }],
			goalComments: []
		}
	});
	const id = (data as { id?: string } | undefined)?.id;
	if (id) rememberEvaluationId(empId, id);
	return id;
}

async function autoCreateRh(empId: string, competencyId: string, level: number, comment?: string, managerComment?: string): Promise<string | undefined> {
	if (!isEditablePhase()) return undefined;
	if (!getActiveCycleId()) {
		try {
			await ensureCycleLoaded();
		} catch {
			return undefined;
		}
	}
	const { data } = await client.POST('/evaluations/{id}/rh-evaluation', {
		params: { path: { id: empId }, header: { 'Idempotency-Key': crypto.randomUUID() } },
		body: {
			competencies: [
				{
					competencyId,
					rating: level,
					comments: comment,
					...(managerComment !== undefined ? { managerComment } : {})
				}
			]
		}
	});
	const id = (data as { id?: string } | undefined)?.id;
	if (id) rememberEvaluationId(empId, id);
	return id;
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
let lastEmployeeId: string | undefined;
export type ViewerMode = 'self' | 'rh' | 'manager';
let lastViewerMode: ViewerMode | undefined;

export async function load(employeeId?: string, viewerMode?: ViewerMode): Promise<void> {
	if (loadPromise) {
		if (employeeId === lastEmployeeId && viewerMode === lastViewerMode) return loadPromise;
		try {
			await loadPromise;
		} catch {
			// ponytail: cambio de empleado/modo → fetch fresco abajo
		}
	}
	if (
		data &&
		Date.now() - lastLoadTime < FRESHNESS_MS &&
		employeeId === lastEmployeeId &&
		viewerMode === lastViewerMode
	)
		return;
	lastEmployeeId = employeeId;
	lastViewerMode = viewerMode;
	loadPromise = _doLoad(employeeId, viewerMode);
	try {
		await loadPromise;
	} finally {
		loadPromise = null;
		lastLoadTime = Date.now();
	}
}

async function _doLoad(employeeId?: string, viewerMode?: ViewerMode): Promise<void> {
	loading = true;
	error = null;

	try {
		if (employeeId) {
			// Employee-specific path (competency network for manager/RH view)
			// cycle_id is optional — backend resolves the active cycle automatically
			try {
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
				const resp = apiData as { employeeId?: string; cycleId?: string; ratings?: Array<{ competencyId?: string; selfRating?: number | null; rhRating?: number | null; comments?: string; selfComment?: string | null; rhComment?: string | null; managerComment?: string | null; managerCommentAuthor?: string | null; managerCommentCreatedAt?: string | null; acceptanceLevel?: number | null; authorName?: string | null; createdAt?: string | null }> } | undefined;
				const mapped = {
					employeeId: resp?.employeeId ?? employeeId,
					competencies: (resp?.ratings ?? []).map((r) => ({
						competencyId: r.competencyId,
						selfRating: r.selfRating,
						rhRating: r.rhRating,
						comments: r.comments,
						selfComment: r.selfComment,
						rhComment: r.rhComment,
						managerComment: r.managerComment,
						managerCommentAuthor: r.managerCommentAuthor,
						managerCommentCreatedAt: r.managerCommentCreatedAt,
						acceptanceLevel: r.acceptanceLevel,
						authorName: r.authorName,
						createdAt: r.createdAt
					})),
					goals: []
				};
				data = normalizeApiData(mapped, employeeId);
				// GET /evaluations/employee/{employeeId} has no evaluation id —
				// resolve it via GET /evaluations?cycle_id&limit=100 (EvaluationListItem.id).
				// Never send employee_id as query: backend ignores it.
				let evalId: string | undefined;
				try {
					evalId = await ensureEvaluationId(employeeId);
				} catch {
					evalId = undefined;
				}
				if (!evalId && isEditablePhase()) {
					await tryAutoCreate(employeeId, viewerMode);
					try {
						evalId = await ensureEvaluationId(employeeId);
					} catch {
						// ponytail: sin evaluationId el PUT lanzará error claro, no fallback silencioso
					}
			}
			if (!evalId) console.warn('[evalId] NOT_FOUND', { employeeId, cycleId: getActiveCycleId() });
				// Manager/RH: hidratar goals desde el detalle (mismo camino que self).
				// GET /evaluations/employee/{employeeId} solo trae competencias.
				if (evalId) {
					try {
						const { data: detailData } = await client.GET('/evaluations/{id}', {
							params: { path: { id: evalId } }
						});
						if (detailData) {
							const normalized = normalizeApiData(
								detailData as Parameters<typeof normalizeApiData>[0],
								employeeId
							);
							data = {
								competencyRatings: data?.competencyRatings ?? [],
								goalClosures: normalized.goalClosures,
								version: normalized.version ?? data?.version ?? null
							};
							syncVersion(detailData as { version?: number | null });
							rememberEvaluationId(
								employeeId,
								(detailData as { id?: string } | undefined)?.id ?? evalId
							);
						}
					} catch {
						// ponytail: sin detalle se conservan competencias; metas quedan vacías
					}
				}
			} catch (e) {
				if (e instanceof HttpNotFoundError) {
					data = normalizeApiData(null, employeeId);
					if (isEditablePhase()) {
						await tryAutoCreate(employeeId, viewerMode);
						try {
							await ensureEvaluationId(employeeId);
						} catch {
							// ponytail: si el POST falla, cae al estado vacío
						}
					}
				} else {
					throw e;
				}
			}
		} else {
			// Self path: resolve evaluationId via list first — never GET
			// /evaluations/{employeeId} ({id} is evaluationId, not employeeId).
			const empId = getSession().user?.employeeId;
			if (!empId) throw new Error('No hay sesión activa');

			try {
				let evalId: string | undefined;
				try {
					evalId = await ensureEvaluationId(empId);
				} catch {
					evalId = undefined;
				}
				if (!evalId && isEditablePhase()) {
					await tryAutoCreate(empId, 'self');
					try {
						evalId = await ensureEvaluationId(empId);
					} catch {
						evalId = undefined;
					}
				}
				if (!evalId) {
					// ponytail: no evaluation yet, show empty state
					data = normalizeApiData(null, empId);
				} else {
					const { data: apiData, error: apiError } = await client.GET('/evaluations/{id}', {
						params: { path: { id: evalId } }
					});

					if (apiError) {
						throw new Error(
							typeof apiError === 'string' ? apiError : 'Error al cargar evaluación'
						);
					}

					data = normalizeApiData(apiData as Parameters<typeof normalizeApiData>[0], empId);
					syncVersion(apiData as { version?: number | null });
					rememberEvaluationId(
						empId,
						(apiData as { id?: string } | undefined)?.id ?? evalId
					);
				}
			} catch (e) {
				if (e instanceof HttpNotFoundError) {
					// ponytail: stale cache → evict, mutations re-resolve; show empty state
					evaluationIds.delete(empId);
					data = normalizeApiData(null, empId);
				} else {
					throw e;
				}
			}
		}
	} catch (e) {
		data = null;
		error = e instanceof Error ? e.message : 'Error desconocido al cargar evaluaciones';
	} finally {
		loading = false;
	}
}

/** Reload keeping the last viewed employeeId (forces fresh fetch). */
export function reload(employeeId?: string, viewerMode?: ViewerMode): Promise<void> {
	lastLoadTime = 0;
	return load(employeeId ?? lastEmployeeId, viewerMode ?? lastViewerMode);
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
	goalIds: string[],
	_phaseKind?: 'avance' | 'cierre'
): EvaluationStatus {
	const empRatings = (data?.competencyRatings ?? []).filter(
		(cr) => cr.employeeId === employeeId
	);
	// Self scope only: distinct competencyIds with selfRating set (never rhRating).
	const ratedIds = new SvelteSet(
		empRatings.filter((cr) => cr.selfRating != null).map((cr) => cr.competencyId)
	);
	const ratingsDone = ratedIds.size;

	const requiredGoals = goalIds ?? [];
	const empClosures = (data?.goalClosures ?? []).filter(
		(gc) => gc.employeeId === employeeId
	);
	const closureByGoal = new SvelteMap(empClosures.map((gc) => [gc.goalId, gc]));
	let closuresDone = 0;
	for (const gid of requiredGoals) {
		const gc = closureByGoal.get(gid);
		if (!gc) continue;
		// ponytail: self signal = non-empty selfAssessment; progress >0 covers
		// progress-only saves while default 0 without comment stays not-done.
		const hasSelf =
			typeof gc.selfAssessment === 'string'
				? gc.selfAssessment.trim() !== ''
				: gc.selfAssessment != null;
		const hasProgress =
			typeof gc.finalProgress === 'number' &&
			Number.isFinite(gc.finalProgress) &&
			gc.finalProgress > 0;
		if (hasSelf || hasProgress) closuresDone++;
	}

	if (ratingsDone === 0 && closuresDone === 0) return 'pending';

	// Pillars still loading (total 0): never report completed.
	if (!totalCompetencies || totalCompetencies <= 0) return 'in-progress';

	if (ratingsDone < totalCompetencies || closuresDone < requiredGoals.length)
		return 'in-progress';

	return 'completed';
}

// ─── Mutations: Employee Self-Evaluation ───────────────────────────────────────

export async function rateCompetency(
	employeeId: string,
	competencyId: string,
	level: 1 | 2 | 3 | 4 | 5,
	comment?: string
): Promise<void> {
	assertValidLevel(level);
	if (!isEditablePhase()) return;

	if (!getSession().user?.employeeId) return;
	const evaluationId = pathIdFor(employeeId);
	const query = upsertQuery(employeeId);

	const allRatings = (data?.competencyRatings ?? []).filter(
		(cr) => cr.employeeId === employeeId
	);
	const competencies = allRatings
		.map((r) => ({
			competencyId: r.competencyId,
			rating:
				r.competencyId === competencyId
					? level
					: r.selfRating,
			comments:
				r.competencyId === competencyId
					? comment
					: r.selfComment
		}))
		.filter((c) => typeof c.rating === 'number' && c.rating >= 1 && c.rating <= 5)
		.map((c) => ({ ...c, rating: c.rating as 1 | 2 | 3 | 4 | 5 }));
	// If the competency isn't in the list yet, add it
	const existingIds = new SvelteSet(competencies.map((c) => c.competencyId));
	if (!existingIds.has(competencyId)) {
		competencies.push({ competencyId, rating: level, comments: comment });
	}

	let respData: { version?: number | null } | undefined;
	try {
		const { data, error: apiError } = await client.PUT('/evaluations/{id}/self-evaluation', {
			params: { path: { id: evaluationId }, query, header: { 'If-Match': currentVersion() } },
			body: { competencies }
		});
		respData = data as { version?: number | null } | undefined;
		if (apiError) {
			const { code } = errorWithCode(apiError);
			if (isNotFoundCode(code, apiError)) {
				// ponytail: caché rancia → re-resolver vía lista; solo entonces intentar crear
				await refreshEvaluationId(employeeId);
				if (!evaluationIds.get(employeeId)) {
					await autoCreateSelf(employeeId, competencyId, level, comment);
				}
				const { data: retryData, error: retryError } = await client.PUT('/evaluations/{id}/self-evaluation', {
					params: { path: { id: resolveEvaluationId(employeeId) }, header: { 'If-Match': currentVersion() } },
					body: { competencies }
				});
			if (retryError) throw new Error('Error al guardar autoevaluación');
				syncVersion(retryData as { version?: number | null });
				patchCompetencyRating(employeeId, competencyId, { selfRating: level, selfComment: comment });
				return;
			}
			throw new Error('Error al guardar autoevaluación');
		}
	} catch (e) {
		if (e instanceof HttpNotFoundError) {
			await refreshEvaluationId(employeeId);
			if (!evaluationIds.get(employeeId)) {
				await autoCreateSelf(employeeId, competencyId, level, comment);
			}
		const { data: catchRetryData, error: retryError } = await client.PUT('/evaluations/{id}/self-evaluation', {
			params: { path: { id: resolveEvaluationId(employeeId) }, query, header: { 'If-Match': currentVersion() } },
			body: { competencies }
		});
			if (retryError) throw new Error('Error al guardar autoevaluación', { cause: e });
			respData = catchRetryData as { version?: number | null } | undefined;
		} else {
			throw e;
		}
	}
	syncVersion(respData as { version?: number | null });
	patchCompetencyRating(employeeId, competencyId, { selfRating: level, selfComment: comment });
}

export async function closeGoal(
	employeeId: string,
	goalId: string,
	finalProgress: number,
	selfAssessment?: string
): Promise<void> {
	if (!isEditablePhase()) return;

	if (!getSession().user?.employeeId) return;
	const query = upsertQuery(employeeId);
	// F3: dirty-only — omit untouched/empty comment so prior text is preserved.
	const body: { goalId: string; finalProgress: number; selfAssessment?: string } = {
		goalId,
		finalProgress
	};
	if (selfAssessment != null && selfAssessment.trim() !== '') {
		body.selfAssessment = selfAssessment;
	}
	try {
		const { data: respData, error: apiError } = await client.PUT('/evaluations/{id}/goal-state', {
			params: { path: { id: pathIdFor(employeeId) }, query, header: { 'If-Match': currentVersion() } },
			body
		});
		if (apiError && !isNotFoundCode(errorWithCode(apiError).code, apiError)) {
			const { message, code } = errorWithCode(apiError);
			if (isVersionConflict(code, message)) await reload();
			const err = new Error(message || 'Error al cerrar meta');
			(err as Error & { code?: string | null }).code = code;
			throw err;
		}
		let resp: { version?: number | null } | undefined = (respData as { version?: number | null } | undefined);
		if (apiError) {
			await refreshEvaluationId(employeeId);
			const { data: retryData, error: retryError } = await client.PUT('/evaluations/{id}/goal-state', {
				params: { path: { id: pathIdFor(employeeId) }, query, header: { 'If-Match': currentVersion() } },
				body
			});
			if (retryError) {
				const { message, code } = errorWithCode(retryError);
				if (isVersionConflict(code, message)) await reload();
				const err = new Error(message || 'Error al cerrar meta');
				(err as Error & { code?: string | null }).code = code;
				throw err;
			}
			resp = retryData as { version?: number | null } | undefined;
		}
		syncVersion(resp);
		patchGoalClosure(employeeId, goalId, {
			finalProgress,
			...(body.selfAssessment !== undefined ? { selfAssessment: body.selfAssessment } : {})
		});
	} catch (e) {
		if (!isNotFoundCode(errorWithCode(e).code, e)) throw e;
		await refreshEvaluationId(employeeId);
		const { data: retryData, error: retryError } = await client.PUT('/evaluations/{id}/goal-state', {
			params: { path: { id: pathIdFor(employeeId) }, query, header: { 'If-Match': currentVersion() } },
			body
		});
		if (retryError) {
			const { message, code } = errorWithCode(retryError);
			if (isVersionConflict(code, message)) await reload();
			throw new Error(message || 'Error al cerrar meta', { cause: e });
		}
		syncVersion(retryData as { version?: number | null });
		patchGoalClosure(employeeId, goalId, {
			finalProgress,
			...(body.selfAssessment !== undefined ? { selfAssessment: body.selfAssessment } : {})
		});
	}
}

// ─── Mutations: RH Evaluation ─────────────────────────────────────────────────

export async function rhRateCompetency(
	employeeId: string,
	competencyId: string,
	level: 1 | 2 | 3 | 4 | 5,
	comment?: string,
	managerComment?: string
): Promise<void> {
	assertValidLevel(level);
	if (!isEditablePhase()) return;

	if (!getSession().user?.employeeId) throw new Error('No hay sesión activa');
	const evaluationId = pathIdFor(employeeId);
	const query = upsertQuery(employeeId);

	const allRatings = (data?.competencyRatings ?? []).filter(
		(cr) => cr.employeeId === employeeId
	);
	const competencies = allRatings.map((r) => ({
		competencyId: r.competencyId,
		rating:
			r.competencyId === competencyId
				? level
				: (r.rhRating ?? r.selfRating),
		comments:
			r.competencyId === competencyId
				? comment
				: (r.rhComment ?? r.selfComment),
		...(r.competencyId === competencyId && managerComment !== undefined
			? { managerComment }
			: r.competencyId !== competencyId && r.managerComment != null
				? { managerComment: r.managerComment }
				: {})
	})).filter((c) => typeof c.rating === 'number' && c.rating >= 1 && c.rating <= 5)
		.map((c) => ({ ...c, rating: c.rating as 1 | 2 | 3 | 4 | 5 }));
	const existingIds = new SvelteSet(competencies.map((c) => c.competencyId));
	if (!existingIds.has(competencyId)) {
		competencies.push({
			competencyId,
			rating: level,
			comments: comment,
			...(managerComment !== undefined ? { managerComment } : {})
		});
	}

	try {
		const { data: respData, error: apiError } = await client.PUT('/evaluations/{id}/rh-evaluation', {
			params: { path: { id: evaluationId }, query, header: { 'If-Match': currentVersion() } },
			body: { competencies }
		});
		if (apiError) {
			const { message, code } = errorWithCode(apiError);
			if (isNotFoundCode(code, apiError)) {
				// ponytail: caché rancia → re-resolver vía lista; solo entonces intentar crear
				await refreshEvaluationId(employeeId);
				if (!evaluationIds.get(employeeId)) {
					await autoCreateRh(employeeId, competencyId, level, comment, managerComment);
				}
				const { data: retryData, error: retryError } = await client.PUT('/evaluations/{id}/rh-evaluation', {
					params: { path: { id: resolveEvaluationId(employeeId) }, query, header: { 'If-Match': currentVersion() } },
					body: { competencies }
				});
				if (retryError) {
					const r = errorWithCode(retryError);
					if (isVersionConflict(r.code, r.message)) await reload(employeeId);
					const err = new Error(r.message || 'Error al guardar evaluación RH');
					(err as Error & { code?: string | null }).code = r.code;
					throw err;
				}
				syncVersion(retryData as { version?: number | null });
				patchCompetencyRating(employeeId, competencyId, {
					rhRating: level,
					rhComment: comment,
					...(managerComment !== undefined ? { managerComment } : {})
				});
				return;
			}
			if (isVersionConflict(code, message)) await reload(employeeId);
			const err = new Error(message || 'Error al guardar evaluación RH');
			(err as Error & { code?: string | null }).code = code;
			throw err;
		}
		syncVersion(respData as { version?: number | null });
		patchCompetencyRating(employeeId, competencyId, {
			rhRating: level,
			rhComment: comment,
			...(managerComment !== undefined ? { managerComment } : {})
		});
	} catch (e) {
		const { message, code } = errorWithCode(e);
		if (isNotFoundCode(code, e)) {
			// ponytail: caché rancia → re-resolver vía lista; solo entonces intentar crear
			await refreshEvaluationId(employeeId);
			if (!evaluationIds.get(employeeId)) {
				await autoCreateRh(employeeId, competencyId, level, comment, managerComment);
			}
			const { data: retryData, error: retryError } = await client.PUT('/evaluations/{id}/rh-evaluation', {
				params: { path: { id: resolveEvaluationId(employeeId) }, header: { 'If-Match': currentVersion() } },
				body: { competencies }
			});
			if (retryError) {
				const r = errorWithCode(retryError);
				if (isVersionConflict(r.code, r.message)) await reload(employeeId);
				const err = new Error(r.message || 'Error al guardar evaluación RH');
				(err as Error & { code?: string | null }).code = r.code;
				throw err;
			}
			syncVersion(retryData as { version?: number | null });
			patchCompetencyRating(employeeId, competencyId, {
				rhRating: level,
				rhComment: comment,
				...(managerComment !== undefined ? { managerComment } : {})
			});
			return;
		}
		if (e instanceof Error && (e as Error & { code?: unknown }).code !== undefined) throw e;
		const err = new Error(message || 'Error al guardar evaluación RH');
		(err as Error & { code?: string | null }).code = code;
		throw err;
	}
}

export async function rhAssessGoal(
	employeeId: string,
	goalId: string,
	rhAssessment?: string
): Promise<void> {
	if (!isEditablePhase()) return;

	if (!getSession().user?.employeeId) throw new Error('No hay sesión activa');
	const query = upsertQuery(employeeId);
	// F3: dirty-only — omit untouched/empty assessment so prior text is preserved.
	const body: { goalId: string; rhAssessment?: string } = {
		goalId
	};
	if (rhAssessment != null && rhAssessment.trim() !== '') {
		body.rhAssessment = rhAssessment;
	}
	try {
		const { data: respData, error: apiError } = await client.PUT('/evaluations/{id}/goal-state', {
			params: { path: { id: pathIdFor(employeeId) }, query, header: { 'If-Match': currentVersion() } },
			body
		});
		if (apiError && !isNotFoundCode(errorWithCode(apiError).code, apiError)) {
			const { message, code } = errorWithCode(apiError);
			if (isVersionConflict(code, message)) await reload();
			const err = new Error(message || 'Error al guardar evaluación RH de la meta');
			(err as Error & { code?: string | null }).code = code;
			throw err;
		}
		let resp: { version?: number | null } | undefined = (respData as { version?: number | null } | undefined);
		if (apiError) {
			await refreshEvaluationId(employeeId);
			const { data: retryData, error: retryError } = await client.PUT('/evaluations/{id}/goal-state', {
				params: { path: { id: pathIdFor(employeeId) }, query, header: { 'If-Match': currentVersion() } },
				body
			});
			if (retryError) {
				const { message, code } = errorWithCode(retryError);
				if (isVersionConflict(code, message)) await reload();
				const err = new Error(message || 'Error al guardar evaluación RH de la meta');
				(err as Error & { code?: string | null }).code = code;
				throw err;
			}
			resp = retryData as { version?: number | null } | undefined;
		}
		syncVersion(resp);
		if (body.rhAssessment !== undefined) {
			patchGoalClosure(employeeId, goalId, { rhAssessment: body.rhAssessment });
		}
	} catch (e) {
		if (!isNotFoundCode(errorWithCode(e).code, e)) throw e;
		await refreshEvaluationId(employeeId);
		const { data: retryData, error: retryError } = await client.PUT('/evaluations/{id}/goal-state', {
			params: { path: { id: pathIdFor(employeeId) }, query, header: { 'If-Match': currentVersion() } },
			body
		});
		if (retryError) {
			const { message, code } = errorWithCode(retryError);
			if (isVersionConflict(code, message)) await reload();
			throw new Error(message || 'Error al guardar evaluación RH de la meta', { cause: e });
		}
		syncVersion(retryData as { version?: number | null });
		if (body.rhAssessment !== undefined) {
			patchGoalClosure(employeeId, goalId, { rhAssessment: body.rhAssessment });
		}
	}
}

// ─── Mutations: Manager ───────────────────────────────────────────────────────

export async function addManagerComment(
	employeeId: string,
	goalId: string,
	comment: string
): Promise<void> {
	if (!isEditablePhase()) return;

	const empId = getSession().user?.employeeId;
	if (!empId) return;
	// F3: dirty-only — never send an empty manager comment (would blank prior text).
	if (comment.trim() === '') return;
	const query = upsertQuery(employeeId);
	const body = {
		goalId,
		role: 'manager' as const,
		comment
	};
	try {
		const { data: respData, error: apiError } = await client.PUT('/evaluations/{id}/goal-comments', {
			params: { path: { id: pathIdFor(employeeId) }, query, header: { 'If-Match': currentVersion() } },
			body
		});
		if (apiError && !isNotFoundCode(errorWithCode(apiError).code, apiError)) {
			const { message, code } = errorWithCode(apiError);
			if (isVersionConflict(code, message)) await reload();
			const err = new Error(message || 'Error al guardar comentario del manager');
			(err as Error & { code?: string | null }).code = code;
			throw err;
		}
		let resp: { version?: number | null } | undefined = (respData as { version?: number | null } | undefined);
		if (apiError) {
			await refreshEvaluationId(employeeId);
			const { data: retryData, error: retryError } = await client.PUT('/evaluations/{id}/goal-comments', {
				params: { path: { id: pathIdFor(employeeId) }, query, header: { 'If-Match': currentVersion() } },
				body
			});
			if (retryError) {
				const { message, code } = errorWithCode(retryError);
				if (isVersionConflict(code, message)) await reload();
				const err = new Error(message || 'Error al guardar comentario del manager');
				(err as Error & { code?: string | null }).code = code;
				throw err;
			}
			resp = retryData as { version?: number | null } | undefined;
		}
		syncVersion(resp);
		patchGoalClosure(employeeId, goalId, { managerComment: comment });
	} catch (e) {
		if (!isNotFoundCode(errorWithCode(e).code, e)) throw e;
		await refreshEvaluationId(employeeId);
		const { data: retryData, error: retryError } = await client.PUT('/evaluations/{id}/goal-comments', {
			params: { path: { id: pathIdFor(employeeId) }, query, header: { 'If-Match': currentVersion() } },
			body
		});
		if (retryError) {
			const { message, code } = errorWithCode(retryError);
			if (isVersionConflict(code, message)) await reload();
			throw new Error(message || 'Error al guardar comentario del manager', { cause: e });
		}
		syncVersion(retryData as { version?: number | null });
		patchGoalClosure(employeeId, goalId, { managerComment: comment });
	}
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

	const evaluationId = await ensureEvaluationId(empId);
	if (!evaluationId) throw new Error('Sin evaluationId: llama load() primero');
	for (const r of empRatings) assertValidLevel(r.selfRating);
	const { error: apiError } = await client.POST('/evaluations/{id}/self-evaluation', {
		params: { path: { id: evaluationId }, header: { 'Idempotency-Key': crypto.randomUUID() } },
		body: {
			competencies: empRatings.map((r) => ({
				competencyId: r.competencyId,
				rating: r.selfRating as 1 | 2 | 3 | 4 | 5,
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
	if (!getSession().user?.employeeId) throw new Error('No hay sesión activa');

	const empRatings = (data?.competencyRatings ?? []).filter(
		(cr) => cr.employeeId === employeeId
	);

	const evaluationId = await ensureEvaluationId(employeeId);
	if (!evaluationId) throw new Error('Sin evaluationId: llama load(employeeId) primero');
	for (const r of empRatings) assertValidLevel(r.rhRating);
	const { error: apiError } = await client.POST('/evaluations/{id}/rh-evaluation', {
		params: { path: { id: evaluationId }, header: { 'Idempotency-Key': crypto.randomUUID() } },
		body: {
			competencies: empRatings.map((r) => ({
				competencyId: r.competencyId,
				rating: r.rhRating as 1 | 2 | 3 | 4 | 5,
				comments: r.rhComment ?? r.selfComment,
				...(r.managerComment != null ? { managerComment: r.managerComment } : {})
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

	const evaluationId = await ensureEvaluationId(empId);
	if (!evaluationId) throw new Error('Sin evaluationId: llama load() primero');
	const { error: apiError } = await client.POST('/evaluations/{id}/finalize', {
		params: { path: { id: evaluationId } },
		body: reason ? { reason } : undefined
	});
	if (apiError) throw new Error('Error al finalizar evaluación');
	await reload();
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

function _applyRating(
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
			id: `${mode === 'self' ? 'sr' : 'rh'}-${employeeId}-${competencyId}`,
			employeeId,
			competencyId,
			...(mode === 'self'
				? { selfRating: level, selfComment: comment }
				: { rhRating: level, rhComment: comment })
		}
	];
}
