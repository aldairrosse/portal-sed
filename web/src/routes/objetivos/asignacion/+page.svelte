<script lang="ts">
	import { Save, Plus, Library, MessageSquare } from '@lucide/svelte';
	import type { Goal, GoalCategory, GoalUnit } from '$lib/types/goal';
	import type { ChangeRequest } from '$lib/types/goal';
	import { MANAGER_MAP } from '$lib/types/goal';
	import type { EvaluationProfile } from '$lib/types/evaluation';
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
		unlinkKpiFromGoal,
		getAssignmentsByProfile
	} from '$lib/stores/goalsStore.svelte';
	import { getProfile } from '$lib/stores/devContext.svelte';
	import WeightIndicator from '$lib/components/goals/WeightIndicator.svelte';
	import CategoryCard from '$lib/components/goals/CategoryCard.svelte';
	import CategoryFormModal from '$lib/components/goals/CategoryFormModal.svelte';
	import GoalFormModal from '$lib/components/goals/GoalFormModal.svelte';
	import KpiFormModal from '$lib/components/goals/KpiFormModal.svelte';
	import ReadOnlyBanner from '$lib/components/goals/ReadOnlyBanner.svelte';
	import AssigneePicker from '$lib/components/goals/AssigneePicker.svelte';
	import RequestChangeModal from '$lib/components/goals/RequestChangeModal.svelte';

	// ─── Mode detection ──────────────────────────────────────────────────────

	const viewerProfile = $derived(getProfile());

	// Build inverse manager map: which profiles report to the current viewer
	const subordinateProfiles = $derived<EvaluationProfile[]>(
		(Object.entries(MANAGER_MAP) as [EvaluationProfile, EvaluationProfile][])
			.filter(([, mgr]) => mgr === viewerProfile)
			.map(([sub]) => sub)
	);

	const ownAssignment = $derived(getAssignmentsByProfile(viewerProfile)[0]);

	const subordinateAssignments = $derived(
		subordinateProfiles.flatMap((p) => getAssignmentsByProfile(p))
	);

	const availableAssignments = $derived(
		ownAssignment ? [ownAssignment, ...subordinateAssignments] : [...subordinateAssignments]
	);

	const showAssigneePicker = $derived(subordinateProfiles.length > 0);

	let selectedEmployeeId = $state('');

	// Reset selected employee when assignment context changes (profile switch)
	$effect(() => {
		if (ownAssignment && !selectedEmployeeId) {
			selectedEmployeeId = ownAssignment.employeeId;
		}
	});

	const targetAssignment = $derived(
		availableAssignments.find((a) => a.employeeId === selectedEmployeeId)
	);

	const mode = $derived<'editor' | 'reader'>(
		targetAssignment?.profileId === viewerProfile ? 'editor' : 'reader'
	);

	const targetEmployeeName = $derived(targetAssignment?.employeeName ?? '');

	// ─── Existing page state ─────────────────────────────────────────────────

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

	// ─── Request change modal state ─────────────────────────────────────────

	let showRequestModal = $state(false);
	let requestEntityType: ChangeRequest['entityType'] = $state('goal');
	let requestEntityId = $state('');
	let requestEntityName = $state('');

	function openRequestModal(type: ChangeRequest['entityType'], id: string, name: string) {
		requestEntityType = type;
		requestEntityId = id;
		requestEntityName = name;
		showRequestModal = true;
	}

	function closeRequestModal() {
		showRequestModal = false;
	}

	// ─── Handlers ────────────────────────────────────────────────────────────

	function handleAssigneeSelect(employeeId: string) {
		selectedEmployeeId = employeeId;
	}

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

	function handleRequestChangeGoal(goal: Goal) {
		openRequestModal('goal', goal.id, goal.name);
	}

	function handleRequestChangeCategory(category: GoalCategory) {
		openRequestModal('category', category.id, category.name);
	}

	function handleRequestAssignmentChange() {
		if (!targetAssignment) return;
		openRequestModal('assignment', targetAssignment.id, targetEmployeeName);
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
			<!-- AssigneePicker for profiles with subordinates -->
			{#if showAssigneePicker}
				<AssigneePicker
					assignments={availableAssignments}
					selectedEmployeeId={selectedEmployeeId}
					onSelect={handleAssigneeSelect}
				/>
			{/if}
			<button
				class="btn btn-ghost btn-sm"
				onclick={() => (showKpiLibrary = true)}
				aria-label="Biblioteca de KPI"
			>
				<Library class="w-4 h-4" />
				Biblioteca de KPI
			</button>
			{#if mode === 'editor'}
				<button class="btn btn-primary btn-sm" disabled={!valid} onclick={handleSaveAssignment}>
					<Save class="w-4 h-4" />
					Guardar asignación
				</button>
			{:else}
				<button
					class="btn btn-warning btn-sm"
					onclick={handleRequestAssignmentChange}
					aria-label="Solicitar cambio en asignación"
				>
					<MessageSquare class="w-4 h-4" />
					Solicitar cambio
				</button>
			{/if}
		</div>
	</div>

	<!-- Read-only banner -->
	{#if mode === 'reader'}
		<ReadOnlyBanner employeeName={targetEmployeeName} />
	{/if}

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
					{mode}
					onRequestChangeCategory={handleRequestChangeCategory}
					onRequestChangeGoal={handleRequestChangeGoal}
				/>
			{/each}
		</div>
	{:else}
		<div class="text-center py-12 text-base-content/50 text-sm">
			No hay categorías registradas. Cree la primera categoría para comenzar.
		</div>
	{/if}

	<!-- Nueva categoría button (editor only) -->
	{#if mode === 'editor'}
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
	{/if}
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

{#if targetAssignment}
	<RequestChangeModal
		open={showRequestModal}
		entityType={requestEntityType}
		entityId={requestEntityId}
		entityName={requestEntityName}
		requestedBy={viewerProfile}
		onClose={closeRequestModal}
	/>
{/if}
