<script lang="ts">
	import { X, Plus, Trash2, Save } from '@lucide/svelte';
	import type { KPI, KpiUnit } from '$lib/types/goal';
	import { getKpis, addKpi, updateKpi, deleteKpi } from '$lib/stores/goalsStore.svelte';

	interface Props {
		open: boolean;
		onClose: () => void;
	}

	let { open, onClose }: Props = $props();

	let dialogEl: HTMLDialogElement | undefined = $state();
	let editingKpi: KPI | null = $state(null);
	let formName = $state('');
	let formDescription = $state('');
	let formUnit = $state<KpiUnit>('porcentaje');
	let formDirection = $state<'ascendente' | 'descendente'>('ascendente');
	let formTargetValue = $state<number | undefined>(undefined);
	let formError = $state('');

	const unitOptions: Array<{ value: KpiUnit; label: string }> = [
		{ value: 'porcentaje', label: 'Porcentaje (%)' },
		{ value: 'moneda', label: 'Moneda ($)' },
		{ value: 'numero', label: 'Número' },
		{ value: 'binario', label: 'Binario (Sí/No)' }
	];

	const kpis = $derived(getKpis());

	$effect(() => {
		if (!dialogEl) return;
		if (open) {
			editingKpi = null;
			resetForm();
			dialogEl.showModal();
		} else {
			dialogEl.close();
		}
	});

	function resetForm() {
		formName = '';
		formDescription = '';
		formUnit = 'porcentaje';
		formDirection = 'ascendente';
		formTargetValue = undefined;
		formError = '';
	}

	function startEdit(kpi: KPI) {
		editingKpi = kpi;
		formName = kpi.name;
		formDescription = kpi.description;
		formUnit = kpi.unit;
		formDirection = kpi.direction;
		formTargetValue = kpi.targetValue;
		formError = '';
	}

	function cancelEdit() {
		editingKpi = null;
		resetForm();
	}

	function handleCancel() {
		onClose();
	}

	function handleBackdropClick(e: MouseEvent) {
		if (e.target === dialogEl) {
			handleCancel();
		}
	}

	function validate(): string | null {
		if (!formName.trim()) return 'El nombre es obligatorio.';
		if (!formDescription.trim()) return 'La descripción es obligatoria.';
		return null;
	}

	function handleFormSubmit(e: Event) {
		e.preventDefault();
		const err = validate();
		if (err) {
			formError = err;
			return;
		}

		if (editingKpi) {
			updateKpi(editingKpi.id, {
				name: formName.trim(),
				description: formDescription.trim(),
				unit: formUnit,
				direction: formDirection,
				targetValue: formTargetValue
			});
		} else {
			const newKpi: KPI = {
				id: `kpi-${Date.now()}`,
				name: formName.trim(),
				description: formDescription.trim(),
				unit: formUnit,
				direction: formDirection,
				targetValue: formTargetValue
			};
			addKpi(newKpi);
		}

		editingKpi = null;
		resetForm();
	}

	function handleDelete(kpiId: string) {
		deleteKpi(kpiId);
	}
</script>

<dialog
	bind:this={dialogEl}
	class="modal"
	class:modal-open={open}
	aria-modal="true"
	aria-labelledby="kpi-library-title"
	onclick={handleBackdropClick}
	onclose={handleCancel}
>
	<div class="modal-box max-w-xl">
		<div class="flex items-center justify-between mb-5">
			<h3 id="kpi-library-title" class="text-lg font-semibold text-base-content">
				Biblioteca de KPI
			</h3>
			<button
				class="btn btn-ghost btn-square btn-sm"
				onclick={handleCancel}
				aria-label="Cerrar"
			>
				<X class="w-4 h-4" />
			</button>
		</div>

		{#if kpis.length > 0}
			<div class="space-y-2 mb-6 max-h-60 overflow-y-auto">
				{#each kpis as kpi (kpi.id)}
					<div class="flex items-center justify-between gap-3 p-3 rounded-lg bg-base-200/50">
						<div class="flex-1 min-w-0">
							<span class="text-sm font-medium block">{kpi.name}</span>
							<span class="text-xs text-base-content/50 truncate block">{kpi.description}</span>
							<div class="flex gap-2 mt-1">
								<span class="badge badge-ghost badge-xs">{kpi.unit}</span>
								<span class="badge badge-ghost badge-xs">{kpi.direction}</span>
								{#if kpi.targetValue !== undefined && kpi.targetValue !== null}
									<span class="badge badge-ghost badge-xs">Meta: {kpi.targetValue}</span>
								{/if}
							</div>
						</div>
						<div class="flex items-center gap-1 flex-shrink-0">
							<button
								class="btn btn-ghost btn-xs"
								onclick={() => startEdit(kpi)}
								aria-label="Editar {kpi.name}"
							>
								Editar
							</button>
							<button
								class="btn btn-ghost btn-xs text-error"
								onclick={() => handleDelete(kpi.id)}
								aria-label="Eliminar {kpi.name}"
							>
								<Trash2 class="w-3 h-3" />
							</button>
						</div>
					</div>
				{/each}
			</div>
		{:else}
			<p class="text-sm text-base-content/50 italic mb-6">No hay KPI registrados.</p>
		{/if}

		<!-- Form for adding/editing -->
		<form onsubmit={handleFormSubmit}>
			{#if formError}
				<div class="alert alert-error mb-4 text-sm" role="alert">
					<span>{formError}</span>
				</div>
			{/if}

			<h4 class="text-sm font-semibold text-base-content mb-3">
				{editingKpi ? 'Editar KPI' : 'Nuevo KPI'}
			</h4>

			<div class="form-control mb-3">
				<label class="label" for="kpi-name">
					<span class="label-text">Nombre</span>
				</label>
				<input
					id="kpi-name"
					type="text"
					class="input input-bordered w-full"
					bind:value={formName}
					placeholder="Ej: Satisfacción del cliente"
					required
					aria-required="true"
				/>
			</div>

			<div class="form-control mb-3">
				<label class="label" for="kpi-description">
					<span class="label-text">Descripción</span>
				</label>
				<textarea
					id="kpi-description"
					class="textarea textarea-bordered w-full"
					rows={2}
					bind:value={formDescription}
					placeholder="Describe el indicador"
					required
					aria-required="true"
				></textarea>
			</div>

			<div class="grid grid-cols-2 gap-3 mb-3">
				<div class="form-control">
					<label class="label" for="kpi-unit">
						<span class="label-text">Unidad</span>
					</label>
					<select
						id="kpi-unit"
						class="select select-bordered w-full"
						bind:value={formUnit}
					>
						{#each unitOptions as opt (opt.value)}
							<option value={opt.value}>{opt.label}</option>
						{/each}
					</select>
				</div>

				<div class="form-control">
					<label class="label" for="kpi-direction">
						<span class="label-text">Dirección</span>
					</label>
					<select
						id="kpi-direction"
						class="select select-bordered w-full"
						bind:value={formDirection}
					>
						<option value="ascendente">Ascendente (mayor es mejor)</option>
						<option value="descendente">Descendente (menor es mejor)</option>
					</select>
				</div>
			</div>

			<div class="form-control mb-4">
				<label class="label" for="kpi-target">
					<span class="label-text">Valor objetivo</span>
					<span class="label-text-alt text-base-content/30">Opcional</span>
				</label>
				<input
					id="kpi-target"
					type="number"
					class="input input-bordered w-full"
					min={0}
					step={0.01}
					bind:value={formTargetValue}
				/>
			</div>

			<div class="flex items-center justify-between">
				{#if editingKpi}
					<div class="flex gap-2">
						<button type="submit" class="btn btn-primary btn-sm">
							<Save class="w-4 h-4" />
							Guardar cambios
						</button>
						<button type="button" class="btn btn-ghost btn-sm" onclick={cancelEdit}>Cancelar</button>
					</div>
				{:else}
					<button type="submit" class="btn btn-primary btn-sm">
						<Plus class="w-4 h-4" />
						Agregar KPI
					</button>
				{/if}
			</div>
		</form>
	</div>
</dialog>
