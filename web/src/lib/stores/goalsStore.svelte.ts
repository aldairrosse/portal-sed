import type {
	Goal,
	GoalCategory,
	GoalKpiLink,
	KPI,
	EmployeeAssignment,
	ChangeRequest,
	GoalComment,
	CyclePhase
} from '$lib/types/goal';
import { MANAGER_MAP } from '$lib/types/goal';
import type { EvaluationProfile } from '$lib/types/evaluation';

import categoriesData from '$lib/fixtures/goals/goal-categories.json';
import goalsData from '$lib/fixtures/goals/goals.json';
import kpisData from '$lib/fixtures/goals/kpis.json';
import goalKpiLinksData from '$lib/fixtures/goals/goal-kpi-links.json';
import assignmentsData from '$lib/fixtures/goals/assignments.json';
import { getPhase as getDevPhase } from '$lib/stores/devContext.svelte';

// ─── Tolerance ────────────────────────────────────────────────────────────────

const EPSILON = 0.01;

// ─── State ────────────────────────────────────────────────────────────────────

let categories = $state<GoalCategory[]>(structuredClone(categoriesData));
let goals = $state<Goal[]>(structuredClone(goalsData as Goal[]));
let kpis = $state<KPI[]>(structuredClone(kpisData as KPI[]));
let goalKpiLinks = $state<GoalKpiLink[]>(structuredClone(goalKpiLinksData as GoalKpiLink[]));
let assignments = $state<EmployeeAssignment[]>(structuredClone(assignmentsData as EmployeeAssignment[]));
let changeRequests = $state<ChangeRequest[]>([]);

// ─── Getters: General ─────────────────────────────────────────────────────────

export function getCategories(): GoalCategory[] {
	return categories;
}

export function getGoals(): Goal[] {
	return goals;
}

export function getKpis(): KPI[] {
	return kpis;
}

export function getGoalKpiLinks(): GoalKpiLink[] {
	return goalKpiLinks;
}

export function getAssignments(): EmployeeAssignment[] {
	return assignments;
}

export function getChangeRequests(): ChangeRequest[] {
	return changeRequests;
}

export function getCyclePhase(): CyclePhase {
	return getDevPhase();
}

// ─── Getters: Progress & Comments ─────────────────────────────────────────────

export function getGoalProgress(goalId: string): number | undefined {
	const goal = goals.find((g) => g.id === goalId);
	return goal?.progress;
}

export function getGoalComments(goalId: string): GoalComment[] {
	const goal = goals.find((g) => g.id === goalId);
	return goal?.comments ?? [];
}

export function getCategoryProgressAverage(categoryId: string): number {
	const catGoals = goals.filter((g) => g.categoryId === categoryId);
	if (catGoals.length === 0) return 0;
	const withProgress = catGoals.filter((g) => g.progress !== undefined);
	if (withProgress.length === 0) return 0;
	const total = withProgress.reduce((acc, g) => {
		const pct = g.unit === 'porcentaje' ? (g.progress ?? 0) : ((g.progress ?? 0) / (g.targetValue || 1)) * 100;
		return acc + Math.min(pct, 100);
	}, 0);
	return total / withProgress.length;
}

export function getGoalPermissions(
	role: EvaluationProfile,
	isOwner: boolean
): { canEditProgress: boolean; canComment: boolean; canEditWeight: boolean; canDelete: boolean } {
	if (getDevPhase() === 'inicio-anio') {
		return {
			canEditProgress: false,
			canComment: false,
			canEditWeight: isOwner,
			canDelete: isOwner
		};
	}
	// phase === 'avance'
	return {
		canEditProgress: true,
		canComment: true,
		canEditWeight: false,
		canDelete: false
	};
}

// ─── Getters: Filtered ────────────────────────────────────────────────────────

export function getGoalsByCategory(categoryId: string): Goal[] {
	return goals.filter((g) => g.categoryId === categoryId);
}

export function getKpisForGoal(goalId: string): KPI[] {
	const linkKpiIds = goalKpiLinks
		.filter((link) => link.goalId === goalId)
		.map((link) => link.kpiId);
	return kpis.filter((kpi) => linkKpiIds.includes(kpi.id));
}

export function getLinksForGoal(goalId: string): GoalKpiLink[] {
	return goalKpiLinks.filter((link) => link.goalId === goalId);
}

export function getLinksForKpi(kpiId: string): GoalKpiLink[] {
	return goalKpiLinks.filter((link) => link.kpiId === kpiId);
}

export function getAssignmentsByProfile(profileId: EvaluationProfile): EmployeeAssignment[] {
	return assignments.filter((a) => a.profileId === profileId);
}

export function getAssignmentByEmployee(employeeId: string): EmployeeAssignment | undefined {
	return assignments.find((a) => a.employeeId === employeeId);
}

// ─── Getters: Validation ──────────────────────────────────────────────────────

/**
 * Sum of all category weights equals 100 ± ε.
 */
function doCategoryWeightsSumTo100(): boolean {
	const sum = categories.reduce((acc, c) => acc + c.weight, 0);
	return Math.abs(sum - 100) <= EPSILON;
}

/**
 * For each category, the goals within it sum to 100 ± ε.
 */
function areAllCategoryGoalWeightsValid(): boolean {
	for (const cat of categories) {
		const catGoals = goals.filter((g) => g.categoryId === cat.id);
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
	const catGoals = goals.filter((g) => g.categoryId === categoryId);
	if (catGoals.length === 0) return true;
	const sum = catGoals.reduce((acc, g) => acc + g.weight, 0);
	return Math.abs(sum - 100) <= EPSILON;
}

// ─── Getters: Manager Hierarchy ───────────────────────────────────────────────

/**
 * Returns the manager profile for a given profile, or undefined if none.
 */
export function getManagerOf(profile: EvaluationProfile): EvaluationProfile | undefined {
	return MANAGER_MAP[profile];
}

// ─── Mutations: Categories ────────────────────────────────────────────────────

export function addCategory(category: GoalCategory): void {
	categories = [...categories, category];
}

export function updateCategory(id: string, updates: Partial<Omit<GoalCategory, 'id'>>): void {
	categories = categories.map((c) => (c.id === id ? { ...c, ...updates } : c));
}

export function deleteCategory(id: string): void {
	// Block deletion in 'avance' phase
	if (getDevPhase() === 'medio-anio') return;
	// Cascade: remove goals of this category
	const deletedGoalIds = goals.filter((g) => g.categoryId === id).map((g) => g.id);
	goals = goals.filter((g) => g.categoryId !== id);
	// Cascade: remove KPI links for deleted goals
	goalKpiLinks = goalKpiLinks.filter((link) => !deletedGoalIds.includes(link.goalId));
	// Remove the category
	categories = categories.filter((c) => c.id !== id);
}

// ─── Mutations: Goals ─────────────────────────────────────────────────────────

export function addGoal(goal: Goal): void {
	goals = [...goals, goal];
}

export function updateGoal(id: string, updates: Partial<Omit<Goal, 'id'>>): void {
	goals = goals.map((g) => (g.id === id ? { ...g, ...updates } : g));
}

export function deleteGoal(id: string): void {
	// Block deletion in 'avance' phase
	if (getDevPhase() === 'medio-anio') return;
	goals = goals.filter((g) => g.id !== id);
	// Cascade: remove KPI links for this goal
	goalKpiLinks = goalKpiLinks.filter((link) => link.goalId !== id);
}

// ─── Mutations: KPIs ──────────────────────────────────────────────────────────

export function addKpi(kpi: KPI): void {
	kpis = [...kpis, kpi];
}

export function updateKpi(id: string, updates: Partial<Omit<KPI, 'id'>>): void {
	kpis = kpis.map((k) => (k.id === id ? { ...k, ...updates } : k));
}

export function deleteKpi(id: string): void {
	kpis = kpis.filter((k) => k.id !== id);
	// Cascade: remove links for this KPI
	goalKpiLinks = goalKpiLinks.filter((link) => link.kpiId !== id);
}

// ─── Mutations: GoalKpiLink (N:M) ─────────────────────────────────────────────

export function linkKpiToGoal(goalId: string, kpiId: string, weight?: number): void {
	// Idempotent: skip if link already exists
	const exists = goalKpiLinks.some((link) => link.goalId === goalId && link.kpiId === kpiId);
	if (exists) return;
	goalKpiLinks = [...goalKpiLinks, { goalId, kpiId, weight }];
}

export function unlinkKpiFromGoal(goalId: string, kpiId: string): void {
	goalKpiLinks = goalKpiLinks.filter(
		(link) => !(link.goalId === goalId && link.kpiId === kpiId)
	);
}

export function updateLinkWeight(goalId: string, kpiId: string, weight: number | undefined): void {
	goalKpiLinks = goalKpiLinks.map((link) =>
		link.goalId === goalId && link.kpiId === kpiId ? { ...link, weight } : link
	);
}

// ─── Mutations: Assignments ───────────────────────────────────────────────────

export function addAssignment(assignment: EmployeeAssignment): void {
	assignments = [...assignments, assignment];
}

export function updateAssignment(
	id: string,
	updates: Partial<Omit<EmployeeAssignment, 'id'>>
): void {
	assignments = assignments.map((a) =>
		a.id === id ? { ...a, ...updates, updatedAt: new Date().toISOString() } : a
	);
}

export function deleteAssignment(id: string): void {
	assignments = assignments.filter((a) => a.id !== id);
}

export function assignGoalToEmployee(employeeId: string, goalId: string): void {
	assignments = assignments.map((a) =>
		a.employeeId === employeeId && !a.goalIds.includes(goalId)
			? { ...a, goalIds: [...a.goalIds, goalId], updatedAt: new Date().toISOString() }
			: a
	);
}

export function unassignGoalFromEmployee(employeeId: string, goalId: string): void {
	assignments = assignments.map((a) =>
		a.employeeId === employeeId
			? {
					...a,
					goalIds: a.goalIds.filter((gid) => gid !== goalId),
					updatedAt: new Date().toISOString()
				}
			: a
	);
}

// ─── Mutations: ChangeRequests ────────────────────────────────────────────────

export function recordChangeRequest(request: ChangeRequest): void {
	changeRequests = [...changeRequests, request];
}

export function approveChangeRequest(id: string, approvedBy: string): void {
	changeRequests = changeRequests.map((cr) =>
		cr.id === id
			? { ...cr, status: 'approved', approvedBy, approvedAt: new Date().toISOString() }
			: cr
	);
}

export function rejectChangeRequest(id: string): void {
	changeRequests = changeRequests.map((cr) =>
		cr.id === id ? { ...cr, status: 'rejected' } : cr
	);
}

// ─── Mutations: Progress & Comments ───────────────────────────────────────────

export function updateGoalProgress(goalId: string, progress: number): void {
	goals = goals.map((g) =>
		g.id === goalId
			? { ...g, progress, progressUpdatedAt: new Date().toISOString() }
			: g
	);
}

export function addGoalComment(
	goalId: string,
	authorId: string,
	authorName: string,
	content: string
): void {
	const comment: GoalComment = {
		id: `comment-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
		authorId,
		authorName,
		content,
		createdAt: new Date().toISOString()
	};
	goals = goals.map((g) =>
		g.id === goalId ? { ...g, comments: [...(g.comments ?? []), comment] } : g
	);
}

export function deleteGoalComment(goalId: string, commentId: string): void {
	goals = goals.map((g) =>
		g.id === goalId
			? { ...g, comments: (g.comments ?? []).filter((c) => c.id !== commentId) }
			: g
	);
}
