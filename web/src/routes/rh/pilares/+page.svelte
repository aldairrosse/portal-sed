<script lang="ts">
	import { Plus } from '@lucide/svelte';
	import { getPillars, addPillar, updatePillar, deletePillar } from '$lib/stores/competencyStore.svelte';
	import type { Pillar } from '$lib/types/competency';
	import PillarTable from '$lib/components/competency/PillarTable.svelte';
	import PillarFormModal from '$lib/components/competency/PillarFormModal.svelte';
	import ConfirmDeleteModal from '$lib/components/competency/ConfirmDeleteModal.svelte';
	import PageSkeleton from '$lib/components/ui/PageSkeleton.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';

	const pillars = $derived(getPillars());
	let loading = $state(true);
	let showFormModal = $state(false);
	let editingPillar: Pillar | null = $state(null);
	let deletingPillar: Pillar | null = $state(null);
	let successMsg = $state('');

	let pillarlen = $derived(pillars.length);

	// Simulate initial load
	$effect(() => {
		const t = setTimeout(() => (loading = false), 300);
		return () => clearTimeout(t);
	});

	$effect(() => {
		if (successMsg) {
			const t = setTimeout(() => (successMsg = ''), 3000);
			return () => clearTimeout(t);
		}
	});

	function generateId(): string {
		return crypto.randomUUID();
	}

	function handleNew() {
		editingPillar = null;
		showFormModal = true;
	}

	function handleEdit(pillar: Pillar) {
		editingPillar = pillar;
		showFormModal = true;
	}

	function handleFormSave(data: { name: string; description: string }) {
		if (editingPillar) {
			updatePillar(editingPillar.id, data);
			successMsg = `Pilar "{data.name}" actualizado correctamente.`;
		} else {
			const newPillar: Pillar = {
				id: generateId(),
				...data
			};
			addPillar(newPillar);
			successMsg = `Pilar "{data.name}" creado correctamente.`;
		}
		showFormModal = false;
		editingPillar = null;
	}

	function handleFormCancel() {
		showFormModal = false;
		editingPillar = null;
	}

	function handleDelete(pillar: Pillar) {
		deletingPillar = pillar;
	}

	function handleDeleteConfirm() {
		if (!deletingPillar) return;
		const name = deletingPillar.name;
		deletePillar(deletingPillar.id);
		successMsg = `Pilar "{name}" eliminado correctamente.`;
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
			<h1 class="text-2xl font-bold text-base-content">Pilares</h1>
			<p class="text-base-content/50 text-sm mt-1">
				Gestiona los pilares del marco de competencias.
			</p>
		</div>
		<button class="btn btn-primary btn-sm" onclick={handleNew}>
			<Plus class="w-4 h-4" />
			Nuevo pilar
		</button>
	</div>

	{#if successMsg}
		<div class="alert alert-success mb-4 text-sm" role="status">
			<span>{successMsg}</span>
		</div>
	{/if}

	{#if loading}
		<PageSkeleton rows={3} />
	{:else if pillarlen === 0}
		<EmptyState
			title="Sin pilares"
			message="Aún no hay pilares creados. Crea el primer pilar para comenzar."
		/>
	{:else}
		<PillarTable {pillars} onEdit={handleEdit} onDelete={handleDelete} />
	{/if}
</div>

<PillarFormModal
	open={showFormModal}
	pillar={editingPillar}
	onSave={handleFormSave}
	onCancel={handleFormCancel}
/>

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
