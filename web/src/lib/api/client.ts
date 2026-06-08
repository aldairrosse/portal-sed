import createClient from 'openapi-fetch';
import type { paths as AuthPaths } from './schemas/auth.d.ts';
import type { paths as CyclePaths } from './schemas/cycle.d.ts';

type AppPaths = AuthPaths & CyclePaths;

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
