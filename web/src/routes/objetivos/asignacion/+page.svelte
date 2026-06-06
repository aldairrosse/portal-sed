<script lang="ts">
    import { goto } from "$app/navigation";
    import { Save, Plus, Library, MessageSquare } from "@lucide/svelte";
    import type {
        Goal,
        GoalCategory,
        GoalUnit,
        GoalComment,
    } from "$lib/types/goal";
    import type { ChangeRequest } from "$lib/types/goal";
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
        getAssignmentsByProfile,
        getAssignments,
        getCyclePhase,
        getGoalPermissions,
        updateGoalProgress,
        addGoalComment,
        deleteGoalComment,
        getGoalComments,
    } from "$lib/stores/goalsStore.svelte";
    import { getProfile } from "$lib/stores/devContext.svelte";
    import { getChildren } from "$lib/stores/orgHierarchyStore.svelte";
    import WeightIndicator from "$lib/components/goals/WeightIndicator.svelte";
    import ProgressIndicator from "$lib/components/goals/ProgressIndicator.svelte";
    import CategoryCard from "$lib/components/goals/CategoryCard.svelte";
    import CategoryFormModal from "$lib/components/goals/CategoryFormModal.svelte";
    import GoalFormModal from "$lib/components/goals/GoalFormModal.svelte";
    import ReadOnlyBanner from "$lib/components/goals/ReadOnlyBanner.svelte";
    import AssigneePicker from "$lib/components/goals/AssigneePicker.svelte";
    import RequestChangeModal from "$lib/components/goals/RequestChangeModal.svelte";
    import CommentPopover from "$lib/components/goals/GoalCommentModal.svelte";

    // ─── Mode detection ──────────────────────────────────────────────────────

    const viewerProfile = $derived(getProfile());

    const ownAssignment = $derived(getAssignmentsByProfile(viewerProfile)[0]);
    const currentUserId = $derived(ownAssignment?.employeeId ?? "");

    const children = $derived(getChildren(currentUserId));
    const childIds = $derived(children.map((c) => c.id));

    const allAssignments = $derived(getAssignments());
    const subordinateAssignments = $derived(
        allAssignments.filter((a) => childIds.includes(a.employeeId)),
    );

    const availableAssignments = $derived(
        ownAssignment
            ? [ownAssignment, ...subordinateAssignments]
            : [...subordinateAssignments],
    );

    const showAssigneePicker = $derived(children.length > 0);

    let selectedEmployeeId = $state("");

    // Reset selected employee when assignment context changes (profile switch)
    $effect(() => {
        if (ownAssignment && !selectedEmployeeId) {
            selectedEmployeeId = ownAssignment.employeeId;
        }
    });

    const targetAssignment = $derived(
        availableAssignments.find((a) => a.employeeId === selectedEmployeeId),
    );

    const mode = $derived<"editor" | "reader">(
        targetAssignment?.profileId === viewerProfile ? "editor" : "reader",
    );

    const targetEmployeeName = $derived(targetAssignment?.employeeName ?? "");

    // ─── Cycle phase & permissions ──────────────────────────────────────────

    const phase = $derived(getCyclePhase());
    const isOwner = $derived(mode === "editor");
    const permissions = $derived(getGoalPermissions(viewerProfile, isOwner));

    // ─── Comment modal state ────────────────────────────────────────────────

    let commentGoal: Goal | null = $state(null);
    let commentGoalComments = $state<GoalComment[]>([]);
    let showCommentModal = $state(false);

    function openComments(goal: Goal) {
        commentGoal = goal;
        commentGoalComments = getGoalComments(goal.id);
        showCommentModal = true;
    }

    function handleAddComment(goalId: string, content: string) {
        addGoalComment(goalId, viewerProfile, viewerProfile, content);
        commentGoalComments = getGoalComments(goalId);
    }

    function handleDeleteComment(goalId: string, commentId: string) {
        deleteGoalComment(goalId, commentId);
        commentGoalComments = getGoalComments(goalId);
    }

    // ─── Existing page state ─────────────────────────────────────────────────

    let showCategoryForm = $state(false);
    let showGoalForm = $state(false);
    let editingCategory: GoalCategory | null = $state(null);
    let editingGoal: Goal | null = $state(null);
    let goalFormCategoryId = $state("");
    let successMsg = $state("");
    let errorMsg = $state("");

    const categories = $derived(getCategories());
    const goals = $derived(getGoals());
    const allKpis = $derived(getKpis());
    const globalSum = $derived(
        categories.reduce((sum, c) => sum + c.weight, 0),
    );
    const valid = $derived(isAssignmentValid());

    // ─── Request change modal state ─────────────────────────────────────────

    let showRequestModal = $state(false);
    let requestEntityType: ChangeRequest["entityType"] = $state("goal");
    let requestEntityId = $state("");
    let requestEntityName = $state("");

    function openRequestModal(
        type: ChangeRequest["entityType"],
        id: string,
        name: string,
    ) {
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

    function handleSaveCategory(data: {
        name: string;
        description: string;
        weight: number;
    }) {
        if (editingCategory) {
            updateCategory(editingCategory.id, data);
        } else {
            const newCat: GoalCategory = {
                id: `cat-${Date.now()}`,
                name: data.name,
                description: data.description,
                weight: data.weight,
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
                targetValue: data.targetValue,
            });
            // Sync KPI links
            const currentLinked = getKpisForGoal(editingGoal.id).map(
                (k) => k.id,
            );
            const toAdd = data.linkedKpiIds.filter(
                (id) => !currentLinked.includes(id),
            );
            const toRemove = currentLinked.filter(
                (id) => !data.linkedKpiIds.includes(id),
            );
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
                targetValue: data.targetValue,
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
        successMsg = "Asignación guardada correctamente.";
        errorMsg = "";
        setTimeout(() => (successMsg = ""), 3000);
    }

    function handleRequestChangeGoal(goal: Goal) {
        openRequestModal("goal", goal.id, goal.name);
    }

    function handleRequestChangeCategory(category: GoalCategory) {
        openRequestModal("category", category.id, category.name);
    }

    function handleRequestAssignmentChange() {
        if (!targetAssignment) return;
        openRequestModal("assignment", targetAssignment.id, targetEmployeeName);
    }

    function handleUpdateProgress(goalId: string, progress: number) {
        updateGoalProgress(goalId, progress);
    }
</script>

<svelte:head>
    <title>Asignación anual — SED</title>
</svelte:head>

<div class="space-y-6 max-w-full min-w-0">
    <!-- Page header -->
    <div>
        <div>
            <h1 class="text-2xl font-bold text-base-content">
                {phase === "medio-anio"
                    ? "Avance de metas"
                    : "Asignación anual"}
            </h1>
            <p class="text-sm text-base-content/50 mt-1">
                {phase === "medio-anio"
                    ? "Registre el avance de sus metas y agregue comentarios."
                    : "Defina las categorías y metas para el período de evaluación."}
            </p>
        </div>
        <div class="flex items-center gap-2 mt-3">
            {#if showAssigneePicker}
                <AssigneePicker
                    assignments={availableAssignments}
                    {selectedEmployeeId}
                    onSelect={handleAssigneeSelect}
                    {currentUserId}
                />
            {/if}
            {#if phase !== "medio-anio"}
                <button
                    class="btn btn-ghost btn-sm"
                    onclick={() => goto("/objetivos/asignacion/biblioteca")}
                    aria-label="Biblioteca de KPI"
                >
                    <Library class="w-4 h-4" />
                    Biblioteca de KPI
                </button>
            {/if}
            {#if mode === "editor" && phase !== "medio-anio" && phase !== "fin-anio"}
                <button
                    class="btn btn-primary btn-sm ml-auto"
                    disabled={!valid}
                    onclick={handleSaveAssignment}
                >
                    <Save class="w-4 h-4" />
                    Guardar asignación
                </button>
            {:else if mode === "reader"}
                <button
                    class="btn btn-warning btn-sm ml-auto"
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
    {#if mode === "reader"}
        <ReadOnlyBanner employeeName={targetEmployeeName} {phase} />
    {/if}

    <!-- Global weight indicator (sticky) -->
    <div
        class="sticky top-2 z-30 bg-base-200/95 backdrop-blur-sm rounded-lg p-4 mt-2 mb-4 border border-base-300 shadow-sm min-w-0"
    >
        <p class="text-sm font-semibold text-base-content mb-2">
            {phase === "medio-anio"
                ? "Avance global de metas"
                : "Distribución global de metas"}
        </p>
        {#if phase === "medio-anio"}
            {@const allGoals = goals}
            {@const withProgress = allGoals.filter(
                (g) => g.progress !== undefined,
            )}
            {@const avgProgress =
                withProgress.length > 0
                    ? withProgress.reduce((acc, g) => {
                          const pct =
                              g.unit === "porcentaje"
                                  ? (g.progress ?? 0)
                                  : ((g.progress ?? 0) / (g.targetValue || 1)) *
                                    100;
                          return acc + Math.min(pct, 100);
                      }, 0) / withProgress.length
                    : 0}
            <ProgressIndicator
                value={avgProgress}
                label="Avance promedio total"
                color="primary"
            />
        {:else}
            <WeightIndicator
                current={globalSum}
                label="Suma total de categorías"
            />
            {#if !valid}
                <p class="text-xs text-warning mt-1">
                    La suma de pesos debe ser 100% tanto a nivel global como en
                    cada categoría.
                </p>
            {/if}
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
        <div class="space-y-4 min-w-0">
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
                    {phase}
                    canDelete={permissions.canDelete}
                    canAddGoal={permissions.canDelete}
                    canEditCategory={permissions.canEditWeight}
                    canEditProgress={permissions.canEditProgress}
                    canComment={permissions.canComment}
                    onUpdateProgress={handleUpdateProgress}
                    onOpenComments={openComments}
                />
            {/each}
        </div>
    {:else}
        <div class="text-center py-12 text-base-content/50 text-sm">
            No hay categorías registradas. Cree la primera categoría para
            comenzar.
        </div>
    {/if}

    <!-- Nueva categoría button (editor only, not in avance or cierre mode) -->
    {#if mode === "editor" && phase !== "medio-anio" && phase !== "fin-anio"}
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
    {allKpis}
    onSave={handleSaveGoal}
    onCancel={() => {
        showGoalForm = false;
        editingGoal = null;
    }}
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

{#if showCommentModal && commentGoal}
    <CommentPopover
        open={showCommentModal}
        goal={commentGoal}
        comments={commentGoalComments}
        onAdd={handleAddComment}
        onDelete={handleDeleteComment}
        onClose={() => (showCommentModal = false)}
        currentUserId={viewerProfile}
    />
{/if}
