<script lang="ts">
	import { getProfile } from '$lib/stores/devContext.svelte';
	import {
		getMatrixEntries,
		getAllEntries,
		getQuadrantDefs
	} from '$lib/stores/nineBoxStore.svelte';
	import { getChildren, getDescendants } from '$lib/stores/orgHierarchyStore.svelte';
	import { type EvaluationProfile } from '$lib/types/evaluation';
	import type { NineBoxEntry, NineBoxScale } from '$lib/types/nine-box';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import NineBoxMatrix from '$lib/components/nine-box/NineBoxMatrix.svelte';
	import NineBoxEntryCard from '$lib/components/nine-box/NineBoxEntryCard.svelte';
	import { Grid3x3 } from '@lucide/svelte';

	// ─── Profile-to-employee mapping (dev fixtures) ───────────────────────────

	const PROFILE_NODE_ID: Partial<Record<EvaluationProfile, string>> = {
		'director-general': 'emp-dg-01',
		director: 'emp-director-01',
		jefe: 'emp-jefe-01',
		rh: 'emp-rh-01'
	};

	const MANAGER_PROFILES: EvaluationProfile[] = [
		'jefe',
		'director',
		'director-general',
		'rh'
	];

	// ─── Reactive state ───────────────────────────────────────────────────────

	const profile = $derived(getProfile());
	const isAuthorized = $derived(MANAGER_PROFILES.includes(profile));

	const scopeIds = $derived.by<string[]>(() => {
		if (!isAuthorized) return [];

		switch (profile) {
			case 'jefe': {
				const nodeId = PROFILE_NODE_ID[profile]!;
				return getChildren(nodeId).map((n) => n.id);
			}
			case 'director': {
				const nodeId = PROFILE_NODE_ID[profile]!;
				return getDescendants(nodeId).map((n) => n.id);
			}
			case 'director-general':
			case 'rh':
				return getAllEntries().map((e) => e.employeeId);
			default:
				return [];
		}
	});

	const matrixEntries = $derived<NineBoxEntry[]>(getMatrixEntries(scopeIds));
	const quadrantDefs = $derived(getQuadrantDefs());

	// ─── Cell modal state ─────────────────────────────────────────────────────

	let modalEntries = $state<NineBoxEntry[]>([]);
	let modalPerf = $state<NineBoxScale>(5);
	let modalPot = $state<NineBoxScale>(5);

	function handleCellClick(cellEntries: NineBoxEntry[], perf: NineBoxScale, pot: NineBoxScale) {
		modalEntries = cellEntries;
		modalPerf = perf;
		modalPot = pot;
	}

	function handleCloseModal() {
		modalEntries = [];
	}
</script>

<svelte:head>
	<title>Matriz 9×9 — SED</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<!-- Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-base-content flex items-center gap-2">
				<Grid3x3 class="w-6 h-6" />
				Matriz 9×9
			</h1>
			<p class="text-sm text-base-content/50 mt-1">
				Desempeño vs Potencial
			</p>
		</div>
		{#if isAuthorized}
			<span class="badge badge-ghost badge-sm">{matrixEntries.length} empleados</span>
		{/if}
	</div>

	{#if !isAuthorized}
		<EmptyState
			title="Sin acceso"
			message="No tienes permisos para ver la matriz 9×9. Esta función está disponible para jefes, directores y RH."
			actionLabel="Volver al inicio"
			actionHref="/"
		/>
	{:else if matrixEntries.length === 0}
		<EmptyState
			title="Sin evaluatees"
			message="No hay empleados en tu scope para mostrar en la matriz."
		/>
	{:else}
		<!-- Matrix -->
		<NineBoxMatrix
			entries={matrixEntries}
			{quadrantDefs}
			onCellClick={handleCellClick}
		/>
	{/if}
</div>

<!-- Cell detail modal -->
{#if modalEntries.length > 0}
	<NineBoxEntryCard
		entries={modalEntries}
		perf={modalPerf}
		pot={modalPot}
		onClose={handleCloseModal}
	/>
{/if}
