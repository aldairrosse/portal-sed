<script lang="ts">
	import { Check } from '@lucide/svelte';
	import { validateCategory } from './goalValidation';

	interface Props {
		onSave: (data: { name: string; description: string; weight: number }) => void;
		onCancel: () => void;
	}

	let { onSave, onCancel }: Props = $props();

	let name = $state('');
	let description = $state('');
	let weight = $state(0);
	let error = $state('');

	function handleSubmit(e: Event) {
		e.preventDefault();
		const err = validateCategory({ name, description, weight });
		if (err) {
			error = err;
			return;
		}
		onSave({ name, description, weight });
	}
</script>

<div class="w-full border border-base-300 rounded-lg p-4 bg-base-200/50">
	<form onsubmit={handleSubmit}>
		{#if error}
			<div class="alert alert-error text-sm mb-3" role="alert">
				<span>{error}</span>
			</div>
		{/if}
		<div class="grid grid-cols-1 md:grid-cols-3 gap-3">
			<div class="form-control">
				<label class="label" for="new-cat-name">
					<span class="label-text text-xs">Nombre</span>
				</label>
				<input
					id="new-cat-name"
					type="text"
					class="input input-bordered input-sm w-full"
					bind:value={name}
					placeholder="Nombre de la categoría"
					required
				/>
			</div>
			<div class="form-control">
				<label class="label" for="new-cat-desc">
					<span class="label-text text-xs">Descripción</span>
				</label>
				<textarea
					id="new-cat-desc"
					class="textarea textarea-bordered textarea-sm w-full"
					rows={1}
					bind:value={description}
					placeholder="Descripción"
					required
				></textarea>
			</div>
			<div class="form-control">
				<label class="label" for="new-cat-weight">
					<span class="label-text text-xs">Peso (%)</span>
				</label>
				<input
					id="new-cat-weight"
					type="number"
					class="input input-bordered input-sm w-full"
					bind:value={weight}
					min={0}
					max={100}
					step={0.1}
					placeholder="0"
					required
				/>
			</div>
		</div>
		<div class="flex justify-end gap-2 mt-3">
			<button type="button" class="btn btn-ghost btn-sm" onclick={onCancel}>
				Cancelar
			</button>
			<button type="submit" class="btn btn-primary btn-sm">
				<Check class="w-4 h-4" /> Guardar categoría
			</button>
		</div>
	</form>
</div>
