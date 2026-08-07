import type {
	Goal,
	GoalCategory,
	GoalKpiLink,
	KPI,
	EmployeeAssignment,
	ChangeRequest,
	GoalComment,
	GoalProposal,
	GoalUnit,
	KpiUnit,
	CyclePhase
} from '$lib/types/goal';
import type { EvaluationProfile } from '$lib/types/evaluation';

import { getActivePhase } from '$lib/api/cycle.svelte';
import { getSession } from '$lib/api/session.svelte';
import { client } from '$lib/api/client';
import { getActiveCycle } from '$lib/stores/cycleStore.svelte';
import { progressPercent } from '$lib/utils/scoring';
import { SvelteDate, SvelteMap } from 'svelte/reactivity';

// ─── Internal data shape ──────────────────────────────────────────────────────

interface StoreData {
	categories: GoalCategory[];
	goals: Goal[];
	kpis: KPI[];
	goalKpiLinks: GoalKpiLink[];
	assignments: EmployeeAssignment[];
	changeRequests: ChangeRequest[];
	assignmentComments: SvelteMap<string, GoalComment[]>;
}

// ─── Tolerance ────────────────────────────────────────────────────────────────

const EPSILON = 0.01;

// ─── Triplete state ───────────────────────────────────────────────────────────
// Svelte 5 idiomatic pattern: use a class with $state fields so that
// reassignable reactive state can be safely exported from a .svelte.ts module.

class StoreState {
	data = $state<StoreData | null>(null);
	loading = $state(true);
	error = $state<string | null>(null);
}

export const storeState = new StoreState();

let loadPromise: Promise<void> | null = null;
let lastLoadTime = 0;
const FRESHNESS_MS = 500;

/** @returns true while load() is in progress. */
export function isLoading(): boolean {
	return storeState.loading;
}

/** @returns the current error message, or null if no error. */
export function getError(): string | null {
	return storeState.error;
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

function getEmployeeId(): string {
	return getSession().user?.employeeId ?? '';
}

function mapToGoalProposal(pp: Record<string, unknown>): GoalProposal {
    return {
        id: (pp.id as string) ?? '',
        goalId: (pp.goal_id as string) ?? '',
        requestedBy: (pp.requested_by as string) ?? '',
        name: (pp.name as string) ?? '',
        description: (pp.description as string) ?? '',
        unit: (pp.unit as GoalUnit) ?? 'numero',
        direction: (pp.direction as 'ascendente' | 'descendente') ?? 'ascendente',
        weight: (pp.weight as number) ?? 0,
        targetValue: (pp.target_value as number) ?? 0,
        baselineValue: pp.baseline_value as number | undefined,
        kpiIds: (pp.kpi_ids as string[]) ?? [],
        status: (pp.status as 'pending' | 'accepted' | 'rejected') ?? 'pending',
        reviewedBy: pp.reviewed_by as string | undefined,
        reviewedAt: pp.reviewed_at as string | undefined,
        createdAt: (pp.created_at as string) ?? '',
        updatedAt: (pp.updated_at as string) ?? '',
    };
}

function mapToGoalComment(raw: unknown): GoalComment {
    const c = (raw ?? {}) as Record<string, unknown>;
    return {
        id: (c.id as string) ?? '',
        authorId: (c.author_id as string) ?? '',
        authorName: (c.author_name as string) ?? '',
        content: (c.content as string) ?? '',
        createdAt: (c.created_at as string) ?? '',
        goalId: c.goal_id as string | undefined,
        categoryId: c.category_id as string | undefined,
        assignmentId: c.assignment_id as string | undefined,
    };
}

/**
 * Normalize API responses into the flat StoreData format that getters consume.
 */
function normalizeApiData(
	apiCategories: Array<{
		id?: string;
		employee_id?: string;
		name?: string;
		description?: string;
		weight?: number;
		goals?: Array<{
			id?: string;
			category_id?: string;
			name?: string;
			description?: string;
			unit?: string;
			weight?: number;
			target_value?: number;
			current_value?: number;
			state?: string;
			version?: number;
			direction?: string;
			baseline_value?: number;
			pending_proposal?: Record<string, unknown>;
			kpis?: Array<{
				id?: string;
				name?: string;
				unit?: string;
				description?: string;
				direction?: string;
				current_value?: number;
				target_value?: number;
			}>;
			created_at?: string;
			updated_at?: string;
		}>;
	}>,
	apiKpis: Array<{
		id?: string;
		name?: string;
		unit?: string;
		description?: string;
		direction?: string;
		current_value?: number;
		target_value?: number;
	}>,
	apiAssignment: {
		id?: string;
		employee_id?: string;
		cycle_id?: string;
		categories?: Array<unknown>;
		created_at?: string;
	} | null,
	profileId?: string
): StoreData {
	const cats: GoalCategory[] = [];
	const goals: Goal[] = [];
	const goalKpiLinks: GoalKpiLink[] = [];
	const assignedGoalIds: string[] = [];

	// 1. Build full KPI catalog from the /kpis endpoint
	const kpisMap = new SvelteMap<string, KPI>();
	for (const ak of apiKpis) {
		const kpi: KPI = {
			id: ak.id ?? crypto.randomUUID(),
			name: ak.name ?? '',
			description: ak.description ?? '',
			unit: (ak.unit as KpiUnit) ?? 'numero',
			direction: (ak.direction as 'ascendente' | 'descendente') ?? 'ascendente',
			currentValue: ak.current_value,
			targetValue: ak.target_value,
			minValue: undefined,
			maxValue: undefined
		};
		kpisMap.set(kpi.id, kpi);
	}

	// 2. Flatten categories → categories + goals + KPI links
	for (const ac of apiCategories) {
		cats.push({
			id: ac.id ?? crypto.randomUUID(),
			name: ac.name ?? '',
			description: ac.description ?? '',
			weight: ac.weight ?? 0,
			pillarId: (ac as Record<string, unknown>)?.pillar_id as string | undefined
		});

		const catId = ac.id ?? '';
		for (const ag of ac.goals ?? []) {
			const goalId = ag.id ?? crypto.randomUUID();
			goals.push({
				id: goalId,
				name: ag.name ?? '',
				description: ag.description ?? '',
				categoryId: ag.category_id ?? catId,
				weight: ag.weight ?? 0,
				unit: (ag.unit as GoalUnit) ?? 'numero',
				direction: (ag.direction as 'ascendente' | 'descendente') ?? 'ascendente',
				targetValue: ag.target_value ?? 0,
				baselineValue: ag.baseline_value,
				progress: ag.current_value,
				progressUpdatedAt: ag.updated_at,
				comments: [],
			version: ag.version ?? 1,
			pendingProposal: ag.pending_proposal ? mapToGoalProposal(ag.pending_proposal) : undefined
		});
			assignedGoalIds.push(goalId);

			// Each nested KPI becomes a link + potentially a new KPI entry
			for (const kpiRef of ag.kpis ?? []) {
				if (kpiRef.id) {
					goalKpiLinks.push({ goalId, kpiId: kpiRef.id });
					if (!kpisMap.has(kpiRef.id)) {
						kpisMap.set(kpiRef.id, {
							id: kpiRef.id,
							name: kpiRef.name ?? '',
							description: kpiRef.description ?? '',
							unit: (kpiRef.unit as KpiUnit) ?? 'numero',
							direction: (kpiRef.direction as 'ascendente' | 'descendente') ?? 'ascendente',
							currentValue: kpiRef.current_value,
							targetValue: kpiRef.target_value,
							minValue: undefined,
							maxValue: undefined
						});
					}
				}
			}
		}
	}

	// 3. Build assignments
	const assignments: EmployeeAssignment[] = [];
	if (apiAssignment?.id) {
		assignments.push({
			id: apiAssignment.id,
			employeeId: apiAssignment.employee_id ?? '',
			employeeName: '',
			profileId: profileId as EvaluationProfile ?? 'colaborador',
			managerId: null,
			goalIds: assignedGoalIds,
			createdAt: apiAssignment.created_at ?? new Date().toISOString(),
			updatedAt: apiAssignment.created_at ?? new Date().toISOString()
		});
	}

	return {
		categories: cats,
		goals,
		kpis: [...kpisMap.values()],
		goalKpiLinks,
		assignments,
		changeRequests: [],
		assignmentComments: new SvelteMap<string, GoalComment[]>(),
	};
}

/**
 * Load goals data.
 *
 * In DEV without VITE_USE_API: loads from fixture files (structured clone).
 * In production / VITE_USE_API=true: fetches from the real API endpoints.
 * @param forceRefresh - if true, bypasses the freshness guard and forces a reload
 */
export async function load(forceRefresh = false): Promise<void> {
	if (loadPromise) return loadPromise;
	if (!forceRefresh && storeState.data && Date.now() - lastLoadTime < FRESHNESS_MS) return;
	loadPromise = _doLoad();
	try {
		await loadPromise;
	} finally {
		loadPromise = null;
		lastLoadTime = Date.now();
	}
}

/**
 * Load goals data for a specific employee (used by boss viewing subordinates).
 * Bypasses the freshness guard to force a reload with different employee data.
 */
export async function loadForEmployee(empId: string): Promise<void> {
	if (loadPromise) return loadPromise;
	loadPromise = _doLoad(empId);
	try {
		await loadPromise;
	} finally {
		loadPromise = null;
		lastLoadTime = Date.now();
	}
}

async function _doLoad(empIdOverride?: string): Promise<void> {
	storeState.loading = true;
	storeState.error = null;

	try {
		const empId = empIdOverride ?? getEmployeeId();

		const [catsRes, kpisRes, assignmentRes] = await Promise.all([
			client.GET('/employees/{empId}/categories', {
				params: { path: { empId } }
			}),
			client.GET('/kpis', {}),
			client.GET('/employees/{empId}/assignments', {
				params: { path: { empId } }
			})
		]);

		if (catsRes.error) {
			throw new Error(
				(catsRes.error as { error?: { message?: string } })?.error?.message ??
					'Error al cargar categorías'
			);
		}

		let apiAssignment = (assignmentRes.data as
			| {
					id?: string;
					employee_id?: string;
					cycle_id?: string;
					categories?: Array<unknown>;
					created_at?: string;
			  }
			| null
			| undefined) ?? null;

		if (empIdOverride && !apiAssignment?.id) {
			const activeCycle = getActiveCycle();
			if (activeCycle?.id) {
				const { error: createErr } = await client.POST('/employees/{empId}/assignments', {
					params: { path: { empId } },
					body: { cycle_id: activeCycle.id }
				});
				if (!createErr) {
					const { data: fresh } = await client.GET('/employees/{empId}/assignments', {
						params: { path: { empId } }
					});
					apiAssignment = (fresh as typeof apiAssignment) ?? null;
				}
			}
		}

		const apiCategories = (catsRes.data as { items?: Array<unknown> })?.items ?? [];
		const apiKpis = (kpisRes.data as { items?: Array<unknown> })?.items ?? [];

		storeState.data = normalizeApiData(
			apiCategories as Parameters<typeof normalizeApiData>[0],
			apiKpis as Parameters<typeof normalizeApiData>[1],
			apiAssignment ?? null,
			getSession().user?.profileId
		);

		// ponytail: load comments for all goals so badges show correct counts
		await loadAllGoalComments(empId);
	} catch (e) {
		storeState.error = e instanceof Error ? e.message : 'Error desconocido al cargar datos de objetivos';
	} finally {
		storeState.loading = false;
	}
}

/**
 * Load comments for all goals and categories in the store so badges show correct counts.
 * Silently ignores errors — comments are non-critical UI enhancements.
 */
export async function loadAllGoalComments(_empId?: string): Promise<void> {
	if (!storeState.data) return;
	const goals = storeState.data.goals;
	const categories = storeState.data.categories;

	// Load goal comments
	if (goals.length > 0) {
		const results = await Promise.allSettled(
			goals.map((g) =>
				client.GET('/goals/{goalId}/comments', {
					params: { path: { goalId: g.id } }
				}).then(({ data, error }) => {
					if (data && !error) {
						return { id: g.id, comments: (data as unknown as GoalComment[]).map(mapToGoalComment) };
					}
					return { id: g.id, comments: [] as GoalComment[] };
				})
			)
		);

		const commentMap = new SvelteMap<string, GoalComment[]>();
		for (const r of results) {
			if (r.status === 'fulfilled') {
				commentMap.set(r.value.id, r.value.comments);
			}
		}

		storeState.data = {
			...storeState.data,
			goals: storeState.data.goals.map((g) => ({
				...g,
				comments: commentMap.get(g.id) ?? g.comments ?? []
			}))
		};
	}

	// Load category comments
	if (categories.length > 0) {
		const catResults = await Promise.allSettled(
			categories.map((c) =>
				client.GET('/categories/{catId}/comments', {
					params: { path: { catId: c.id } }
				}).then(({ data, error }) => {
					if (data && !error) {
						return { id: c.id, comments: (data as unknown as GoalComment[]).map(mapToGoalComment) };
					}
					return { id: c.id, comments: [] as GoalComment[] };
				})
			)
		);

		const catCommentMap = new SvelteMap<string, GoalComment[]>();
		for (const r of catResults) {
			if (r.status === 'fulfilled') {
				catCommentMap.set(r.value.id, r.value.comments);
			}
		}

		storeState.data = {
			...storeState.data,
			categories: storeState.data.categories.map((c) => ({
				...c,
				comments: catCommentMap.get(c.id) ?? c.comments ?? []
			}))
		};
	}
}


/** Alias for load(). */
export function reload(forceRefresh = false): Promise<void> {
	return load(forceRefresh);
}

// ─── Getters: General ─────────────────────────────────────────────────────────

export function getCategories(): GoalCategory[] {
	return storeState.data?.categories ?? [];
}

export function getGoals(): Goal[] {
	return storeState.data?.goals ?? [];
}

export function getKpis(): KPI[] {
	return storeState.data?.kpis ?? [];
}

export function getGoalKpiLinks(): GoalKpiLink[] {
	return storeState.data?.goalKpiLinks ?? [];
}

export function getAssignments(): EmployeeAssignment[] {
	return storeState.data?.assignments ?? [];
}

export function getChangeRequests(): ChangeRequest[] {
	return storeState.data?.changeRequests ?? [];
}

export function getCyclePhase(): CyclePhase {
	return getActivePhase() ?? 'inicio-anio';
}

// ─── Getters: Progress & Comments ─────────────────────────────────────────────

export function getGoalProgress(goalId: string): number | undefined {
	return storeState.data?.goals.find((g) => g.id === goalId)?.progress;
}

export function getGoalComments(goalId: string): GoalComment[] {
	return storeState.data?.goals.find((g) => g.id === goalId)?.comments ?? [];
}

export function getCategoryProgressAverage(categoryId: string): number {
	const catGoals = (storeState.data?.goals ?? []).filter((g) => g.categoryId === categoryId);
	if (catGoals.length === 0) return 0;
	const withProgress = catGoals.filter((g) => g.progress !== undefined);
	if (withProgress.length === 0) return 0;
	const total = withProgress.reduce((acc, g) => {
		const pct = progressPercent(g.progress ?? 0, g.targetValue, g.baselineValue, g.direction);
		return acc + pct;
	}, 0);
	return total / withProgress.length;
}

/**
 * Calculate the weighted score for the current employee across all categories.
 *
 * Formula: Σ(cat.weight/100 × Σ(goal.weight/100 × progressPercent(goal)))
 *
 * Only goals with progress data are included in the calculation.
 */
export function getWeightedScore(): number {
	const cats = storeState.data?.categories ?? [];
	const allGoals = storeState.data?.goals ?? [];
	let total = 0;

	for (const cat of cats) {
		const catGoals = allGoals.filter((g) => g.categoryId === cat.id);
		const withProgress = catGoals.filter((g) => g.progress !== undefined);
		if (withProgress.length === 0) continue;

		const catGoalSum = withProgress.reduce((acc, g) => {
			const pct = progressPercent(g.progress ?? 0, g.targetValue, g.baselineValue, g.direction);
			return acc + (g.weight / 100) * pct;
		}, 0);

		total += (cat.weight / 100) * catGoalSum;
	}

	return total;
}

export function getGoalPermissions(
	role: EvaluationProfile,
	isOwner: boolean
): {
	canEditProgress: boolean;
	canComment: boolean;
	canEditWeight: boolean;
	canEditDirection: boolean;
	canEditBaseline: boolean;
	canDelete: boolean;
	canClose: boolean;
} {
	const phase = getActivePhase() ?? 'inicio-anio';
	if (phase === 'inicio-anio') {
		return {
			canEditProgress: false,
			canComment: false,
			canEditWeight: isOwner,
			canEditDirection: isOwner,
			canEditBaseline: isOwner,
			canDelete: isOwner,
			canClose: false
		};
	}
	if (phase === 'fin-anio') {
		return {
			canEditProgress: false,
			canComment: false,
			canEditWeight: false,
			canEditDirection: false,
			canEditBaseline: false,
			canDelete: false,
			canClose: isOwner
		};
	}
	// phase === 'medio-anio' (avance)
	return {
		canEditProgress: true,
		canComment: true,
		canEditWeight: false,
		canEditDirection: false,
		canEditBaseline: false,
		canDelete: false,
		canClose: false
	};
}

// ─── Getters: Filtered ────────────────────────────────────────────────────────

export function getGoalsByCategory(categoryId: string): Goal[] {
	return (storeState.data?.goals ?? []).filter((g) => g.categoryId === categoryId);
}

export function getKpisForGoal(goalId: string): KPI[] {
	const linkKpiIds = (storeState.data?.goalKpiLinks ?? [])
		.filter((link) => link.goalId === goalId)
		.map((link) => link.kpiId);
	return (storeState.data?.kpis ?? []).filter((kpi) => linkKpiIds.includes(kpi.id));
}

export function getLinksForGoal(goalId: string): GoalKpiLink[] {
	return (storeState.data?.goalKpiLinks ?? []).filter((link) => link.goalId === goalId);
}

export function getLinksForKpi(kpiId: string): GoalKpiLink[] {
	return (storeState.data?.goalKpiLinks ?? []).filter((link) => link.kpiId === kpiId);
}

export function getAssignmentsByProfile(profileId: EvaluationProfile): EmployeeAssignment[] {
	return (storeState.data?.assignments ?? []).filter((a) => a.profileId === profileId);
}

export function getAssignmentByEmployee(employeeId: string): EmployeeAssignment | undefined {
	return (storeState.data?.assignments ?? []).find((a) => a.employeeId === employeeId);
}

// ─── Getters: Validation ──────────────────────────────────────────────────────

/**
 * Sum of all category weights equals 100 ± ε.
 */
function doCategoryWeightsSumTo100(): boolean {
	const sum = (storeState.data?.categories ?? []).reduce((acc, c) => acc + c.weight, 0);
	return Math.abs(sum - 100) <= EPSILON;
}

/**
 * For each category, the goals within it sum to 100 ± ε.
 */
function areAllCategoryGoalWeightsValid(): boolean {
	for (const cat of storeState.data?.categories ?? []) {
		const catGoals = (storeState.data?.goals ?? []).filter((g) => g.categoryId === cat.id);
		if (catGoals.length === 0) continue;
		const sum = catGoals.reduce((acc, g) => acc + g.weight, 0);
		if (Math.abs(sum - 100) > EPSILON) return false;
	}
	return true;
}

/**
 * Returns true when categories sum to 100% AND
 * goals within each category also sum to 100%.
 */
export function isAssignmentValid(): boolean {
	return doCategoryWeightsSumTo100() && areAllCategoryGoalWeightsValid();
}

/**
 * Returns true when the goals in the given category sum to 100 ± ε.
 */
export function isCategoryGoalsWeightValid(categoryId: string): boolean {
	const catGoals = (storeState.data?.goals ?? []).filter((g) => g.categoryId === categoryId);
	if (catGoals.length === 0) return true;
	const sum = catGoals.reduce((acc, g) => acc + g.weight, 0);
	return Math.abs(sum - 100) <= EPSILON;
}

// ─── Mutations: Categories ────────────────────────────────────────────────────

export async function addCategory(category: GoalCategory): Promise<void> {
	const empId = getEmployeeId();
	const { error: apiError } = await client.POST('/employees/{empId}/categories', {
		params: { path: { empId } },
		body: { name: category.name, description: category.description, weight: category.weight, ...(category.pillarId ? { pillar_id: category.pillarId } : {}) }
	});
	if (apiError) throw new Error((apiError as { error?: { message?: string } })?.error?.message ?? 'Error al crear categoría');
	await reload(true);
}

export async function updateCategory(id: string, updates: Partial<Omit<GoalCategory, 'id'>>): Promise<void> {
	const empId = getEmployeeId();
	const { error: apiError } = await client.PUT('/employees/{empId}/categories/{catId}', {
		params: { path: { empId, catId: id } },
		body: { name: updates.name ?? '', description: updates.description ?? '', weight: updates.weight ?? 0, ...(updates.pillarId ? { pillar_id: updates.pillarId } : {}) }
	});
	if (apiError) throw new Error((apiError as { error?: { message?: string } })?.error?.message ?? 'Error al actualizar categoría');
	await reload(true);
}

export async function deleteCategory(id: string, opts?: { skipReload?: boolean }): Promise<void> {
	// Block deletion outside 'inicio-anio' phase
	const phase = getActivePhase() ?? 'inicio-anio';
	if (phase === 'medio-anio' || phase === 'fin-anio') return;

	const empId = getEmployeeId();
	const { error: apiError } = await client.DELETE('/employees/{empId}/categories/{catId}', {
		params: { path: { empId, catId: id } }
	});
	if (apiError) throw new Error((apiError as { error?: { message?: string } })?.error?.message ?? 'Error al eliminar categoría');
	if (!opts?.skipReload) await reload(true);
}

// ─── Mutations: Goals ─────────────────────────────────────────────────────────

export async function addGoal(goal: Goal): Promise<string> {
	const empId = getEmployeeId();
	const { data, error: apiError } = await client.POST('/employees/{empId}/categories/{catId}/goals', {
		params: { path: { empId, catId: goal.categoryId } },
		body: {
			name: goal.name,
			description: goal.description,
			unit: goal.unit as 'porcentaje' | 'moneda' | 'numero',
			weight: goal.weight,
			target_value: goal.targetValue,
			direction: goal.direction as 'ascendente' | 'descendente',
			baseline_value: goal.baselineValue
		}
	});
	if (apiError) throw new Error((apiError as { error?: { message?: string } })?.error?.message ?? 'Error al crear meta');
	await reload(true);
	return data?.id ?? goal.id;
}

export async function updateGoal(id: string, updates: Partial<Omit<Goal, 'id'>>): Promise<void> {
	const { error: apiError } = await client.PUT('/goals/{goalId}', {
		params: { path: { goalId: id } },
		body: {
			name: updates.name ?? '',
			description: updates.description ?? '',
			unit: (updates.unit as 'porcentaje' | 'moneda' | 'numero') ?? 'numero',
			weight: updates.weight ?? 0,
			target_value: updates.targetValue ?? 0,
			direction: (updates.direction as 'ascendente' | 'descendente') ?? 'ascendente',
			baseline_value: updates.baselineValue,
			version: updates.version ?? 1
		}
	});
	if (apiError) throw new Error((apiError as { error?: { message?: string } })?.error?.message ?? 'Error al actualizar meta');
	await reload(true);
}

export async function deleteGoal(id: string, opts?: { skipReload?: boolean }): Promise<void> {
	// Block deletion outside 'inicio-anio' phase
	const phase = getActivePhase() ?? 'inicio-anio';
	if (phase === 'medio-anio' || phase === 'fin-anio') return;

	const { error: apiError } = await client.DELETE('/goals/{goalId}', {
		params: { path: { goalId: id } }
	});
	if (apiError) throw new Error((apiError as { error?: { message?: string } })?.error?.message ?? 'Error al eliminar meta');
	if (!opts?.skipReload) await reload(true);
}

// ─── Mutations: Goal Proposals ─────────────────────────────────────────────────

export async function createGoalProposal(goalId: string, data: {
	name: string;
	description: string;
	unit: GoalUnit;
	weight: number;
	targetValue: number;
	direction: 'ascendente' | 'descendente';
	baselineValue?: number;
	kpiIds: string[];
}): Promise<void> {
	const { error } = await client.POST('/goals/{goalId}/proposals', {
		params: { path: { goalId } },
		body: {
			name: data.name,
			description: data.description,
			unit: data.unit,
			weight: data.weight,
			target_value: data.targetValue,
			direction: data.direction,
			baseline_value: data.baselineValue,
			kpi_ids: data.kpiIds,
		}
	});
	if (error) throw new Error(
		(error as { error?: { message?: string } })?.error?.message ?? 'Error al crear propuesta'
	);
	await reload(true);
}

export async function acceptGoalProposal(goalId: string, proposalId: string, reviewedBy: string): Promise<void> {
	const { error } = await client.PATCH('/goals/{goalId}/proposals/{propId}', {
		params: { path: { goalId, propId: proposalId } },
		body: { status: 'accepted', reviewed_by: reviewedBy }
	});
	if (error) throw new Error(
		(error as { error?: { message?: string } })?.error?.message ?? 'Error al aceptar propuesta'
	);
	await reload(true);
}

export async function rejectGoalProposal(goalId: string, proposalId: string): Promise<void> {
	const { error } = await client.PATCH('/goals/{goalId}/proposals/{propId}', {
		params: { path: { goalId, propId: proposalId } },
		body: { status: 'rejected', reviewed_by: '' }
	});
	if (error) throw new Error(
		(error as { error?: { message?: string } })?.error?.message ?? 'Error al rechazar propuesta'
	);
	await reload(true);
}

// ─── Getters: Proposals ───────────────────────────────────────────────────────

export function getPendingProposal(goalId: string): GoalProposal | undefined {
	return storeState.data?.goals.find(g => g.id === goalId)?.pendingProposal;
}

// ─── Mutations: KPIs ──────────────────────────────────────────────────────────

export async function addKpi(kpi: KPI): Promise<void> {
	const { error: apiError } = await client.POST('/kpis', {
		body: {
			name: kpi.name,
			description: kpi.description,
			unit: kpi.unit as 'porcentaje' | 'moneda' | 'numero' | 'binario',
			direction: kpi.direction as 'ascendente' | 'descendente',
			target_value: kpi.targetValue
		}
	});
	if (apiError) throw new Error((apiError as { error?: { message?: string } })?.error?.message ?? 'Error al crear KPI');
	await reload(true);
}

export async function updateKpi(id: string, updates: Partial<Omit<KPI, 'id'>>): Promise<void> {
	const { error: apiError } = await client.PUT('/kpis/{kpiId}', {
		params: { path: { kpiId: id } },
		body: {
			name: updates.name ?? '',
			description: updates.description ?? '',
			unit: (updates.unit as 'porcentaje' | 'moneda' | 'numero' | 'binario') ?? 'numero',
			direction: (updates.direction as 'ascendente' | 'descendente') ?? 'ascendente',
			target_value: updates.targetValue
		}
	});
	if (apiError) throw new Error((apiError as { error?: { message?: string } })?.error?.message ?? 'Error al actualizar KPI');
	await reload(true);
}

export async function deleteKpi(id: string): Promise<void> {
	const { error: apiError } = await client.DELETE('/kpis/{kpiId}', {
		params: { path: { kpiId: id } }
	});
	if (apiError) throw new Error((apiError as { error?: { message?: string } })?.error?.message ?? 'Error al eliminar KPI');
	await reload(true);
}

// ─── Mutations: GoalKpiLink (N:M) ─────────────────────────────────────────────

export async function linkKpiToGoal(goalId: string, kpiId: string): Promise<void> {
	// Idempotent: skip if link already exists
	const exists = (storeState.data?.goalKpiLinks ?? []).some((link) => link.goalId === goalId && link.kpiId === kpiId);
	if (exists) return;

	const { error: apiError } = await client.POST('/goals/{goalId}/kpis', {
		params: { path: { goalId } },
		body: { kpi_id: kpiId }
	});
	if (apiError) throw new Error((apiError as { error?: { message?: string } })?.error?.message ?? 'Error al vincular KPI');
	await reload(true);
}

export async function unlinkKpiFromGoal(goalId: string, kpiId: string): Promise<void> {
	const { error: apiError } = await client.DELETE('/goals/{goalId}/kpis/{kpiId}', {
		params: { path: { goalId, kpiId } }
	});
	if (apiError) throw new Error((apiError as { error?: { message?: string } })?.error?.message ?? 'Error al desvincular KPI');
	await reload(true);
}

export async function updateLinkWeight(goalId: string, kpiId: string, weight: number | undefined): Promise<void> {
	// No dedicated API endpoint for link weight. Local-only for now.
	storeState.data = {
		...storeState.data!,
		goalKpiLinks: (storeState.data?.goalKpiLinks ?? []).map((link) =>
			link.goalId === goalId && link.kpiId === kpiId ? { ...link, weight } : link
		)
	};
}

// ─── Mutations: Assignments ───────────────────────────────────────────────────

export async function addAssignment(assignment: EmployeeAssignment): Promise<void> {
	const empId = getEmployeeId();
	const activeCycle = getActiveCycle();
	if (!activeCycle?.id) throw new Error('No hay un ciclo activo para asignar');
	const { error: apiError } = await client.POST('/employees/{empId}/assignments', {
		params: { path: { empId } },
		body: { cycle_id: activeCycle.id }
	});
	if (apiError) throw new Error((apiError as { error?: { message?: string } })?.error?.message ?? 'Error al crear asignación');
	await reload(true);
}

export async function updateAssignment(
	id: string,
	updates: Partial<Omit<EmployeeAssignment, 'id'>>
): Promise<void> {
	// No dedicated API endpoint for partial assignment update. Local-only for now.
	storeState.data = {
		...storeState.data!,
		assignments: (storeState.data?.assignments ?? []).map((a) =>
			a.id === id
				? { ...a, ...updates, updatedAt: new SvelteDate().toISOString() }
				: a
		)
	};
}

export async function deleteAssignment(id: string): Promise<void> {
	// No dedicated API endpoint for assignment deletion. Local-only for now.
	storeState.data = {
		...storeState.data!,
		assignments: (storeState.data?.assignments ?? []).filter((a) => a.id !== id)
	};
}

export async function assignGoalToEmployee(employeeId: string, goalId: string): Promise<void> {
	// No dedicated API endpoint. Local-only for now.
	storeState.data = {
		...storeState.data!,
		assignments: (storeState.data?.assignments ?? []).map((a) =>
			a.employeeId === employeeId && !a.goalIds.includes(goalId)
				? { ...a, goalIds: [...a.goalIds, goalId], updatedAt: new SvelteDate().toISOString() }
				: a
		)
	};
}

export async function unassignGoalFromEmployee(employeeId: string, goalId: string): Promise<void> {
	// No dedicated API endpoint. Local-only for now.
	storeState.data = {
		...storeState.data!,
		assignments: (storeState.data?.assignments ?? []).map((a) =>
			a.employeeId === employeeId
				? {
						...a,
						goalIds: a.goalIds.filter((gid) => gid !== goalId),
						updatedAt: new SvelteDate().toISOString()
					}
				: a
		)
	};
}

// ─── Mutations: ChangeRequests ────────────────────────────────────────────────

export async function recordChangeRequest(request: ChangeRequest): Promise<void> {
	const { data, error: apiError } = await client.POST('/change-requests', {
		body: {
			entity_type: request.entityType,
			entity_id: request.entityId,
			requested_by: request.requestedBy,
		}
	});
	if (apiError) throw new Error('Error al crear solicitud de cambio');
	if (data) {
		storeState.data = { ...storeState.data!, changeRequests: [...(storeState.data?.changeRequests ?? []), data as unknown as ChangeRequest] };
	}
}

export async function approveChangeRequest(id: string, approvedBy: string): Promise<void> {
	const { error: apiError } = await client.PATCH('/change-requests/{crId}', {
		params: { path: { crId: id } },
		body: { status: 'approved', approved_by: approvedBy }
	});
	if (apiError) throw new Error('Error al aprobar solicitud');
	await reload(true);
}

export async function rejectChangeRequest(id: string): Promise<void> {
	const { error: apiError } = await client.PATCH('/change-requests/{crId}', {
		params: { path: { crId: id } },
		body: { status: 'rejected', approved_by: '' }
	});
	if (apiError) throw new Error('Error al rechazar solicitud');
	await reload(true);
}

// ─── Mutations: Progress & Comments ───────────────────────────────────────────

const debounceTimers = new SvelteMap<string, ReturnType<typeof setTimeout>>();

export async function updateGoalProgress(goalId: string, progress: number): Promise<void> {
	// Clear pending debounce for this goal
	const existing = debounceTimers.get(goalId);
	if (existing) clearTimeout(existing);

	return new Promise((resolve, reject) => {
		const timer = setTimeout(async () => {
			debounceTimers.delete(goalId);

			const goal = storeState.data?.goals.find(g => g.id === goalId);
			if (!goal) return resolve();

			// Guard: no change
			if (goal.progress === progress) return resolve();

			// Snapshot for rollback
			const previousProgress = goal.progress;

			// Optimistic update (same pattern as addGoalComment)
			storeState.data = {
				...storeState.data!,
				goals: storeState.data!.goals.map(g =>
					g.id === goalId ? { ...g, progress, progressUpdatedAt: new SvelteDate().toISOString() } : g
				)
			};

			try {
				const { error: apiError } = await client.PATCH('/goals/{goalId}/progress', {
					params: { path: { goalId } },
					body: { current_value: progress }
				});
				if (apiError) throw new Error(
					(apiError as { error?: { message?: string } })?.error?.message ?? 'Error al actualizar progreso'
				);
				resolve();
			} catch (e) {
				// Rollback
				storeState.data = {
					...storeState.data!,
					goals: storeState.data!.goals.map(g =>
						g.id === goalId ? { ...g, progress: previousProgress } : g
					)
				};
				reject(e);
			}
		}, 500);

		debounceTimers.set(goalId, timer);
	});
}

export async function addGoalComment(
	goalId: string,
	authorId: string,
	authorName: string,
	content: string
): Promise<void> {
	const { data, error: apiError } = await client.POST('/goals/{goalId}/comments', {
		params: { path: { goalId } },
		body: { content, author_id: authorId, author_name: authorName }
	});
	if (apiError) throw new Error('Error al guardar comentario');
	if (data) {
		const comment = mapToGoalComment(data);
		storeState.data = {
			...storeState.data!,
			goals: (storeState.data?.goals ?? []).map((g) =>
				g.id === goalId ? { ...g, comments: [...(g.comments ?? []), comment] } : g
			)
		};
	}
}

export async function deleteGoalComment(goalId: string, commentId: string): Promise<void> {
	const { error: apiError } = await client.DELETE('/goals/{goalId}/comments/{commentId}', {
		params: { path: { goalId, commentId } }
	});
	if (apiError) throw new Error('Error al eliminar comentario');
	storeState.data = {
		...storeState.data!,
		goals: (storeState.data?.goals ?? []).map((g) =>
			g.id === goalId
				? { ...g, comments: (g.comments ?? []).filter((c) => c.id !== commentId) }
				: g
		)
	};
}

// ─── Mutations: Category Comments ─────────────────────────────────────────────

export function getCategoryComments(categoryId: string): GoalComment[] {
	return storeState.data?.categories.find((c) => c.id === categoryId)?.comments ?? [];
}

export async function addCategoryComment(
	categoryId: string,
	authorId: string,
	authorName: string,
	content: string
): Promise<void> {
	const { data, error: apiError } = await client.POST('/categories/{catId}/comments', {
		params: { path: { catId: categoryId } },
		body: { content, author_id: authorId, author_name: authorName }
	});
	if (apiError) throw new Error('Error al guardar comentario');
	if (data) {
		const comment = mapToGoalComment(data);
		storeState.data = {
			...storeState.data!,
			categories: (storeState.data?.categories ?? []).map((c) =>
				c.id === categoryId ? { ...c, comments: [...(c.comments ?? []), comment] } : c
			)
		};
	}
}

export async function deleteCategoryComment(categoryId: string, commentId: string): Promise<void> {
	const { error: apiError } = await client.DELETE('/categories/{catId}/comments/{commentId}', {
		params: { path: { catId: categoryId, commentId } }
	});
	if (apiError) throw new Error('Error al eliminar comentario');
	storeState.data = {
		...storeState.data!,
		categories: (storeState.data?.categories ?? []).map((c) =>
			c.id === categoryId
				? { ...c, comments: (c.comments ?? []).filter((cm) => cm.id !== commentId) }
				: c
		)
	};
}

// ─── Mutations: Assignment Comments ──────────────────────────────────────────

export function getAssignmentComments(assignmentId: string): GoalComment[] {
	return storeState.data?.assignmentComments.get(assignmentId) ?? [];
}

export async function loadAssignmentComments(assignmentId: string): Promise<void> {
	if (!storeState.data) return;
	if (storeState.data.assignmentComments.has(assignmentId)) return;
	await fetchAndSetAssignmentComments(assignmentId);
}

export async function refreshAssignmentComments(assignmentId: string): Promise<void> {
	if (!storeState.data) return;
	await fetchAndSetAssignmentComments(assignmentId);
}

async function fetchAndSetAssignmentComments(assignmentId: string): Promise<void> {
	if (!storeState.data) return;
	const { data, error: apiError } = await client.GET('/assignments/{assignId}/comments', {
		params: { path: { assignId: assignmentId } }
	});
	if (apiError || !data) return;
	const comments = (data as unknown as GoalComment[]).map(mapToGoalComment);
	const newMap = new SvelteMap(storeState.data.assignmentComments);
	newMap.set(assignmentId, comments);
	storeState.data = { ...storeState.data, assignmentComments: newMap };
}

export async function addAssignmentComment(
	assignmentId: string,
	authorId: string,
	authorName: string,
	content: string
): Promise<void> {
	const { data, error: apiError } = await client.POST('/assignments/{assignId}/comments', {
		params: { path: { assignId: assignmentId } },
		body: { content, author_id: authorId, author_name: authorName }
	});
	if (apiError) throw new Error('Error al guardar comentario');
	if (data) {
		const comment = mapToGoalComment(data);
		const existing = storeState.data?.assignmentComments.get(assignmentId) ?? [];
		const newMap = new SvelteMap<string, GoalComment[]>(storeState.data?.assignmentComments ?? new SvelteMap());
		newMap.set(assignmentId, [...existing, comment]);
		storeState.data = { ...storeState.data!, assignmentComments: newMap };
	}
}

export async function deleteAssignmentComment(assignmentId: string, commentId: string): Promise<void> {
	const { error: apiError } = await client.DELETE('/assignments/{assignId}/comments/{commentId}', {
		params: { path: { assignId: assignmentId, commentId } }
	});
	if (apiError) throw new Error('Error al eliminar comentario');
	const existing = storeState.data?.assignmentComments.get(assignmentId) ?? [];
	const newMap = new SvelteMap<string, GoalComment[]>(storeState.data?.assignmentComments ?? new SvelteMap());
	newMap.set(assignmentId, existing.filter((c) => c.id !== commentId));
	storeState.data = { ...storeState.data!, assignmentComments: newMap };
}
