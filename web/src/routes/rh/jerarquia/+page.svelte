<script lang="ts">
	import OrgHierarchyTree from '$lib/components/org-hierarchy/OrgHierarchyTree.svelte';
	import { getRoot, getNodeById } from '$lib/stores/orgHierarchyStore.svelte';
	import { getProfile, getPhase } from '$lib/stores/devContext.svelte';
	import {
		selectNode,
		getMetrics,
		getEmployeeList,
		getSelectedNodeId
	} from '$lib/stores/rhHierarchyStore.svelte';
	import { PROFILE_LABELS } from '$lib/types/evaluation';
	import type { OrgNode } from '$lib/types/org-hierarchy';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import { Network, Users, Target, CheckCircle2, Clock, Star, Briefcase, MapPin } from '@lucide/svelte';

	// ─── Profile guard ─────────────────────────────────────────────────────

	const profile = $derived(getProfile());
	const isAuthorized = $derived(profile === 'rh');

	// ─── Phase detection ────────────────────────────────────────────────────

	const phase = $derived(getPhase());
	const metricType = $derived(
		phase === 'medio-anio'
			? 'progress'
			: phase === 'fin-anio'
				? 'rating'
				: 'unavailable'
	);

	// ─── Tree root — full corporate tree for RRHH ──────────────────────────

	const treeRoot: OrgNode | null = $derived(isAuthorized ? getRoot() : null);

	// ─── Reactive bindings to store ────────────────────────────────────────

	const selectedNodeId = $derived(getSelectedNodeId());
	const metrics = $derived(getMetrics());
	const employeeList = $derived(getEmployeeList());

	const selectedNode = $derived(
		selectedNodeId && isAuthorized ? getNodeById(selectedNodeId) : null
	);

	function handleNodeSelect(node: OrgNode) {
		selectNode(node.id);
	}
</script>

<svelte:head>
	<title>Jerarquía organizacional — SED</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<!-- Header -->
	<div>
		<h1 class="text-2xl font-bold text-base-content flex items-center gap-2">
			<Network class="w-6 h-6" />
			Jerarquía organizacional
		</h1>
		<p class="text-sm text-base-content/50 mt-1">Vista transversal de métricas por área</p>
	</div>

	{#if !isAuthorized}
		<EmptyState
			title="Sin acceso"
			message="Solo el área de RRHH puede ver las métricas de la organización."
			actionLabel="Volver al inicio"
			actionHref="/"
		/>
	{:else if !treeRoot}
		<EmptyState
			title="Sin datos"
			message="No se encontró la jerarquía organizacional."
		/>
	{:else}
		<div class="flex flex-col lg:flex-row gap-8">
			<!-- Tree panel — ~40vw -->
			<div class="lg:w-2/5">
				<div class="card bg-base-100 border border-base-300">
					<div class="card-body p-0">
						<h2 class="card-title text-xs font-semibold text-base-content/50 tracking-wide px-4 pt-4">
							Organigrama
						</h2>
						<OrgHierarchyTree
							node={treeRoot}
							onNodeSelect={handleNodeSelect}
							{selectedNodeId}
						/>
					</div>
				</div>
			</div>

			<!-- Detail panel — ~60vw -->
			<div class="lg:w-3/5">
				{#if selectedNode}
					{@const nodeLabel = PROFILE_LABELS[selectedNode.profileId as keyof typeof PROFILE_LABELS] ?? selectedNode.profileId}

					<div class="card bg-base-100 border border-base-300">
						<div class="card-body p-0">
							<!-- Header: node name + badge -->
							<div class="flex items-center gap-3 px-4 pt-4 pb-3 border-b border-base-200">
								<div class="flex items-center gap-2 min-w-0">
									<h3 class="text-lg font-bold truncate">{selectedNode.name}</h3>
									<span class="badge badge-ghost badge-sm shrink-0 capitalize">{nodeLabel}</span>
								</div>
							</div>

							<!-- Phase-conditional metrics -->
							{#if metricType === 'progress' && metrics}
								<div class="grid grid-cols-2 gap-4 p-4">
									<!-- Avg progress -->
									<div class="flex flex-col gap-1 bg-base-200 rounded-lg p-4">
										<div class="flex items-center gap-2 text-base-content/40">
											<Target class="w-4 h-4" />
											<span class="text-xs font-medium uppercase tracking-wide">Avance promedio</span>
										</div>
										<p class="text-2xl font-bold">
											{metrics.avgProgress !== null ? `${Math.round(metrics.avgProgress)}%` : '—'}
										</p>
									</div>

									<!-- Completed goals -->
									<div class="flex flex-col gap-1 bg-base-200 rounded-lg p-4">
										<div class="flex items-center gap-2 text-base-content/40">
											<CheckCircle2 class="w-4 h-4 text-success" />
											<span class="text-xs font-medium uppercase tracking-wide">Completadas</span>
										</div>
										<p class="text-2xl font-bold text-success">{metrics.completedGoals}</p>
									</div>

									<!-- Pending goals -->
									<div class="flex flex-col gap-1 bg-base-200 rounded-lg p-4">
										<div class="flex items-center gap-2 text-base-content/40">
											<Clock class="w-4 h-4 text-warning" />
											<span class="text-xs font-medium uppercase tracking-wide">Pendientes</span>
										</div>
										<p class="text-2xl font-bold text-warning">{metrics.pendingGoals}</p>
									</div>

									<!-- Employees with goals -->
									<div class="flex flex-col gap-1 bg-base-200 rounded-lg p-4">
										<div class="flex items-center gap-2 text-base-content/40">
											<Users class="w-4 h-4" />
											<span class="text-xs font-medium uppercase tracking-wide">Colaboradores</span>
										</div>
										<p class="text-2xl font-bold">
											{metrics.employeesWithGoals}
											<span class="text-sm font-normal text-base-content/40">/ {metrics.employeeCount}</span>
										</p>
									</div>
								</div>

							{:else if metricType === 'rating' && metrics}
								<div class="grid grid-cols-2 gap-4 p-4">
									<!-- Avg rating -->
									<div class="flex flex-col gap-1 bg-base-200 rounded-lg p-4">
										<div class="flex items-center gap-2 text-base-content/40">
											<Star class="w-4 h-4 text-warning" />
											<span class="text-xs font-medium uppercase tracking-wide">Rating promedio</span>
										</div>
										<p class="text-2xl font-bold">
											{metrics.avgRating !== null ? metrics.avgRating.toFixed(1) : '—'}
										</p>
									</div>

									<!-- Ratings count -->
									<div class="flex flex-col gap-1 bg-base-200 rounded-lg p-4">
										<div class="flex items-center gap-2 text-base-content/40">
											<Star class="w-4 h-4" />
											<span class="text-xs font-medium uppercase tracking-wide">Evaluaciones</span>
										</div>
										<p class="text-2xl font-bold">{metrics.ratingsCount}</p>
									</div>

									<!-- Employee count -->
									<div class="flex flex-col gap-1 bg-base-200 rounded-lg p-4">
										<div class="flex items-center gap-2 text-base-content/40">
											<Users class="w-4 h-4" />
											<span class="text-xs font-medium uppercase tracking-wide">Total colaboradores</span>
										</div>
										<p class="text-2xl font-bold">{metrics.employeeCount}</p>
									</div>

									<!-- Employees with goals (contextual) -->
									<div class="flex flex-col gap-1 bg-base-200 rounded-lg p-4">
										<div class="flex items-center gap-2 text-base-content/40">
											<Target class="w-4 h-4" />
											<span class="text-xs font-medium uppercase tracking-wide">Con metas</span>
										</div>
										<p class="text-2xl font-bold">{metrics.employeesWithGoals}</p>
									</div>
								</div>

							{:else if metricType === 'unavailable'}
								<div class="px-4 pb-4">
									<EmptyState
										title="Métricas no disponibles"
										message="Las métricas de rendimiento aún no están disponibles en la fase de inicio de año."
									/>
								</div>
							{/if}

							<!-- Employee table -->
							{#if employeeList.length > 0}
								<div class="border-t border-base-200 px-4 py-3">
									<h4 class="text-sm font-semibold text-base-content/60 mb-3 flex items-center gap-2">
										<Users class="w-4 h-4" />
										Colaboradores del área
									</h4>
									<div class="overflow-x-auto">
										<table class="table table-sm">
											<thead>
												<tr>
													<th class="text-xs font-semibold uppercase tracking-wide text-base-content/40">Nombre</th>
													<th class="text-xs font-semibold uppercase tracking-wide text-base-content/40">Puesto</th>
													<th class="text-xs font-semibold uppercase tracking-wide text-base-content/40">Perfil</th>
												</tr>
											</thead>
											<tbody>
												{#each employeeList as emp (emp.id)}
													<tr class="hover">
														<td class="font-medium">{emp.name}</td>
														<td class="text-base-content/60">{emp.position}</td>
														<td><span class="badge badge-ghost badge-sm capitalize">{emp.profile}</span></td>
													</tr>
												{/each}
											</tbody>
										</table>
									</div>
								</div>
							{/if}
						</div>
					</div>
				{:else}
					<div class="card bg-base-100 border border-base-300">
						<div class="card-body p-8 text-center">
							<Users class="w-10 h-10 text-base-content/20 mx-auto mb-3" />
							<p class="text-sm text-base-content/40">
								Selecciona un área del organigrama para ver sus métricas.
							</p>
						</div>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>
