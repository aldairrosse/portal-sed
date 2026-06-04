<script lang="ts">
	import { Pencil, Trash2, Plus, MessageSquare } from '@lucide/svelte';
	import type { Goal, GoalCategory, KPI, CyclePhase } from '$lib/types/goal';
	import WeightIndicator from './WeightIndicator.svelte';
	import ProgressIndicator from './ProgressIndicator.svelte';
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
		mode?: 'editor' | 'reader';
		onRequestChangeCategory?: (category: GoalCategory) => void;
		onRequestChangeGoal?: (goal: Goal) => void;
		phase?: CyclePhase;
		canDelete?: boolean;
		canAddGoal?: boolean;
		canEditCategory?: boolean;
		canEditProgress?: boolean;
		canComment?: boolean;
		onUpdateProgress?: (goalId: string, progress: number) => void;
		onOpenComments?: (goal: Goal) => void;
	}

	let {
		category,
		goals,
		getKpisForGoal,
		onEditCategory,
		onDeleteCategory,
		onAddGoal,
		onEditGoal,
		onDeleteGoal,
		mode = 'editor',
		onRequestChangeCategory,
		onRequestChangeGoal,
		phase = 'asignacion',
		canDelete = true,
		canAddGoal = true,
		canEditCategory = true,
		canEditProgress = false,
		canComment = false,
		onUpdateProgress,
		onOpenComments
	}: Props = $props();

	let categoryProgress = $derived.by(() => {
		if (goals.length === 0) return 0;
		const withProgress = goals.filter((g) => g.progress !== undefined);
		if (withProgress.length === 0) return 0;
		const total = withProgress.reduce((acc, g) => {
			const pct = g.unit === 'porcentaje' ? (g.progress ?? 0) : ((g.progress ?? 0) / (g.targetValue || 1)) * 100;
			return acc + Math.min(pct, 100);
		}, 0);
		return total / withProgress.length;
	});
</script>

<div class="card bg-base-100 border border-base-300 max-w-full">
	<div class="card-body px-0 py-5 min-w-0">
		<!-- Header -->
		<div class="flex flex-wrap items-start justify-between gap-4 mb-4">
			<div class="flex-1 min-w-0">
				<div class="flex items-center gap-2 mb-1">
					<h3 class="text-lg font-semibold text-base-content">{category.name}</h3>
					<span class="badge badge-md font-mono">{category.weight}%</span>
				</div>
				<p class="text-xs text-base-content/50 truncate">{category.description}</p>
			</div>
			<div class="flex items-center gap-1">
				{#if phase === 'avance'}
					<!-- No category edit/delete in avance mode -->
				{:else if mode === 'editor' && canEditCategory}
					<button
						class="btn btn-ghost btn-square btn-sm"
						title="Editar"
						onclick={() => onEditCategory(category)}
						aria-label="Editar categoría {category.name}"
					>
						<Pencil class="w-4 h-4" />
					</button>
					{#if canDelete}
						<button
							class="btn btn-ghost btn-square btn-sm text-error"
							title="Eliminar"
							onclick={() => onDeleteCategory(category.id)}
							aria-label="Eliminar categoría {category.name}"
						>
							<Trash2 class="w-4 h-4" />
						</button>
					{/if}
				{:else if onRequestChangeCategory}
					<button
						class="btn btn-ghost btn-sm text-warning"
						title="Solicitar cambio"
						onclick={() => onRequestChangeCategory(category)}
						aria-label="Solicitar cambio en categoría {category.name}"
					>
						<MessageSquare class="w-4 h-4" />
						Solicitar cambio
					</button>
				{/if}
			</div>
		</div>

		<!-- Indicators -->
		<div class="mb-4 flex items-center gap-4">
			{#if phase === 'avance'}
				<ProgressIndicator value={categoryProgress} label="Avance promedio" />
			{:else}
				<WeightIndicator
					current={goals.reduce((sum, g) => sum + g.weight, 0)}
					label="Peso de metas en {category.name}"
				/>
			{/if}
		</div>

		<!-- Goals table -->
		{#if goals.length > 0}
			<div class="w-full max-w-full overflow-x-auto">
				<table class="table table-sm w-full" aria-label="Metas de {category.name}">
					<thead>
						<tr>
							<th class="text-xs font-semibold text-base-content/60">Meta</th>
							<th class="text-xs font-semibold text-base-content/60">Valor objetivo</th>
							<th class="text-xs font-semibold text-base-content/60">Peso</th>
							<th class="text-xs font-semibold text-base-content/60">
								{phase === 'avance' ? 'Avance' : 'KPI'}
							</th>
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
								{mode}
								onRequestChange={onRequestChangeGoal}
								{phase}
								{canEditProgress}
								{canComment}
								{canDelete}
								{onUpdateProgress}
								{onOpenComments}
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

		<!-- Add goal button (editor only, not in avance mode) -->
		{#if mode === 'editor' && canAddGoal && phase !== 'avance'}
			<div class="mt-3">
				<button
					class="btn btn-outline btn-primary btn-sm"
					onclick={() => onAddGoal(category.id)}
				>
					<Plus class="w-4 h-4" />
					Nueva meta
				</button>
			</div>
		{/if}
	</div>
</div>
