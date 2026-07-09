<script lang="ts">
	import { Plus, Award } from '@lucide/svelte';
	import {
		getPillars,
		addPillar,
		updatePillar,
		deletePillar,
		load,
		isLoading,
		getError
	} from '$lib/stores/competencyStore.svelte';
	import type { Pillar } from '$lib/types/competency';
	import PillarTable from '$lib/components/competency/PillarTable.svelte';
	import ConfirmDeleteModal from '$lib/components/competency/ConfirmDeleteModal.svelte';
	import PageSkeleton from '$lib/components/ui/PageSkeleton.svelte';
	import ErrorState from '$lib/components/ui/ErrorState.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import * as notifications from '$lib/stores/notifications.svelte';

	const pillars = $derived(getPillars());
	const loading = $derived(isLoading());
	const storeError = $derived(getError());
	let editingId = $state<string | null>(null);
	let deletingPillar: Pillar | null = $state(null);

	let isAnyInlineEditing = $derived(editingId !== null);
	let pillarlen = $derived(pillars.length);

	$effect(() => {
		load();
	});

	function generateId(): string {
		return crypto.randomUUID();
	}

	function handleNew() {
		editingId = '__new__';
	}

	async function handlePillarSave(data: { name: string; description: string; id?: string }) {
		if (data.id) {
			try {
				await updatePillar(data.id, { name: data.name, description: data.description });
				notifications.success(`Pilar "${data.name}" actualizado correctamente.`);
			} catch (e) {
				notifications.error(e instanceof Error ? e.message : 'Error al actualizar pilar');
			}
		} else {
			const newPillar: Pillar = { id: generateId(), name: data.name, description: data.description, updatedAt: new Date().toISOString() };
			try {
				await addPillar(newPillar);
				notifications.success(`Pilar "${data.name}" creado correctamente.`);
			} catch (e) {
				notifications.error(e instanceof Error ? e.message : 'Error al crear pilar');
			}
		}
	}

	function handleDelete(pillar: Pillar) {
		deletingPillar = pillar;
	}

	async function handleDeleteConfirm() {
		if (!deletingPillar) return;
		const { name } = deletingPillar;
		try {
			await deletePillar(deletingPillar.id);
			notifications.success(`Pilar "${name}" eliminado correctamente.`);
		} catch (e) {
			notifications.error(e instanceof Error ? e.message : 'Error al eliminar pilar');
		}
		deletingPillar = null;
	}

	function handleDeleteCancel() {
		deletingPillar = null;
	}
</script>

<svelte:head>
	<title>Pilares — SED</title>
</svelte:head>

<div class="max-w-4xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-base-content flex items-center gap-2">
				<Award class="w-6 h-6" />
				Pilares
			</h1>
			<p class="text-base-content/50 text-sm mt-1">
				Gestiona los pilares del marco de competencias.
			</p>
		</div>
		<button class="btn btn-primary btn-sm" onclick={handleNew} disabled={isAnyInlineEditing}>
			<Plus class="w-4 h-4" />
			Nuevo pilar
		</button>
	</div>

	{#if loading}
		<PageSkeleton rows={3} />
	{:else if storeError}
		<ErrorState message={storeError} onretry={load} />
	{:else if pillarlen === 0 && editingId !== '__new__'}
		<EmptyState
			title="Sin pilares"
			message="Aún no hay pilares creados. Crea el primer pilar para comenzar."
		/>
	{:else}
		<PillarTable {pillars} bind:editingId onSave={handlePillarSave} onDelete={handleDelete} />
	{/if}
</div>

{#if deletingPillar}
	<ConfirmDeleteModal
		open={true}
		title="Eliminar pilar"
		message="Se eliminará este pilar y todas sus competencias asociadas. Esta acción no se puede deshacer."
		itemName={deletingPillar.name}
		onConfirm={handleDeleteConfirm}
		onCancel={handleDeleteCancel}
	/>
{/if}
