import { client } from './client';
import type { EvaluationProfile } from '$lib/types/evaluation';

export interface AuthUser {
	employeeId: string;
	email: string;
	name: string;
	profileId: EvaluationProfile;
	organizationId: string;
}

const FIXTURE_USER: AuthUser = {
	employeeId: '00000000-0000-0000-0000-000000000001',
	email: 'dev@sed.local',
	name: 'Usuario Desarrollo',
	profileId: 'colaborador',
	organizationId: '00000000-0000-0000-0000-000000000001'
};

// ponytail: set by logout() right before the full reload; consumed by the
// next ensureSession() so we don't fire a redundant /auth/me that we know
// will 400 (cookie is already cleared).
const JUST_LOGGED_OUT_KEY = 'sed-just-logged-out';

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

	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		user = { ...FIXTURE_USER };
		loading = false;
		return;
	}

	try {
		const { data, error: apiError } = await (client as any).GET('/auth/me');
		if (apiError) {
			throw new Error(typeof apiError === 'string' ? apiError : 'Error de autenticación');
		}
		const raw = data as {
			employee: { id: string; email?: string; first_name?: string; last_name?: string };
			role: string;
			profile?: { id?: string; name?: string };
			organization_id?: string;
		};
		user = {
			employeeId: raw.employee.id ?? '',
			email: raw.employee.email ?? '',
			name: [raw.employee.first_name, raw.employee.last_name].filter(Boolean).join(' ') || 'Usuario',
			profileId: raw.role as EvaluationProfile,
			organizationId: raw.organization_id ?? raw.employee.id ?? ''
		};
	} catch (e) {
		error = e instanceof Error ? e.message : 'Error al cargar sesión';
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

	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		// ponytail: hardcoded test profiles for local dev. Replace with SSO when ready.
		const userMap: Record<string, AuthUser> = {
			'fgarcia@mobo.mx': {
				employeeId: '00000000-0000-0000-0000-000000000001',
				email: 'fgarcia@mobo.mx',
				name: 'Fernando García Domínguez',
				profileId: 'director',
				organizationId: '00000000-0000-0000-0000-000000000001'
			},
			'alberto@mobo.mx': {
				employeeId: '00000000-0000-0000-0000-000000000002',
				email: 'alberto@mobo.mx',
				name: 'Alberto Cohen',
				profileId: 'director-general',
				organizationId: '00000000-0000-0000-0000-000000000001'
			},
			'abraham@mobo.mx': {
				employeeId: '00000000-0000-0000-0000-000000000003',
				email: 'abraham@mobo.mx',
				name: 'Abraham Esses Cohen',
				profileId: 'jefe',
				organizationId: '00000000-0000-0000-0000-000000000001'
			},
			'agil@mobo.mx': {
				employeeId: '00000000-0000-0000-0000-000000000004',
				email: 'agil@mobo.mx',
				name: 'Cristiann Gil Ruíz',
				profileId: 'rh',
				organizationId: '00000000-0000-0000-0000-000000000001'
			},
			'fperez@mobo.com.mx': {
				employeeId: '00000000-0000-0000-0000-000000000005',
				email: 'fperez@mobo.com.mx',
				name: 'Frankil Aldair Pérez Rosales',
				profileId: 'colaborador',
				organizationId: '00000000-0000-0000-0000-000000000001'
			}
		};

		const match = userMap[email];
		if (!match) {
			error = 'Usuario demo no válido';
			loading = false;
			throw new Error('Usuario demo no válido');
		}

		user = { ...match };
		loading = false;
		return;
	}

	const { error: apiError } = await (client as any).POST('/auth/dev-login', {
		body: { email }
	});
	if (apiError) {
		error = typeof apiError === 'string' ? apiError : 'Error al iniciar sesión';
		loading = false;
		throw new Error(error);
	}
	await ensureSession();
}

export async function logout(): Promise<void> {
	loggingOut = true;
	try {
		if (!(import.meta.env.DEV && !import.meta.env.VITE_USE_API)) {
			await (client as any).POST('/auth/logout');
		}
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

	// loading = true BEFORE user = null so the layout's auth-guard $effect
	// sees session.loading = true and returns early — it won't fire
	// goto('/login') which would cause a SPA-nav flash of the demo before
	// the full reload takes over.
	loading = true;
	user = null;
	error = null;
	window.location.href = '/login';
}
