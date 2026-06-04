<script lang="ts">
	import { X } from '@lucide/svelte';
	import type { GoalCategory } from '$lib/types/goal';
	import { getCategories } from '$lib/stores/goalsStore.svelte';

	interface Props {
		open: boolean;
		category: GoalCategory | null;
		onSave: (data: { name: string; description: string; weight: number }) => void;
		onCancel: () => void;
	}

	let { open, category, onSave, onCancel }: Props = $props();

	let dialogEl: HTMLDialogElement | undefined = $state();
	let name = $state('');
	let description = $state('');
	let weight = $state(0);
	let error = $state('');

	const isEditing = $derived(category !== null);
	const title = $derived(isEditing ? 'Editar categoría' : 'Nueva categoría');

	$effect(() => {
		if (!dialogEl) return;
		if (open) {
			name = category?.name ?? '';
			description = category?.description ?? '';
			weight = category?.weight ?? 0;
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

	function validate(): string | null {
		if (!name.trim()) return 'El nombre es obligatorio.';
		if (!description.trim()) return 'La descripción es obligatoria.';
		if (weight < 0 || weight > 100) return 'El peso debe estar entre 0 y 100.';
		const trimmed = name.trim();
		const existing = getCategories();
		const duplicate = existing.find(
			(c) => c.name.toLowerCase() === trimmed.toLowerCase() && c.id !== category?.id
		);
		if (duplicate) return 'Ya existe una categoría con ese nombre.';
		return null;
	}

	function handleSubmit(e: Event) {
		e.preventDefault();
		const err = validate();
		if (err) {
			error = err;
			return;
		}
		onSave({ name: name.trim(), description: description.trim(), weight });
	}
</script>

<dialog
	bind:this={dialogEl}
	class="modal"
	class:modal-open={open}
	aria-modal="true"
	aria-labelledby="category-form-title"
	onclick={handleBackdropClick}
	onclose={handleCancel}
>
	<div class="modal-box">
		<div class="flex items-center justify-between mb-5">
			<h3 id="category-form-title" class="text-lg font-semibold text-base-content">{title}</h3>
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
				<label class="label" for="category-name">
					<span class="label-text">Nombre</span>
				</label>
				<input
					id="category-name"
					type="text"
					class="input input-bordered w-full"
					bind:value={name}
					placeholder="Ej: Ventas y resultados financieros"
					required
					aria-required="true"
				/>
			</div>

			<div class="form-control mb-3">
				<label class="label" for="category-description">
					<span class="label-text">Descripción</span>
				</label>
				<textarea
					id="category-description"
					class="textarea textarea-bordered w-full"
					rows={3}
					bind:value={description}
					placeholder="Describe el propósito de esta categoría"
					required
					aria-required="true"
				></textarea>
			</div>

			<div class="form-control mb-3">
				<label class="label" for="category-weight">
					<span class="label-text">Peso (%)</span>
				</label>
				<input
					id="category-weight"
					type="number"
					class="input input-bordered w-full"
					min={0}
					max={100}
					step={0.1}
					bind:value={weight}
					required
					aria-required="true"
				/>
			</div>

			<div class="modal-action mt-6">
				<button type="button" class="btn btn-ghost btn-sm" onclick={handleCancel}>Cancelar</button>
				<button type="submit" class="btn btn-primary btn-sm">
					{isEditing ? 'Guardar cambios' : 'Crear categoría'}
				</button>
			</div>
		</form>
	</div>
</dialog>
