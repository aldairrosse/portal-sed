<script lang="ts">
	import { Pencil, Trash2, Plus } from '@lucide/svelte';
	import type { Competency } from '$lib/types/competency';

	interface Props {
		competencies: Competency[];
		pillarName: string;
		onEdit: (competency: Competency) => void;
		onDelete: (competency: Competency) => void;
		onAdd: () => void;
	}

	let { competencies, pillarName, onEdit, onDelete, onAdd }: Props = $props();
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
				<tr>
					<td class="font-medium">{competency.name}</td>
					<td class="text-base-content/60 text-sm">{competency.description}</td>
					<td class="text-right">
						<button
							class="btn btn-ghost btn-square btn-sm"
							onclick={() => onEdit(competency)}
							aria-label="Editar {competency.name}"
						>
							<Pencil class="w-4 h-4" />
						</button>
						<button
							class="btn btn-ghost btn-square btn-sm text-error"
							onclick={() => onDelete(competency)}
							aria-label="Eliminar {competency.name}"
						>
							<Trash2 class="w-4 h-4" />
						</button>
					</td>
				</tr>
			{/each}
		</tbody>
	</table>

	{#if competencies.length === 0}
		<div class="text-center py-12">
			<p class="text-base-content/50 text-sm mb-4">
				Este pilar aún no tiene competencias.
			</p>
			<button class="btn btn-primary btn-sm" onclick={onAdd}>
				<Plus class="w-4 h-4" />
				Nueva competencia
			</button>
		</div>
	{/if}
</div>
