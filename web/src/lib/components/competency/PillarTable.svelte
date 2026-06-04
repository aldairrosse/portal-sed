<script lang="ts">
	import { Pencil, Trash2, ChevronRight } from '@lucide/svelte';
	import type { Pillar } from '$lib/types/competency';

	interface Props {
		pillars: Pillar[];
		onEdit: (pillar: Pillar) => void;
		onDelete: (pillar: Pillar) => void;
	}

	let { pillars, onEdit, onDelete }: Props = $props();
</script>

<div class="overflow-x-auto">
	<table class="table table-zebra" aria-label="Lista de pilares">
		<thead>
			<tr>
				<th class="w-1/3">Nombre</th>
				<th class="w-1/2">Descripción</th>
				<th class="w-[140px] text-right">Acciones</th>
			</tr>
		</thead>
		<tbody>
			{#each pillars as pillar (pillar.id)}
				<tr>
					<td class="font-medium">
						<a
							href="/rh/pilares/{pillar.id}/competencias"
							class="link link-hover text-primary flex items-center gap-1.5"
						>
							{pillar.name}
							<ChevronRight class="w-3.5 h-3.5 flex-shrink-0" strokeWidth={2} />
						</a>
					</td>
					<td class="text-base-content/60 text-sm">{pillar.description}</td>
					<td class="text-right">
						<button
							class="btn btn-ghost btn-square btn-sm"
							onclick={() => onEdit(pillar)}
							aria-label="Editar {pillar.name}"
						>
							<Pencil class="w-4 h-4" />
						</button>
						<button
							class="btn btn-ghost btn-square btn-sm text-error"
							onclick={() => onDelete(pillar)}
							aria-label="Eliminar {pillar.name}"
						>
							<Trash2 class="w-4 h-4" />
						</button>
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
</div>
