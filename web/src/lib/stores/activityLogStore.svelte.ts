import { client } from '$lib/api/client';
import type { components } from '$lib/api/schemas/activity-logs';

// ─── Types ──────────────────────────────────────────────────────────────────────

export interface ActivityLog {
	id: string;
	action: string;
	description: string;
	module: string;
	metadata?: Record<string, unknown> | null;
	created_at: string;
}

// ─── Module state ────────────────────────────────────────────────────────────────

let logs = $state<ActivityLog[]>([]);
let loading = $state(false);
let error = $state<string | null>(null);

// ─── Getters ─────────────────────────────────────────────────────────────────────

export function getActivityLogs(): ActivityLog[] {
	return logs;
}

export function isLoading(): boolean {
	return loading;
}

export function getError(): string | null {
	return error;
}

// ─── Load ────────────────────────────────────────────────────────────────────────

export async function loadActivityLogs(employeeId: string, limit = 50): Promise<void> {
	loading = true;
	error = null;

	try {
		const { data, error: apiError } = await client.GET(
			'/employees/{employeeId}/activity-logs',
			{ params: { path: { employeeId }, query: { limit } } }
		);

		if (apiError) {
			throw new Error(
				typeof apiError === 'string' ? apiError : 'Error al cargar actividad'
			);
		}

		const raw = data as { data?: Array<components['schemas']['ActivityLog']> };
		logs = (raw?.data ?? []).map((l) => ({
			id: l.id,
			action: l.action,
			description: l.description,
			module: l.module,
			metadata: l.metadata ?? null,
			created_at: l.created_at
		}));
	} catch (e) {
		error = e instanceof Error ? e.message : 'Error al cargar actividad';
		logs = [];
	} finally {
		loading = false;
	}
}

/** Alias for loadActivityLogs. */
export function reload(employeeId: string): Promise<void> {
	return loadActivityLogs(employeeId);
}

export function clearLogs(): void {
	logs = [];
}
