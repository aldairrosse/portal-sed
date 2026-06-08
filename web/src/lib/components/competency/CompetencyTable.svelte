<script lang="ts">
	import { Check, Pencil, Trash2 } from '@lucide/svelte';
	import type { Competency } from '$lib/types/competency';
	import { getCompetenciesByPillar } from '$lib/stores/competencyStore.svelte';

	interface Props {
		competencies: Competency[];
		pillarId: string;
		pillarName: string;
		editingId?: string | null;
		onSave: (data: { name: string; description: string; id?: string }) => void;
		onDelete: (competency: Competency) => void;
	}

	let { competencies, pillarId, pillarName, editingId = $bindable(null), onSave, onDelete }: Props = $props();

	let editName = $state('');
	let editDescription = $state('');
	let localError = $state('');

	const isEditing = $derived(editingId !== null);

	$effect(() => {
		if (editingId === '__new__') {
			editName = '';
			editDescription = '';
		} else if (editingId) {
			const c = competencies.find((c) => c.id === editingId);
			if (c) {
				editName = c.name;
				editDescription = c.description;
			}
		}
		localError = '';
	});

	function startEdit(competency: Competency) {
		editingId = competency.id;
	}

	function cancel() {
		editingId = null;
		localError = '';
	}

	function validate(nameVal: string, idToExclude?: string): string | null {
		if (!nameVal.trim()) return 'El nombre es obligatorio.';
		if (!editDescription.trim()) return 'La descripción es obligatoria.';
		const trimmed = nameVal.trim();
		const existing = getCompetenciesByPillar(pillarId);
		const duplicate = existing.find(
			(c) => c.name.toLowerCase() === trimmed.toLowerCase() && c.id !== idToExclude
		);
		if (duplicate) return 'Ya existe una competencia con ese nombre en este pilar.';
		return null;
	}

	function saveNew() {
		const err = validate(editName);
		if (err) {
			localError = err;
			return;
		}
		onSave({ name: editName.trim(), description: editDescription.trim() });
		editingId = null;
		localError = '';
	}

	function saveEdit(id: string) {
		const err = validate(editName, id);
		if (err) {
			localError = err;
			return;
		}
		onSave({ name: editName.trim(), description: editDescription.trim(), id });
		editingId = null;
		localError = '';
	}
</script>

<div class="overflow-x-auto">
	<table class="table table-zebra" aria-label="Competencias de {pillarName}">
		<thead>
			<tr>
				<th class="w-1/3">Nombre</th>
				<th class="w-1/2">Descripción</th>
				<th class="w-[140px] text-right">Acciones</th>
			</tr>
		</thead>
		<tbody>
			{#each competencies as competency (competency.id)}
				{#if editingId === competency.id}
					<!-- Edit mode -->
					<tr>
						<td colspan="3" class="p-4">
							<div class="border border-base-300 rounded-lg p-4 bg-base-200/50">
								{#if localError}
									<div class="alert alert-error text-sm mb-3" role="alert">
										<span>{localError}</span>
									</div>
								{/if}
								<div class="grid grid-cols-1 md:grid-cols-2 gap-3 mb-3">
									<div class="form-control">
										<label class="label" for="edit-name-{competency.id}">
											<span class="label-text text-xs">Nombre</span>
										</label>
										<input
											id="edit-name-{competency.id}"
											type="text"
											class="input input-bordered input-sm w-full"
											bind:value={editName}
											required
										/>
									</div>
									<div class="form-control">
										<label class="label" for="edit-desc-{competency.id}">
											<span class="label-text text-xs">Descripción</span>
										</label>
										<textarea
											id="edit-desc-{competency.id}"
											class="textarea textarea-bordered textarea-sm w-full"
											rows={1}
											bind:value={editDescription}
											required
										></textarea>
									</div>
								</div>
								<div class="flex justify-end gap-2 mt-3">
									<button class="btn btn-ghost btn-sm" onclick={cancel}>Cancelar</button>
									<button class="btn btn-primary btn-sm" onclick={() => saveEdit(competency.id)}>
										<Check class="w-4 h-4" /> Guardar competencia
									</button>
								</div>
							</div>
						</td>
					</tr>
				{:else}
					<!-- Display mode -->
					<tr>
						<td class="font-medium">{competency.name}</td>
						<td class="text-base-content/60 text-sm">{competency.description}</td>
						<td class="text-right">
							<button
								class="btn btn-ghost btn-square btn-sm"
								onclick={() => startEdit(competency)}
								disabled={isEditing}
								aria-label="Editar {competency.name}"
							>
								<Pencil class="w-4 h-4" />
							</button>
							<button
								class="btn btn-ghost btn-square btn-sm text-error"
								onclick={() => onDelete(competency)}
								disabled={isEditing}
								aria-label="Eliminar {competency.name}"
							>
								<Trash2 class="w-4 h-4" />
							</button>
						</td>
					</tr>
				{/if}
			{/each}
			{#if editingId === '__new__'}
				<!-- New row form -->
				<tr>
					<td colspan="3" class="p-4">
						<div class="border border-base-300 rounded-lg p-4 bg-base-200/50">
							{#if localError}
								<div class="alert alert-error text-sm mb-3" role="alert">
									<span>{localError}</span>
								</div>
							{/if}
							<div class="grid grid-cols-1 md:grid-cols-2 gap-3 mb-3">
								<div class="form-control">
									<label class="label" for="edit-name-__new__">
										<span class="label-text text-xs">Nombre</span>
									</label>
									<input
										id="edit-name-__new__"
										type="text"
										class="input input-bordered input-sm w-full"
										bind:value={editName}
										required
									/>
								</div>
								<div class="form-control">
									<label class="label" for="edit-desc-__new__">
										<span class="label-text text-xs">Descripción</span>
									</label>
									<textarea
										id="edit-desc-__new__"
										class="textarea textarea-bordered textarea-sm w-full"
										rows={1}
										bind:value={editDescription}
										required
									></textarea>
								</div>
							</div>
							<div class="flex justify-end gap-2 mt-3">
								<button class="btn btn-ghost btn-sm" onclick={cancel}>Cancelar</button>
								<button class="btn btn-primary btn-sm" onclick={saveNew}>
									<Check class="w-4 h-4" /> Guardar competencia
								</button>
							</div>
						</div>
					</td>
				</tr>
			{/if}
		</tbody>
	</table>
</div>
