<script lang="ts">
	import { onMount } from 'svelte';
	import EmployeeEvaluationDetail from '$lib/components/evaluation/EmployeeEvaluationDetail.svelte';
	import EmployeeEvaluationTable from '$lib/components/evaluation/EmployeeEvaluationTable.svelte';
	import PageSkeleton from '$lib/components/ui/PageSkeleton.svelte';
	import ErrorState from '$lib/components/ui/ErrorState.svelte';
	import { getActivePhase } from '$lib/api/cycle.svelte';
	import {
		getItems,
		isLoading,
		getError,
		hasMoreItems,
		hasPrevItems,
		getCurrentPage,
		getTotalCount,
		getSearchQuery,
		load,
		search,
		next,
		prev,
	} from '$lib/stores/rhEvaluadosStore.svelte';
	import { titleCase } from '$lib/utils/text';

	import { ClipboardList, ChevronLeft, ChevronRight } from '@lucide/svelte';

	const items = $derived(getItems());
	const loading = $derived(isLoading());
	const storeError = $derived(getError());
	const hasMore = $derived(hasMoreItems());
	const hasPrev = $derived(hasPrevItems());
	const currentPage = $derived(getCurrentPage());
	const totalCount = $derived(getTotalCount());
	const searchQuery = $derived(getSearchQuery());

	const phase = $derived(getActivePhase() ?? 'inicio-anio');
	const isFinAnio = $derived(phase === 'fin-anio');
	const isMedioAnio = $derived(phase === 'medio-anio');

	const phaseDescription = $derived(
		isFinAnio
			? 'Evaluación formal de competencias y cierre de metas'
			: isMedioAnio
				? 'Revisión de avance de objetivos y competencias de todos los empleados'
				: 'Seguimiento de objetivos de todos los empleados para el ciclo actual'
	);

	let selectedEmployeeId = $state('');
	let inputQuery = $state('');

	onMount(() => {
		load();
	});

	function handleSelect(employeeId: string) {
		selectedEmployeeId = employeeId;
	}

	function handleBack() {
		selectedEmployeeId = '';
	}

	function handleSearch(e: Event) {
		const val = (e.target as HTMLInputElement).value;
		inputQuery = val;
		search(val);
	}
</script>

<svelte:head>
	<title>Evaluaciones RH — SED</title>
</svelte:head>

<div class="flex flex-col gap-6">
	{#if !selectedEmployeeId}
		<div>
			<h1 class="text-2xl font-bold text-base-content flex items-center gap-2">
				<ClipboardList class="w-6 h-6" />
				Evaluaciones RH
			</h1>
			<p class="text-sm text-base-content/50 mt-1">
				{phaseDescription}
			</p>
		</div>
	{/if}

	{#if !selectedEmployeeId}
		<div class="flex items-center justify-between gap-2">
			<div class="w-full max-w-sm">
				<input
					type="text"
					class="input input-bordered input-sm w-full"
					placeholder="Buscar por nombre..."
					value={inputQuery}
					oninput={handleSearch}
					aria-label="Buscar empleado"
				/>
			</div>
			{#if totalCount > 0 || (inputQuery.trim() && items.length > 0)}
				<span class="text-xs text-base-content/50 whitespace-nowrap">
					{#if inputQuery.trim()}
						Viendo {items.length} resultado{items.length !== 1 ? 's' : ''}
					{:else}
						Viendo {items.length} de {totalCount} empleado{totalCount !== 1 ? 's' : ''}
					{/if}
				</span>
			{/if}
			<div class="flex items-center gap-1">
				<button
					type="button"
					class="btn btn-outline btn-xs"
					onclick={prev}
					disabled={!hasPrev || loading}
				>
					<ChevronLeft class="w-4 h-4" />
					Anterior
				</button>
				<span class="text-xs text-base-content/50 px-1">
					Pág. {currentPage + 1}
				</span>
				<button
					type="button"
					class="btn btn-outline btn-xs"
					onclick={next}
					disabled={!hasMore || loading}
				>
					Siguiente
					<ChevronRight class="w-4 h-4" />
				</button>
			</div>
		</div>
	{/if}

	{#if loading && items.length === 0}
		<PageSkeleton variant="table" rows={5} />
	{:else if storeError}
		<ErrorState message={storeError} onretry={() => load()} />
	{:else if items.length === 0 && !loading}
		<p class="text-sm text-base-content/30 italic text-center py-8">
			Sin empleados para mostrar
		</p>
	{:else}
		<EmployeeEvaluationTable
			mode="rh"
			rows={items}
			onSelect={handleSelect}
			selectedEmployeeId={selectedEmployeeId}
			disabled={!isFinAnio}
		>
			{#snippet detail()}
				{#if selectedEmployeeId}
					<EmployeeEvaluationDetail
						employeeId={selectedEmployeeId}
						viewerMode="rh"
						showBreadcrumb={true}
						onBack={handleBack}
					/>
				{/if}
			{/snippet}
		</EmployeeEvaluationTable>
	{/if}
</div>
