<script lang="ts">
	import { X } from '@lucide/svelte';
	import type { Goal, KPI, GoalUnit } from '$lib/types/goal';
	import { getGoals, getKpisForGoal, linkKpiToGoal, unlinkKpiFromGoal } from '$lib/stores/goalsStore.svelte';

	interface Props {
		open: boolean;
		goal: Goal | null;
		categoryId: string;
		allKpis: KPI[];
		onSave: (data: { name: string; description: string; unit: GoalUnit; weight: number; targetValue: number; linkedKpiIds: string[] }) => void;
		onCancel: () => void;
	}

	let { open, goal, categoryId, allKpis, onSave, onCancel }: Props = $props();

	let dialogEl: HTMLDialogElement | undefined = $state();
	let name = $state('');
	let description = $state('');
	let unit = $state<GoalUnit>('porcentaje');
	let weight = $state(0);
	let targetValue = $state(0);
	let selectedKpiIds = $state<string[]>([]);
	let error = $state('');

	const isEditing = $derived(goal !== null);
	const title = $derived(isEditing ? 'Editar meta' : 'Nueva meta');

	const unitOptions: Array<{ value: GoalUnit; label: string }> = [
		{ value: 'porcentaje', label: 'Porcentaje (%)' },
		{ value: 'moneda', label: 'Moneda ($)' },
		{ value: 'numero', label: 'Número' },
		{ value: 'binario', label: 'Binario (Sí/No)' }
	];

	$effect(() => {
		if (!dialogEl) return;
		if (open) {
			name = goal?.name ?? '';
			description = goal?.description ?? '';
			unit = goal?.unit ?? 'porcentaje';
			weight = goal?.weight ?? 0;
			targetValue = goal?.targetValue ?? 0;
			if (goal) {
				selectedKpiIds = getKpisForGoal(goal.id).map((k) => k.id);
			} else {
				selectedKpiIds = [];
			}
			error = '';
			dialogEl.showModal();
		} else {
			dialogEl.close();
		}
	});

	function toggleKpi(kpiId: string) {
		if (selectedKpiIds.includes(kpiId)) {
			selectedKpiIds = selectedKpiIds.filter((id) => id !== kpiId);
		} else {
			selectedKpiIds = [...selectedKpiIds, kpiId];
		}
	}

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
		if (targetValue <= 0) return 'El valor objetivo debe ser mayor a 0.';
		const trimmed = name.trim();
		const existing = getGoals().filter((g) => g.categoryId === categoryId);
		const duplicate = existing.find(
			(g) => g.name.toLowerCase() === trimmed.toLowerCase() && g.id !== goal?.id
		);
		if (duplicate) return 'Ya existe una meta con ese nombre en esta categoría.';
		return null;
	}

	function handleSubmit(e: Event) {
		e.preventDefault();
		const err = validate();
		if (err) {
			error = err;
			return;
		}
		onSave({
			name: name.trim(),
			description: description.trim(),
			unit,
			weight,
			targetValue,
			linkedKpiIds: selectedKpiIds
		});
	}
</script>

<dialog
	bind:this={dialogEl}
	class="modal"
	class:modal-open={open}
	aria-modal="true"
	aria-labelledby="goal-form-title"
	onclick={handleBackdropClick}
	onclose={handleCancel}
>
	<div class="modal-box max-w-xl">
		<div class="flex items-center justify-between mb-5">
			<h3 id="goal-form-title" class="text-lg font-semibold text-base-content">{title}</h3>
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
				<label class="label" for="goal-name">
					<span class="label-text">Nombre</span>
				</label>
				<input
					id="goal-name"
					type="text"
					class="input input-bordered w-full"
					bind:value={name}
					placeholder="Ej: Alcanzar meta de ventas mensuales"
					required
					aria-required="true"
				/>
			</div>

			<div class="form-control mb-3">
				<label class="label" for="goal-description">
					<span class="label-text">Descripción</span>
				</label>
				<textarea
					id="goal-description"
					class="textarea textarea-bordered w-full"
					rows={2}
					bind:value={description}
					placeholder="Describe el objetivo de la meta"
					required
					aria-required="true"
				></textarea>
			</div>

			<div class="grid grid-cols-2 gap-3 mb-3">
				<div class="form-control">
					<label class="label" for="goal-unit">
						<span class="label-text">Unidad</span>
					</label>
					<select
						id="goal-unit"
						class="select select-bordered w-full"
						bind:value={unit}
						aria-required="true"
					>
						{#each unitOptions as opt (opt.value)}
							<option value={opt.value}>{opt.label}</option>
						{/each}
					</select>
				</div>

				<div class="form-control">
					<label class="label" for="goal-weight">
						<span class="label-text">Peso (%)</span>
					</label>
					<input
						id="goal-weight"
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
			</div>

			<div class="form-control mb-4">
				<label class="label" for="goal-target">
					<span class="label-text">Valor objetivo</span>
				</label>
				<input
					id="goal-target"
					type="number"
					class="input input-bordered w-full"
					min={0}
					step={0.01}
					bind:value={targetValue}
					required
					aria-required="true"
				/>
			</div>

			{#if allKpis.length > 0}
				<fieldset class="border border-base-300 rounded-lg p-4 mb-2">
					<legend class="text-sm font-semibold text-base-content px-1">Indicadores clave (KPI)</legend>
					<p class="text-xs text-base-content/50 mb-3">Seleccione los KPI que aplican a esta meta</p>
					<div class="space-y-2 max-h-48 overflow-y-auto">
						{#each allKpis as kpi (kpi.id)}
							<label class="flex items-center gap-3 cursor-pointer p-2 rounded hover:bg-base-200/50">
								<input
									type="checkbox"
									class="checkbox checkbox-sm checkbox-primary"
									checked={selectedKpiIds.includes(kpi.id)}
									onchange={() => toggleKpi(kpi.id)}
								/>
								<div class="flex-1 min-w-0">
									<span class="text-sm font-medium block">{kpi.name}</span>
									<span class="text-xs text-base-content/50 truncate block">{kpi.description}</span>
								</div>
								<span class="badge badge-ghost badge-xs">{kpi.unit}</span>
							</label>
						{/each}
					</div>
				</fieldset>
			{/if}

			<div class="modal-action mt-6">
				<button type="button" class="btn btn-ghost btn-sm" onclick={handleCancel}>Cancelar</button>
				<button type="submit" class="btn btn-primary btn-sm">
					{isEditing ? 'Guardar cambios' : 'Crear meta'}
				</button>
			</div>
		</form>
	</div>
</dialog>
