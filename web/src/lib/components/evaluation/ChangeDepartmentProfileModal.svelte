<script lang="ts">
	import { onMount } from 'svelte';
	import { client } from '$lib/api/client';
	import { getProfiles, load as loadCompetencyData } from '$lib/stores/competencyStore.svelte';
	import { getRoot, updateEmployeeAssignment, load as loadOrgHierarchy } from '$lib/stores/orgHierarchyStore.svelte';
	import { titleCase } from '$lib/utils/text';
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
		getProfiles().map((p) => ({ value: p.id, label: titleCase(p.name) })),
	);
	const rootNode = $derived(getRoot());

	onMount(async () => {
		try {
			if (!getRoot()) {
				await loadOrgHierarchy();
			}
			if (getProfiles().length === 0) {
				await loadCompetencyData();
			}

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
			<div class="skeleton h-6 w-48 mb-4"></div>
			<div class="skeleton h-8 w-full mb-2"></div>
			<div class="skeleton h-64 w-full"></div>
		</div>
	</dialog>
{:else}
	<dialog class="modal modal-open" onclick={(e) => { if (e.target === e.currentTarget) onclose(); }}>
		<div
			class="modal-box max-w-lg"
		>
			<h3 class="font-bold text-lg mb-4">
				Cambiar departamento y perfil
			</h3>
			<p class="text-sm text-base-content/30">
				{employeeName}
			</p>

			{#if error}
				<div class="alert alert-error text-sm mb-4">{error}</div>
			{/if}

			<div class="flex flex-col gap-4">
				<div>
					<span class="label text-xs font-semibold text-base-content/60"
						>Perfil</span
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
					<span class="label text-xs font-semibold text-base-content/60"
						>Departamento</span
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