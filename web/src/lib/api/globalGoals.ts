// API client for global goals

const API_BASE = '/api/v1';

export interface GlobalGoal {
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
  assignments: GlobalAssignment[];
  rules: GlobalRule[];
}

export interface GlobalAssignment {
  id: string;
  goal_id: string;
  employee_id: string;
  weight: number;
  target_value: number;
  baseline_value?: number;
}

export interface GlobalRule {
  id: string;
  goal_id: string;
  rule_type: string;
  department_id?: string;
  min_direct_reports?: number;
  default_weight: number;
}

export interface CreateGlobalGoalRequest {
  name: string;
  description: string;
  unit: string;
  direction: string;
  goal_kind: string;
  weight: number;
  target_value: number;
  assignments?: CreateAssignmentRequest[];
  rules?: CreateRuleRequest[];
}

export interface CreateAssignmentRequest {
  employee_id: string;
  weight: number;
  target_value: number;
  baseline_value?: number;
}

export interface CreateRuleRequest {
  rule_type: string;
  department_id?: string;
  min_direct_reports?: number;
  default_weight: number;
}

export interface UpdateGlobalGoalRequest {
  name: string;
  description: string;
  unit: string;
  direction: string;
  goal_kind: string;
  weight: number;
  target_value: number;
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

export async function listGlobalGoals(cycleId?: string): Promise<GlobalGoal[]> {
  const params = new URLSearchParams();
  if (cycleId) params.set('cycleId', cycleId);
  const url = `${API_BASE}/goals/global${params.toString() ? '?' + params.toString() : ''}`;
  return fetchJSON<GlobalGoal[]>(url);
}

export async function getGlobalGoal(goalId: string): Promise<GlobalGoal> {
  return fetchJSON<GlobalGoal>(`${API_BASE}/goals/global/${goalId}`);
}

export async function createGlobalGoal(request: CreateGlobalGoalRequest): Promise<GlobalGoal> {
  return fetchJSON<GlobalGoal>(`${API_BASE}/goals/global`, {
    method: 'POST',
    body: JSON.stringify(request),
  });
}

export async function updateGlobalGoal(goalId: string, request: UpdateGlobalGoalRequest): Promise<GlobalGoal> {
  return fetchJSON<GlobalGoal>(`${API_BASE}/goals/global/${goalId}`, {
    method: 'PUT',
    body: JSON.stringify(request),
  });
}

export async function deleteGlobalGoal(goalId: string): Promise<void> {
  return fetchJSON<void>(`${API_BASE}/goals/global/${goalId}`, {
    method: 'DELETE',
  });
}

export async function executeRules(goalId: string): Promise<{ assignments_created: number }> {
  return fetchJSON<{ assignments_created: number }>(`${API_BASE}/goals/global/${goalId}/execute-rules`, {
    method: 'POST',
  });
}
