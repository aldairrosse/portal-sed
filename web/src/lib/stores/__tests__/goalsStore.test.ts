import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

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

// ─── DEV mode ──────────────────────────────────────────────────────────────────

describe('goalsStore – DEV mode', () => {
	beforeEach(() => {
		vi.resetModules();
		vi.stubEnv('DEV', true);
		vi.stubEnv('VITE_USE_API', '');
	});

	it('load() populates from fixture files when DEV && !VITE_USE_API', async () => {
		const store = await import('../goalsStore.svelte.ts');

		// Initial state before load
		expect(store.storeState.loading).toBe(true);

		await store.load();

		expect(store.storeState.loading).toBe(false);
		expect(store.storeState.error).toBeNull();
		expect(store.getCategories()).toHaveLength(4);
		expect(store.getGoals()).toHaveLength(10);
		expect(store.getKpis()).toHaveLength(6);
		expect(store.getGoalKpiLinks()).toHaveLength(11);
		expect(store.getAssignments()).toHaveLength(10);
		expect(store.getChangeRequests()).toHaveLength(0);
	}, 15000);

	it('does not call client API in DEV mode', async () => {
		const store = await import('../goalsStore.svelte.ts');
		await store.load();

		const cm = await getClientMock();
		expect(cm.GET).not.toHaveBeenCalled();
	});

	it('getCategories() returns fixture categories', async () => {
		const store = await import('../goalsStore.svelte.ts');
		await store.load();

		const cats = store.getCategories();
		expect(cats.find((c: { id: string }) => c.id === 'cat-ventas-finanzas')?.name).toBe(
			'Ventas y resultados financieros'
		);
	});

	it('getGoalsByCategory() filters correctly', async () => {
		const store = await import('../goalsStore.svelte.ts');
		await store.load();

		const goals = store.getGoalsByCategory('cat-ventas-finanzas');
		expect(goals).toHaveLength(3);
		expect(goals[0].name).toBe('Alcanzar meta de ventas mensuales');
	});

	it('getKpisForGoal() returns linked KPIs', async () => {
		const store = await import('../goalsStore.svelte.ts');
		await store.load();

		const kpis = store.getKpisForGoal('goal-alcanzar-ventas');
		expect(kpis.length).toBeGreaterThanOrEqual(1);
		expect(kpis.some((k: { id: string }) => k.id === 'kpi-ventas-mensuales')).toBe(true);
	});

	it('getCategoryProgressAverage() computes average from goals with progress', async () => {
		const store = await import('../goalsStore.svelte.ts');
		await store.load();

		// cat-ventas-finanzas has 3 goals:
		//   - goal-alcanzar-ventas:   unit=moneda     progress=650000   target=1_000_000 → (650000/1_000_000)*100 = 65
		//   - goal-mejorar-margen:    unit=porcentaje progress=18                         → 18 (raw value, not relative)
		//   - goal-reducir-costos:    no progress → excluded from average
		// avg = (65 + 18) / 2 = 41.5
		const avg = store.getCategoryProgressAverage('cat-ventas-finanzas');
		expect(avg).toBeCloseTo(41.5, 0);
	});

	it('getGoalsByCategory() returns empty array for unknown category', async () => {
		const store = await import('../goalsStore.svelte.ts');
		await store.load();

		expect(store.getGoalsByCategory('cat-nonexistent')).toEqual([]);
	});

	it('isAssignmentValid() returns true for fixture data', async () => {
		const store = await import('../goalsStore.svelte.ts');
		await store.load();

		expect(store.isAssignmentValid()).toBe(true);
	});

	it('addCategory() mutates local state without API call', async () => {
		const store = await import('../goalsStore.svelte.ts');
		await store.load();

		const before = store.getCategories().length;
		await store.addCategory({
			id: 'cat-test',
			name: 'Test',
			description: 'Test category',
			weight: 10
		});

		expect(store.getCategories()).toHaveLength(before + 1);
		expect(store.getCategories().find((c: { id: string }) => c.id === 'cat-test')).toBeTruthy();

		const cm = await getClientMock();
		expect(cm.POST).not.toHaveBeenCalled();
	});

	it('updateCategory() mutates local state without API call', async () => {
		const store = await import('../goalsStore.svelte.ts');
		await store.load();

		await store.updateCategory('cat-ventas-finanzas', { name: 'Ventas actualizado' });

		const cat = store.getCategories().find((c: { id: string }) => c.id === 'cat-ventas-finanzas');
		expect(cat?.name).toBe('Ventas actualizado');
	});

	it('deleteCategory() removes category and cascade deletes goals and links', async () => {
		const store = await import('../goalsStore.svelte.ts');
		await store.load();

		await store.deleteCategory('cat-operaciones-procesos');

		expect(store.getCategories().find((c: { id: string }) => c.id === 'cat-operaciones-procesos')).toBeUndefined();
		// Cascade: goals in that category should be gone
		expect(store.getGoalsByCategory('cat-operaciones-procesos')).toHaveLength(0);
	});

	it('getChangeRequests() returns empty until a change is recorded', async () => {
		const store = await import('../goalsStore.svelte.ts');
		await store.load();

		expect(store.getChangeRequests()).toHaveLength(0);
	});

	it('recordChangeRequest() appends to change requests', async () => {
		const store = await import('../goalsStore.svelte.ts');
		await store.load();

		await store.recordChangeRequest({
			id: 'cr-test',
			entityType: 'goal',
			entityId: 'goal-api-1',
			action: 'create',
			changes: {},
			reason: 'Test',
			requestedBy: 'tester',
			requestedAt: new Date().toISOString(),
			status: 'pending'
		});

		expect(store.getChangeRequests()).toHaveLength(1);
		expect(store.getChangeRequests()[0].id).toBe('cr-test');
	});
});

// ─── API mode ──────────────────────────────────────────────────────────────────

describe('goalsStore – API mode', () => {
	beforeEach(() => {
		vi.resetModules();
		vi.clearAllMocks();
		vi.stubEnv('DEV', false);
		vi.stubEnv('VITE_USE_API', 'true');
	});

	it('load() calls three GET endpoints and populates store from API response', async () => {
		const cm = await getClientMock();
		cm.GET.mockImplementation((url: string) => {
			if (url.includes('categories')) {
				return Promise.resolve(okGetResponse(SAMPLE_API_CATEGORIES));
			}
			if (url.includes('kpis')) {
				return Promise.resolve(okGetResponse(SAMPLE_API_KPIS));
			}
			// assignments endpoint returns a single object, not { items }
			return Promise.resolve({ data: SAMPLE_API_ASSIGNMENT, error: null });
		});

		const store = await import('../goalsStore.svelte.ts');
		await store.load();

		expect(store.storeState.loading).toBe(false);
		expect(store.storeState.error).toBeNull();
		expect(cm.GET).toHaveBeenCalledTimes(3);
		expect(store.getCategories()).toHaveLength(2);
		expect(store.getGoals()).toHaveLength(1);
	});

	it('addCategory() calls POST then reloads (three GETs again)', async () => {
		const cm = await getClientMock();
		// Set up GET responses for the reload that happens after POST
		cm.GET.mockImplementation((url: string) => {
			if (url.includes('categories')) {
				return Promise.resolve(okGetResponse(SAMPLE_API_CATEGORIES));
			}
			if (url.includes('kpis')) {
				return Promise.resolve(okGetResponse(SAMPLE_API_KPIS));
			}
			return Promise.resolve({ data: SAMPLE_API_ASSIGNMENT, error: null });
		});
		cm.POST.mockResolvedValue({ error: null });

		const store = await import('../goalsStore.svelte.ts');
		await store.addCategory({
			id: 'cat-new',
			name: 'Nueva',
			description: 'Nueva categoría',
			weight: 15
		});

		// POST called once
		expect(cm.POST).toHaveBeenCalledTimes(1);
		// Reload triggers GET again: 3 calls per reload
		expect(cm.GET).toHaveBeenCalledTimes(3);
	});

	it('handles API error in load() gracefully', async () => {
		const cm = await getClientMock();
		cm.GET.mockResolvedValue({
			data: null,
			error: { error: { message: 'Error al cargar categorías' } }
		});

		const store = await import('../goalsStore.svelte.ts');
		await store.load();

		expect(store.storeState.loading).toBe(false);
		expect(store.storeState.error).toBe('Error al cargar categorías');
		expect(store.getCategories()).toEqual([]);
	});

	it('handles empty API responses', async () => {
		const cm = await getClientMock();
		cm.GET.mockImplementation((url: string) => {
			if (url.includes('assignments')) {
				return Promise.resolve({ data: null, error: null });
			}
			return Promise.resolve(okGetResponse([]));
		});

		const store = await import('../goalsStore.svelte.ts');
		await store.load();

		expect(store.storeState.loading).toBe(false);
		expect(store.storeState.error).toBeNull();
		expect(store.getCategories()).toEqual([]);
		expect(store.getGoals()).toEqual([]);
		expect(store.getKpis()).toEqual([]);
		expect(store.getAssignments()).toEqual([]);
	});
});
