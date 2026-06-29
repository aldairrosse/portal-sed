import { client } from '$lib/api/client';
import { getActiveCycle } from '$lib/stores/cycleStore.svelte';

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
	employees: AreaMetricsEmployee[];
}

export interface AreaMetricsEmployee {
	id: string;
	firstName: string;
	lastName: string;
	jobTitle: string;
	profileId: string;
	profileDescription: string;
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
let metricsLoading = $state(false);
let metricsError = $state<string | null>(null);

// ─── Actions ────────────────────────────────────────────────────────────────────

/**
 * Fetch area metrics for an org node from the API.
 */
async function fetchAreaMetrics(nodeId: string): Promise<void> {
	metricsLoading = true;
	metricsError = null;

	try {
		const activeCycle = getActiveCycle();
		const params: Record<string, string> = {};

		if (activeCycle?.id) {
			params.cycleId = activeCycle.id;
		}

		const query = Object.keys(params).length > 0 ? params : undefined;
		const res = await client.GET('/org-nodes/{nodeId}/area-metrics', {
			params: {
				path: { nodeId },
				query
			}
		});

		if (res.error) {
			throw new Error('Error al cargar métricas del área');
		}

		const raw = res.data as AreaMetrics;
		if (!raw) {
			throw new Error('No se recibieron métricas');
		}

		metrics = {
			nodeId: raw.nodeId,
			employeeCount: raw.employeeCount ?? 0,
			employeesWithGoals: raw.employeesWithGoals ?? 0,
			avgProgress: raw.avgProgress ?? null,
			completedGoals: raw.completedGoals ?? 0,
			pendingGoals: raw.pendingGoals ?? 0,
			avgRating: raw.avgRating ?? null,
			ratingsCount: raw.ratingsCount ?? 0,
			employees: raw.employees ?? []
		};

		console.log('Employees from API:', raw.employees);
		employeeList = buildEmployeeListFromApi(raw.employees ?? []);
	} catch (e) {
		metricsError = e instanceof Error ? e.message : 'Error al cargar métricas';
		metrics = null;
		// Keep last known employee list on failure
	} finally {
		metricsLoading = false;
	}
}

/**
 * Map API employee array to EmployeeRow[]. Sorted A–Z.
 */
function buildEmployeeListFromApi(employees: AreaMetricsEmployee[]): EmployeeRow[] {
	return employees
		.map((emp) => ({
			id: emp.id,
			name: `${emp.firstName} ${emp.lastName}`,
			position: emp.jobTitle ?? '',
			profile: emp.profileDescription ?? ''
		}))
		.sort((a, b) => a.name.localeCompare(b.name));
}

/**
 * Select an org node and fetch its metrics from the API.
 *
 * Called when RRHH clicks a node in `OrgHierarchyTree`.
 */
export function selectNode(nodeId: string): void {
	selectedNodeId = nodeId;

	if (!nodeId) {
		metrics = null;
		employeeList = [];
		metricsError = null;
		return;
	}

	fetchAreaMetrics(nodeId);
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

export function isLoadingMetrics(): boolean {
	return metricsLoading;
}

export function getMetricsError(): string | null {
	return metricsError;
}
