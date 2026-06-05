<script lang="ts">
	import EvaluationStatusBadge from './EvaluationStatusBadge.svelte';
	import { getEvaluationStatus } from '$lib/stores/evaluationStore.svelte';
	import { getPillars, getCompetenciesByPillar } from '$lib/stores/competencyStore.svelte';
	import { getNodeById } from '$lib/stores/orgHierarchyStore.svelte';
	import type { EmployeeAssignment } from '$lib/types/goal';
	import type { Snippet } from 'svelte';

	interface Props {
		employees: EmployeeAssignment[];
		onSelect: (employeeId: string) => void;
		selectedEmployeeId?: string;
		detail?: Snippet;
	}

	let { employees, onSelect, selectedEmployeeId = '', detail }: Props = $props();

	let searchQuery = $state('');

	const pillars = $derived(getPillars());
	const allCompetencies = $derived(pillars.flatMap((p) => getCompetenciesByPillar(p.id)));

	const filteredEmployees = $derived(
		searchQuery.trim() === ''
			? employees
			: employees.filter(
					(e) =>
						e.employeeName.toLowerCase().includes(searchQuery.toLowerCase()) ||
						getArea(e.employeeId).toLowerCase().includes(searchQuery.toLowerCase())
				)
	);

	function getArea(employeeId: string): string {
		const node = getNodeById(employeeId);
		if (!node) return '—';
		return node.profileId;
	}

	function getStatus(employeeId: string) {
		const assignment = employees.find((e) => e.employeeId === employeeId);
		return getEvaluationStatus(employeeId, allCompetencies.length, assignment?.goalIds ?? []);
	}
</script>

<div class="flex flex-col gap-6">
	<!-- Search input -->
	<div class="w-full max-w-sm">
		<input
			id="employee-search"
			type="text"
			class="input input-bordered input-sm w-full"
			placeholder="Buscar por nombre o área..."
			bind:value={searchQuery}
			aria-label="Buscar empleado"
		/>
	</div>

	{#if selectedEmployeeId}
		{@render detail?.()}
	{:else if filteredEmployees.length === 0}
		<p class="text-sm text-base-content/30 italic text-center py-8">
			No se encontraron empleados.
		</p>
	{:else}
		<!-- Table -->
		<div class="overflow-x-auto">
			<table class="table table-sm">
				<thead>
					<tr>
						<th class="text-xs font-semibold text-base-content/60">Nombre</th>
						<th class="text-xs font-semibold text-base-content/60">Área</th>
						<th class="text-xs font-semibold text-base-content/60">Estado</th>
						<th class="text-xs font-semibold text-base-content/60">Acción</th>
					</tr>
				</thead>
				<tbody>
					{#each filteredEmployees as emp (emp.employeeId)}
						<tr class="hover:bg-base-200">
							<td class="font-medium">{emp.employeeName}</td>
							<td>{getArea(emp.employeeId)}</td>
							<td>
								<EvaluationStatusBadge status={getStatus(emp.employeeId)} />
							</td>
							<td>
								<button
									type="button"
									class="btn btn-primary btn-xs"
									onclick={() => onSelect(emp.employeeId)}
								>
									Evaluar
								</button>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
