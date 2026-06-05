<script lang="ts">
	import EmployeeEvaluationDetail from '$lib/components/evaluation/EmployeeEvaluationDetail.svelte';
	import EmployeeEvaluationTable from '$lib/components/evaluation/EmployeeEvaluationTable.svelte';
	import { getProfile } from '$lib/stores/devContext.svelte';
	import { getAssignments, getAssignmentsByProfile } from '$lib/stores/goalsStore.svelte';
	import { getChildren } from '$lib/stores/orgHierarchyStore.svelte';

	const profile = $derived(getProfile());
	const currentUserId = $derived(getAssignmentsByProfile(profile)[0]?.employeeId ?? '');

	const children = $derived(getChildren(currentUserId));
	const childIds = $derived(children.map((c) => c.id));

	const allAssignments = $derived(getAssignments());
	const subordinateAssignments = $derived(
		allAssignments.filter((a) => childIds.includes(a.employeeId))
	);

	let selectedEmployeeId = $state('');

	function handleSelect(employeeId: string) {
		selectedEmployeeId = employeeId;
	}

	function handleBack() {
		selectedEmployeeId = '';
	}
</script>

<svelte:head>
	<title>Mis evaluados — SED</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<div>
		<h1 class="text-2xl font-bold text-base-content">Mis evaluados</h1>
		<p class="text-sm text-base-content/50 mt-1">
			Revisión de evaluaciones de tu equipo
		</p>
	</div>

	{#if subordinateAssignments.length > 0}
		<EmployeeEvaluationTable
			employees={subordinateAssignments}
			onSelect={handleSelect}
			selectedEmployeeId={selectedEmployeeId}
		>
			{#snippet detail()}
				{#if selectedEmployeeId}
					<EmployeeEvaluationDetail
						employeeId={selectedEmployeeId}
						viewerMode="manager"
						showBreadcrumb={true}
						onBack={handleBack}
					/>
				{/if}
			{/snippet}
		</EmployeeEvaluationTable>
	{:else}
		<p class="text-sm text-base-content/30 italic text-center py-8">
			No tenés evaluados asignados.
		</p>
	{/if}
</div>
