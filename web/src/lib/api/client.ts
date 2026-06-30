import createClient from 'openapi-fetch';
import type { paths as AuthPaths } from './schemas/auth.d.ts';
import type { paths as CyclePaths } from './schemas/cycle.d.ts';
import type { paths as GoalsPaths } from './schemas/goals.d.ts';
import type { paths as CompetencyPaths } from './schemas/competency.d.ts';
import type { paths as OrgHierarchyPaths } from './schemas/org-hierarchy.d.ts';
import type { paths as ActivityLogsPaths } from './schemas/activity-logs.d.ts';
import type { paths as EvalPaths } from './schemas/evaluations.d.ts';

type AppPaths = AuthPaths & CyclePaths & GoalsPaths & CompetencyPaths & OrgHierarchyPaths & ActivityLogsPaths & EvalPaths;

export const baseURL: string = import.meta.env.VITE_API_URL ?? '/api/v1';

async function fetchWithCredentials(input: RequestInfo | URL, init?: RequestInit): Promise<Response> {
	const response = await fetch(input, { ...init, credentials: 'include' });
	if (response.status === 401) {
		window.location.href = '/login';
	}
	return response;
}

export const client = createClient<AppPaths>({
	baseUrl: baseURL,
	headers: { 'Content-Type': 'application/json' },
	fetch: fetchWithCredentials
});
