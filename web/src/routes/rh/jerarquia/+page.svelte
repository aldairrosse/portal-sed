<script lang="ts">
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import OrgHierarchyTree from '$lib/components/org-hierarchy/OrgHierarchyTree.svelte';
	import { getRoot, getNodeById, load, isLoading, getError } from '$lib/stores/orgHierarchyStore.svelte';
import { getProfile } from '$lib/stores/devContext.svelte';
import { getActivePhase } from '$lib/api/cycle.svelte';
	import {
		selectNode,
		getMetrics,
		getEmployeeList,
		getSelectedNodeId,
		isLoadingMetrics,
		getMetricsError
	} from '$lib/stores/rhHierarchyStore.svelte';
	import type { OrgNode } from '$lib/types/org-hierarchy';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import ErrorState from '$lib/components/ui/ErrorState.svelte';
	import PageSkeleton from '$lib/components/ui/PageSkeleton.svelte';
	import { Network, Users, Target, CheckCircle2, Clock, Star } from '@lucide/svelte';

	// ─── Profile guard ─────────────────────────────────────────────────────

	const profile = $derived(getProfile());

	// Redirect jefe to the evaluacion hierarchy view
	onMount(() => {
		if (profile === 'jefe' && browser) {
			goto('/evaluacion/9x9/jerarquia');
			return;
		}
		if (isAuthorized) {
			load();
		}
	});

	const isAuthorized = $derived(profile === 'rh');

	// ─── Phase detection ────────────────────────────────────────────────────

	const phase = $derived(getActivePhase() ?? 'inicio-anio');
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
	const metricsLoading = $derived(isLoadingMetrics());
	const metricsError = $derived(getMetricsError());

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
			Jerarquía de departamentos
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
	{:else if isLoading()}
		<div class="flex flex-col lg:flex-row gap-8">
			<div class="lg:w-2/5">
				<div class="card bg-base-100 border border-base-300">
					<div class="card-body p-0">
						<h2 class="card-title text-xs font-semibold text-base-content/50 tracking-wide px-4 pt-4">
							Departamentos
						</h2>
						<OrgHierarchyTree loading={true} />
					</div>
				</div>
			</div>
			<div class="lg:w-3/5">
				<div class="card bg-base-100 border border-base-300">
					<div class="card-body p-8 text-center">
						<PageSkeleton variant="default" rows={5} />
					</div>
				</div>
			</div>
		</div>
	{:else if getError()}
		<ErrorState message={getError() ?? undefined} onretry={load} />
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
						<h2 class="card-title text-xs font-semibold text-base-content/50 tracking-wide pt-4">
							Departamentos
						</h2>
						<OrgHierarchyTree
							viewType='departments'
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

					<div class="card bg-base-100 border border-base-300">
						<div class="card-body p-0">
							<!-- Header: node name + badge -->
							<div class="flex items-center gap-3 px-4 pt-4 pb-3">
								<div class="flex items-center gap-2 min-w-0">
									<h3 class="text-xs font-semibold truncate text-base-content/50 tracking-wide">
										{selectedNode.name}
									</h3>
								</div>
							</div>

							<!-- Loading state -->
							{#if metricsLoading}
								<div class="p-6 text-center text-sm text-base-content/40">Cargando métricas…</div>
							{:else if metricsError}
								<div class="px-4 pb-4">
									<EmptyState
										title="Error al cargar métricas"
										message={metricsError}
									/>
									<div class="mt-4 text-center">
										<button class="btn btn-outline btn-sm" onclick={() => selectNode(selectedNodeId)}>
											Reintentar
										</button>
									</div>
								</div>
							{:else if metrics && metrics.employeeCount === 0}
								<div class="p-6 text-center text-sm text-base-content/40">
									Sin empleados en esta área
								</div>
							{:else if metricType === 'progress' && metrics}
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

							{:else if !metrics}
								<div class="p-6 text-center text-sm text-base-content/40">
									Selecciona un área para ver sus métricas.
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
								<div class="px-4 py-3">
									<div class="overflow-x-auto">
										<table class="table table-sm">
											<thead>
												<tr>
													<th class="text-xs font-semibold tracking-wide text-base-content/40">Nombre</th>
													<th class="text-xs font-semibold tracking-wide text-base-content/40">Puesto</th>
													<th class="text-xs font-semibold tracking-wide text-base-content/40">Perfil</th>
												</tr>
											</thead>
											<tbody>
												{#each employeeList as emp (emp.id)}
													<tr class="hover">
														<td class="font-medium">{emp.name}</td>
														<td class="text-xs text-base-content/50">{emp.position}</td>
														<td><span class="text-xs text-base-content/50 capitalize">{emp.profile}</span></td>
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
