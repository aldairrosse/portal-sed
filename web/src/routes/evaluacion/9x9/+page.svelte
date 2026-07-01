<script lang="ts">
	import { untrack } from 'svelte';
	import { getProfile } from '$lib/stores/devContext.svelte';
	import {
		load,
		getMatrixEntries,
		getAllEntries,
		getQuadrantDefs,
		getQuadrantDef,
		isLoading,
		getError,
		reload,
		markReady
	} from '$lib/stores/nineBoxStore.svelte';
	import { getDescendants } from '$lib/stores/orgHierarchyStore.svelte';
	import { loadCycles, getActiveCycle, getError as cycleError } from '$lib/stores/cycleStore.svelte';
	import { loadPhases, getPhaseId } from '$lib/stores/phaseStore.svelte';
	import { type EvaluationProfile } from '$lib/types/evaluation';
	import type { NineBoxEntry, NineBoxTier } from '$lib/types/nine-box';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import NineBoxMatrix from '$lib/components/nine-box/NineBoxMatrix.svelte';
	import NineBoxEntryCard from '$lib/components/nine-box/NineBoxEntryCard.svelte';
	import NineBoxCellConfig from '$lib/components/nine-box/NineBoxCellConfig.svelte';
	import { Grid3x3, Settings } from '@lucide/svelte';

	// ─── Phase options ─────────────────────────────────────────────────────────

	const NINEBOX_PHASES = [
		{ id: 'medio-anio', phaseEnum: 'avance', label: 'Avance medio año', shortLabel: 'Avance' },
		{ id: 'fin-anio', phaseEnum: 'cierre', label: 'Cierre fin de año', shortLabel: 'Evaluación' }
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

	let selectedPhase = $state<NineBoxPhaseId>('medio-anio');

	// Resolve UUIDs for cycle and phase
	const activeCycle = $derived(getActiveCycle());
	const selectedPhaseEnum = $derived(
		NINEBOX_PHASES.find((p) => p.id === selectedPhase)?.phaseEnum
	);
	const phaseUUID = $derived(selectedPhaseEnum ? getPhaseId(selectedPhaseEnum) : undefined);

	// Load prerequisite data once, then nine-box data when UUIDs are ready
	$effect(() => {
		if (isAuthorized) {
			// ponytail: untrack prevents $effect from tracking reads inside loadCycles/loadPhases
			// (phaseDefinitions, cycles are $state — without untrack, reassignment triggers
			// infinite effect re-fire → 489+ requests to /phases).
			untrack(() => {
				loadCycles();
				loadPhases();
			});
		}
	});

	$effect(() => {
		if (isAuthorized) {
			if (activeCycle && phaseUUID) {
				load(activeCycle.id, phaseUUID);
			} else {
				// No cycle/phase available — show empty grid instead of infinite skeleton
				markReady();
			}
		}
	});

	const scopeIds = $derived.by<string[]>(() => {
		if (!isAuthorized) return [];

		switch (profile) {
		case 'jefe': {
			const nodeId = PROFILE_NODE_ID[profile]!;
			return getDescendants(nodeId).map((n) => n.id);
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

	// Prereq state: cycles and phases loaded (even if empty — show matrix anyway)
	const prereqError = $derived(cycleError());

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
	<title>{phaseLabel} — Matriz 9-Box</title>
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
				Desempeño vs Potencial
			</p>
		</div>
		<!-- Phase selector -->
	<div role="tablist" class="tabs tabs-box gap-0">
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
	</div>

	{#if !isAuthorized}
		<EmptyState
			title="Sin acceso"
			message="No tienes permisos para ver la matriz 9-Box. Esta función está disponible para jefes, directores y RH."
			actionLabel="Volver al inicio"
			actionHref="/"
		/>
	{:else}
		<span class="badge badge-ghost badge-sm mx-auto">{matrixEntries.length} empleados</span>

		<!-- Warning banners (non-blocking) -->
		{#if prereqError}
			<div class="alert alert-warning">
				<span>{prereqError}</span>
			</div>
		{/if}
		{#if error}
			<div class="alert alert-error">
				<span>{error}</span>
				<button class="btn btn-xs btn-ghost" onclick={reload}>Reintentar</button>
			</div>
		{/if}
		{#if loading}
			<div class="text-xs text-base-content/40">Cargando datos...</div>
		{/if}

		<!-- Always show matrix (empty or populated) -->
		<NineBoxMatrix
			entries={matrixEntries}
			{quadrantDefs}
			onCellClick={handleCellClick}
		/>

		{#if matrixEntries.length === 0 && !loading}
			<p class="text-sm text-base-content/50 text-center mt-2">
				No hay empleados para mostrar en la matriz.
			</p>
		{/if}

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
