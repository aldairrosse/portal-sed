<script lang="ts">
	import { Plus, Trash2, Save, X, Pencil, Library, TrendingUp, TrendingDown } from '@lucide/svelte';
	import type { KPI, KpiUnit } from '$lib/types/goal';
	import { getKpis, addKpi, updateKpi, deleteKpi } from '$lib/stores/goalsStore.svelte';
	import CustomSelect from '$lib/components/ui/CustomSelect.svelte';

	// ─── KPI list ──────────────────────────────────────────────────────────────

	let kpis = $derived(getKpis());

	// ─── Add form state ────────────────────────────────────────────────────────

	let showAddForm = $state(false);
	let addFormName = $state('');
	let addFormDescription = $state('');
	let addFormUnit = $state<KpiUnit>('porcentaje');
	let addFormDirection = $state<'ascendente' | 'descendente'>('ascendente');
	let addFormTargetValue = $state<number | undefined>(undefined);
	let addFormError = $state('');

	const unitOptions: Array<{ value: KpiUnit; label: string }> = [
		{ value: 'porcentaje', label: 'Porcentaje (%)' },
		{ value: 'moneda', label: 'Moneda ($)' },
		{ value: 'numero', label: 'Número' },
		{ value: 'binario', label: 'Binario (Sí/No)' }
	];

	const directionOptions = [
		{ value: 'ascendente', label: 'Ascendente (mayor es mejor)' },
		{ value: 'descendente', label: 'Descendente (menor es mejor)' }
	];

	function resetAddForm() {
		addFormName = '';
		addFormDescription = '';
		addFormUnit = 'porcentaje';
		addFormDirection = 'ascendente';
		addFormTargetValue = undefined;
		addFormError = '';
	}

	function openAddForm() {
		resetAddForm();
		showAddForm = true;
	}

	function cancelAddForm() {
		showAddForm = false;
		resetAddForm();
	}

	function validateAddForm(): string | null {
		if (!addFormName.trim()) return 'El nombre es obligatorio.';
		if (!addFormDescription.trim()) return 'La descripción es obligatoria.';
		return null;
	}

	function handleAddSubmit(e: Event) {
		e.preventDefault();
		const err = validateAddForm();
		if (err) {
			addFormError = err;
			return;
		}
		const newKpi: KPI = {
			id: `kpi-${Date.now()}`,
			name: addFormName.trim(),
			description: addFormDescription.trim(),
			unit: addFormUnit,
			direction: addFormDirection,
			targetValue: addFormTargetValue
		};
		addKpi(newKpi);
		showAddForm = false;
		resetAddForm();
	}

	// ─── Inline edit state ─────────────────────────────────────────────────────

	let editingId = $state<string | null>(null);
	let editName = $state('');
	let editDescription = $state('');
	let editUnit = $state<KpiUnit>('porcentaje');
	let editDirection = $state<'ascendente' | 'descendente'>('ascendente');
	let editTargetValue = $state<number | undefined>(undefined);
	let editError = $state('');

	function startEdit(kpi: KPI) {
		editingId = kpi.id;
		editName = kpi.name;
		editDescription = kpi.description;
		editUnit = kpi.unit;
		editDirection = kpi.direction;
		editTargetValue = kpi.targetValue;
		editError = '';
	}

	function cancelEdit() {
		editingId = null;
		editError = '';
	}

	function handleEditSubmit(e: Event) {
		e.preventDefault();
		if (!editName.trim()) {
			editError = 'El nombre es obligatorio.';
			return;
		}
		if (!editDescription.trim()) {
			editError = 'La descripción es obligatoria.';
			return;
		}
		if (!editingId) return;
		updateKpi(editingId, {
			name: editName.trim(),
			description: editDescription.trim(),
			unit: editUnit,
			direction: editDirection,
			targetValue: editTargetValue
		});
		editingId = null;
		editError = '';
	}

	// ─── Delete confirmation ───────────────────────────────────────────────────

	let deleteTargetId = $state<string | null>(null);
	let deleteTargetName = $state('');

	function openDeleteConfirm(kpi: KPI) {
		deleteTargetId = kpi.id;
		deleteTargetName = kpi.name;
	}

	function confirmDelete() {
		if (deleteTargetId) {
			deleteKpi(deleteTargetId);
		}
		deleteTargetId = null;
		deleteTargetName = '';
	}

	function cancelDelete() {
		deleteTargetId = null;
		deleteTargetName = '';
	}

	// ─── Helpers ───────────────────────────────────────────────────────────────

	const unitLabels: Record<KpiUnit, string> = {
		porcentaje: 'Porcentaje',
		moneda: 'Moneda',
		numero: 'Número',
		binario: 'Binario'
	};

	const directionLabels = {
		ascendente: 'Ascendente',
		descendente: 'Descendente'
	};
</script>

<svelte:head>
	<title>Biblioteca de KPI — SED</title>
</svelte:head>

<div class="space-y-6 max-w-full min-w-0">
	<!-- Breadcrumbs -->
	<div class="text-sm breadcrumbs">
		<ul>
			<li><a href="/objetivos/asignacion">Asignación anual</a></li>
			<li class="text-base-content font-medium">Biblioteca</li>
		</ul>
	</div>

	<!-- Page header -->
	<div>
		<h1 class="text-2xl font-bold text-base-content flex items-center gap-2">
			<Library class="w-6 h-6" />
			Biblioteca de KPI
		</h1>
		<p class="text-sm text-base-content/50 mt-1">
			Gestione los indicadores clave de desempeño reutilizables para vincular a metas.
		</p>
	</div>

	<!-- KPI Table -->
	{#if kpis.length > 0}
		<div class="overflow-x-auto">
			<table class="table table-zebra w-full">
				<thead>
					<tr>
						<th class="text-xs font-semibold text-base-content/60">Nombre</th>
						<th class="text-xs font-semibold text-base-content/60">Descripción</th>
						<th class="text-xs font-semibold text-base-content/60">Unidad</th>
						<th class="text-xs font-semibold text-base-content/60">Dirección</th>
						<th class="text-xs font-semibold text-base-content/60 text-center">Meta</th>
						<th class="text-xs font-semibold text-base-content/60 text-center">Acciones</th>
					</tr>
				</thead>
				<tbody>
					{#each kpis as kpi (kpi.id)}
						{#if editingId === kpi.id}
							<!-- Edit mode row -->
							<tr>
								<td>
									<input
										type="text"
										class="input input-bordered input-sm w-full"
										bind:value={editName}
										required
									/>
								</td>
								<td>
									<input
										type="text"
										class="input input-bordered input-sm w-full"
										bind:value={editDescription}
										required
									/>
								</td>
								<td>
									<CustomSelect
										options={unitOptions}
										value={editUnit}
										onChange={(v) => (editUnit = v as KpiUnit)}
										ariaLabel="Unidad"
									/>
								</td>
								<td>
									<CustomSelect
										options={directionOptions}
										value={editDirection}
										onChange={(v) => (editDirection = v as 'ascendente' | 'descendente')}
										ariaLabel="Dirección"
									/>
								</td>
								<td>
									<input
										type="number"
										class="input input-bordered input-sm w-full text-center"
										min={0}
										step={0.01}
										bind:value={editTargetValue}
									/>
								</td>
								<td>
									{#if editError}
										<p class="text-xs text-error mb-1">{editError}</p>
									{/if}
									<div class="flex items-center justify-center gap-1">
										<button
											class="btn btn-primary btn-xs"
											onclick={handleEditSubmit}
											aria-label="Guardar cambios"
										>
											<Save class="w-3 h-3" />
										</button>
										<button
											class="btn btn-ghost btn-xs"
											onclick={cancelEdit}
											aria-label="Cancelar edición"
										>
											<X class="w-3 h-3" />
										</button>
									</div>
								</td>
							</tr>
						{:else}
							<!-- Display mode row -->
							<tr>
								<td class="font-medium">{kpi.name}</td>
								<td class="text-sm text-base-content/70">{kpi.description}</td>
							<td>
								<span class="badge badge-sm bg-primary/20 text-primary">{unitLabels[kpi.unit]}</span>
							</td>
							<td>
								<span class="badge badge-sm bg-primary/20 text-primary">
									{#if kpi.direction === 'ascendente'}
										<TrendingUp class="w-3 h-3" />
									{:else}
										<TrendingDown class="w-3 h-3" />
									{/if}
									{directionLabels[kpi.direction]}
								</span>
							</td>
								<td class="text-center">
									{#if kpi.targetValue !== undefined && kpi.targetValue !== null}
										{kpi.targetValue}
									{:else}
										<span class="text-base-content/30">—</span>
									{/if}
								</td>
								<td>
									<div class="flex items-center justify-center gap-1">
										<button
											class="btn btn-ghost btn-xs"
											onclick={() => startEdit(kpi)}
											aria-label="Editar {kpi.name}"
										>
											<Pencil class="w-3 h-3" />
										</button>
										<button
											class="btn btn-ghost btn-xs text-error"
											onclick={() => openDeleteConfirm(kpi)}
											aria-label="Eliminar {kpi.name}"
										>
											<Trash2 class="w-3 h-3" />
										</button>
									</div>
								</td>
							</tr>
						{/if}
					{/each}
				</tbody>
			</table>
		</div>
	{:else}
		<div class="text-center py-12 text-base-content/50 text-sm">
			No hay KPI registrados. Agregue el primer KPI para comenzar.
		</div>
	{/if}

	<!-- Inline add form -->
	{#if showAddForm}
		<div class="border border-base-300 rounded-lg p-4 bg-base-200/50">
			<h3 class="text-sm font-semibold text-base-content mb-3">Nuevo KPI</h3>

			{#if addFormError}
				<div class="alert alert-error mb-3 text-sm" role="alert">
					<span>{addFormError}</span>
				</div>
			{/if}

			<form onsubmit={handleAddSubmit} class="space-y-3">
				<div class="grid grid-cols-1 md:grid-cols-2 gap-3">
					<div class="form-control">
						<label class="label" for="add-kpi-name">
							<span class="label-text text-xs">Nombre</span>
						</label>
						<input
							id="add-kpi-name"
							type="text"
							class="input input-bordered input-sm w-full"
							bind:value={addFormName}
							placeholder="Ej: Satisfacción del cliente"
							required
							aria-required="true"
						/>
					</div>

					<div class="form-control">
						<label class="label" for="add-kpi-description">
							<span class="label-text text-xs">Descripción</span>
						</label>
						<input
							id="add-kpi-description"
							type="text"
							class="input input-bordered input-sm w-full"
							bind:value={addFormDescription}
							placeholder="Describe el indicador"
							required
							aria-required="true"
						/>
					</div>
				</div>

				<div class="grid grid-cols-1 md:grid-cols-3 gap-3">
					<div class="form-control">
						<label class="label" for="add-kpi-unit">
							<span class="label-text text-xs">Unidad</span>
						</label>
						<CustomSelect
							options={unitOptions}
							value={addFormUnit}
							onChange={(v) => (addFormUnit = v as KpiUnit)}
							ariaLabel="Unidad"
						/>
					</div>

					<div class="form-control">
						<label class="label" for="add-kpi-direction">
							<span class="label-text text-xs">Dirección</span>
						</label>
						<CustomSelect
							options={directionOptions}
							value={addFormDirection}
							onChange={(v) => (addFormDirection = v as 'ascendente' | 'descendente')}
							ariaLabel="Dirección"
						/>
					</div>

					<div class="form-control">
						<label class="label" for="add-kpi-target">
							<span class="label-text text-xs">Valor objetivo</span>
							<span class="label-text-alt text-base-content/30 text-xs">Opcional</span>
						</label>
						<input
							id="add-kpi-target"
							type="number"
							class="input input-bordered input-sm w-full"
							min={0}
							step={0.01}
							bind:value={addFormTargetValue}
						/>
					</div>
				</div>

				<div class="flex items-center gap-2">
					<button type="submit" class="btn btn-primary btn-sm">
						<Save class="w-4 h-4" />
						Guardar KPI
					</button>
					<button type="button" class="btn btn-ghost btn-sm" onclick={cancelAddForm}>
						Cancelar
					</button>
				</div>
			</form>
		</div>
	{/if}

	<!-- Add KPI button (disabled while form is open) -->
	{#if !showAddForm}
		<div class="flex justify-center pt-2">
			<button class="btn btn-outline btn-primary" onclick={openAddForm}>
				<Plus class="w-4 h-4" />
				Agregar KPI
			</button>
		</div>
	{/if}
</div>

<!-- Delete confirmation modal -->
<dialog class="modal" class:modal-open={deleteTargetId !== null}>
	<div class="modal-box">
		<h3 class="text-lg font-bold">Eliminar KPI</h3>
		<p class="py-4">
			¿Está seguro que desea eliminar el KPI <strong>{deleteTargetName}</strong>? Esta acción no se
			puede deshacer.
		</p>
		<div class="modal-action">
			<button class="btn btn-ghost" onclick={cancelDelete}>Cancelar</button>
			<button class="btn btn-error" onclick={confirmDelete}>Eliminar</button>
		</div>
	</div>
	<form method="dialog" class="modal-backdrop">
		<button onclick={cancelDelete}>close</button>
	</form>
</dialog>
