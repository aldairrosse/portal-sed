<script lang="ts">
	import { X } from '@lucide/svelte';
	import type { Competency } from '$lib/types/competency';
	import { getCompetenciesByPillar } from '$lib/stores/competencyStore.svelte';

	interface Props {
		open: boolean;
		competency: Competency | null;
		pillarId: string;
		onSave: (data: { name: string; description: string }) => void;
		onCancel: () => void;
	}

	let { open, competency, pillarId, onSave, onCancel }: Props = $props();

	let dialogEl: HTMLDialogElement | undefined = $state();
	let name = $state('');
	let description = $state('');
	let error = $state('');

	const isEditing = $derived(competency !== null);
	const title = $derived(isEditing ? 'Editar competencia' : 'Nueva competencia');

	$effect(() => {
		if (!dialogEl) return;
		if (open) {
			name = competency?.name ?? '';
			description = competency?.description ?? '';
			error = '';
			dialogEl.showModal();
		} else {
			dialogEl.close();
		}
	});

	function handleCancel() {
		onCancel();
	}

	function handleBackdropClick(e: MouseEvent) {
		if (e.target === dialogEl) {
			handleCancel();
		}
	}

	function validate(nameVal: string, idToExclude?: string): string | null {
		if (!nameVal.trim()) return 'El nombre es obligatorio.';
		if (!description.trim()) return 'La descripción es obligatoria.';
		const trimmed = nameVal.trim();
		const existing = getCompetenciesByPillar(pillarId);
		const duplicate = existing.find(
			(c) => c.name.toLowerCase() === trimmed.toLowerCase() && c.id !== idToExclude
		);
		if (duplicate) return 'Ya existe una competencia con ese nombre en este pilar.';
		return null;
	}

	function handleSubmit(e: Event) {
		e.preventDefault();
		const err = validate(name, competency?.id);
		if (err) {
			error = err;
			return;
		}
		onSave({ name: name.trim(), description: description.trim() });
	}
</script>

<dialog
	bind:this={dialogEl}
	class="modal"
	class:modal-open={open}
	aria-modal="true"
	aria-labelledby="competency-form-title"
	onclick={handleBackdropClick}
	onclose={handleCancel}
>
	<div class="modal-box">
		<div class="flex items-center justify-between mb-5">
			<h3 id="competency-form-title" class="text-lg font-semibold text-base-content">{title}</h3>
			<button
				class="btn btn-ghost btn-square btn-sm"
				onclick={handleCancel}
				aria-label="Cerrar"
			>
				<X class="w-4 h-4" />
			</button>
		</div>

		<form onsubmit={handleSubmit}>
			{#if error}
				<div class="alert alert-error mb-4 text-sm" role="alert">
					<span>{error}</span>
				</div>
			{/if}

			<div class="form-control mb-3">
				<label class="label" for="competency-name">
					<span class="label-text">Nombre</span>
				</label>
				<input
					id="competency-name"
					type="text"
					class="input input-bordered w-full"
					bind:value={name}
					placeholder="Ej: Comunicación efectiva"
					required
					aria-required="true"
				/>
			</div>

			<div class="form-control mb-3">
				<label class="label" for="competency-description">
					<span class="label-text">Descripción</span>
				</label>
				<textarea
					id="competency-description"
					class="textarea textarea-bordered w-full"
					rows={3}
					bind:value={description}
					placeholder="Describe esta competencia"
					required
					aria-required="true"
				></textarea>
			</div>

			<div class="modal-action mt-6">
				<button type="button" class="btn btn-ghost btn-sm" onclick={handleCancel}>Cancelar</button>
				<button type="submit" class="btn btn-primary btn-sm">
					{isEditing ? 'Guardar cambios' : 'Crear competencia'}
				</button>
			</div>
		</form>
	</div>
</dialog>
