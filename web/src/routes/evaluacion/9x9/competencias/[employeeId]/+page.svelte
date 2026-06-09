<script lang="ts">
	import { page } from '$app/stores';
	import CompetencyNetworkView from '$lib/components/evaluation/CompetencyNetworkView.svelte';
	import { getNodeById } from '$lib/stores/orgHierarchyStore.svelte';
	import { PROFILE_LABELS } from '$lib/types/evaluation';
	import { getProfile } from '$lib/stores/devContext.svelte';
	import { getAssignmentsByProfile } from '$lib/stores/goalsStore.svelte';
	import { Users, Network, Table } from '@lucide/svelte';

	const employeeId = $derived($page.params.employeeId);
	const employeeNode = $derived(getNodeById(employeeId));
	const employeeName = $derived(employeeNode?.name ?? 'Empleado');
	const profileLabel = $derived(employeeNode ? PROFILE_LABELS[employeeNode.profileId as keyof typeof PROFILE_LABELS] ?? '' : '');

	const profile = $derived(getProfile());
	const myEmployeeId = $derived(getAssignmentsByProfile(profile)[0]?.employeeId ?? '');
	const isOwnProfile = $derived(employeeId === myEmployeeId);

	let activeTab: 'radar' | 'table' = $state('radar');

	function handleTabKeydown(e: KeyboardEvent) {
		if (e.key === 'ArrowRight' || e.key === 'ArrowLeft') {
			e.preventDefault();
			activeTab = activeTab === 'radar' ? 'table' : 'radar';
		}
	}
</script>

<svelte:head>
	<title>Competencias de {employeeName} — SED</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<!-- Breadcrumb -->
	<nav class="breadcrumbs text-sm" aria-label="Navegación">
		<ul>
			{#if isOwnProfile}
				<li>
					<a href="/mi-evaluacion" class="link link-hover text-base-content/50">
						Mi evaluación
					</a>
				</li>
				<li class="text-base-content/70">
					<span class="font-medium">Yo</span>
				</li>
			{:else}
				<li>
					<a href="/evaluacion/9x9/competencias" class="link link-hover text-base-content/50">
						Competencias
					</a>
				</li>
				<li class="text-base-content/70">
					<span class="font-medium">{employeeName}</span>
				</li>
			{/if}
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
					<Users class="w-3.5 h-3.5" />
					{profileLabel}
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

	<!-- Competency view -->
	<CompetencyNetworkView {employeeId} {employeeName} {activeTab} />
</div>
