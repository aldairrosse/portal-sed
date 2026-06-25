<script lang="ts">
	import { getProfile } from '$lib/stores/devContext.svelte';
	import {
		load,
		getMatrixEntries,
		getAllEntries,
		getQuadrantDefs,
		getQuadrantDef,
		getCurrentCycleId,
		getCurrentPhaseId,
		isLoading,
		getError,
		reload
	} from '$lib/stores/nineBoxStore.svelte';
	import { getChildren, getDescendants } from '$lib/stores/orgHierarchyStore.svelte';
	import { type EvaluationProfile, CYCLE_PHASES } from '$lib/types/evaluation';
	import type { NineBoxEntry, NineBoxTier } from '$lib/types/nine-box';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import PageSkeleton from '$lib/components/ui/PageSkeleton.svelte';
	import ErrorState from '$lib/components/ui/ErrorState.svelte';
	import NineBoxMatrix from '$lib/components/nine-box/NineBoxMatrix.svelte';
	import NineBoxEntryCard from '$lib/components/nine-box/NineBoxEntryCard.svelte';
	import NineBoxCellConfig from '$lib/components/nine-box/NineBoxCellConfig.svelte';
	import { Grid3x3, Settings } from '@lucide/svelte';

	// ─── Phase options ─────────────────────────────────────────────────────────

	const NINEBOX_PHASES = [
		{ id: 'medio-anio', label: 'Avance medio año', shortLabel: 'Avance' },
		{ id: 'fin-anio', label: 'Cierre fin de año', shortLabel: 'Evaluación' }
	] as const;

	type NineBoxPhaseId = (typeof NINEBOX_PHASES)[number]['id'];

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

	// ─── Reactive state ──────────────────────────────────────────────────────

	const profile = $derived(getProfile());
	const isAuthorized = $derived(MANAGER_PROFILES.includes(profile));
	const isRH = $derived(profile === 'rh');

	const DEFAULT_CYCLE_ID = '2026';

	let selectedPhase = $state<NineBoxPhaseId>('medio-anio');

	// Load data on mount and when phase changes
	$effect(() => {
		if (isAuthorized) {
			load(DEFAULT_CYCLE_ID, selectedPhase);
		}
	});

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

	const loading = $derived(isLoading());
	const error = $derived(getError());
	const matrixEntries = $derived<NineBoxEntry[]>(getMatrixEntries(scopeIds));
	const quadrantDefs = $derived(getQuadrantDefs());

	const phaseLabel = $derived(
		NINEBOX_PHASES.find((p) => p.id === selectedPhase)?.shortLabel ?? ''
	);

	// ─── Cell modal state ────────────────────────────────────────────────────

	let modalEntries = $state<NineBoxEntry[]>([]);
	let modalPerfTier = $state<NineBoxTier>(2);
	let modalPotTier = $state<NineBoxTier>(2);

	function handleCellClick(cellEntries: NineBoxEntry[], perfTier: NineBoxTier, potTier: NineBoxTier) {
		modalEntries = cellEntries;
		modalPerfTier = perfTier;
		modalPotTier = potTier;
	}

	function handleCloseModal() {
		modalEntries = [];
	}

	// ─── Quadrant config modal state ─────────────────────────────────────────

	let configQuadrant = $state<number | null>(null);

	function handleOpenConfig(quadrant: number) {
		configQuadrant = quadrant;
	}

	function handleCloseConfig() {
		configQuadrant = null;
	}
</script>

<svelte:head>
	<title>{phaseLabel} — Matriz 9-Box — SED</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<!-- Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-base-content flex items-center gap-2">
				<Grid3x3 class="w-6 h-6" />
				Matriz 9-Box
			</h1>
			<p class="text-sm text-base-content/50 mt-1">
				{phaseLabel} — Desempeño vs Potencial
			</p>
		</div>
		{#if isAuthorized}
			<span class="badge badge-ghost badge-sm">{matrixEntries.length} empleados</span>
		{/if}
	</div>

	<!-- Phase selector -->
	<div role="tablist" class="tabs tabs-bordered gap-0">
		{#each NINEBOX_PHASES as phase (phase.id)}
			<button
				role="tab"
				type="button"
				class="tab {selectedPhase === phase.id ? 'tab-active' : ''}"
				onclick={() => { selectedPhase = phase.id; }}
			>
				{phase.label}
			</button>
		{/each}
	</div>

	{#if loading}
		<PageSkeleton variant="card" rows={3} />
	{:else if error}
		<ErrorState
			title="Error al cargar la matriz"
			message={error}
			retryLabel="Reintentar"
			onretry={reload}
		/>
	{:else if !isAuthorized}
		<EmptyState
			title="Sin acceso"
			message="No tienes permisos para ver la matriz 9-Box. Esta función está disponible para jefes, directores y RH."
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

		<!-- RH: Config button per quadrant -->
		{#if isRH}
			<div class="flex flex-wrap gap-2 mt-2">
				{#each quadrantDefs as qd (qd.quadrant)}
					<button
						type="button"
						class="btn btn-xs btn-ghost gap-1"
						style="border-left: 3px solid {qd.colorHex};"
						onclick={() => handleOpenConfig(qd.quadrant)}
					>
						<Settings class="w-3 h-3" />
						Q{qd.quadrant}: {qd.label}
					</button>
				{/each}
			</div>
		{/if}
	{/if}
</div>

<!-- Cell detail modal -->
{#if modalEntries.length > 0}
	<NineBoxEntryCard
		entries={modalEntries}
		perfTier={modalPerfTier}
		potTier={modalPotTier}
		onClose={handleCloseModal}
	/>
{/if}

<!-- Quadrant config modal (RH only) -->
{#if isRH && configQuadrant !== null}
	{@const qDef = getQuadrantDef(configQuadrant)}
	{#if qDef}
		<NineBoxCellConfig
			quadrantDef={qDef}
			onClose={handleCloseConfig}
		/>
	{/if}
{/if}
