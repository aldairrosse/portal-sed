<script lang="ts">
	import { page } from '$app/stores';
	import CompetencyNetworkView from '$lib/components/evaluation/CompetencyNetworkView.svelte';
	import { getNodeById } from '$lib/stores/orgHierarchyStore.svelte';
	import { PROFILE_LABELS } from '$lib/types/evaluation';
	import { Users } from '@lucide/svelte';

	const employeeId = $derived($page.params.employeeId);
	const employeeNode = $derived(getNodeById(employeeId));
	const employeeName = $derived(employeeNode?.name ?? 'Empleado');
	const profileLabel = $derived(employeeNode ? PROFILE_LABELS[employeeNode.profileId as keyof typeof PROFILE_LABELS] ?? '' : '');
</script>

<svelte:head>
	<title>Competencias de {employeeName} — SED</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<!-- Breadcrumb -->
	<nav class="breadcrumbs text-sm" aria-label="Navegación">
		<ul>
			<li>
				<a href="/evaluacion/9x9/competencias" class="link link-hover text-base-content/50">
					Competencias
				</a>
			</li>
			<li class="text-base-content/70">
				<span class="font-medium">{employeeName}</span>
			</li>
		</ul>
	</nav>

	<!-- Employee header -->
	<div class="flex items-center gap-4">
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
	</div>

	<!-- Competency table -->
	<CompetencyNetworkView {employeeId} {employeeName} />
</div>
