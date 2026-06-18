import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { OrgNode } from '$lib/types/org-hierarchy';
import type { Goal, EmployeeAssignment } from '$lib/types/goal';
import type { CompetencyRating } from '$lib/types/evaluation-result';

// The pure functions are imported from the store. Vitest resolves .svelte.ts
// via SvelteKit's vite plugin. We mock the fixture imports and store
// dependencies to keep tests isolated and deterministic.
vi.mock('$lib/fixtures/goals/goals.json', () => ({ default: [] }));
vi.mock('$lib/fixtures/goals/assignments.json', () => ({ default: [] }));
vi.mock('$lib/fixtures/evaluations/rh-evaluations.json', () => ({ default: [] }));
vi.mock('$lib/stores/orgHierarchyStore.svelte', () => ({
	getRoot: vi.fn(() => null),
	getScopeIds: vi.fn(() => [])
}));

// ─── Synthetic fixtures ─────────────────────────────────────────────────────────

function makeGoal(overrides: Partial<Goal> & { id: string }): Goal {
	return {
		name: '',
		description: '',
		categoryId: 'cat-1',
		weight: 0,
		unit: 'numero',
		targetValue: 100,
		progress: 50,
		...overrides
	};
}

function makeAssignment(overrides: Partial<EmployeeAssignment> & { id: string }): EmployeeAssignment {
	return {
		employeeId: '',
		employeeName: '',
		profileId: 'colaborador',
		managerId: null,
		goalIds: [],
		createdAt: '',
		updatedAt: '',
		...overrides
	};
}

function makeRating(overrides: Partial<CompetencyRating> & { id: string }): CompetencyRating {
	return {
		employeeId: '',
		competencyId: 'comp-1',
		...overrides
	};
}

function makeNode(overrides: Partial<OrgNode> & { id: string }): OrgNode {
	return {
		name: '',
		profileId: 'colaborador',
		managerId: null,
		children: [],
		...overrides
	};
}

// ========================================================================
// computeAreaProgress
// ========================================================================

describe('computeAreaProgress', () => {
	it('returns correct average and counts for employees with goals', async () => {
		const { computeAreaProgress } = await import('../rhHierarchyStore.svelte.ts');

		const goals: Goal[] = [
			makeGoal({ id: 'g1', targetValue: 100, progress: 80 }),   // 80%
			makeGoal({ id: 'g2', targetValue: 50, progress: 60 }),    // 120% → capped 100% → completed
			makeGoal({ id: 'g3', targetValue: 10, progress: 5 })      // 50% → pending
		];

		const assignments: EmployeeAssignment[] = [
			makeAssignment({ id: 'a1', employeeId: 'emp-1', goalIds: ['g1', 'g2'] }),
			makeAssignment({ id: 'a2', employeeId: 'emp-2', goalIds: ['g3'] }),
			makeAssignment({ id: 'a3', employeeId: 'emp-3', goalIds: [] }) // 0 goals → excluded
		];

		const result = computeAreaProgress(['emp-1', 'emp-2', 'emp-3'], assignments, goals);

		// emp-1: (80 + 100) / 2 = 90
		// emp-2: 50 / 1 = 50
		// emp-3: excluded (0 goals)
		// avg = (90 + 50) / 2 = 70
		expect(result.avgProgress).toBeCloseTo(70, 1);
		// g2: progress >= targetValue → completed
		expect(result.completed).toBe(1);
		// g1: progress < targetValue, g3: progress < targetValue → pending
		expect(result.pending).toBe(2);
	});

	it('excludes employees with zero valid goals from average', async () => {
		const { computeAreaProgress } = await import('../rhHierarchyStore.svelte.ts');

		const goals: Goal[] = [
			makeGoal({ id: 'g1', targetValue: 10, progress: 9 })
		];

		const assignments: EmployeeAssignment[] = [
			makeAssignment({ id: 'a1', employeeId: 'emp-1', goalIds: ['g1'] }),
			makeAssignment({ id: 'a2', employeeId: 'emp-2', goalIds: [] }) // no goals
		];

		const result = computeAreaProgress(['emp-1', 'emp-2'], assignments, goals);

		// Only emp-1 contributes: 9/10 * 100 = 90%
		expect(result.avgProgress).toBeCloseTo(90, 1);
		expect(result.completed).toBe(0);
		expect(result.pending).toBe(1);
	});

	it('skips goals with targetValue === 0', async () => {
		const { computeAreaProgress } = await import('../rhHierarchyStore.svelte.ts');

		const goals: Goal[] = [
			makeGoal({ id: 'g1', targetValue: 0, progress: 100 }), // skipped
			makeGoal({ id: 'g2', targetValue: 10, progress: 8 })    // included
		];

		const assignments: EmployeeAssignment[] = [
			makeAssignment({ id: 'a1', employeeId: 'emp-1', goalIds: ['g1', 'g2'] })
		];

		const result = computeAreaProgress(['emp-1'], assignments, goals);

		// Only g2 counts: 8/10 * 100 = 80%
		expect(result.avgProgress).toBeCloseTo(80, 1);
		expect(result.completed).toBe(0);
		expect(result.pending).toBe(1);
	});

	it('returns zeros for empty employeeIds', async () => {
		const { computeAreaProgress } = await import('../rhHierarchyStore.svelte.ts');

		const result = computeAreaProgress([], [], []);
		expect(result.avgProgress).toBe(0);
		expect(result.completed).toBe(0);
		expect(result.pending).toBe(0);
	});

	it('returns zeros when no assignments match', async () => {
		const { computeAreaProgress } = await import('../rhHierarchyStore.svelte.ts');

		const result = computeAreaProgress(
			['emp-unknown'],
			[makeAssignment({ id: 'a1', employeeId: 'emp-1', goalIds: ['g1'] })],
			[makeGoal({ id: 'g1', targetValue: 10, progress: 5 })]
		);
		expect(result.avgProgress).toBe(0);
		expect(result.completed).toBe(0);
		expect(result.pending).toBe(0);
	});

	it('caps progress at 100% per goal', async () => {
		const { computeAreaProgress } = await import('../rhHierarchyStore.svelte.ts');

		const goals: Goal[] = [
			makeGoal({ id: 'g1', targetValue: 10, progress: 20 }) // 200% → capped 100%
		];

		const assignments: EmployeeAssignment[] = [
			makeAssignment({ id: 'a1', employeeId: 'emp-1', goalIds: ['g1'] })
		];

		const result = computeAreaProgress(['emp-1'], assignments, goals);
		expect(result.avgProgress).toBeCloseTo(100, 1);
		expect(result.completed).toBe(1);
		expect(result.pending).toBe(0);
	});
});

// ========================================================================
// computeAreaRating
// ========================================================================

describe('computeAreaRating', () => {
	it('returns correct average rating', async () => {
		const { computeAreaRating } = await import('../rhHierarchyStore.svelte.ts');

		const evaluations: CompetencyRating[] = [
			makeRating({ id: 'r1', employeeId: 'emp-1', competencyId: 'comp-a', rhRating: 4 }),
			makeRating({ id: 'r2', employeeId: 'emp-1', competencyId: 'comp-b', rhRating: 3 }),
			makeRating({ id: 'r3', employeeId: 'emp-2', competencyId: 'comp-a', rhRating: 5 })
		];

		const result = computeAreaRating(['emp-1', 'emp-2'], evaluations);
		// (4 + 3 + 5) / 3 = 4
		expect(result.avgRating).toBeCloseTo(4, 1);
		expect(result.ratingsCount).toBe(3);
	});

	it('returns null and zero count when no ratings exist', async () => {
		const { computeAreaRating } = await import('../rhHierarchyStore.svelte.ts');

		const result = computeAreaRating(
			['emp-1'],
			[makeRating({ id: 'r1', employeeId: 'emp-1', competencyId: 'comp-a', rhRating: undefined })]
		);
		expect(result.avgRating).toBeNull();
		expect(result.ratingsCount).toBe(0);
	});

	it('returns null for empty employeeIds', async () => {
		const { computeAreaRating } = await import('../rhHierarchyStore.svelte.ts');

		const result = computeAreaRating([], []);
		expect(result.avgRating).toBeNull();
		expect(result.ratingsCount).toBe(0);
	});

	it('excludes employees with no rhRating from the mean', async () => {
		const { computeAreaRating } = await import('../rhHierarchyStore.svelte.ts');

		const evaluations: CompetencyRating[] = [
			makeRating({ id: 'r1', employeeId: 'emp-1', competencyId: 'comp-a', rhRating: 4 }),
			makeRating({ id: 'r2', employeeId: 'emp-2', competencyId: 'comp-a', rhRating: undefined }),
			makeRating({ id: 'r3', employeeId: 'emp-3', competencyId: 'comp-a', rhRating: 5 })
		];

		const result = computeAreaRating(['emp-1', 'emp-2', 'emp-3'], evaluations);
		// Only emp-1 (4) and emp-3 (5) count: (4 + 5) / 2 = 4.5
		expect(result.avgRating).toBeCloseTo(4.5, 1);
		expect(result.ratingsCount).toBe(2);
	});

	it('returns null when employees exist but have no ratings', async () => {
		const { computeAreaRating } = await import('../rhHierarchyStore.svelte.ts');

		const result = computeAreaRating(['emp-1', 'emp-2'], []);
		expect(result.avgRating).toBeNull();
		expect(result.ratingsCount).toBe(0);
	});
});

// ========================================================================
// buildEmployeeList
// ========================================================================

describe('buildEmployeeList', () => {
	it('includes the area manager and all descendants, sorted A-Z', async () => {
		const { buildEmployeeList } = await import('../rhHierarchyStore.svelte.ts');

		const nodes: OrgNode[] = [
			makeNode({ id: 'emp-b', name: 'Bob', profileId: 'jefe' }),
			makeNode({ id: 'emp-a', name: 'Alice', profileId: 'colaborador' }),
			makeNode({ id: 'emp-c', name: 'Charlie', profileId: 'vendedor' })
		];

		const result = buildEmployeeList(['emp-a', 'emp-b', 'emp-c'], nodes);

		expect(result).toHaveLength(3);
		// A-Z sort
		expect(result[0].name).toBe('Alice');
		expect(result[1].name).toBe('Bob');
		expect(result[2].name).toBe('Charlie');
	});

	it('maps profileId to correct PROFILE_LABELS position', async () => {
		const { buildEmployeeList } = await import('../rhHierarchyStore.svelte.ts');

		const nodes: OrgNode[] = [
			makeNode({ id: 'emp-1', name: 'Carmen', profileId: 'director' }),
			makeNode({ id: 'emp-2', name: 'Luis', profileId: 'gerente-tienda' }),
			makeNode({ id: 'emp-3', name: 'Laura', profileId: 'rh' })
		];

		const result = buildEmployeeList(['emp-1', 'emp-2', 'emp-3'], nodes);

		expect(result.find((r) => r.id === 'emp-1')?.position).toBe('Director');
		expect(result.find((r) => r.id === 'emp-2')?.position).toBe('Gerente de tienda');
		expect(result.find((r) => r.id === 'emp-3')?.position).toBe('Recursos Humanos');
	});

	it('falls back to profileId when label is unknown', async () => {
		const { buildEmployeeList } = await import('../rhHierarchyStore.svelte.ts');

		const nodes: OrgNode[] = [
			makeNode({ id: 'emp-x', name: 'X', profileId: 'unknown-role' })
		];

		const result = buildEmployeeList(['emp-x'], nodes);
		expect(result[0].position).toBe('unknown-role');
	});

	it('returns empty array for empty employeeIds', async () => {
		const { buildEmployeeList } = await import('../rhHierarchyStore.svelte.ts');

		const result = buildEmployeeList([], []);
		expect(result).toEqual([]);
	});

	it('skips employeeIds that have no matching node', async () => {
		const { buildEmployeeList } = await import('../rhHierarchyStore.svelte.ts');

		const nodes: OrgNode[] = [
			makeNode({ id: 'emp-1', name: 'Alice', profileId: 'colaborador' })
		];

		// emp-2 has no matching node → excluded
		const result = buildEmployeeList(['emp-1', 'emp-2'], nodes);
		expect(result).toHaveLength(1);
		expect(result[0].id).toBe('emp-1');
	});
});
