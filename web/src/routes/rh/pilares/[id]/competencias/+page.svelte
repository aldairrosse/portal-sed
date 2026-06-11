<script lang="ts">
	import { ArrowLeft, Plus } from '@lucide/svelte';
	import { page } from '$app/stores';
	import {
		getPillars,
		getCompetenciesByPillar,
		addCompetency,
		updateCompetency,
		deleteCompetency,
		load,
		isLoading,
		getError
	} from '$lib/stores/competencyStore.svelte';
	import type { Competency } from '$lib/types/competency';
	import CompetencyTable from '$lib/components/competency/CompetencyTable.svelte';
	import ConfirmDeleteModal from '$lib/components/competency/ConfirmDeleteModal.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import ErrorState from '$lib/components/ui/ErrorState.svelte';
	import PageSkeleton from '$lib/components/ui/PageSkeleton.svelte';
	import * as notifications from '$lib/stores/notifications.svelte';

	const pillarId = $derived($page.params.id);
	const pillars = $derived(getPillars());
	const pillar = $derived(pillars.find((p) => p.id === pillarId));
	const competencies = $derived(getCompetenciesByPillar(pillarId));
	const loading = $derived(isLoading());
	const storeError = $derived(getError());

	let editingId = $state<string | null>(null);
	let deletingCompetency: Competency | null = $state(null);

	let isAnyInlineEditing = $derived(editingId !== null);
	let compLen = $derived(competencies.length);

	$effect(() => {
		load();
	});

	function generateId(): string {
		return crypto.randomUUID();
	}

	function handleNew() {
		editingId = '__new__';
	}

	function handleSave(data: { name: string; description: string; id?: string }) {
		if (data.id) {
			updateCompetency(data.id, { name: data.name, description: data.description });
			notifications.success(`Competencia "${data.name}" actualizada correctamente.`);
		} else {
			const newComp: Competency = {
				id: generateId(),
				pillarId,
				name: data.name,
				description: data.description
			};
			addCompetency(newComp);
			notifications.success(`Competencia "${data.name}" creada correctamente.`);
		}
	}

	function handleDelete(competency: Competency) {
		deletingCompetency = competency;
	}

	function handleDeleteConfirm() {
		if (!deletingCompetency) return;
		const { name } = deletingCompetency;
		deleteCompetency(deletingCompetency.id);
		notifications.success(`Competencia "${name}" eliminada correctamente.`);
		deletingCompetency = null;
	}

	function handleDeleteCancel() {
		deletingCompetency = null;
	}
</script>

<svelte:head>
	<title>{pillar?.name ?? 'Cargando...'} — Competencias — SED</title>
</svelte:head>

<div class="max-w-4xl mx-auto">
	<!-- Back link -->
	<a href="/rh/pilares" class="link link-hover text-base-content/50 text-sm flex items-center gap-1 mb-4">
		<ArrowLeft class="w-4 h-4" />
		Volver a pilares
	</a>

	{#if pillar}
		<div class="flex items-center justify-between mb-6">
			<div>
				<h1 class="text-2xl font-bold text-base-content">{pillar.name}</h1>
				<p class="text-base-content/50 text-sm mt-1">{pillar.description}</p>
			</div>
			<button class="btn btn-primary btn-sm" onclick={handleNew} disabled={isAnyInlineEditing}>
				<Plus class="w-4 h-4" />
				Nueva competencia
			</button>
		</div>
	{:else if !loading}
		<EmptyState title="Pilar no encontrado" message="El pilar especificado no existe." />
	{/if}

	{#if loading}
		<PageSkeleton rows={3} />
	{:else if storeError}
		<ErrorState message={storeError} onretry={load} />
	{:else if pillar && compLen === 0 && editingId !== '__new__'}
		<EmptyState
			title="Sin competencias"
			message="Este pilar aún no tiene competencias definidas."
		/>
	{:else if pillar}
		<CompetencyTable
			competencies={competencies}
			{pillarId}
			pillarName={pillar.name}
			bind:editingId
			onSave={handleSave}
			onDelete={handleDelete}
		/>
	{/if}
</div>

{#if deletingCompetency}
	<ConfirmDeleteModal
		open={true}
		title="Eliminar competencia"
		message="Se eliminará esta competencia y todos sus criterios de escala asociados. Esta acción no se puede deshacer."
		itemName={deletingCompetency.name}
		onConfirm={handleDeleteConfirm}
		onCancel={handleDeleteCancel}
	/>
{/if}
