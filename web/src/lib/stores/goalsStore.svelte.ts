import type {
	Goal,
	GoalCategory,
	GoalKpiLink,
	KPI,
	EmployeeAssignment,
	ChangeRequest,
	GoalComment,
	GoalUnit,
	KpiUnit,
	CyclePhase
} from '$lib/types/goal';
import type { EvaluationProfile } from '$lib/types/evaluation';

import categoriesData from '$lib/fixtures/goals/goal-categories.json';
import goalsData from '$lib/fixtures/goals/goals.json';
import kpisData from '$lib/fixtures/goals/kpis.json';
import goalKpiLinksData from '$lib/fixtures/goals/goal-kpi-links.json';
import assignmentsData from '$lib/fixtures/goals/assignments.json';
import { getActivePhase } from '$lib/api/cycle.svelte';
import { getSession } from '$lib/api/session.svelte';
import { client } from '$lib/api/client';
import { progressPercent } from '$lib/utils/scoring';

// ─── Internal data shape ──────────────────────────────────────────────────────

interface StoreData {
	categories: GoalCategory[];
	goals: Goal[];
	kpis: KPI[];
	goalKpiLinks: GoalKpiLink[];
	assignments: EmployeeAssignment[];
	changeRequests: ChangeRequest[];
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

// ─── Helpers ──────────────────────────────────────────────────────────────────

function getEmployeeId(): string {
	return getSession().user?.employeeId ?? '';
}

function loadFixtures(): StoreData {
	return {
		categories: structuredClone(categoriesData),
		goals: structuredClone(goalsData as Goal[]),
		kpis: structuredClone(kpisData as KPI[]),
		goalKpiLinks: structuredClone(goalKpiLinksData as GoalKpiLink[]),
		assignments: structuredClone(assignmentsData as EmployeeAssignment[]),
		changeRequests: []
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
			kpis?: Array<{
				id?: string;
				name?: string;
				unit?: string;
				description?: string;
				direction?: string;
				current_value?: number;
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
	}>,
	apiAssignment: {
		id?: string;
		employee_id?: string;
		cycle_id?: string;
		categories?: Array<unknown>;
		created_at?: string;
	} | null
): StoreData {
	const cats: GoalCategory[] = [];
	const goals: Goal[] = [];
	const goalKpiLinks: GoalKpiLink[] = [];
	const assignedGoalIds: string[] = [];

	// 1. Build full KPI catalog from the /kpis endpoint
	const kpisMap = new Map<string, KPI>();
	for (const ak of apiKpis) {
		const kpi: KPI = {
			id: ak.id ?? crypto.randomUUID(),
			name: ak.name ?? '',
			description: ak.description ?? '',
			unit: (ak.unit as KpiUnit) ?? 'numero',
			direction: (ak.direction as 'ascendente' | 'descendente') ?? 'ascendente',
			currentValue: ak.current_value,
			targetValue: undefined,
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
			weight: ac.weight ?? 0
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
				comments: []
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
							targetValue: undefined,
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
			profileId: 'colaborador',
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
		changeRequests: []
	};
}

/**
 * Load goals data.
 *
 * In DEV without VITE_USE_API: loads from fixture files (structured clone).
 * In production / VITE_USE_API=true: fetches from the real API endpoints.
 */
export async function load(): Promise<void> {
	storeState.loading = true;
	storeState.error = null;

	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		storeState.data = loadFixtures();
		storeState.loading = false;
		return;
	}

	const empId = getEmployeeId();

	try {
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

		const apiCategories = (catsRes.data as { items?: Array<unknown> })?.items ?? [];
		const apiKpis = (kpisRes.data as { items?: Array<unknown> })?.items ?? [];
		const apiAssignment = assignmentRes.data as
			| {
					id?: string;
					employee_id?: string;
					cycle_id?: string;
					categories?: Array<unknown>;
					created_at?: string;
			  }
			| null
			| undefined;

		storeState.data = normalizeApiData(
			apiCategories as Parameters<typeof normalizeApiData>[0],
			apiKpis as Parameters<typeof normalizeApiData>[1],
			apiAssignment ?? null
		);
	} catch (e) {
		storeState.error = e instanceof Error ? e.message : 'Error desconocido al cargar datos de objetivos';
	} finally {
		storeState.loading = false;
	}
}

/** Alias for load(). */
export function reload(): Promise<void> {
	return load();
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
	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		storeState.data = { ...storeState.data!, categories: [...(storeState.data?.categories ?? []), category] };
		return;
	}
	const empId = getEmployeeId();
	const { error: apiError } = await client.POST('/employees/{empId}/categories', {
		params: { path: { empId } },
		body: { name: category.name, description: category.description, weight: category.weight }
	});
	if (apiError) throw new Error((apiError as { error?: { message?: string } })?.error?.message ?? 'Error al crear categoría');
	await reload();
}

export async function updateCategory(id: string, updates: Partial<Omit<GoalCategory, 'id'>>): Promise<void> {
	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		storeState.data = {
			...storeState.data!,
			categories: (storeState.data?.categories ?? []).map((c) => (c.id === id ? { ...c, ...updates } : c))
		};
		return;
	}
	const empId = getEmployeeId();
	const { error: apiError } = await client.PUT('/employees/{empId}/categories/{catId}', {
		params: { path: { empId, catId: id } },
		body: {
			name: updates.name ?? '',
			description: updates.description ?? '',
			weight: updates.weight ?? 0
		}
	});
	if (apiError) throw new Error((apiError as { error?: { message?: string } })?.error?.message ?? 'Error al actualizar categoría');
	await reload();
}

export async function deleteCategory(id: string): Promise<void> {
	// Block deletion outside 'inicio-anio' phase
	const phase = getActivePhase() ?? 'inicio-anio';
	if (phase === 'medio-anio' || phase === 'fin-anio') return;

	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		// Cascade: remove goals of this category
		const deletedGoalIds = (storeState.data?.goals ?? []).filter((g) => g.categoryId === id).map((g) => g.id);
		storeState.data = {
			...storeState.data!,
			goals: (storeState.data?.goals ?? []).filter((g) => g.categoryId !== id),
			goalKpiLinks: (storeState.data?.goalKpiLinks ?? []).filter((link) => !deletedGoalIds.includes(link.goalId)),
			categories: (storeState.data?.categories ?? []).filter((c) => c.id !== id)
		};
		return;
	}
	const empId = getEmployeeId();
	const { error: apiError } = await client.DELETE('/employees/{empId}/categories/{catId}', {
		params: { path: { empId, catId: id } }
	});
	if (apiError) throw new Error((apiError as { error?: { message?: string } })?.error?.message ?? 'Error al eliminar categoría');
	await reload();
}

// ─── Mutations: Goals ─────────────────────────────────────────────────────────

export async function addGoal(goal: Goal): Promise<void> {
	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		storeState.data = { ...storeState.data!, goals: [...(storeState.data?.goals ?? []), goal] };
		return;
	}
	const empId = getEmployeeId();
	const { error: apiError } = await client.POST('/employees/{empId}/categories/{catId}/goals', {
		params: { path: { empId, catId: goal.categoryId } },
		body: {
			name: goal.name,
			description: goal.description,
			unit: goal.unit as 'porcentaje' | 'moneda' | 'numero',
			weight: goal.weight,
			target_value: goal.targetValue
		}
	});
	if (apiError) throw new Error((apiError as { error?: { message?: string } })?.error?.message ?? 'Error al crear meta');
	await reload();
}

export async function updateGoal(id: string, updates: Partial<Omit<Goal, 'id'>>): Promise<void> {
	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		storeState.data = {
			...storeState.data!,
			goals: (storeState.data?.goals ?? []).map((g) => (g.id === id ? { ...g, ...updates } : g))
		};
		return;
	}
	const { error: apiError } = await client.PUT('/goals/{goalId}', {
		params: { path: { goalId: id } },
		body: {
			name: updates.name ?? '',
			description: updates.description ?? '',
			unit: (updates.unit as 'porcentaje' | 'moneda' | 'numero') ?? 'numero',
			weight: updates.weight ?? 0,
			target_value: updates.targetValue ?? 0,
			version: 1
		}
	});
	if (apiError) throw new Error((apiError as { error?: { message?: string } })?.error?.message ?? 'Error al actualizar meta');
	await reload();
}

export async function deleteGoal(id: string): Promise<void> {
	// Block deletion outside 'inicio-anio' phase
	const phase = getActivePhase() ?? 'inicio-anio';
	if (phase === 'medio-anio' || phase === 'fin-anio') return;

	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		storeState.data = {
			...storeState.data!,
			goals: (storeState.data?.goals ?? []).filter((g) => g.id !== id),
			goalKpiLinks: (storeState.data?.goalKpiLinks ?? []).filter((link) => link.goalId !== id)
		};
		return;
	}
	const { error: apiError } = await client.DELETE('/goals/{goalId}', {
		params: { path: { goalId: id } }
	});
	if (apiError) throw new Error((apiError as { error?: { message?: string } })?.error?.message ?? 'Error al eliminar meta');
	await reload();
}

// ─── Mutations: KPIs ──────────────────────────────────────────────────────────

export async function addKpi(kpi: KPI): Promise<void> {
	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		storeState.data = { ...storeState.data!, kpis: [...(storeState.data?.kpis ?? []), kpi] };
		return;
	}
	const { error: apiError } = await client.POST('/kpis', {
		body: {
			name: kpi.name,
			description: kpi.description,
			unit: kpi.unit as 'porcentaje' | 'moneda' | 'numero'
		}
	});
	if (apiError) throw new Error((apiError as { error?: { message?: string } })?.error?.message ?? 'Error al crear KPI');
	await reload();
}

export async function updateKpi(id: string, updates: Partial<Omit<KPI, 'id'>>): Promise<void> {
	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		storeState.data = {
			...storeState.data!,
			kpis: (storeState.data?.kpis ?? []).map((k) => (k.id === id ? { ...k, ...updates } : k))
		};
		return;
	}
	const { error: apiError } = await client.PUT('/kpis/{kpiId}', {
		params: { path: { kpiId: id } },
		body: {
			name: updates.name ?? '',
			description: updates.description ?? '',
			unit: (updates.unit as 'porcentaje' | 'moneda' | 'numero') ?? 'numero'
		}
	});
	if (apiError) throw new Error((apiError as { error?: { message?: string } })?.error?.message ?? 'Error al actualizar KPI');
	await reload();
}

export async function deleteKpi(id: string): Promise<void> {
	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		storeState.data = {
			...storeState.data!,
			kpis: (storeState.data?.kpis ?? []).filter((k) => k.id !== id),
			goalKpiLinks: (storeState.data?.goalKpiLinks ?? []).filter((link) => link.kpiId !== id)
		};
		return;
	}
	const { error: apiError } = await client.DELETE('/kpis/{kpiId}', {
		params: { path: { kpiId: id } }
	});
	if (apiError) throw new Error((apiError as { error?: { message?: string } })?.error?.message ?? 'Error al eliminar KPI');
	await reload();
}

// ─── Mutations: GoalKpiLink (N:M) ─────────────────────────────────────────────

export async function linkKpiToGoal(goalId: string, kpiId: string, weight?: number): Promise<void> {
	// Idempotent: skip if link already exists
	const exists = (storeState.data?.goalKpiLinks ?? []).some((link) => link.goalId === goalId && link.kpiId === kpiId);
	if (exists) return;

	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		storeState.data = { ...storeState.data!, goalKpiLinks: [...(storeState.data?.goalKpiLinks ?? []), { goalId, kpiId, weight }] };
		return;
	}
	const { error: apiError } = await client.POST('/goals/{goalId}/kpis', {
		params: { path: { goalId } },
		body: { kpi_id: kpiId }
	});
	if (apiError) throw new Error((apiError as { error?: { message?: string } })?.error?.message ?? 'Error al vincular KPI');
	await reload();
}

export async function unlinkKpiFromGoal(goalId: string, kpiId: string): Promise<void> {
	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		storeState.data = {
			...storeState.data!,
			goalKpiLinks: (storeState.data?.goalKpiLinks ?? []).filter((link) => !(link.goalId === goalId && link.kpiId === kpiId))
		};
		return;
	}
	const { error: apiError } = await client.DELETE('/goals/{goalId}/kpis/{kpiId}', {
		params: { path: { goalId, kpiId } }
	});
	if (apiError) throw new Error((apiError as { error?: { message?: string } })?.error?.message ?? 'Error al desvincular KPI');
	await reload();
}

export async function updateLinkWeight(goalId: string, kpiId: string, weight: number | undefined): Promise<void> {
	// No dedicated API endpoint for link weight. Local-only for now.
	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		storeState.data = {
			...storeState.data!,
			goalKpiLinks: (storeState.data?.goalKpiLinks ?? []).map((link) =>
				link.goalId === goalId && link.kpiId === kpiId ? { ...link, weight } : link
			)
		};
		return;
	}
	// In API mode, update locally and rely on next reload() for consistency
	storeState.data = {
		...storeState.data!,
		goalKpiLinks: (storeState.data?.goalKpiLinks ?? []).map((link) =>
			link.goalId === goalId && link.kpiId === kpiId ? { ...link, weight } : link
		)
	};
}

// ─── Mutations: Assignments ───────────────────────────────────────────────────

export async function addAssignment(assignment: EmployeeAssignment): Promise<void> {
	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		storeState.data = { ...storeState.data!, assignments: [...(storeState.data?.assignments ?? []), assignment] };
		return;
	}
	const empId = getEmployeeId();
	const { error: apiError } = await client.POST('/employees/{empId}/assignments', {
		params: { path: { empId } },
		body: { cycle_id: assignment.id }
	});
	if (apiError) throw new Error((apiError as { error?: { message?: string } })?.error?.message ?? 'Error al crear asignación');
	await reload();
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
				? { ...a, ...updates, updatedAt: new Date().toISOString() }
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
				? { ...a, goalIds: [...a.goalIds, goalId], updatedAt: new Date().toISOString() }
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
						updatedAt: new Date().toISOString()
					}
				: a
		)
	};
}

// ─── Mutations: ChangeRequests ────────────────────────────────────────────────

export async function recordChangeRequest(request: ChangeRequest): Promise<void> {
	// UI-only concept, no API endpoint. Local-only.
	storeState.data = { ...storeState.data!, changeRequests: [...(storeState.data?.changeRequests ?? []), request] };
}

export async function approveChangeRequest(id: string, approvedBy: string): Promise<void> {
	// UI-only concept, no API endpoint. Local-only.
	storeState.data = {
		...storeState.data!,
		changeRequests: (storeState.data?.changeRequests ?? []).map((cr) =>
			cr.id === id
				? { ...cr, status: 'approved' as const, approvedBy, approvedAt: new Date().toISOString() }
				: cr
		)
	};
}

export async function rejectChangeRequest(id: string): Promise<void> {
	// UI-only concept, no API endpoint. Local-only.
	storeState.data = {
		...storeState.data!,
		changeRequests: (storeState.data?.changeRequests ?? []).map((cr) =>
			cr.id === id ? { ...cr, status: 'rejected' as const } : cr
		)
	};
}

// ─── Mutations: Progress & Comments ───────────────────────────────────────────

export async function updateGoalProgress(goalId: string, progress: number): Promise<void> {
	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		storeState.data = {
			...storeState.data!,
			goals: (storeState.data?.goals ?? []).map((g) =>
				g.id === goalId
					? { ...g, progress, progressUpdatedAt: new Date().toISOString() }
					: g
			)
		};
		return;
	}
	const { error: apiError } = await client.PATCH('/goals/{goalId}/progress', {
		params: { path: { goalId } },
		body: { current_value: progress }
	});
	if (apiError) throw new Error((apiError as { error?: { message?: string } })?.error?.message ?? 'Error al actualizar progreso');
	await reload();
}

export async function addGoalComment(
	goalId: string,
	authorId: string,
	authorName: string,
	content: string
): Promise<void> {
	// No dedicated API endpoint for comments. Local-only for now.
	const comment = {
		id: `comment-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
		authorId,
		authorName,
		content,
		createdAt: new Date().toISOString()
	};
	storeState.data = {
		...storeState.data!,
		goals: (storeState.data?.goals ?? []).map((g) =>
			g.id === goalId ? { ...g, comments: [...(g.comments ?? []), comment] } : g
		)
	};
}

export async function deleteGoalComment(goalId: string, commentId: string): Promise<void> {
	// No dedicated API endpoint for comments. Local-only for now.
	storeState.data = {
		...storeState.data!,
		goals: (storeState.data?.goals ?? []).map((g) =>
			g.id === goalId
				? { ...g, comments: (g.comments ?? []).filter((c) => c.id !== commentId) }
				: g
		)
	};
}
