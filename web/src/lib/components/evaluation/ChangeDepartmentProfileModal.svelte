<script lang="ts">
	import { onMount } from 'svelte';
	import { client } from '$lib/api/client';
	import { getProfiles } from '$lib/stores/competencyStore.svelte';
	import { getRoot, updateEmployeeAssignment } from '$lib/stores/orgHierarchyStore.svelte';
	import CustomSelect from '$lib/components/ui/CustomSelect.svelte';
	import OrgHierarchyTree from '$lib/components/org-hierarchy/OrgHierarchyTree.svelte';

	interface EmployeeDetailResponse {
		id: string;
		profileId: string;
		orgNodeId: string;
		firstName: string;
		lastName: string;
	}

	interface Props {
		employeeId: string;
		onsave: () => void;
		onclose: () => void;
	}

	let {
		employeeId,
		onsave = () => {},
		onclose = () => {},
	}: Props = $props();

	let loading = $state(true);
	let saving = $state(false);
	let error = $state<string | null>(null);
	let employeeName = $state('');
	let selectedProfileId = $state('');
	let selectedOrgNodeId = $state('');
	let initialProfileId = $state('');
	let initialOrgNodeId = $state('');

	const profileOptions = $derived(
		getProfiles().map((p) => ({ value: p.id, label: p.name })),
	);
	const rootNode = $derived(getRoot());

	onMount(async () => {
		try {
			const res = await client.GET('/employees/{empId}', {
				params: { path: { empId: employeeId } },
			});
			const data = (res.data as { data?: EmployeeDetailResponse })?.data;
			if (!data) throw new Error('Empleado no encontrado');

			employeeName = `${data.firstName} ${data.lastName}`;
			selectedProfileId = data.profileId;
			selectedOrgNodeId = data.orgNodeId;
			initialProfileId = data.profileId;
			initialOrgNodeId = data.orgNodeId;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Error al cargar datos del empleado';
		} finally {
			loading = false;
		}
	});

	const hasChanges = $derived(
		selectedProfileId !== '' && selectedOrgNodeId !== '' &&
		(selectedProfileId !== initialProfileId || selectedOrgNodeId !== initialOrgNodeId),
	);

	async function handleSave() {
		if (saving || !hasChanges) return;
		saving = true;
		error = null;
		try {
			await updateEmployeeAssignment(employeeId, selectedProfileId, selectedOrgNodeId);
			onsave();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Error al guardar cambios';
		} finally {
			saving = false;
		}
	}
</script>

{#if loading}
	<dialog class="modal modal-open">
		<div class="modal-box">
			<div class="skeleton h-6 w-48 mb-4" />
			<div class="skeleton h-8 w-full mb-2" />
			<div class="skeleton h-64 w-full" />
		</div>
	</dialog>
{:else}
	<dialog class="modal modal-open" onclick={onclose}>
		<div
			class="modal-box max-w-lg"
			onclick={(e) => e.stopPropagation()}
		>
			<h3 class="font-bold text-lg mb-4">
				Cambiar departamento y perfil — {employeeName}
			</h3>

			{#if error}
				<div class="alert alert-error text-sm mb-4">{error}</div>
			{/if}

			<div class="flex flex-col gap-4">
				<div>
					<label class="label text-xs font-semibold text-base-content/60"
						>Perfil</label
					>
					<CustomSelect
						options={profileOptions}
						value={selectedProfileId}
						onChange={(v) => (selectedProfileId = v)}
						placeholder="Seleccionar perfil"
						ariaLabel="Seleccionar perfil de evaluación"
					/>
				</div>

				<div>
					<label class="label text-xs font-semibold text-base-content/60"
						>Departamento</label
					>
					{#if rootNode}
						<div class="border border-base-300 rounded-box max-h-64 overflow-y-auto">
							<OrgHierarchyTree
								node={rootNode}
								viewType="departments"
								selectable={true}
								selectableSelectedId={selectedOrgNodeId}
								onSelectableClick={(id) => (selectedOrgNodeId = id)}
								initialExpandedIds={[rootNode.id]}
							/>
						</div>
					{:else}
						<p class="text-sm text-base-content/30 italic">Cargando árbol organizacional...</p>
					{/if}
				</div>
			</div>

			<div class="modal-action">
				<button
					type="button"
					class="btn btn-ghost btn-sm"
					disabled={saving}
					onclick={onclose}
				>
					Cancelar
				</button>
				<button
					type="button"
					class="btn btn-primary btn-sm"
					disabled={!hasChanges || saving}
					onclick={handleSave}
				>
					{saving ? 'Guardando...' : 'Guardar'}
				</button>
			</div>
		</div>
	</dialog>
{/if}