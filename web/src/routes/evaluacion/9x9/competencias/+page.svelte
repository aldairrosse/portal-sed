<script lang="ts">
	import { onMount } from 'svelte';
	import { ChartColumn, ChevronLeft, ChevronRight } from '@lucide/svelte';
	import PageSkeleton from '$lib/components/ui/PageSkeleton.svelte';
	import ErrorState from '$lib/components/ui/ErrorState.svelte';
	import { titleCase } from '$lib/utils/text';
	import { getActiveCycle, loadCycles } from '$lib/stores/cycleStore.svelte';
	import {
		getItems,
		isLoading,
		getError,
		hasMoreItems,
		hasPrevItems,
		getCurrentPage,
		getTotalCount,
		init,
		load,
		search,
		next,
		prev,
	} from '$lib/stores/competencyResultsStore.svelte';

	const items = $derived(getItems());
	const loading = $derived(isLoading());
	const storeError = $derived(getError());
	const hasMore = $derived(hasMoreItems());
	const hasPrev = $derived(hasPrevItems());
	const currentPage = $derived(getCurrentPage());
	const totalCount = $derived(getTotalCount());

	let inputQuery = $state('');

	onMount(async () => {
		await loadCycles();
		const cycle = getActiveCycle();
		if (cycle?.id) {
			init(cycle.id);
		}
	});

	function handleSearch(e: Event) {
		const val = (e.target as HTMLInputElement).value;
		inputQuery = val;
		search(val);
	}

	function statusBadge(status: string): { label: string; class: string } {
		switch (status) {
			case 'completada':
				return { label: 'Completada', class: 'badge-success' };
			case 'autoevaluacion':
				return { label: 'Autoevaluación', class: 'badge-warning' };
			case 'pendiente':
				return { label: 'Pendiente', class: 'badge-ghost' };
			default:
				return { label: 'Sin datos', class: 'badge-ghost' };
		}
	}
</script>

<svelte:head>
	<title>Resultados de competencias — SED</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<!-- Header -->
	<div>
		<h1 class="text-2xl font-bold text-base-content flex items-center gap-2">
			<ChartColumn class="w-6 h-6" />
			Resultados de competencias
		</h1>
		<p class="text-sm text-base-content/50 mt-1">
			Promedios de autoevaluación y evaluación RH por empleado
		</p>
	</div>

	<!-- Controls bar — ALWAYS visible outside loading conditional -->
	<div class="flex items-center justify-between gap-2">
		<div class="flex items-center gap-3">
			<input
				type="text"
				class="input input-bordered input-sm max-w-xs"
				placeholder="Buscar empleado..."
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

	<!-- Content -->
	{#if loading && items.length === 0}
		<PageSkeleton variant="table" rows={5} />
	{:else if storeError}
		<ErrorState message={storeError} onretry={() => load()} />
	{:else if items.length === 0 && !loading}
		<p class="text-sm text-base-content/30 italic text-center py-8">
			Sin resultados de competencias para mostrar
		</p>
	{:else}
		<div class="card bg-base-100 shadow-sm border border-base-200">
			<div class="overflow-x-auto">
				<table class="table table-sm">
					<thead>
						<tr>
							<th class="text-xs font-semibold text-base-content/60">Empleado</th>
							<th class="text-xs font-semibold text-base-content/60">Perfil</th>
							<th class="text-xs font-semibold text-base-content/60 text-center">Autoevaluación</th>
							<th class="text-xs font-semibold text-base-content/60 text-center">RH</th>
							<th class="text-xs font-semibold text-base-content/60 text-center">Estado</th>
							<th class="w-10"></th>
						</tr>
					</thead>
					<tbody>
						{#each items as item (item.id)}
							{@const badge = statusBadge(item.status)}
							<tr class="hover:bg-base-200/50 transition-colors">
								<td>
									<div class="flex items-center gap-2.5">
										<div class="avatar avatar-placeholder">
											<div
												class="bg-primary text-primary-content w-8 rounded-full flex items-center justify-center"
											>
												<span class="text-xs font-bold">
													{item.name.charAt(0).toUpperCase()}
												</span>
											</div>
										</div>
										<span class="font-medium text-sm">{item.name}</span>
									</div>
								</td>
								<td>
									<span class="text-xs text-base-content/50">{titleCase(item.profileName)}</span>
								</td>
								<td class="text-center">
									<span
										class="text-sm font-mono {item.selfRatingAvg != null
											? 'text-base-content'
											: 'text-base-content/30'}"
									>
										{item.selfRatingAvg?.toFixed(1) ?? '—'}
									</span>
								</td>
								<td class="text-center">
									<span
										class="text-sm font-mono {item.rhRatingAvg != null
											? 'text-base-content'
											: 'text-base-content/30'}"
									>
										{item.rhRatingAvg?.toFixed(1) ?? '—'}
									</span>
								</td>
								<td class="text-center">
									<span class="badge badge-sm {badge.class}">{badge.label}</span>
								</td>
								<td>
									<a
										href="/evaluacion/9x9/competencias/{item.id}"
										class="btn btn-ghost btn-square btn-xs"
										aria-label="Ver competencias de {item.name}"
									>
										<ChevronRight class="w-4 h-4" />
									</a>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>
