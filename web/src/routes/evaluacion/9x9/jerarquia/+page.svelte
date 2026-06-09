<script lang="ts">
	import OrgHierarchyTree from '$lib/components/org-hierarchy/OrgHierarchyTree.svelte';
	import { getRoot, getNodeById } from '$lib/stores/orgHierarchyStore.svelte';
	import { getProfile } from '$lib/stores/devContext.svelte';
	import { PROFILE_LABELS, type EvaluationProfile } from '$lib/types/evaluation';
	import type { OrgNode } from '$lib/types/org-hierarchy';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import { MapPin, Target, ArrowUpRight, Briefcase, Network, Users } from '@lucide/svelte';

	// ─── Profile guard ─────────────────────────────────────────────────────

	const profile = $derived(getProfile());
	const ALLOWED_PROFILES: EvaluationProfile[] = ['director', 'director-general'];
	const isAuthorized = $derived(ALLOWED_PROFILES.includes(profile));

	// ─── Tree root ─────────────────────────────────────────────────────────

	const PROFILE_NODE_ID: Partial<Record<EvaluationProfile, string>> = {
		'director-general': 'emp-dg-01',
		director: 'emp-director-01'
	};

	const treeRoot: OrgNode | null = $derived.by(() => {
		if (!isAuthorized) return null;
		if (profile === 'director-general') return getRoot();
		const nodeId = PROFILE_NODE_ID[profile];
		if (!nodeId) return null;
		return getNodeById(nodeId);
	});

	// ─── Node selection ────────────────────────────────────────────────────

	let selectedNodeId = $state<string>('');
	let selectedNode = $state<OrgNode | null>(null);

	// ─── Area mapping (mock data) ──────────────────────────────────────────

	const AREA_MAP: Record<string, string> = {
		'emp-dg-01': 'Dirección General',
		'emp-director-01': 'Comercial · Sucursal Centro',
		'emp-director-02': 'Operaciones · Sucursal Norte',
		'emp-director-03': 'Administración · Sucursal Sur',
		'emp-jefe-01': 'Comercial · Ventas Mayoristas',
		'emp-jefe-02': 'Comercial · Ventas Minoristas',
		'emp-jefe-03': 'Operaciones · Logística',
		'emp-jefe-04': 'Operaciones · Almacén',
		'emp-jefe-05': 'Administración · Finanzas',
		'emp-jefe-06': 'Administración · RRHH',
		'emp-jefe-directo-dg-01': 'Estrategia · Proyectos Especiales'
	};

	// ─── Goal percentage (mock data) ───────────────────────────────────────

	const GOAL_PERCENT_MAP: Record<string, number> = {
		'emp-dg-01': 78,
		'emp-director-01': 65,
		'emp-director-02': 82,
		'emp-director-03': 54,
		'emp-jefe-01': 71,
		'emp-jefe-02': 45,
		'emp-jefe-03': 88,
		'emp-jefe-04': 33,
		'emp-jefe-05': 62,
		'emp-jefe-06': 91,
		'emp-jefe-directo-dg-01': 56
	};

	function getArea(node: OrgNode | null): string {
		if (!node) return '—';
		return AREA_MAP[node.id] ?? 'Sin asignar';
	}

	function getGoalPercent(node: OrgNode | null): number {
		if (!node) return 0;
		return GOAL_PERCENT_MAP[node.id] ?? 0;
	}

	function handleNodeSelect(node: OrgNode) {
		selectedNodeId = node.id;
		selectedNode = node;
	}
</script>

<svelte:head>
	<title>Jerarquía organizacional — SED</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<!-- Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-base-content flex items-center gap-2">
				<Network class="w-6 h-6" />
				Jerarquía organizacional
			</h1>
			<p class="text-sm text-base-content/50 mt-1">Explora la estructura de tu organización</p>
		</div>
		<a href="/evaluacion/9x9" class="btn btn-ghost btn-sm gap-1.5">
			<ArrowUpRight class="w-4 h-4" />
			Matriz 9×9
		</a>
	</div>

	{#if !isAuthorized}
		<EmptyState
			title="Sin acceso"
			message="Solo directores y director general pueden ver la jerarquía organizacional."
			actionLabel="Volver al inicio"
			actionHref="/"
		/>
	{:else if !treeRoot}
		<EmptyState
			title="Sin datos"
			message="No se encontró la jerarquía para tu perfil."
		/>
	{:else}
		<div class="flex flex-col lg:flex-row gap-8">
			<!-- Tree -->
			<div class="lg:w-1/2 xl:w-2/5">
				<div class="card bg-base-100 border border-base-300">
					<div class="card-body p-0">
						<h2 class="card-title text-xs font-semibold text-base-content/50 tracking-wide">
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

			<!-- Node summary panel -->
			<div class="lg:w-1/2 xl:w-3/5">
				{#if selectedNode}
					<div class="card bg-base-100 border border-base-300">
						<div class="card-body p-0">
							<h2 class="card-title text-lg mb-4">
								{selectedNode.name}
							</h2>

							<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
								<!-- Profile -->
								<div class="flex items-center gap-3 bg-base-200 rounded-lg p-3">
									<Briefcase class="w-5 h-5 text-base-content/40" />
									<div>
										<p class="text-xs text-base-content/40">Perfil</p>
										<p class="text-sm font-medium capitalize">
											{PROFILE_LABELS[selectedNode.profileId as keyof typeof PROFILE_LABELS] ?? selectedNode.profileId}
										</p>
									</div>
								</div>

								<!-- Area -->
								<div class="flex items-center gap-3 bg-base-200 rounded-lg p-3">
									<MapPin class="w-5 h-5 text-base-content/40" />
									<div>
										<p class="text-xs text-base-content/40">Area</p>
										<p class="text-sm font-medium">{getArea(selectedNode)}</p>
									</div>
								</div>

								<!-- Goal percentage -->
								<div class="flex items-center gap-3 bg-base-200 rounded-lg p-3">
									<Target class="w-5 h-5 text-base-content/40" />
									<div>
										<p class="text-xs text-base-content/40">Meta global</p>
										<p class="text-sm font-medium">{getGoalPercent(selectedNode)} %</p>
									</div>
								</div>

								<!-- Link to competencies -->
								<div class="flex items-center gap-3 bg-base-200 rounded-lg p-3">
									<ArrowUpRight class="w-5 h-5 text-base-content/40" />
									<div>
										<p class="text-xs text-base-content/40">Competencias</p>
										<a
											href="/evaluacion/9x9/competencias/{selectedNode.id}"
											class="link link-primary text-sm font-medium"
										>
											Ver red
										</a>
									</div>
								</div>
							</div>
						</div>
					</div>
				{:else}
					<div class="card bg-base-100 border border-base-300">
						<div class="card-body p-8 text-center">
							<Users class="w-10 h-10 text-base-content/20 mx-auto mb-3" />
							<p class="text-sm text-base-content/40">
								Selecciona un nodo del organigrama para ver sus detalles.
							</p>
						</div>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>
