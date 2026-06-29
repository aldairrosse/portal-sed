import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import type { AreaMetrics } from '../rhHierarchyStore.svelte';

// ─── Mock dependencies (relative paths required for vi.mock resolution) ────────

vi.mock('../../api/client', () => ({
	client: {
		GET: vi.fn()
	}
}));

vi.mock('../cycleStore.svelte', () => ({
	getActiveCycle: vi.fn(() => null)
}));

// ─── Helper types ───────────────────────────────────────────────────────────────

type MockClient = {
	GET: ReturnType<typeof vi.fn>;
};

async function getMockClient(): Promise<MockClient> {
	const mod = await import('../../api/client');
	return mod.client as unknown as MockClient;
}

async function mockGetActiveCycle(): Promise<ReturnType<typeof vi.fn>> {
	const mod = await import('../cycleStore.svelte');
	return mod.getActiveCycle as unknown as ReturnType<typeof vi.fn>;
}

// ─── Fixture helpers ────────────────────────────────────────────────────────────

function makeAreaMetrics(overrides?: Partial<AreaMetrics>): AreaMetrics {
	return {
		nodeId: 'node-1',
		employeeCount: 2,
		employeesWithGoals: 1,
		avgProgress: 75,
		completedGoals: 3,
		pendingGoals: 1,
		avgRating: 4.2,
		ratingsCount: 5,
		employees: [
			{
				id: 'emp-1',
				firstName: 'Alice',
				lastName: 'Smith',
				jobTitle: 'Developer',
				profileId: 'colaborador',
				profileDescription: 'Colaborador'
			},
			{
				id: 'emp-2',
				firstName: 'Bob',
				lastName: 'Jones',
				jobTitle: 'Designer',
				profileId: 'colaborador',
				profileDescription: 'Colaborador'
			}
		],
		...overrides
	};
}

// ─── Tests ──────────────────────────────────────────────────────────────────────

describe('rhHierarchyStore', () => {
	beforeEach(() => {
		vi.clearAllMocks();
		vi.resetModules();
	});

	afterEach(() => {
		vi.restoreAllMocks();
	});

	describe('initial state', () => {
		it('returns null for getMetrics() before selection', async () => {
			const store = await import('../rhHierarchyStore.svelte');
			expect(store.getMetrics()).toBeNull();
		}, 15000);

		it('returns empty array for getEmployeeList() initially', async () => {
			const store = await import('../rhHierarchyStore.svelte');
			expect(store.getEmployeeList()).toEqual([]);
		});

		it('returns empty string for getSelectedNodeId() initially', async () => {
			const store = await import('../rhHierarchyStore.svelte');
			expect(store.getSelectedNodeId()).toBe('');
		});

		it('returns false for isLoadingMetrics() initially', async () => {
			const store = await import('../rhHierarchyStore.svelte');
			expect(store.isLoadingMetrics()).toBe(false);
		});

		it('returns null for getMetricsError() initially', async () => {
			const store = await import('../rhHierarchyStore.svelte');
			expect(store.getMetricsError()).toBeNull();
		});
	});

	describe('selectNode (synchronous paths)', () => {
		it('clears all state when nodeId is empty', async () => {
			const cm = await getMockClient();
			cm.GET.mockResolvedValue({ data: makeAreaMetrics(), error: undefined });

			const store = await import('../rhHierarchyStore.svelte');

			// First select a node to set non-null state, then clear it
			store.selectNode('node-1');
			await new Promise((r) => setTimeout(r, 0));

			store.selectNode('');

			expect(store.getSelectedNodeId()).toBe('');
			expect(store.getMetrics()).toBeNull();
			expect(store.getEmployeeList()).toEqual([]);
			expect(store.getMetricsError()).toBeNull();
			expect(store.isLoadingMetrics()).toBe(false);
		}, 15000);

		it('sets selectedNodeId immediately and marks loading', async () => {
			const cm = await getMockClient();
			cm.GET.mockResolvedValue({ data: makeAreaMetrics(), error: undefined });

			const store = await import('../rhHierarchyStore.svelte');

			store.selectNode('node-1');

			expect(store.getSelectedNodeId()).toBe('node-1');
			expect(store.isLoadingMetrics()).toBe(true);
		}, 15000);
	});

	describe('selectNode (async fetch)', () => {
		it('populates metrics and employee list on success', async () => {
			const cm = await getMockClient();
			const mockData = makeAreaMetrics();
			cm.GET.mockResolvedValue({ data: mockData, error: undefined });

			const store = await import('../rhHierarchyStore.svelte');
			store.selectNode('node-1');
			await new Promise((r) => setTimeout(r, 0));

			expect(store.isLoadingMetrics()).toBe(false);
			expect(store.getMetricsError()).toBeNull();

			const metrics = store.getMetrics();
			expect(metrics).not.toBeNull();
			expect(metrics?.nodeId).toBe('node-1');
			expect(metrics?.employeeCount).toBe(2);
			expect(metrics?.avgProgress).toBe(75);
			expect(metrics?.completedGoals).toBe(3);
			expect(metrics?.avgRating).toBe(4.2);

			const list = store.getEmployeeList();
			expect(list).toHaveLength(2);
			expect(list[0].name).toBe('Alice Smith');
			expect(list[1].name).toBe('Bob Jones');
		}, 15000);

		it('handles API error gracefully', async () => {
			const cm = await getMockClient();
			cm.GET.mockResolvedValue({
				data: undefined,
				error: { error: { message: 'Node not found' } }
			});

			const store = await import('../rhHierarchyStore.svelte');
			store.selectNode('node-missing');
			await new Promise((r) => setTimeout(r, 0));

			expect(store.isLoadingMetrics()).toBe(false);
			expect(store.getMetrics()).toBeNull();
			expect(store.getMetricsError()).toBe('Error al cargar métricas del área');
		}, 15000);

		it('handles null response data', async () => {
			const cm = await getMockClient();
			cm.GET.mockResolvedValue({ data: null, error: undefined });

			const store = await import('../rhHierarchyStore.svelte');
			store.selectNode('node-1');
			await new Promise((r) => setTimeout(r, 0));

			expect(store.isLoadingMetrics()).toBe(false);
			expect(store.getMetricsError()).toBe('No se recibieron métricas');
			expect(store.getMetrics()).toBeNull();
		}, 15000);
	});

	describe('cycleId integration', () => {
		it('includes cycleId in query when active cycle is available', async () => {
			const cycleMock = await mockGetActiveCycle();
			cycleMock.mockReturnValue({ id: 'cycle-2026' });

			const cm = await getMockClient();
			cm.GET.mockResolvedValue({ data: makeAreaMetrics(), error: undefined });

			const store = await import('../rhHierarchyStore.svelte');
			store.selectNode('node-1');
			await new Promise((r) => setTimeout(r, 0));

			expect(cm.GET).toHaveBeenCalledWith(
				'/org-nodes/{nodeId}/area-metrics',
				expect.objectContaining({
					params: expect.objectContaining({
						query: { cycleId: 'cycle-2026' }
					})
				})
			);
		}, 15000);
	});
});
