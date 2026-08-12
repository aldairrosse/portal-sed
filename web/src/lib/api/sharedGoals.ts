// API client for shared goals

const API_BASE = '/api/v1';

export interface SharedGoal {
  id: string;
  name: string;
  description: string;
  unit: string;
  direction: string;
  weight: number;
  target_value: number;
  goal_kind: string;
  state: string;
  created_by: string;
  created_at: string;
  updated_at: string;
  group?: SharedGroup;
  members: SharedMember[];
}

export interface SharedGroup {
  id: string;
  goal_id: string;
  created_by: string;
  name: string;
  description: string;
}

export interface SharedMember {
  id: string;
  group_id: string;
  employee_id: string;
  weight: number;
  target_value: number;
  baseline_value?: number;
}

export interface CreateSharedGoalRequest {
  name: string;
  description: string;
  unit: string;
  direction: string;
  goal_kind: string;
  weight: number;
  target_value: number;
  group_name: string;
  group_description: string;
  members: CreateMemberRequest[];
}

export interface CreateMemberRequest {
  employee_id: string;
  weight: number;
  target_value: number;
  baseline_value?: number;
}

export interface UpdateSharedGoalRequest {
  name: string;
  description: string;
  unit: string;
  direction: string;
  goal_kind: string;
  weight: number;
  target_value: number;
}

export interface AddMemberRequest {
  employee_id: string;
  weight: number;
  target_value: number;
  baseline_value?: number;
}

export interface UpdateProgressRequest {
  current_value: number;
}

async function fetchJSON<T>(url: string, options?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    ...options,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Unknown error' }));
    throw new Error(error.error || `HTTP ${response.status}`);
  }

  if (response.status === 204) {
    return null as T;
  }

  return response.json();
}

export async function listSharedGoals(view?: 'creator' | 'member'): Promise<SharedGoal[]> {
  const params = new URLSearchParams();
  if (view) params.set('view', view);
  const url = `${API_BASE}/goals/shared${params.toString() ? '?' + params.toString() : ''}`;
  return fetchJSON<SharedGoal[]>(url);
}

export async function getSharedGoal(goalId: string): Promise<SharedGoal> {
  return fetchJSON<SharedGoal>(`${API_BASE}/goals/shared/${goalId}`);
}

export async function createSharedGoal(request: CreateSharedGoalRequest): Promise<SharedGoal> {
  return fetchJSON<SharedGoal>(`${API_BASE}/goals/shared`, {
    method: 'POST',
    body: JSON.stringify(request),
  });
}

export async function updateSharedGoal(goalId: string, request: UpdateSharedGoalRequest): Promise<SharedGoal> {
  return fetchJSON<SharedGoal>(`${API_BASE}/goals/shared/${goalId}`, {
    method: 'PUT',
    body: JSON.stringify(request),
  });
}

export async function deleteSharedGoal(goalId: string): Promise<void> {
  return fetchJSON<void>(`${API_BASE}/goals/shared/${goalId}`, {
    method: 'DELETE',
  });
}

export async function addMember(goalId: string, request: AddMemberRequest): Promise<SharedMember> {
  return fetchJSON<SharedMember>(`${API_BASE}/goals/shared/${goalId}/members`, {
    method: 'POST',
    body: JSON.stringify(request),
  });
}

export async function removeMember(goalId: string, employeeId: string): Promise<void> {
  return fetchJSON<void>(`${API_BASE}/goals/shared/${goalId}/members/${employeeId}`, {
    method: 'DELETE',
  });
}

export async function updateProgress(goalId: string, employeeId: string, request: UpdateProgressRequest): Promise<{ status: string }> {
  return fetchJSON<{ status: string }>(`${API_BASE}/goals/shared/${goalId}/progress/${employeeId}`, {
    method: 'PUT',
    body: JSON.stringify(request),
  });
}
