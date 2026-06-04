<script lang="ts">
	import { Pencil, Trash2 } from '@lucide/svelte';
	import type { Goal, KPI } from '$lib/types/goal';
	import KpiBadge from './KpiBadge.svelte';

	interface Props {
		goal: Goal;
		kpis: KPI[];
		onEdit: (goal: Goal) => void;
		onDelete: (goalId: string) => void;
	}

	let { goal, kpis, onEdit, onDelete }: Props = $props();

	const unitLabels: Record<string, string> = {
		porcentaje: '%',
		moneda: '$',
		numero: '#',
		binario: 'Sí/No'
	};
</script>

<tr>
	<td class="font-medium text-sm">{goal.name}</td>
	<td class="text-sm text-base-content/70">
		{goal.targetValue}{unitLabels[goal.unit] ?? goal.unit}
	</td>
	<td class="text-sm text-base-content/70">{goal.weight}%</td>
	<td>
		{#if kpis.length > 0}
			<div class="flex flex-wrap gap-1">
				{#each kpis as kpi (kpi.id)}
					<KpiBadge {kpi} />
				{/each}
			</div>
		{:else}
			<span class="text-xs text-base-content/30 italic">Sin KPI</span>
		{/if}
	</td>
	<td class="text-right">
		<div class="flex items-center justify-end gap-1">
			<button
				class="btn btn-ghost btn-square btn-xs"
				onclick={() => onEdit(goal)}
				aria-label="Editar {goal.name}"
			>
				<Pencil class="w-3.5 h-3.5" />
			</button>
			<button
				class="btn btn-ghost btn-square btn-xs text-error"
				onclick={() => onDelete(goal.id)}
				aria-label="Eliminar {goal.name}"
			>
				<Trash2 class="w-3.5 h-3.5" />
			</button>
		</div>
	</td>
</tr>
