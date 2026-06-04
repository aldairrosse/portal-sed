<script lang="ts">
	import { X, Plus, Trash2, Save } from '@lucide/svelte';
	import {
		getScaleCriteriaForCell,
		updateScaleCriterion,
		addScaleCriterion,
		removeScaleCriterion
	} from '$lib/stores/competencyStore.svelte';
	import type { ScaleCriterion } from '$lib/types/competency';

	interface Props {
		open: boolean;
		competencyName: string;
		pillarName: string;
		competencyId: string;
		pillarId: string;
		onSave: () => void;
		onCancel: () => void;
	}

	let { open, competencyName, pillarName, competencyId, pillarId, onSave, onCancel }: Props = $props();

	let dialogEl: HTMLDialogElement | undefined = $state();
	let tempIdCounter = $state(0);

	const levels = [1, 2, 3, 4, 5] as const;

	// Each entry: { localId: string, serverId?: string, level: number, description: string }
	let entries: Array<{ localId: string; serverId: string | null; level: number; description: string }> = $state([]);

	$effect(() => {
		if (!dialogEl) return;
		if (open) {
			const existing = getScaleCriteriaForCell(competencyId, pillarId);
			entries = existing.map((c) => ({
				localId: c.id,
				serverId: c.id,
				level: c.level,
				description: c.description
			}));
			tempIdCounter = 0;
			dialogEl.showModal();
		} else {
			dialogEl.close();
		}
	});

	function getEntriesByLevel(level: number) {
		return entries.filter((e) => e.level === level);
	}

	function addEntry(level: number) {
		tempIdCounter++;
		entries = [
			...entries,
			{ localId: `new-${tempIdCounter}`, serverId: null, level, description: '' }
		];
	}

	function removeEntry(localId: string) {
		entries = entries.filter((e) => e.localId !== localId);
	}

	function updateEntry(localId: string, description: string) {
		entries = entries.map((e) => (e.localId === localId ? { ...e, description } : e));
	}

	function handleCancel() {
		onCancel();
	}

	function handleBackdropClick(e: MouseEvent) {
		if (e.target === dialogEl) {
			handleCancel();
		}
	}

	function handleSubmit(e: Event) {
		e.preventDefault();
		// Get existing criteria IDs before save (snapshot for diff)
		const existingIds = new Set(
			getScaleCriteriaForCell(competencyId, pillarId).map((c) => c.id)
		);
		// eslint-disable-next-line svelte/prefer-svelte-reactivity
		const finalIds = new Set<string>();
		const newEntries: Array<Omit<ScaleCriterion, 'id'>> = [];
		const updatePairs: Array<{ id: string; description: string }> = [];

		for (const entry of entries) {
			if (entry.serverId) {
				finalIds.add(entry.serverId);
				updatePairs.push({ id: entry.serverId, description: entry.description.trim() });
			} else {
				newEntries.push({
					competencyId,
					pillarId,
					level: entry.level as 1 | 2 | 3 | 4 | 5,
					description: entry.description.trim()
				});
			}
		}

		// Remove deleted
		for (const id of existingIds) {
			if (!finalIds.has(id)) {
				removeScaleCriterion(id);
			}
		}

		// Update changed
		for (const pair of updatePairs) {
			updateScaleCriterion(pair.id, pair.description);
		}

		// Add new
		for (const nc of newEntries) {
			addScaleCriterion(nc);
		}

		onSave();
	}
</script>

<dialog
	bind:this={dialogEl}
	class="modal"
	class:modal-open={open}
	aria-modal="true"
	aria-labelledby="scale-criterion-title"
	onclick={handleBackdropClick}
	onclose={handleCancel}
>
	<div class="modal-box max-w-2xl">
		<div class="flex items-center justify-between mb-2">
			<h3 id="scale-criterion-title" class="text-lg font-semibold text-base-content">
				Criterios de escala
			</h3>
			<button
				class="btn btn-ghost btn-square btn-sm"
				onclick={handleCancel}
				aria-label="Cerrar"
			>
				<X class="w-4 h-4" />
			</button>
		</div>

		<p class="text-sm text-base-content/50 mb-5">
			{competencyName} &mdash; {pillarName}
		</p>

		<form onsubmit={handleSubmit}>
			<div class="space-y-6">
				{#each levels as level (level)}
					{@const levelEntries = getEntriesByLevel(level)}
					<fieldset class="rounded-lg border border-base-300 p-4">
						<legend class="text-sm font-semibold text-base-content px-1">Nivel {level}</legend>

						{#if levelEntries.length === 0}
							<p class="text-xs text-base-content/30 italic mb-3">Sin criterios definidos</p>
						{/if}

						<div class="space-y-3">
							{#each levelEntries as entry (entry.localId)}
								<div class="flex items-start gap-2">
									<textarea
										class="textarea textarea-bordered w-full text-sm"
										rows={2}
										value={entry.description}
										oninput={(e) => updateEntry(entry.localId, (e.target as HTMLTextAreaElement).value)}
										placeholder="Describe el comportamiento esperado en este nivel"
										aria-label="Descripción nivel {level}"
									></textarea>
									<button
										type="button"
										class="btn btn-ghost btn-square btn-sm text-error flex-shrink-0 mt-1"
										onclick={() => removeEntry(entry.localId)}
										aria-label="Eliminar criterio nivel {level}"
									>
										<Trash2 class="w-4 h-4" />
									</button>
								</div>
							{/each}
						</div>

						<button
							type="button"
							class="btn btn-ghost btn-xs mt-2 text-primary"
							onclick={() => addEntry(level)}
						>
							<Plus class="w-3 h-3" />
							Agregar criterio
						</button>
					</fieldset>
				{/each}
			</div>

			<div class="modal-action mt-6">
				<button type="button" class="btn btn-ghost btn-sm" onclick={handleCancel}>Cancelar</button>
				<button type="submit" class="btn btn-primary btn-sm">
					<Save class="w-4 h-4" />
					Guardar cambios
				</button>
			</div>
		</form>
	</div>
</dialog>
