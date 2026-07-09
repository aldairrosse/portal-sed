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

// ponytail: propagate HTTP 404 as a catchable error so stores can treat "no record" as empty data
export class HttpNotFoundError extends Error {
	readonly statusCode = 404;
}

async function fetchWithCredentials(input: RequestInfo | URL, init?: RequestInit): Promise<Response> {
	const method = (init?.method ?? 'GET').toUpperCase();
	// ponytail: when openapi-fetch passes a Request object without init (common for
	// PUT/POST with typed headers like If-Match), init?.headers is undefined — but
	// the Request already has the merged headers. Use those as the fallback base so
	// we don't lose headers when creating a new Headers object below.
	const headersBase = init?.headers ?? (input instanceof Request ? input.headers : undefined);
	const headers = new Headers(headersBase);
	if (method === 'GET' || method === 'HEAD') {
		headers.delete('Content-Type');
	}
	const response = await fetch(input, { ...init, headers, credentials: 'include' });
	if (response.status === 401) {
		window.location.href = '/login';
	}
	if (response.status === 404) {
		throw new HttpNotFoundError();
	}
	return response;
}

export const client = createClient<AppPaths>({
	baseUrl: baseURL,
	headers: { 'Content-Type': 'application/json' },
	fetch: fetchWithCredentials
});
