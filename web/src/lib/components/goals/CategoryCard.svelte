<script lang="ts">
	import { Pencil, Trash2, Plus, MessageSquare, MessageCircle, Check } from '@lucide/svelte';
	import type { Goal, GoalCategory, GoalUnit, KPI, CyclePhase } from '$lib/types/goal';
	import { validateCategory } from './goalValidation';
	import WeightIndicator from './WeightIndicator.svelte';
	import ProgressIndicator from './ProgressIndicator.svelte';
	import GoalRow from './GoalRow.svelte';
	import GoalForm from './GoalForm.svelte';

	interface Props {
		category: GoalCategory;
		goals: Goal[];
		getKpisForGoal: (goalId: string) => KPI[];
		onSaveCategory: (data: { id?: string; name: string; description: string; weight: number }) => void;
		onDeleteCategory: (categoryId: string) => void;
		onSaveGoal: (data: { id?: string; categoryId: string; name: string; description: string; unit: GoalUnit; weight: number; targetValue: number; direction: 'ascendente' | 'descendente'; baselineValue?: number; linkedKpiIds: string[] }) => void;
		onDeleteGoal: (goalId: string) => void;
		mode?: 'editor' | 'reader';
		onRequestChangeCategory?: (category: GoalCategory) => void;
		onSaveProposal?: (goalId: string, data: { name: string; description: string; unit: GoalUnit; weight: number; targetValue: number; direction: 'ascendente' | 'descendente'; baselineValue?: number; kpiIds: string[] }) => void | Promise<void>;
		onAcceptProposal?: (goalId: string, proposalId: string) => void | Promise<void>;
		onRejectProposal?: (goalId: string, proposalId: string) => void | Promise<void>;
		phase?: CyclePhase;
		canDelete?: boolean;
		canAddGoal?: boolean;
		canEditCategory?: boolean;
		canEditProgress?: boolean;
		canComment?: boolean;
		allKpis: KPI[];
		isAnyInlineEditing?: boolean;
		onUpdateProgress?: (goalId: string, progress: number) => void;
		onOpenComments?: (goal: Goal) => void;
		onOpenCategoryComments?: (category: GoalCategory) => void;
	}

	let {
		category,
		goals,
		getKpisForGoal,
		onSaveCategory,
		onDeleteCategory,
		onSaveGoal,
		onDeleteGoal,
		mode = 'editor',
		onRequestChangeCategory,
		onSaveProposal,
		onAcceptProposal,
		onRejectProposal,
		phase = 'inicio-anio',
		canDelete = true,
		canAddGoal = true,
		canEditCategory = true,
		canEditProgress = false,
		canComment = false,
		allKpis,
		isAnyInlineEditing = $bindable(false),
		onUpdateProgress,
		onOpenComments,
		onOpenCategoryComments
	}: Props = $props();

	// ─── Category inline edit state ────────────────────────────────────────

	let isEditingCategory = $state(false);
	let editCatName = $state('');
	let editCatDesc = $state('');
	let editCatWeight = $state(0);
	let editCatError = $state('');

	function handleStartEditCategory() {
		editCatName = category.name;
		editCatDesc = category.description;
		editCatWeight = category.weight;
		editCatError = '';
		isEditingCategory = true;
	}

	function handleCancelEditCategory() {
		isEditingCategory = false;
		editCatError = '';
	}

	async function handleSaveCategoryInline() {
		const err = validateCategory({ name: editCatName, description: editCatDesc, weight: editCatWeight, categoryId: category.id });
		if (err) { editCatError = err; return; }
		await onSaveCategory({ id: category.id, name: editCatName.trim(), description: editCatDesc.trim(), weight: editCatWeight });
		isEditingCategory = false;
	}

	// ─── Goal creation inline state ────────────────────────────────────────

	let isCreatingGoal = $state(false);
	let newGoalError = $state('');

	function handleStartCreateGoal() {
		newGoalError = '';
		isCreatingGoal = true;
	}

	function handleCancelCreateGoal() {
		isCreatingGoal = false;
		newGoalError = '';
	}

	async function handleSaveNewGoal(data: {
		name: string; description: string; unit: GoalUnit;
		weight: number; targetValue: number;
		direction: 'ascendente' | 'descendente';
		baselineValue?: number; kpiIds: string[];
	}) {
		try {
			await onSaveGoal({ categoryId: category.id, ...data, linkedKpiIds: data.kpiIds });
			isCreatingGoal = false;
		} catch (e) {
			newGoalError = e instanceof Error ? e.message : 'Error al crear meta';
		}
	}

	// ─── Computed ──────────────────────────────────────────────────────────

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

	// ─── GoalRow coordination state ────────────────────────────────────────

	let editingGoalId = $state<string | null>(null);
</script>

<div class="card bg-base-100 border border-base-300 max-w-full">
	<div class="card-body px-0 py-5 min-w-0">
		<!-- Header -->
		<div class="flex flex-wrap items-start justify-between gap-4 mb-4">
			{#if isEditingCategory}
				<div class="flex-1 min-w-0 w-full">
					<div class="border border-base-300 rounded-lg p-4 bg-base-200/50 w-full">
						{#if editCatError}<div class="alert alert-error text-sm mb-3" role="alert"><span>{editCatError}</span></div>{/if}
						<div class="grid grid-cols-1 md:grid-cols-3 gap-3">
							<div class="form-control">
								<label class="label" for="edit-cat-name-{category.id}"><span class="label-text text-xs">Nombre</span></label>
								<input id="edit-cat-name-{category.id}" type="text" class="input input-bordered input-sm w-full" bind:value={editCatName} placeholder="Nombre" required />
							</div>
							<div class="form-control">
								<label class="label" for="edit-cat-desc-{category.id}"><span class="label-text text-xs">Descripción</span></label>
								<textarea id="edit-cat-desc-{category.id}" class="textarea textarea-bordered textarea-sm w-full" rows={1} bind:value={editCatDesc} placeholder="Descripción" required></textarea>
							</div>
							<div class="form-control">
								<label class="label" for="edit-cat-weight-{category.id}"><span class="label-text text-xs">Peso (%)</span></label>
								<input id="edit-cat-weight-{category.id}" type="number" class="input input-bordered input-sm w-full" bind:value={editCatWeight} min={0} max={100} step={0.1} required />
							</div>
						</div>
						<div class="flex justify-end gap-2 mt-3">
							<button class="btn btn-ghost btn-sm" onclick={handleCancelEditCategory}>Cancelar</button>
							<button class="btn btn-primary btn-sm" onclick={handleSaveCategoryInline}><Check class="w-4 h-4" /> Guardar categoría</button>
						</div>
					</div>
				</div>
			{:else}
				<div class="flex-1 min-w-0">
					<div class="flex items-center gap-2 mb-1">
						<h3 class="text-lg font-semibold text-base-content">{category.name}</h3>
						<span class="badge badge-md font-mono">{category.weight}%</span>
					</div>
					<p class="text-xs text-base-content/50 truncate">{category.description}</p>
				</div>
				<div class="flex items-center gap-1">
					{#if phase === 'medio-anio' || phase === 'fin-anio'}
						<!-- No category edit/delete in avance or cierre mode -->
				{:else if mode === 'editor' && canEditCategory}
					{#if (category.comments?.length ?? 0) > 0}
						<button
							class="btn btn-ghost btn-square btn-sm relative"
							title="Comentarios"
							onclick={() => onOpenCategoryComments?.(category)}
							aria-label="Comentarios de categoría {category.name}"
						>
							<MessageCircle class="w-4 h-4" />
							<span class="badge badge-xs badge-primary absolute -top-1.5 -right-1.5">{category.comments?.length}</span>
						</button>
					{/if}
					<button
						class="btn btn-ghost btn-square btn-sm"
						title="Editar"
						onclick={handleStartEditCategory}
						disabled={isAnyInlineEditing}
						aria-label="Editar categoría {category.name}"
					>
						<Pencil class="w-4 h-4" />
					</button>
						{#if canDelete}
							<button
								class="btn btn-ghost btn-square btn-sm text-error"
								title="Eliminar"
								onclick={() => onDeleteCategory(category.id)}
								disabled={isAnyInlineEditing}
								aria-label="Eliminar categoría {category.name}"
							>
								<Trash2 class="w-4 h-4" />
							</button>
						{/if}
				{:else if onOpenCategoryComments}
					<button
						class="btn btn-sm relative"
						title="Comentar"
						onclick={() => onOpenCategoryComments?.(category)}
						aria-label="Comentar en categoría {category.name}"
					>
						<MessageCircle class="w-4 h-4" />
						Comentar
						{#if (category.comments?.length ?? 0) > 0}
							<span class="badge badge-xs badge-primary absolute -top-2 -right-2">{category.comments?.length}</span>
						{/if}
					</button>
				{/if}
				</div>
			{/if}
		</div>

		<!-- Indicators -->
		<div class="mb-4 flex items-center gap-4">
			{#if phase === 'medio-anio' || phase === 'fin-anio'}
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
								{phase === 'medio-anio' || phase === 'fin-anio' ? 'Avance' : 'KPI'}
							</th>
							{#if phase !== 'fin-anio'}
							<th class="text-xs font-semibold text-base-content/60 text-right">Acciones</th>
						{/if}
						</tr>
					</thead>
					<tbody>
						{#each goals as goal (goal.id)}
							<GoalRow
								{goal}
								kpis={getKpisForGoal(goal.id)}
								{mode}
								{onSaveProposal}
								{onAcceptProposal}
								{onRejectProposal}
								{phase}
								onSaveGoal={onSaveGoal}
								onDeleteGoal={onDeleteGoal}
								{allKpis}
								{editingGoalId}
								onEditingChange={(id) => { editingGoalId = id; }}
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

		<!-- Inline goal creation -->
		{#if isCreatingGoal}
			<GoalForm
				mode="create"
				categoryId={category.id}
				{allKpis}
				submitLabel="Guardar meta"
				error={newGoalError}
				onsave={handleSaveNewGoal}
				oncancel={handleCancelCreateGoal}
			/>
		{:else if mode === 'editor' && canAddGoal && phase !== 'medio-anio' && phase !== 'fin-anio'}
			<div class="mt-3">
				<button class="btn btn-outline btn-primary btn-sm" disabled={isAnyInlineEditing || isEditingCategory} onclick={handleStartCreateGoal}>
					<Plus class="w-4 h-4" /> Nueva meta
				</button>
			</div>
		{/if}
	</div>
</div>
