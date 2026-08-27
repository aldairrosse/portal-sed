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
		try {
			const cloned = response.clone();
			const body = await cloned.json();
			// Only a fatal session-expiry 401 (refresh token permanently invalid)
			// forces re-login. Transient failures (5xx or other codes) fall
			// through so the caller can surface the error without losing the session.
			if (body?.error?.code === 'SSO_SESSION_EXPIRED') {
				// Allow unauthenticated dev view: bypass redirect when dev mode enabled
				try {
					const devRes = await fetch('/api/v1/dev/status', { credentials: 'include' });
					if (devRes.ok) return response;
				} catch {
					// dev probe failed — fall through to redirect
				}
				console.warn('[sso] 401 SSO_SESSION_EXPIRED -> redirect /login', window.location.pathname);
				sessionStorage.setItem('return_to', window.location.pathname + window.location.search);
				sessionStorage.removeItem('sso_redirect_count');
				sessionStorage.setItem('sso_session_expired', '1');
				window.location.href = '/login';
			}
		} catch {
			// Unparseable 401 body — fall through, caller handles the error
		}
	}
	if (response.status === 404) {
		throw new HttpNotFoundError();
	}
	if (response.status === 403) {
		try {
			const cloned = response.clone();
			const body = await cloned.json();
			if (body?.error?.code === 'OTP_REQUIRED') {
				console.warn('[sso] 403 OTP_REQUIRED -> redirect step-up', window.location.pathname);
				window.location.href = '/api/v1/auth/sso-step-up?return_to=' + encodeURIComponent(window.location.pathname + window.location.search);
				throw new Error('OTP_REQUIRED');
			}
		} catch (e) {
			if (e instanceof Error && e.message === 'OTP_REQUIRED') throw e;
		}
	}
	return response;
}

export const client = createClient<AppPaths>({
	baseUrl: baseURL,
	headers: { 'Content-Type': 'application/json' },
	fetch: fetchWithCredentials
});
