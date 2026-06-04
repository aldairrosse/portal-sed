<script lang="ts">
	import { Save, X } from '@lucide/svelte';
	import { getLevelDefinitions, updateLevelDefinition } from '$lib/stores/competencyStore.svelte';

	interface Props {
		open: boolean;
		onClose: () => void;
	}

	let { open, onClose }: Props = $props();

	let dialogEl: HTMLDialogElement | undefined = $state();

	const levels = [1, 2, 3, 4, 5] as const;

	// Local editing state per level
	let editLabels: Record<number, string> = $state({});
	let editDescriptions: Record<number, string> = $state({});
	let hasChanges = $state(false);
	let successMsg = $state('');

	function resetForm() {
		const defs = getLevelDefinitions();
		levels.forEach((l) => {
			const existing = defs.find((d) => d.level === l);
			editLabels[l] = existing?.label ?? '';
			editDescriptions[l] = existing?.description ?? '';
		});
		hasChanges = false;
	}

	$effect(() => {
		if (!dialogEl) return;
		if (open) {
			resetForm();
			dialogEl.showModal();
		} else {
			dialogEl.close();
		}
	});

	function markChanged() {
		hasChanges = true;
	}

	function handleSave() {
		levels.forEach((level) => {
			updateLevelDefinition(level, editLabels[level].trim(), editDescriptions[level].trim());
		});
		hasChanges = false;
		successMsg = 'Definiciones de nivel guardadas correctamente.';
		setTimeout(() => (successMsg = ''), 3000);
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

		{#if successMsg}
			<div class="alert alert-success mb-4 text-sm flex-shrink-0" role="status">
				<span>{successMsg}</span>
			</div>
		{/if}

		<div class="overflow-y-auto flex-1 pr-1 space-y-5">
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
							oninput={markChanged}
							placeholder="Etiqueta del nivel"
							aria-label="Etiqueta nivel {level}"
						/>
						<textarea
							class="textarea textarea-bordered textarea-sm w-full"
							rows={2}
							bind:value={editDescriptions[level]}
							oninput={markChanged}
							placeholder="Descripción del nivel"
							aria-label="Descripción nivel {level}"
						></textarea>
					</div>
				</div>
			{/each}
		</div>

		<div class="modal-action mt-4 flex-shrink-0">
			<button class="btn btn-ghost btn-sm" onclick={handleClose}>Cancelar</button>
			<button class="btn btn-primary btn-sm" onclick={handleSave} disabled={!hasChanges}>
				<Save class="w-4 h-4" />
				Guardar cambios
			</button>
		</div>
	</div>
</dialog>
