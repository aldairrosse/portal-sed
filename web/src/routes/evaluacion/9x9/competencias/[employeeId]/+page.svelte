<script lang="ts">
	import { page } from '$app/stores';
	import CompetencyNetworkView from '$lib/components/evaluation/CompetencyNetworkView.svelte';
	import { load, isLoading } from '$lib/stores/evaluationStore.svelte';
	import { load as loadCompetencies } from '$lib/stores/competencyStore.svelte';
	import { client } from '$lib/api/client';
	import PageSkeleton from '$lib/components/ui/PageSkeleton.svelte';
	import { Network, Table, Briefcase } from '@lucide/svelte';
    import { titleCase } from '$lib/utils/text';

	const employeeId = $derived($page.params.employeeId);

	let employeeData = $state<{ firstName: string; lastName: string; jobTitle: string; profileName: string } | null>(null);
	const employeeName = $derived(employeeData ? `${employeeData.firstName} ${employeeData.lastName}` : 'Empleado');

	let activeTab: 'radar' | 'table' = $state('radar');

	function handleTabKeydown(e: KeyboardEvent) {
		if (e.key === 'ArrowRight' || e.key === 'ArrowLeft') {
			e.preventDefault();
			activeTab = activeTab === 'radar' ? 'table' : 'radar';
		}
	}

	$effect(() => {
		if (employeeId) {
			// Load competency catalog (pillars, competencies, levels) — needed for radar/table
			loadCompetencies();
			// cycle_id is optional — backend resolves the active cycle automatically
			load(employeeId);
			// Fetch employee details for header
			client.GET('/employees/{empId}', { params: { path: { empId: employeeId } } })
				.then(({ data }) => {
					if (data?.data) {
						employeeData = {
							firstName: data.data.firstName ?? '',
							lastName: data.data.lastName ?? '',
							jobTitle: data.data.jobTitle ?? '',
							profileName: data.data.profileName ?? ''
						};
					}
				});
		}
	});
</script>

<svelte:head>
	<title>Competencias de {employeeName} — SED</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<!-- Breadcrumb -->
	<nav class="breadcrumbs text-sm" aria-label="Navegación">
		<ul>
			<li>
				<button onclick={() => window.history.back()} class="link link-hover text-base-content/50">
					Evaluaciones
				</button>
			</li>
			<li class="text-base-content/70">
				<span class="font-medium">{titleCase(employeeData?.profileName ?? 'Colaborador')}</span>
			</li>
		</ul>
	</nav>

	<!-- Employee header + tabs -->
	<div class="flex items-center justify-between flex-wrap gap-3">
		<div class="flex items-center gap-3">
			<div class="avatar placeholder">
				<div class="bg-primary text-primary-content w-10 rounded-full flex items-center justify-center">
					<span class="text-sm font-bold">
						{employeeName.charAt(0).toUpperCase()}
					</span>
				</div>
			</div>
			<div>
				<h1 class="text-xl font-bold text-base-content">{employeeName}</h1>
				<p class="text-sm text-base-content/50 flex items-center gap-1.5 mt-0.5">
					<Briefcase class="w-3.5 h-3.5" />
					{employeeData?.jobTitle ?? 'Empleado'}
				</p>
			</div>
		</div>

		<div class="tabs tabs-box" role="tablist" aria-label="Selector de vista" onkeydown={handleTabKeydown} tabindex="0">
			<button role="tab"
				id="view-radar"
				class="tab"
				class:tab-active={activeTab === 'radar'}
				aria-selected={activeTab === 'radar'}
				aria-controls="panel-radar"
				tabindex={activeTab === 'radar' ? 0 : -1}
				onclick={() => activeTab = 'radar'}>
				<Network class="w-4 h-4" />
				Radar
			</button>
			<button role="tab"
				id="view-table"
				class="tab"
				class:tab-active={activeTab === 'table'}
				aria-selected={activeTab === 'table'}
				aria-controls="panel-table"
				tabindex={activeTab === 'table' ? 0 : -1}
				onclick={() => activeTab = 'table'}>
				<Table class="w-4 h-4" />
				Tabla
			</button>
		</div>
	</div>

	<!-- Loading skeleton -->
	{#if isLoading()}
		<PageSkeleton variant="card" rows={3} avatar />
	{:else}
		<!-- Competency view -->
		<CompetencyNetworkView {employeeId} {employeeName} {activeTab} />
	{/if}
</div>
