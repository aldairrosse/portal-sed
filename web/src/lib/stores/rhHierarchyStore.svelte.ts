import type { OrgNode } from '$lib/types/org-hierarchy';
import type { Goal, EmployeeAssignment } from '$lib/types/goal';
import type { CompetencyRating } from '$lib/types/evaluation-result';
import type { EvaluationProfile } from '$lib/types/evaluation';
import { PROFILE_LABELS } from '$lib/types/evaluation';
import { getRoot, getScopeIds } from '$lib/stores/orgHierarchyStore.svelte';

import goalsData from '$lib/fixtures/goals/goals.json';
import assignmentsData from '$lib/fixtures/goals/assignments.json';
import rhEvaluationsData from '$lib/fixtures/evaluations/rh-evaluations.json';

// ─── Types ──────────────────────────────────────────────────────────────────────

export interface AreaMetrics {
	nodeId: string;
	employeeCount: number;
	employeesWithGoals: number;
	avgProgress: number | null;
	completedGoals: number;
	pendingGoals: number;
	avgRating: number | null;
	ratingsCount: number;
}

export interface EmployeeRow {
	id: string;
	name: string;
	position: string;
	profile: string;
}

// ─── Module state ───────────────────────────────────────────────────────────────

let selectedNodeId = $state<string>('');
let metrics = $state<AreaMetrics | null>(null);
let employeeList = $state<EmployeeRow[]>([]);

// ─── Internal helpers ───────────────────────────────────────────────────────────

function collectNodesByIds(nodeIds: string[]): OrgNode[] {
	const root = getRoot();
	if (!root) return [];
	const idSet = new Set(nodeIds);
	const result: OrgNode[] = [];
	const stack = [root];
	while (stack.length > 0) {
		const node = stack.pop()!;
		if (idSet.has(node.id)) result.push(node);
		stack.push(...node.children);
	}
	return result;
}

// ─── Pure aggregation functions (exported, independently testable) ───────────────

/**
 * Compute goal progress metrics for a set of employees.
 *
 * For each employee with at least one goal (targetValue > 0), their
 * individual mean progress is computed. Employees with zero goals are
 * excluded from the average.
 *
 * Completed / pending counts are per-goal across the entire subtree.
 */
export function computeAreaProgress(
	employeeIds: string[],
	assignments: EmployeeAssignment[],
	goals: Goal[]
): { avgProgress: number; completed: number; pending: number } {
	const idSet = new Set(employeeIds);
	const relevantAssignments = assignments.filter((a) => idSet.has(a.employeeId));

	if (relevantAssignments.length === 0) {
		return { avgProgress: 0, completed: 0, pending: 0 };
	}

	const goalMap = new Map<string, Goal>();
	for (const g of goals) {
		goalMap.set(g.id, g);
	}

	const employeeProgresses: number[] = [];
	let completed = 0;
	let pending = 0;

	for (const assignment of relevantAssignments) {
		const employeeGoals: Goal[] = [];
		for (const gid of assignment.goalIds) {
			const goal = goalMap.get(gid);
			// Skip goals with targetValue === 0 (avoid division by zero)
			if (goal && goal.targetValue > 0) {
				employeeGoals.push(goal);
			}
		}

		// Skip employees with zero valid goals
		if (employeeGoals.length === 0) continue;

		let sum = 0;
		for (const goal of employeeGoals) {
			const pct = ((goal.progress ?? 0) / goal.targetValue) * 100;
			sum += Math.min(pct, 100);

			if (goal.progress !== undefined && goal.progress >= goal.targetValue) {
				completed++;
			} else {
				pending++;
			}
		}
		employeeProgresses.push(sum / employeeGoals.length);
	}

	if (employeeProgresses.length === 0) {
		return { avgProgress: 0, completed: 0, pending: 0 };
	}

	const total = employeeProgresses.reduce((a, b) => a + b, 0);
	return { avgProgress: total / employeeProgresses.length, completed, pending };
}

/**
 * Compute the mean RH evaluation rating for a set of employees.
 *
 * Only employees with at least one `rhRating` contribute to the mean.
 * Returns `avgRating: null` when no ratings exist (not zero).
 */
export function computeAreaRating(
	employeeIds: string[],
	rhEvaluations: CompetencyRating[]
): { avgRating: number | null; ratingsCount: number } {
	const idSet = new Set(employeeIds);
	const ratings: number[] = [];
	for (const e of rhEvaluations) {
		if (idSet.has(e.employeeId) && e.rhRating !== undefined) {
			ratings.push(e.rhRating);
		}
	}

	if (ratings.length === 0) {
		return { avgRating: null, ratingsCount: 0 };
	}

	const sum = ratings.reduce((a, b) => a + b, 0);
	return { avgRating: sum / ratings.length, ratingsCount: ratings.length };
}

/**
 * Build the employee list for a subtree.
 *
 * Includes the area manager (the node itself). Sorted A–Z by name.
 * Profile IDs are mapped to Spanish labels via `PROFILE_LABELS`.
 */
export function buildEmployeeList(
	employeeIds: string[],
	nodes: OrgNode[]
): EmployeeRow[] {
	const nodeMap = new Map<string, OrgNode>();
	for (const node of nodes) {
		nodeMap.set(node.id, node);
	}

	return employeeIds
		.map((id) => nodeMap.get(id))
		.filter((n): n is OrgNode => n !== undefined)
		.map((n) => ({
			id: n.id,
			name: n.name,
			position: PROFILE_LABELS[n.profileId as EvaluationProfile] ?? n.profileId,
			profile: n.profileId
		}))
		.sort((a, b) => a.name.localeCompare(b.name));
}

// ─── Actions ────────────────────────────────────────────────────────────────────

/**
 * Select an org node and compute its metrics + employee list.
 *
 * Called when RRHH clicks a node in `OrgHierarchyTree`. Reads from
 * existing fixture data and the `orgHierarchyStore` tree.
 */
export function selectNode(nodeId: string): void {
	selectedNodeId = nodeId;

	if (!nodeId) {
		metrics = null;
		employeeList = [];
		return;
	}

	const employeeIds = getScopeIds(nodeId);
	if (employeeIds.length === 0) {
		metrics = null;
		employeeList = [];
		return;
	}

	const assignments = assignmentsData as EmployeeAssignment[];
	const goals = goalsData as Goal[];
	const evaluations = rhEvaluationsData as CompetencyRating[];
	const nodes = collectNodesByIds(employeeIds);

	const progress = computeAreaProgress(employeeIds, assignments, goals);
	const rating = computeAreaRating(employeeIds, evaluations);

	// Count employees with at least one valid goal (targetValue > 0)
	const employeesWithGoals = assignments.filter((a) => {
		if (!employeeIds.includes(a.employeeId)) return false;
		return a.goalIds.some((gid) => {
			const goal = goals.find((g) => g.id === gid);
			return goal !== undefined && goal.targetValue > 0;
		});
	}).length;

	metrics = {
		nodeId,
		employeeCount: employeeIds.length,
		employeesWithGoals,
		avgProgress: progress.avgProgress > 0 || progress.completed > 0 || progress.pending > 0 ? progress.avgProgress : null,
		completedGoals: progress.completed,
		pendingGoals: progress.pending,
		avgRating: rating.avgRating,
		ratingsCount: rating.ratingsCount
	};

	employeeList = buildEmployeeList(employeeIds, nodes);
}

// ─── Getters ────────────────────────────────────────────────────────────────────

export function getMetrics(): AreaMetrics | null {
	return metrics;
}

export function getEmployeeList(): EmployeeRow[] {
	return employeeList;
}

export function getSelectedNodeId(): string {
	return selectedNodeId;
}
