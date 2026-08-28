const API_BASE = '/api/v1';

export interface CycleWeightConfig {
	g_weight: number;
	p_weight: number;
}

export interface TeamWeightConfig {
	j_weight: number;
	pj_weight: number;
}

async function fetchJSON<T>(url: string, options?: RequestInit): Promise<T> {
	const res = await fetch(url, { credentials: 'include', ...options, headers: { 'Content-Type': 'application/json', ...(options?.headers ?? {}) } });
	if (!res.ok) {
		const body = await res.json().catch(() => ({ error: res.statusText }));
		const msg = (typeof body.error === 'string' ? body.error : body.error?.message) || body.message || `HTTP ${res.status}`;
		throw new Error(msg);
	}
	if (res.status === 204) return null as T;
	return res.json();
}

export function getCycleWeightConfig(): Promise<CycleWeightConfig> {
	return fetchJSON<CycleWeightConfig>(`${API_BASE}/weights/cycle-config`);
}

export function saveCycleWeightConfig(g: number): Promise<CycleWeightConfig> {
	const clamped = Math.min(100, Math.max(0, g));
	return fetchJSON<CycleWeightConfig>(`${API_BASE}/weights/cycle-config`, {
		method: 'PUT',
		body: JSON.stringify({ g_weight: clamped })
	});
}

export function getTeamWeightConfig(teamId?: string): Promise<TeamWeightConfig> {
	const q = teamId ? `?team_id=${encodeURIComponent(teamId)}` : '';
	return fetchJSON<TeamWeightConfig>(`${API_BASE}/weights/team-config${q}`);
}

export function saveTeamWeightConfig(j: number, teamId?: string): Promise<TeamWeightConfig> {
	const clamped = Math.min(100, Math.max(0, j));
	const q = teamId ? `?team_id=${encodeURIComponent(teamId)}` : '';
	return fetchJSON<TeamWeightConfig>(`${API_BASE}/weights/team-config${q}`, {
		method: 'PUT',
		body: JSON.stringify({ j_weight: clamped })
	});
}
