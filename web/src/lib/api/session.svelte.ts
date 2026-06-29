import { client } from './client';
import { goto } from '$app/navigation';
import { humanizeError } from '$lib/utils/error';
import type { EvaluationProfile } from '$lib/types/evaluation';

export interface AuthUser {
	employeeId: string;
	email: string;
	name: string;
	profileId: EvaluationProfile;
	profileName: string;
	organizationId: string;
	jobTitle: string;
	orgNodeId: string;
	orgNodeName: string;
}


// ponytail: set by logout() right before the full reload; consumed by the
// next ensureSession() so we don't fire a redundant /auth/me that we know
// will 400 (cookie is already cleared).
const JUST_LOGGED_OUT_KEY = 'sed-just-logged-out';

// ponytail: hold the loader for at least this long on successful login so
// fast responses (local fixture, localhost API) don't flash. Errors bail out
// before the wait so failures surface immediately.
export const MIN_LOGIN_MS = 400;

let user = $state<AuthUser | null>(null);
let loading = $state(true);
let error = $state<string | null>(null);
let loggingOut = $state(false);

export async function ensureSession(): Promise<void> {
	loading = true;
	error = null;

	// We just logged out — the cookie is gone, no need to call /me
	if (sessionStorage.getItem(JUST_LOGGED_OUT_KEY)) {
		sessionStorage.removeItem(JUST_LOGGED_OUT_KEY);
		user = null;
		loading = false;
		return;
	}

	try {
		const { data, error: apiError } = await (client as any).GET('/auth/me');
		if (apiError) {
			throw apiError;
		}
		const raw = data as {
			employee: { id: string; email?: string; first_name?: string; last_name?: string; job_title?: string; org_node_id?: string; org_node_name?: string };
			role: string;
			profile?: { id?: string; name?: string };
			organization_id?: string;
		};
		user = {
			employeeId: raw.employee.id ?? '',
			email: raw.employee.email ?? '',
			name: [raw.employee.first_name, raw.employee.last_name].filter(Boolean).join(' ') || 'Usuario',
			profileId: raw.role as EvaluationProfile,
			profileName: raw.profile?.name ?? raw.role ?? '',
			organizationId: raw.organization_id ?? raw.employee.id ?? '',
			jobTitle: raw.employee.job_title ?? '',
			orgNodeId: raw.employee.org_node_id ?? '',
			orgNodeName: raw.employee.org_node_name ?? ''
		};
	} catch (e) {
		error = humanizeError(e, 'Error al cargar sesión');
		user = null;
	} finally {
		loading = false;
	}
}

export function getSession(): {
	user: AuthUser | null;
	loading: boolean;
	error: string | null;
	loggingOut: boolean;
} {
	return { user, loading, error, loggingOut };
}

export async function devLogin(email: string): Promise<void> {
	loading = true;
	error = null;
	const start = Date.now();

	try {
		const { error: apiError } = await (client as any).POST('/auth/dev-login', {
			body: { email }
		});
		if (apiError) {
			error = humanizeError(apiError, 'Error al iniciar sesión');
			throw new Error(error);
		}
		// ponytail: wait BEFORE ensureSession so user stays null during the
		// hold — same reason as the fixture path above. ensureSession is what
		// sets user, so doing it after the wait keeps the loader on the login
		// page (with text) instead of swapping to the layout's bare spinner.
		const elapsed = Date.now() - start;
		if (elapsed < MIN_LOGIN_MS) {
			await new Promise((r) => setTimeout(r, MIN_LOGIN_MS - elapsed));
		}
		// ponytail: logout sets JUST_LOGGED_OUT_KEY to skip a redundant /auth/me
		// after the next full-reload — but we just POSTed a fresh login and need
		// ensureSession to actually fetch /auth/me and populate the user.
		sessionStorage.removeItem(JUST_LOGGED_OUT_KEY);
		await ensureSession();
	} catch (e) {
		// ponytail: fetch throws TypeError on network failure (Failed to fetch, etc.)
		// — the apiError branch above already humanized API errors, so guard with error===null
		if (error === null) {
			error = humanizeError(e, 'Error al iniciar sesión');
			throw new Error(error);
		}
		throw e;
	} finally {
		// ponytail: finally guarantees loading is reset no matter how the request dies
		// (success, API error, network failure, timeout) so the UI can recover.
		loading = false;
	}
}

export async function logout(): Promise<void> {
	const start = Date.now();
	loggingOut = true;
	try {
		await (client as any).POST('/auth/logout');
	} catch {
		// Even if backend fails, clear local state
	}

	// ponytail: dev-only state persists in storage across reloads — wipe it on logout
	try {
		sessionStorage.removeItem('sed-dev-context');
		localStorage.removeItem('sed-dev-comentarios');
		sessionStorage.setItem(JUST_LOGGED_OUT_KEY, '1');
	} catch {
		// storage may be unavailable (private mode, etc.)
	}

	// ponytail: hold the "Cerrando sesión" loader for at least MIN_LOGIN_MS so
	// fast responses don't flash the same way login did before the fix.
	const elapsed = Date.now() - start;
	if (elapsed < MIN_LOGIN_MS) {
		await new Promise((r) => setTimeout(r, MIN_LOGIN_MS - elapsed));
	}

	// ponytail: loading=true BEFORE user=null so the layout's auth-guard
	// $effect sees session.loading=true and returns early mid-navigation —
	// otherwise it would fire goto('/login') on its own and cause a SPA-nav
	// flash of the previous route before we land on /login.
	//
	// After goto lands, reset loading so the login page can render its
	// actual content (demo list / SSO message) instead of a stuck loader.
	// The layout's onMount doesn't re-fire on SPA nav, so ensureSession()
	// is not called automatically — that's why we flip loading here.
	loading = true;
	user = null;
	error = null;
	await goto('/login');
	loading = false;

	// ponytail: reset loggingOut so the next login's Sidebar doesn't render
	// "Cerrando sesión…" forever (the flag was only meant for the in-flight
	// logout, not the post-navigation state).
	loggingOut = false;
}
