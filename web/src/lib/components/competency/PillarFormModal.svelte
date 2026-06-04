<script lang="ts">
	import { X } from '@lucide/svelte';
	import type { Pillar } from '$lib/types/competency';
	import { getPillars } from '$lib/stores/competencyStore.svelte';

	interface Props {
		open: boolean;
		pillar: Pillar | null;
		onSave: (data: { name: string; description: string }) => void;
		onCancel: () => void;
	}

	let { open, pillar, onSave, onCancel }: Props = $props();

	let dialogEl: HTMLDialogElement | undefined = $state();
	let name = $state('');
	let description = $state('');
	let error = $state('');

	const isEditing = $derived(pillar !== null);
	const title = $derived(isEditing ? 'Editar pilar' : 'Nuevo pilar');

	$effect(() => {
		if (!dialogEl) return;
		if (open) {
			name = pillar?.name ?? '';
			description = pillar?.description ?? '';
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
		const existing = getPillars();
		const duplicate = existing.find(
			(p) => p.name.toLowerCase() === trimmed.toLowerCase() && p.id !== idToExclude
		);
		if (duplicate) return 'Ya existe un pilar con ese nombre.';
		return null;
	}

	function handleSubmit(e: Event) {
		e.preventDefault();
		const err = validate(name, pillar?.id);
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
	aria-labelledby="pillar-form-title"
	onclick={handleBackdropClick}
	onclose={handleCancel}
>
	<div class="modal-box">
		<div class="flex items-center justify-between mb-5">
			<h3 id="pillar-form-title" class="text-lg font-semibold text-base-content">{title}</h3>
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
				<label class="label" for="pillar-name">
					<span class="label-text">Nombre</span>
				</label>
				<input
					id="pillar-name"
					type="text"
					class="input input-bordered w-full"
					bind:value={name}
					placeholder="Ej: Liderazgo"
					required
					aria-required="true"
				/>
			</div>

			<div class="form-control mb-3">
				<label class="label" for="pillar-description">
					<span class="label-text">Descripción</span>
				</label>
				<textarea
					id="pillar-description"
					class="textarea textarea-bordered w-full"
					rows={3}
					bind:value={description}
					placeholder="Describe el propósito de este pilar"
					required
					aria-required="true"
				></textarea>
			</div>

			<div class="modal-action mt-6">
				<button type="button" class="btn btn-ghost btn-sm" onclick={handleCancel}>Cancelar</button>
				<button type="submit" class="btn btn-primary btn-sm">
					{isEditing ? 'Guardar cambios' : 'Crear pilar'}
				</button>
			</div>
		</form>
	</div>
</dialog>
