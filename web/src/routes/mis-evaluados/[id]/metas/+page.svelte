<script lang="ts">
	import { page } from '$app/stores';
	import { client } from '$lib/api/client';
	import {
		getCategories,
		getGoals,
		getGoalsByCategory,
		getInstitutionalGoals,
		loadForEmployee,
		storeState
	} from '$lib/stores/goalsStore.svelte';
	import PageSkeleton from '$lib/components/ui/PageSkeleton.svelte';
	import { ArrowLeft, Target, Briefcase } from '@lucide/svelte';
	import { titleCase } from '$lib/utils/text';

	const employeeId = $derived($page.params.id);

	let employeeData = $state<{ firstName: string; lastName: string; jobTitle: string; profileName: string } | null>(null);
	const employeeName = $derived(employeeData ? `${employeeData.firstName} ${employeeData.lastName}` : 'Empleado');

	const categories = $derived(getCategories());
	const institutionalGoals = $derived(getInstitutionalGoals());
	const loading = $derived(storeState.loading);

	$effect(() => {
		if (employeeId) {
			loadForEmployee(employeeId);
			client.GET('/employees/{empId}', { params: { path: { empId: employeeId } } }).then(({ data }) => {
				if (data?.data) {
					employeeData = {
						firstName: (data.data as Record<string, string>).firstName ?? '',
						lastName: (data.data as Record<string, string>).lastName ?? '',
						jobTitle: (data.data as Record<string, string>).jobTitle ?? '',
						profileName: (data.data as Record<string, string>).profileName ?? ''
					};
				}
			});
		}
	});
</script>

<svelte:head>
	<title>Metas de {employeeName} — SED</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<nav class="breadcrumbs text-sm" aria-label="Navegación">
		<ul>
			<li><a href="/mis-evaluados" class="link link-hover text-base-content/50">Mis evaluados</a></li>
			<li class="text-base-content/70"><span class="font-medium">{employeeName}</span></li>
		</ul>
	</nav>

	<div class="flex items-center justify-between flex-wrap gap-3">
		<div class="flex items-center gap-3">
			<div class="avatar placeholder">
				<div class="bg-primary text-primary-content w-10 rounded-full flex items-center justify-center">
					<span class="text-sm font-bold">{employeeName.charAt(0).toUpperCase()}</span>
				</div>
			</div>
			<div>
				<h1 class="text-xl font-bold text-base-content flex items-center gap-2">
					<Target class="w-5 h-5" /> Metas de {employeeName}
				</h1>
				<p class="text-sm text-base-content/50 flex items-center gap-1.5 mt-0.5">
					<Briefcase class="w-3.5 h-3.5" />
					{employeeData?.jobTitle ?? 'Empleado'} — {titleCase(employeeData?.profileName ?? 'colaborador')}
				</p>
			</div>
		</div>
		<a href="/mis-evaluados" class="btn btn-ghost btn-sm">
			<ArrowLeft class="w-4 h-4" /> Volver
		</a>
	</div>

	{#if loading}
		<PageSkeleton variant="card" rows={3} />
	{:else}
		{#if institutionalGoals.length > 0}
			<div class="space-y-3">
				<h2 class="text-sm font-semibold text-base-content/70">Metas institucionales</h2>
				{#each institutionalGoals as goal (goal.id)}
					<div class="card bg-base-100 border border-base-300 p-4">
						<div class="flex items-start justify-between gap-3">
							<div>
								<h3 class="font-semibold flex items-center gap-2">
									{goal.name}
									<span class="badge badge-ghost badge-sm">{goal.weight}%</span>
									<span class="badge badge-outline badge-sm">{goal.source === 'global' ? 'Global' : 'Compartida'}</span>
								</h3>
								{#if goal.description}<p class="text-sm text-base-content/60 mt-1">{goal.description}</p>{/if}
							</div>
						</div>
						<div class="mt-2 text-xs text-base-content/50">
							Avance: {goal.progressPercent?.toFixed(1) ?? '—'}% · Objetivo: {goal.targetValue ?? '—'}
						</div>
					</div>
				{/each}
			</div>
		{/if}

		<div class="space-y-4">
			<h2 class="text-sm font-semibold text-base-content/70">Categorías y metas personales</h2>
			{#each categories as cat (cat.id)}
				<div class="card bg-base-100 border border-base-300">
					<div class="card-body p-4">
						<h3 class="font-semibold flex items-center gap-2">
							{cat.name}
							<span class="badge badge-ghost badge-sm">{cat.weight}%</span>
						</h3>
						{#if cat.description}<p class="text-sm text-base-content/60">{cat.description}</p>{/if}
						<div class="mt-3 space-y-2">
							{#each getGoalsByCategory(cat.id) as goal (goal.id)}
								<div class="border border-base-200 rounded-lg p-3">
									<div class="flex items-start justify-between gap-2">
										<p class="font-medium text-sm">{goal.name}</p>
										<span class="badge badge-ghost badge-xs">{goal.weight}%</span>
									</div>
									{#if goal.description}<p class="text-xs text-base-content/60 mt-1">{goal.description}</p>{/if}
									<div class="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-xs text-base-content/50">
										<span>Unidad: {goal.unit}</span>
										<span>Objetivo: {goal.targetValue}</span>
										{#if goal.baselineValue !== undefined}<span>Línea base: {goal.baselineValue}</span>{/if}
										<span>Avance: {goal.progress ?? 0}</span>
									</div>
								</div>
							{:else}
								<p class="text-xs text-base-content/30 italic">Sin metas en esta categoría</p>
							{/each}
						</div>
					</div>
				</div>
			{:else}
				<div class="card bg-base-100 border border-dashed border-base-300 p-8 text-center">
					<p class="text-sm text-base-content/50">Este colaborador aún no tiene categorías registradas.</p>
					<a href={`/objetivos/asignacion?empId=${employeeId}`} class="btn btn-primary btn-sm mt-3">Formular metas</a>
				</div>
			{/each}
		</div>
	{/if}
</div>
