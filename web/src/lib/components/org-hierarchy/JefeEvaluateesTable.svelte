<script lang="ts">
	import EmptyState from '$lib/components/ui/EmptyState.svelte';

	interface Evaluatee {
		id: string;
		firstName: string;
		lastName: string;
		email?: string;
		employeeNumber?: string;
		orgNodeId?: string;
		profileId?: string;
	}

	interface Props {
		evaluatees?: Evaluatee[];
		loading?: boolean;
		error?: string | null;
		onSelect?: (id: string) => void;
	}

	let {
		evaluatees = [],
		loading = false,
		error = null,
		onSelect = () => {}
	}: Props = $props();
</script>

{#if loading}
	<div class="p-6 text-center text-sm text-base-content/40">Cargando evaluados…</div>
{:else if error}
	<EmptyState title="Error" message={error} />
{:else if evaluatees.length === 0}
	<EmptyState
		title="Sin evaluados"
		message="No tienes evaluados asignados."
	/>
{:else}
	<div class="overflow-x-auto">
		<table class="table table-sm">
			<thead>
				<tr>
					<th class="text-xs font-semibold uppercase tracking-wide text-base-content/40">Nombre</th>
					<th class="text-xs font-semibold uppercase tracking-wide text-base-content/40">Puesto</th>
					<th class="text-xs font-semibold uppercase tracking-wide text-base-content/40">Perfil</th>
				</tr>
			</thead>
			<tbody>
				{#each evaluatees as emp (emp.id)}
					<tr class="hover cursor-pointer" onclick={() => onSelect(emp.id)}>
						<td class="font-medium">{emp.firstName} {emp.lastName}</td>
						<td class="text-base-content/60">{emp.employeeNumber ?? '—'}</td>
						<td><span class="badge badge-ghost badge-sm">{emp.profileId ?? '—'}</span></td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{/if}
