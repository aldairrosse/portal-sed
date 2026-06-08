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

let user = $state<AuthUser | null>(null);
let loading = $state(true);
let error = $state<string | null>(null);

export async function ensureSession(): Promise<void> {
	loading = true;
	error = null;

	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		user = { ...FIXTURE_USER };
		loading = false;
		return;
	}

	try {
		const { data, error: apiError } = await client.GET('/auth/me' as never);
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

export function getSession(): { user: AuthUser | null; loading: boolean; error: string | null } {
	return { user, loading, error };
}
