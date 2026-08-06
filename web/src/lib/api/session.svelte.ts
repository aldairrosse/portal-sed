import { client } from './client';
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

export const MIN_LOGIN_MS = 400;

let user = $state<AuthUser | null>(null);
let loading = $state(true);
let error = $state<string | null>(null);
const loggingOut = $state(false);

export async function ensureSession(): Promise<void> {
	loading = true;
	error = null;

	try {
		const { data, error: apiError } = await client.GET('/auth/me');
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

export async function logout(): Promise<void> {
	window.location.href = '/api/v1/auth/logout';
}
