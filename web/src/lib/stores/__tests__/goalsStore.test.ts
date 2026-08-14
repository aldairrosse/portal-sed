import { describe, it, expect, vi, beforeEach } from 'vitest';

// ─── Mock dependencies ─────────────────────────────────────────────────────────
// These are hoisted by Vitest and apply before any import.

vi.mock('../../api/client', () => ({
	client: {
		GET: vi.fn(),
		POST: vi.fn(),
		PUT: vi.fn(),
		DELETE: vi.fn(),
		PATCH: vi.fn()
	}
}));

vi.mock('../../api/session.svelte', () => ({
	getSession: () => ({ user: { employeeId: 'emp-test-01' }, loading: false, error: null })
}));

vi.mock('../../api/cycle.svelte', () => ({
	getActivePhase: () => 'inicio-anio' as const
}));

// ─── Test-scoped helpers ───────────────────────────────────────────────────────

type MockClient = {
	GET: ReturnType<typeof vi.fn>;
	POST: ReturnType<typeof vi.fn>;
	PUT: ReturnType<typeof vi.fn>;
	DELETE: ReturnType<typeof vi.fn>;
	PATCH: ReturnType<typeof vi.fn>;
};

async function getClientMock(): Promise<MockClient> {
	const mod = await import('../../api/client');
	return mod.client as unknown as MockClient;
}

/**
 * Build a standard success response matching what openapi-fetch client.GET returns.
 */
function okGetResponse(items: unknown[]) {
	return { data: { items }, error: null };
}

/**
 * Build a category with goals for API-mode response shape.
 */
type ApiCategory = {
	id: string;
	name: string;
	description: string;
	weight: number;
	goals?: Array<{
		id: string;
		name: string;
		description: string;
		unit: string;
		weight: number;
		target_value: number;
		current_value?: number;
		kpis?: Array<{ id: string; name?: string; unit?: string }>;
	}>;
};

/**
 * Sample API categories to inject into mock responses.
 */
const SAMPLE_API_CATEGORIES: ApiCategory[] = [
	{
		id: 'cat-api-1',
		name: 'Ventas',
		description: 'Objetivos de ventas',
		weight: 60,
		goals: [
			{
				id: 'goal-api-1',
				name: 'Alcanzar cuota',
				description: 'Cumplir cuota mensual',
				unit: 'moneda',
				weight: 100,
				target_value: 500000,
				current_value: 300000,
				kpis: [{ id: 'kpi-api-1', name: 'Ventas mensuales', unit: 'moneda' }]
			}
		]
	},
	{
		id: 'cat-api-2',
		name: 'Calidad',
		description: 'Objetivos de calidad',
		weight: 40,
		goals: []
	}
];

const SAMPLE_API_KPIS = [
	{
		id: 'kpi-api-1',
		name: 'Ventas mensuales',
		unit: 'moneda',
		description: 'Ingresos del mes'
	}
];

const SAMPLE_API_ASSIGNMENT = {
	id: 'asign-api-1',
	employee_id: 'emp-test-01',
	cycle_id: 'cycle-2026',
	created_at: '2026-01-01T00:00:00Z'
};

/**
 * Route every GET the store can issue to a realistic response.
 * A full load() is: 3 main endpoints + 1 comment GET per goal + 1 per category.
 */
function mockApiRoutes(cm: MockClient) {
	cm.GET.mockImplementation((url: string) => {
		if (url.includes('/comments')) {
			return Promise.resolve({ data: [], error: null });
		}
		if (url.includes('categories')) {
			return Promise.resolve(okGetResponse(SAMPLE_API_CATEGORIES));
		}
		if (url.includes('kpis')) {
			return Promise.resolve(okGetResponse(SAMPLE_API_KPIS));
		}
		// assignments endpoint returns a single object, not { items }
		return Promise.resolve({ data: SAMPLE_API_ASSIGNMENT, error: null });
	});
}

// ─── Store contract (API-only) ─────────────────────────────────────────────────

describe('goalsStore – API contract', () => {
	beforeEach(() => {
		vi.resetModules();
		vi.clearAllMocks();
	});

	it('load() populates store from API responses', async () => {
		const cm = await getClientMock();
		mockApiRoutes(cm);
		const store = await import('../goalsStore.svelte');

		// Initial state before load
		expect(store.storeState.loading).toBe(true);

		await store.load();

		expect(store.storeState.loading).toBe(false);
		expect(store.storeState.error).toBeNull();
		expect(store.getCategories()).toHaveLength(2);
		expect(store.getGoals()).toHaveLength(1);
		expect(store.getKpis()).toHaveLength(1);
		expect(store.getGoalKpiLinks()).toHaveLength(1);
		expect(store.getAssignments()).toHaveLength(1);
		expect(store.getChangeRequests()).toHaveLength(0);
	}, 15000);

	it('load() fetches the main endpoints plus one comments GET per goal/category', async () => {
		const cm = await getClientMock();
		mockApiRoutes(cm);
		const store = await import('../goalsStore.svelte');
		await store.load();

		// 3 main endpoints + 1 goal comment (goal-api-1) + 2 category comments
		expect(cm.GET).toHaveBeenCalledTimes(6);
		expect(cm.GET).toHaveBeenCalledWith(
			'/employees/{empId}/categories',
			expect.objectContaining({ params: { path: { empId: 'emp-test-01' } } })
		);
		expect(cm.GET).toHaveBeenCalledWith('/kpis', {});
		expect(cm.GET).toHaveBeenCalledWith(
			'/goals/{goalId}/comments',
			expect.objectContaining({ params: { path: { goalId: 'goal-api-1' } } })
		);
	});

	it('getCategories() returns normalized categories', async () => {
		const cm = await getClientMock();
		mockApiRoutes(cm);
		const store = await import('../goalsStore.svelte');
		await store.load();

		const cats = store.getCategories();
		expect(cats.find((c: { id: string }) => c.id === 'cat-api-1')?.name).toBe('Ventas');
	});

	it('getGoalsByCategory() filters correctly', async () => {
		const cm = await getClientMock();
		mockApiRoutes(cm);
		const store = await import('../goalsStore.svelte');
		await store.load();

		const goals = store.getGoalsByCategory('cat-api-1');
		expect(goals).toHaveLength(1);
		expect(goals[0].name).toBe('Alcanzar cuota');
	});

	it('getKpisForGoal() returns linked KPIs', async () => {
		const cm = await getClientMock();
		mockApiRoutes(cm);
		const store = await import('../goalsStore.svelte');
		await store.load();

		const kpis = store.getKpisForGoal('goal-api-1');
		expect(kpis.length).toBeGreaterThanOrEqual(1);
		expect(kpis.some((k: { id: string }) => k.id === 'kpi-api-1')).toBe(true);
	});

	it('getCategoryProgressAverage() computes direction-aware average from goals with progress', async () => {
		const cm = await getClientMock();
		mockApiRoutes(cm);
		const store = await import('../goalsStore.svelte');
		await store.load();

		// cat-api-1 has 1 goal: goal-api-1 ascendente progress=300000 target=500000 → 60
		const avg = store.getCategoryProgressAverage('cat-api-1');
		expect(avg).toBeCloseTo(60, 0);
	});

	it('getWeightedScore() calculates total weighted score across all categories', async () => {
		const cm = await getClientMock();
		mockApiRoutes(cm);
		const store = await import('../goalsStore.svelte');
		await store.load();

		// cat-api-1 (w=60): goal-api-1 (w=100): (100/100)*60 = 60 → 60*(60/100) = 36
		const score = store.getWeightedScore();
		expect(score).toBeCloseTo(36, 0);
	});

	it('getGoalsByCategory() returns empty array for unknown category', async () => {
		const cm = await getClientMock();
		mockApiRoutes(cm);
		const store = await import('../goalsStore.svelte');
		await store.load();

		expect(store.getGoalsByCategory('cat-nonexistent')).toEqual([]);
	});

	it('isAssignmentValid() returns true for valid API data', async () => {
		const cm = await getClientMock();
		mockApiRoutes(cm);
		const store = await import('../goalsStore.svelte');
		await store.load();

		expect(store.isAssignmentValid()).toBe(true);
	});

	it('isAssignmentValid() returns false when category weights do not sum to 100', async () => {
		const cm = await getClientMock();
		cm.GET.mockImplementation((url: string) => {
			if (url.includes('/comments')) {
				return Promise.resolve({ data: [], error: null });
			}
			if (url.includes('categories')) {
				return Promise.resolve(okGetResponse([{ ...SAMPLE_API_CATEGORIES[0], weight: 30 }]));
			}
			if (url.includes('kpis')) {
				return Promise.resolve(okGetResponse(SAMPLE_API_KPIS));
			}
			return Promise.resolve({ data: SAMPLE_API_ASSIGNMENT, error: null });
		});

		const store = await import('../goalsStore.svelte');
		await store.load();

		expect(store.isAssignmentValid()).toBe(false);
	});

	it('getAssignmentByEmployee() returns the normalized assignment', async () => {
		const cm = await getClientMock();
		mockApiRoutes(cm);
		const store = await import('../goalsStore.svelte');
		await store.load();

		const assignment = store.getAssignmentByEmployee('emp-test-01');
		expect(assignment?.id).toBe('asign-api-1');
		expect(assignment?.goalIds).toEqual(['goal-api-1']);
	});

	it('load() maps goal comments into getGoalComments()', async () => {
		const cm = await getClientMock();
		cm.GET.mockImplementation((url: string) => {
			if (url.includes('/comments')) {
				return Promise.resolve({
					data: [
						{ id: 'cm-1', author_id: 'a1', author_name: 'Ana', content: 'Hola', created_at: '2026-01-01T00:00:00Z' }
					],
					error: null
				});
			}
			if (url.includes('categories')) {
				return Promise.resolve(okGetResponse(SAMPLE_API_CATEGORIES));
			}
			if (url.includes('kpis')) {
				return Promise.resolve(okGetResponse(SAMPLE_API_KPIS));
			}
			return Promise.resolve({ data: SAMPLE_API_ASSIGNMENT, error: null });
		});

		const store = await import('../goalsStore.svelte');
		await store.load();

		expect(store.getGoalComments('goal-api-1')).toHaveLength(1);
		expect(store.getGoalComments('goal-api-1')[0].content).toBe('Hola');
	});

	it('addCategory() calls POST then reloads', async () => {
		const cm = await getClientMock();
		mockApiRoutes(cm);
		cm.POST.mockResolvedValue({ error: null });

		const store = await import('../goalsStore.svelte');
		await store.addCategory({
			id: 'cat-new',
			name: 'Nueva',
			description: 'Nueva categoría',
			weight: 15
		});

		expect(cm.POST).toHaveBeenCalledTimes(1);
		// Reload triggers a full load again: 6 GETs
		expect(cm.GET).toHaveBeenCalledTimes(6);
	});

	it('updateCategory() calls PUT then reloads', async () => {
		const cm = await getClientMock();
		mockApiRoutes(cm);
		cm.PUT.mockResolvedValue({ error: null });

		const store = await import('../goalsStore.svelte');
		await store.updateCategory('cat-api-1', { name: 'Ventas actualizado' });

		expect(cm.PUT).toHaveBeenCalledTimes(1);
		expect(cm.GET).toHaveBeenCalledTimes(6);
	});

	it('deleteCategory() calls DELETE then reloads', async () => {
		const cm = await getClientMock();
		mockApiRoutes(cm);
		cm.DELETE.mockResolvedValue({ error: null });

		const store = await import('../goalsStore.svelte');
		await store.deleteCategory('cat-api-1');

		expect(cm.DELETE).toHaveBeenCalledTimes(1);
		expect(cm.GET).toHaveBeenCalledTimes(6);
	});

	it('recordChangeRequest() POSTs and appends the returned change request', async () => {
		const cm = await getClientMock();
		mockApiRoutes(cm);
		const cr = {
			id: 'cr-test',
			entityType: 'goal' as const,
			entityId: 'goal-api-1',
			action: 'create' as const,
			changes: {},
			reason: 'Test',
			requestedBy: 'tester',
			requestedAt: new Date().toISOString(),
			status: 'pending' as const
		};
		cm.POST.mockResolvedValue({ data: cr, error: null });

		const store = await import('../goalsStore.svelte');
		await store.load();
		await store.recordChangeRequest(cr);

		expect(store.getChangeRequests()).toHaveLength(1);
		expect(store.getChangeRequests()[0].id).toBe('cr-test');
	});
});

// ─── Error / empty handling ────────────────────────────────────────────────────

describe('goalsStore – error handling', () => {
	beforeEach(() => {
		vi.resetModules();
		vi.clearAllMocks();
	});

	it('handles API error in load() gracefully', async () => {
		const cm = await getClientMock();
		cm.GET.mockResolvedValue({
			data: null,
			error: { error: { message: 'Error al cargar categorías' } }
		});

		const store = await import('../goalsStore.svelte');
		await store.load();

		expect(store.storeState.loading).toBe(false);
		expect(store.storeState.error).toBe('Error al cargar categorías');
		expect(store.getCategories()).toEqual([]);
	});

	it('handles empty API responses', async () => {
		const cm = await getClientMock();
		cm.GET.mockImplementation((url: string) => {
			if (url.includes('/comments')) {
				return Promise.resolve({ data: [], error: null });
			}
			if (url.includes('assignments')) {
				return Promise.resolve({ data: null, error: null });
			}
			return Promise.resolve(okGetResponse([]));
		});

		const store = await import('../goalsStore.svelte');
		await store.load();

		expect(store.storeState.loading).toBe(false);
		expect(store.storeState.error).toBeNull();
		expect(store.getCategories()).toEqual([]);
		expect(store.getGoals()).toEqual([]);
		expect(store.getKpis()).toEqual([]);
		expect(store.getAssignments()).toEqual([]);
	});
});
