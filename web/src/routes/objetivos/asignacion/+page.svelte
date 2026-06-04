<script lang="ts">
	import { Save, Plus, Library } from '@lucide/svelte';
	import type { Goal, GoalCategory, GoalUnit } from '$lib/types/goal';
	import {
		getCategories,
		getGoals,
		getKpis,
		getGoalsByCategory,
		getKpisForGoal,
		addCategory,
		updateCategory,
		deleteCategory,
		addGoal,
		updateGoal,
		deleteGoal,
		isAssignmentValid,
		linkKpiToGoal,
		unlinkKpiFromGoal
	} from '$lib/stores/goalsStore.svelte';
	import WeightIndicator from '$lib/components/goals/WeightIndicator.svelte';
	import CategoryCard from '$lib/components/goals/CategoryCard.svelte';
	import CategoryFormModal from '$lib/components/goals/CategoryFormModal.svelte';
	import GoalFormModal from '$lib/components/goals/GoalFormModal.svelte';
	import KpiFormModal from '$lib/components/goals/KpiFormModal.svelte';

	let showCategoryForm = $state(false);
	let showGoalForm = $state(false);
	let showKpiLibrary = $state(false);
	let editingCategory: GoalCategory | null = $state(null);
	let editingGoal: Goal | null = $state(null);
	let goalFormCategoryId = $state('');
	let successMsg = $state('');
	let errorMsg = $state('');

	const categories = $derived(getCategories());
	const goals = $derived(getGoals());
	const allKpis = $derived(getKpis());
	const globalSum = $derived(categories.reduce((sum, c) => sum + c.weight, 0));
	const valid = $derived(isAssignmentValid());

	function handleSaveCategory(data: { name: string; description: string; weight: number }) {
		if (editingCategory) {
			updateCategory(editingCategory.id, data);
		} else {
			const newCat: GoalCategory = {
				id: `cat-${Date.now()}`,
				name: data.name,
				description: data.description,
				weight: data.weight
			};
			addCategory(newCat);
		}
		showCategoryForm = false;
		editingCategory = null;
	}

	function handleEditCategory(cat: GoalCategory) {
		editingCategory = cat;
		showCategoryForm = true;
	}

	function handleDeleteCategory(catId: string) {
		deleteCategory(catId);
	}

	function handleAddGoal(catId: string) {
		goalFormCategoryId = catId;
		editingGoal = null;
		showGoalForm = true;
	}

	function handleEditGoal(goal: Goal) {
		goalFormCategoryId = goal.categoryId;
		editingGoal = goal;
		showGoalForm = true;
	}

	function handleDeleteGoal(goalId: string) {
		deleteGoal(goalId);
	}

	function handleSaveGoal(data: {
		name: string;
		description: string;
		unit: GoalUnit;
		weight: number;
		targetValue: number;
		linkedKpiIds: string[];
	}) {
		if (editingGoal) {
			updateGoal(editingGoal.id, {
				name: data.name,
				description: data.description,
				unit: data.unit,
				weight: data.weight,
				targetValue: data.targetValue
			});
			// Sync KPI links
			const currentLinked = getKpisForGoal(editingGoal.id).map((k) => k.id);
			const toAdd = data.linkedKpiIds.filter((id) => !currentLinked.includes(id));
			const toRemove = currentLinked.filter((id) => !data.linkedKpiIds.includes(id));
			for (const kpiId of toAdd) {
				linkKpiToGoal(editingGoal.id, kpiId);
			}
			for (const kpiId of toRemove) {
				unlinkKpiFromGoal(editingGoal.id, kpiId);
			}
		} else {
			const newGoal: Goal = {
				id: `goal-${Date.now()}`,
				name: data.name,
				description: data.description,
				categoryId: goalFormCategoryId,
				weight: data.weight,
				unit: data.unit,
				targetValue: data.targetValue
			};
			addGoal(newGoal);
			// Link selected KPIs
			for (const kpiId of data.linkedKpiIds) {
				linkKpiToGoal(newGoal.id, kpiId);
			}
		}
		showGoalForm = false;
		editingGoal = null;
	}

	function handleSaveAssignment() {
		successMsg = 'Asignación guardada correctamente.';
		errorMsg = '';
		setTimeout(() => (successMsg = ''), 3000);
	}
</script>

<svelte:head>
	<title>Asignación anual — SED</title>
</svelte:head>

<div class="space-y-6">
	<!-- Page header -->
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div>
			<h1 class="text-2xl font-bold text-base-content">Asignación anual</h1>
			<p class="text-sm text-base-content/50 mt-1">
				Defina las categorías y metas para el período de evaluación.
			</p>
		</div>
		<div class="flex items-center gap-2">
			<button
				class="btn btn-ghost btn-sm"
				onclick={() => (showKpiLibrary = true)}
				aria-label="Biblioteca de KPI"
			>
				<Library class="w-4 h-4" />
				Biblioteca de KPI
			</button>
			<button class="btn btn-primary btn-sm" disabled={!valid} onclick={handleSaveAssignment}>
				<Save class="w-4 h-4" />
				Guardar asignación
			</button>
		</div>
	</div>

	<!-- Global weight indicator -->
	<div class="bg-base-200/50 rounded-lg p-4">
		<p class="text-sm font-semibold text-base-content mb-2">Distribución global de pesos</p>
		<WeightIndicator current={globalSum} label="Suma total de categorías" />
		{#if !valid}
			<p class="text-xs text-warning mt-1">
				La suma de pesos debe ser 100% tanto a nivel global como en cada categoría.
			</p>
		{/if}
	</div>

	<!-- Success/error alerts -->
	{#if successMsg}
		<div class="alert alert-success text-sm" role="status">
			<span>{successMsg}</span>
		</div>
	{/if}
	{#if errorMsg}
		<div class="alert alert-error text-sm" role="alert">
			<span>{errorMsg}</span>
		</div>
	{/if}

	<!-- Category cards -->
	{#if categories.length > 0}
		<div class="space-y-4">
			{#each categories as cat (cat.id)}
				{@const catGoals = getGoalsByCategory(cat.id)}
				<CategoryCard
					category={cat}
					goals={catGoals}
					{getKpisForGoal}
					onEditCategory={handleEditCategory}
					onDeleteCategory={handleDeleteCategory}
					onAddGoal={handleAddGoal}
					onEditGoal={handleEditGoal}
					onDeleteGoal={handleDeleteGoal}
				/>
			{/each}
		</div>
	{:else}
		<div class="text-center py-12 text-base-content/50 text-sm">
			No hay categorías registradas. Cree la primera categoría para comenzar.
		</div>
	{/if}

	<!-- Nueva categoría button -->
	<div class="flex justify-center pt-2">
		<button
			class="btn btn-outline btn-primary"
			onclick={() => {
				editingCategory = null;
				showCategoryForm = true;
			}}
		>
			<Plus class="w-4 h-4" />
			Nueva categoría
		</button>
	</div>
</div>

<!-- Modals -->
<CategoryFormModal
	open={showCategoryForm}
	category={editingCategory}
	onSave={handleSaveCategory}
	onCancel={() => {
		showCategoryForm = false;
		editingCategory = null;
	}}
/>

<GoalFormModal
	open={showGoalForm}
	goal={editingGoal}
	categoryId={goalFormCategoryId}
	allKpis={allKpis}
	onSave={handleSaveGoal}
	onCancel={() => {
		showGoalForm = false;
		editingGoal = null;
	}}
/>

<KpiFormModal
	open={showKpiLibrary}
	onClose={() => (showKpiLibrary = false)}
/>
