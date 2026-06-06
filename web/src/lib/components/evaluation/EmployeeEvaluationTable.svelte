<script lang="ts">
	import EvaluationStatusBadge from './EvaluationStatusBadge.svelte';
	import { getEvaluationStatus } from '$lib/stores/evaluationStore.svelte';
	import { getPillars, getCompetenciesByPillar } from '$lib/stores/competencyStore.svelte';
	import { getNodeById } from '$lib/stores/orgHierarchyStore.svelte';
	import { PROFILE_LABELS } from '$lib/types/evaluation';
	import type { EmployeeAssignment } from '$lib/types/goal';
	import type { Snippet } from 'svelte';

	interface Props {
		employees: EmployeeAssignment[];
		onSelect: (employeeId: string) => void;
		selectedEmployeeId?: string;
		disabled?: boolean;
		detail?: Snippet;
	}

	let { employees, onSelect, selectedEmployeeId = '', disabled = false, detail }: Props = $props();

	let searchQuery = $state('');

	const pillars = $derived(getPillars());
	const allCompetencies = $derived(pillars.flatMap((p) => getCompetenciesByPillar(p.id)));

	const filteredEmployees = $derived(
		searchQuery.trim() === ''
			? employees
			: employees.filter(
					(e) =>
						e.employeeName.toLowerCase().includes(searchQuery.toLowerCase()) ||
						getProfileLabel(e.employeeId).toLowerCase().includes(searchQuery.toLowerCase())
				)
	);

	function getProfileLabel(employeeId: string): string {
		const node = getNodeById(employeeId);
		if (!node) return '—';
		return PROFILE_LABELS[node.profileId as keyof typeof PROFILE_LABELS] ?? node.profileId;
	}

	function getStatus(employeeId: string) {
		const assignment = employees.find((e) => e.employeeId === employeeId);
		return getEvaluationStatus(employeeId, allCompetencies.length, assignment?.goalIds ?? []);
	}
</script>

<div class="flex flex-col gap-6">
	{#if !selectedEmployeeId}
		<!-- Search input -->
		<div class="w-full max-w-sm">
			<input
				id="employee-search"
				type="text"
				class="input input-bordered input-sm w-full"
				placeholder="Buscar por nombre o perfil..."
				bind:value={searchQuery}
				aria-label="Buscar empleado"
			/>
		</div>
	{/if}

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
						<th class="text-xs font-semibold text-base-content/60">Empleado</th>
						<th class="text-xs font-semibold text-base-content/60">Perfil</th>
						<th class="text-xs font-semibold text-base-content/60">Estado</th>
						<th class="text-xs font-semibold text-base-content/60">Acción</th>
					</tr>
				</thead>
				<tbody>
					{#each filteredEmployees as emp (emp.employeeId)}
						<tr class="hover:bg-base-200">
							<td>
								<div class="flex items-center gap-2.5">
									<div class="avatar avatar-placeholder">
										<div class="bg-primary text-primary-content w-8 rounded-full flex items-center justify-center">
											<span class="text-xs font-bold">
												{emp.employeeName.charAt(0).toUpperCase()}
											</span>
										</div>
									</div>
									<span class="font-medium text-sm">{emp.employeeName}</span>
								</div>
							</td>
							<td>
								<span class="text-xs text-base-content/50">{getProfileLabel(emp.employeeId)}</span>
							</td>
							<td>
								<EvaluationStatusBadge status={getStatus(emp.employeeId)} />
							</td>
							<td>
							<button
								type="button"
								class="btn btn-primary btn-xs"
								onclick={() => onSelect(emp.employeeId)}
								disabled={disabled}
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
