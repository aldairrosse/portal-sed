<script lang="ts">
	import EmployeeEvaluationDetail from '$lib/components/evaluation/EmployeeEvaluationDetail.svelte';
	import EmployeeEvaluationTable from '$lib/components/evaluation/EmployeeEvaluationTable.svelte';
	import { getPhase } from '$lib/stores/devContext.svelte';
	import { getAssignments } from '$lib/stores/goalsStore.svelte';

	const phase = $derived(getPhase());
	const isFinAnio = $derived(phase === 'fin-anio');
	const isMedioAnio = $derived(phase === 'medio-anio');

	const phaseDescription = $derived(
		isFinAnio
			? 'Evaluación formal de competencias y cierre de metas'
			: isMedioAnio
				? 'Revisión de avance de objetivos y competencias de todos los empleados'
				: 'Seguimiento de objetivos de todos los empleados para el ciclo actual'
	);

	const allAssignments = $derived(getAssignments());

	let selectedEmployeeId = $state('');

	function handleSelect(employeeId: string) {
		selectedEmployeeId = employeeId;
	}

	function handleBack() {
		selectedEmployeeId = '';
	}
</script>

<svelte:head>
	<title>Evaluaciones RH — SED</title>
</svelte:head>

<div class="flex flex-col gap-6">
	{#if !selectedEmployeeId}
		<div>
			<h1 class="text-2xl font-bold text-base-content">Evaluaciones RH</h1>
			<p class="text-sm text-base-content/50 mt-1">
				{phaseDescription}
			</p>
		</div>
	{/if}

	{#if allAssignments.length > 0}
		<EmployeeEvaluationTable
			employees={allAssignments}
			onSelect={handleSelect}
			selectedEmployeeId={selectedEmployeeId}
			disabled={!isFinAnio}
		>
			{#snippet detail()}
				{#if selectedEmployeeId}
					<EmployeeEvaluationDetail
						employeeId={selectedEmployeeId}
						viewerMode="rh"
						showBreadcrumb={true}
						onBack={handleBack}
					/>
				{/if}
			{/snippet}
		</EmployeeEvaluationTable>
	{:else}
		<p class="text-sm text-base-content/30 italic text-center py-8">
			No hay evaluados registrados.
		</p>
	{/if}
</div>
