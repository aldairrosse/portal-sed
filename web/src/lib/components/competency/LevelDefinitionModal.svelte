<script lang="ts">
	import { Save, X } from '@lucide/svelte';
	import { untrack } from 'svelte';
	import { getLevelDefinitions, updateLevelDefinition } from '$lib/stores/competencyStore.svelte';
	import * as notifications from '$lib/stores/notifications.svelte';

	interface Props {
		open: boolean;
		onClose: () => void;
		onSaved?: () => void;
	}

	let { open, onClose, onSaved }: Props = $props();

	let dialogEl: HTMLDialogElement | undefined = $state();

	const levels = [1, 2, 3, 4, 5] as const;

	// Initialise all 5 levels up-front so editLabels[l] is never undefined
	// before resetForm runs (the $effect can fire after the first render).
	const empty = (): Record<number, string> =>
		levels.reduce<Record<number, string>>((acc, l) => ((acc[l] = ''), acc), {});

	let editLabels: Record<number, string> = $state(empty());
	let editDescriptions: Record<number, string> = $state(empty());
	// Original values at modal-open time, used to detect changes
	let originalLabels: Record<number, string> = $state(empty());
	let originalDescriptions: Record<number, string> = $state(empty());

	let saving = $state(false);

	function resetForm() {
		// ponytail: untrack — reading getLevelDefinitions() here must NOT create
		// a reactive dependency, or the $effect re-runs on every data change
		// and wipes user input.
		const defs = untrack(() => getLevelDefinitions());
		levels.forEach((l) => {
			const existing = defs.find((d) => d.level === l);
			const label = existing?.label ?? '';
			const description = existing?.description ?? '';
			editLabels[l] = label;
			editDescriptions[l] = description;
			originalLabels[l] = label;
			originalDescriptions[l] = description;
		});
	}

	// Derived: which levels differ from the original snapshot.
	// Includes levels with a label even if unchanged — matches the user's
	// intent: "tenga al menos una etiqueta" OR "uno que cambió".
	const changedLevels = $derived(
		levels.filter(
			(l) =>
				editLabels[l].trim() !== originalLabels[l].trim() ||
				editDescriptions[l].trim() !== originalDescriptions[l].trim()
		)
	);

	const hasChanges = $derived(changedLevels.length > 0);

	// Reset only when the modal opens. Depends only on `open` and `dialogEl`.
	$effect(() => {
		if (!dialogEl) return;
		if (open) {
			resetForm();
			dialogEl.showModal();
		} else {
			dialogEl.close();
		}
	});

	async function handleSaveAll() {
		// Only save levels that have a label (the API rejects empty labels).
		const toSave = changedLevels.filter((l) => editLabels[l].trim().length > 0);
		if (toSave.length === 0) {
			notifications.warning('Ingresa al menos una etiqueta para guardar.');
			return;
		}

		saving = true;
		try {
			for (const level of toSave) {
				await updateLevelDefinition(
					level,
					editLabels[level].trim(),
					editDescriptions[level].trim()
				);
			}
			// Refresh the original snapshot so the button disables after save.
			toSave.forEach((l) => {
				originalLabels[l] = editLabels[l].trim();
				originalDescriptions[l] = editDescriptions[l].trim();
			});
			notifications.success(
				toSave.length === 1
					? `Nivel ${toSave[0]} guardado correctamente.`
					: `${toSave.length} niveles guardados correctamente.`
			);
			onSaved?.();
		} catch (err) {
			console.error('[LevelDef] Save failed:', err);
			notifications.error(
				`Error al guardar: ${err instanceof Error ? err.message : String(err)}`
			);
		} finally {
			saving = false;
		}
	}

	function handleClose() {
		onClose();
	}

	function handleBackdropClick(e: MouseEvent) {
		if (e.target === dialogEl) {
			handleClose();
		}
	}
</script>

<dialog
	bind:this={dialogEl}
	class="modal"
	class:modal-open={open}
	aria-modal="true"
	aria-labelledby="level-def-title"
	onclick={handleBackdropClick}
	onclose={handleClose}
>
	<div class="modal-box max-w-3xl my-8 flex flex-col max-h-[calc(100vh-4rem)]">
		<div class="flex items-center justify-between mb-4 flex-shrink-0">
			<h3 id="level-def-title" class="text-lg font-semibold text-base-content">
				Definiciones de nivel
			</h3>
			<button
				class="btn btn-ghost btn-square btn-sm"
				onclick={handleClose}
				aria-label="Cerrar"
			>
				<X class="w-4 h-4" />
			</button>
		</div>

		<p class="text-xs text-base-content/50 mb-4 flex-shrink-0">
			Estas definiciones aplican a todos los perfiles de evaluación.
		</p>

		<div class="overflow-y-auto flex-1 pr-1 space-y-5 py-2 px-1">
		{#each levels as level (level)}
			<div class="flex items-start gap-3">
					<div
						class="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0 mt-1"
					>
						<span class="text-sm font-bold text-primary">{level}</span>
					</div>
					<div class="flex-1 space-y-2">
						<input
							type="text"
							class="input input-bordered input-sm w-full"
							bind:value={editLabels[level]}
							placeholder="Etiqueta del nivel"
							aria-label="Etiqueta nivel {level}"
						/>
						<textarea
							class="textarea textarea-bordered textarea-sm w-full"
							rows={2}
							bind:value={editDescriptions[level]}
							placeholder="Descripción del nivel"
							aria-label="Descripción nivel {level}"
						></textarea>
					</div>
				</div>
			{/each}
		</div>

		<div class="modal-action mt-4 flex-shrink-0">
			<button class="btn btn-ghost btn-sm" onclick={handleClose}>Cerrar</button>
			<button
				class="btn btn-primary btn-sm"
				onclick={handleSaveAll}
				disabled={!hasChanges || saving}
			>
				<Save class="w-4 h-4" />
				{saving ? 'Guardando...' : 'Guardar cambios'}
			</button>
		</div>
	</div>
</dialog>
