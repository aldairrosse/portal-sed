<script lang="ts">
	import { Pencil, Trash2, Plus } from '@lucide/svelte';
	import type { Goal, GoalCategory, KPI } from '$lib/types/goal';
	import WeightIndicator from './WeightIndicator.svelte';
	import GoalRow from './GoalRow.svelte';

	interface Props {
		category: GoalCategory;
		goals: Goal[];
		getKpisForGoal: (goalId: string) => KPI[];
		onEditCategory: (category: GoalCategory) => void;
		onDeleteCategory: (categoryId: string) => void;
		onAddGoal: (categoryId: string) => void;
		onEditGoal: (goal: Goal) => void;
		onDeleteGoal: (goalId: string) => void;
	}

	let {
		category,
		goals,
		getKpisForGoal,
		onEditCategory,
		onDeleteCategory,
		onAddGoal,
		onEditGoal,
		onDeleteGoal
	}: Props = $props();
</script>

<div class="card bg-base-100 border border-base-300">
	<div class="card-body p-5">
		<!-- Header -->
		<div class="flex items-start justify-between gap-4 mb-4">
			<div class="flex-1 min-w-0">
				<div class="flex items-center gap-2 mb-1">
					<h3 class="text-base font-semibold text-base-content">{category.name}</h3>
					<span class="badge badge-md font-mono">{category.weight}%</span>
				</div>
				<p class="text-xs text-base-content/50 truncate">{category.description}</p>
			</div>
			<div class="flex items-center gap-1 flex-shrink-0">
				<button
					class="btn btn-ghost btn-square btn-sm"
					onclick={() => onEditCategory(category)}
					aria-label="Editar categoría {category.name}"
				>
					<Pencil class="w-4 h-4" />
				</button>
				<button
					class="btn btn-ghost btn-square btn-sm text-error"
					onclick={() => onDeleteCategory(category.id)}
					aria-label="Eliminar categoría {category.name}"
				>
					<Trash2 class="w-4 h-4" />
				</button>
			</div>
		</div>

		<!-- Weight indicator for goals within this category -->
		<div class="mb-4">
			<WeightIndicator
				current={goals.reduce((sum, g) => sum + g.weight, 0)}
				label="Peso de metas en {category.name}"
			/>
		</div>

		<!-- Goals table -->
		{#if goals.length > 0}
			<div class="overflow-x-auto">
				<table class="table table-sm" aria-label="Metas de {category.name}">
					<thead>
						<tr>
							<th class="text-xs font-semibold text-base-content/60">Meta</th>
							<th class="text-xs font-semibold text-base-content/60">Valor objetivo</th>
							<th class="text-xs font-semibold text-base-content/60">Peso</th>
							<th class="text-xs font-semibold text-base-content/60">KPI</th>
							<th class="text-xs font-semibold text-base-content/60 text-right">Acciones</th>
						</tr>
					</thead>
					<tbody>
						{#each goals as goal (goal.id)}
							<GoalRow
								{goal}
								kpis={getKpisForGoal(goal.id)}
								onEdit={onEditGoal}
								onDelete={onDeleteGoal}
							/>
						{/each}
					</tbody>
				</table>
			</div>
		{:else}
			<p class="text-sm text-base-content/30 italic text-center py-4">
				Sin metas registradas en esta categoría.
			</p>
		{/if}

		<!-- Add goal button -->
		<div class="mt-3">
			<button
				class="btn btn-outline btn-primary btn-sm"
				onclick={() => onAddGoal(category.id)}
			>
				<Plus class="w-4 h-4" />
				Nueva meta
			</button>
		</div>
	</div>
</div>
