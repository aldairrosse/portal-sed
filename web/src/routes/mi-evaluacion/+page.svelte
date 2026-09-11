<script lang="ts">
	import { onMount } from 'svelte';
	import EmployeeEvaluationDetail from '$lib/components/evaluation/EmployeeEvaluationDetail.svelte';
	import { getSession } from '$lib/api/session.svelte';
	import { getProfile } from '$lib/stores/devContext.svelte';
	import {
		getAssignmentsByProfile,
		load as loadGoals,
		isLoading as goalsLoading,
	} from '$lib/stores/goalsStore.svelte';
	import PageSkeleton from '$lib/components/ui/PageSkeleton.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import { ClipboardCheck } from '@lucide/svelte';

	const profile = $derived(getProfile());
	const assignments = $derived(getAssignmentsByProfile(profile));
	const assignment = $derived(assignments[0]);
	const assignmentStatus = $derived(assignment?.status);
	const employeeId = $derived(assignment?.employeeId ?? '');
	const employeeName = $derived(
		getSession().user?.name ?? getSession().user?.profileName ?? '',
	);
	const loading = $derived(goalsLoading());
	const showEmptyGate = $derived(
		!employeeId ||
			assignmentStatus === 'no_iniciado' ||
			assignmentStatus === undefined,
	);

	onMount(() => {
		loadGoals();
	});
</script>

<svelte:head>
	<title>Mi evaluación — SED</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<div>
		<h1 class="text-2xl font-bold text-base-content flex items-center gap-2">
			<ClipboardCheck class="w-6 h-6" />
			Mi evaluación
		</h1>
		<p class="text-sm text-base-content/50 mt-1">
			Autoevaluación de competencias y cierre de metas
		</p>
	</div>

	{#if loading}
		<PageSkeleton variant="card" rows={3} />
	{:else if employeeId && !showEmptyGate}
		<EmployeeEvaluationDetail
			{employeeId}
			viewerMode="self"
			showBreadcrumb={false}
			{employeeName}
		/>
	{:else}
		<EmptyState
			title="Sin asignación de metas"
			message="Aún no tienes metas asignadas para este ciclo. Crea tus categorías y objetivos para empezar."
			actionLabel="Ir a metas"
			actionHref="/objetivos/asignacion"
		/>
	{/if}
</div>
