<script lang="ts">
	import { Check } from '@lucide/svelte';
	import { validateCategory } from './goalValidation';
	import type { GoalCategory } from '$lib/types/goal';
	import CustomSelect from '$lib/components/ui/CustomSelect.svelte';

	interface Props {
		mode?: 'create' | 'edit';
		category?: GoalCategory;
		pillars: { value: string; label: string }[];
		onSave: (data: { id?: string; name: string; description: string; weight: number; pillarId?: string }) => void;
		onCancel: () => void;
		error?: string;
		submitLabel?: string;
	}

	let {
		mode = 'create',
		category,
		pillars,
		onSave,
		onCancel,
		error: externalError = '',
		submitLabel = 'Guardar categoría'
	}: Props = $props();

	let name = $state(category?.name ?? '');
	let description = $state(category?.description ?? '');
	let weight = $state(category?.weight ?? 0);
	let pillarId = $state<string>(category?.pillarId ?? '');
	let localError = $state('');

	let error = $derived(localError || externalError);

	const pillarOptions = $derived([
		{ value: '', label: 'Sin pilar' },
		...pillars
	]);

	function handleSubmit(e: Event) {
		e.preventDefault();
		const err = validateCategory({ name, description, weight, categoryId: category?.id });
		if (err) {
			localError = err;
			return;
		}
		onSave({
			id: category?.id,
			name: name.trim(),
			description: description.trim(),
			weight,
			pillarId: pillarId || undefined
		});
	}
</script>

<div class="w-full border border-base-300 rounded-lg p-4 bg-base-200/50">
	<form onsubmit={handleSubmit}>
		{#if error}
			<div class="alert alert-error text-sm mb-3" role="alert">
				<span>{error}</span>
			</div>
		{/if}
		<div class="grid grid-cols-1 md:grid-cols-2 gap-3">
			<div class="form-control">
				<label class="label" for="cat-name">
					<span class="label-text text-xs">Nombre</span>
				</label>
				<input
					id="cat-name"
					type="text"
					class="input input-bordered input-sm w-full"
					bind:value={name}
					placeholder="Nombre de la categoría"
					required
				/>
			</div>
			<div class="form-control">
				<label class="label" for="cat-desc">
					<span class="label-text text-xs">Descripción</span>
				</label>
				<textarea
					id="cat-desc"
					class="textarea textarea-bordered textarea-sm w-full"
					rows={1}
					bind:value={description}
					placeholder="Descripción"
					required
				></textarea>
			</div>
			<div class="form-control">
				<label class="label" for="cat-weight">
					<span class="label-text text-xs">Ponderación (%)</span>
				</label>
				<input
					id="cat-weight"
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
			<div class="form-control">
				<label class="label">
					<span class="label-text text-xs">Pilar</span>
				</label>
				<CustomSelect
					options={pillarOptions}
					value={pillarId}
					onChange={(v) => { pillarId = v; }}
					placeholder="Sin pilar"
					ariaLabel="Pilar"
				/>
			</div>
		</div>
		<div class="flex justify-end gap-2 mt-3">
			<button type="button" class="btn btn-ghost btn-sm" onclick={onCancel}>
				Cancelar
			</button>
			<button type="submit" class="btn btn-primary btn-sm">
				<Check class="w-4 h-4" /> {submitLabel}
			</button>
		</div>
	</form>
</div>
