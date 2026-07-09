import { client } from '$lib/api/client';

// ─── Types ──────────────────────────────────────────────────────────────────

export interface TeamMember {
	id: string;
	firstName: string;
	lastName: string;
	employeeNumber: string;
	orgNodeId: string;
}

// ─── State ──────────────────────────────────────────────────────────────────

let members = $state<TeamMember[]>([]);
let loading = $state(false);
let loadedForId = $state('');

// ─── Load ───────────────────────────────────────────────────────────────────

export async function loadTeam(empId: string): Promise<void> {
	if (!empId || loadedForId === empId) return;
	loading = true;
	try {
		const res = await client.GET('/employees/{empId}/team', {
			params: { path: { empId } }
		});
		const data = (res.data as { data?: Array<TeamMember> })?.data;
		if (data) {
			members = data;
			loadedForId = empId;
		}
	} catch {
		members = [];
	} finally {
		loading = false;
	}
}

// ─── Getters ────────────────────────────────────────────────────────────────

export function getTeamMembers(): TeamMember[] {
	return members;
}

export function isLoading(): boolean {
	return loading;
}
