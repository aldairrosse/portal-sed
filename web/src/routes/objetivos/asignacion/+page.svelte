<script lang="ts">
    import { goto } from "$app/navigation";
    import { Save, Plus, Library, MessageSquare, Check } from "@lucide/svelte";
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
    import { validateCategory, validateGoal, UNIT_OPTIONS } from "$lib/components/goals/goalValidation";
    import CategoryCard from "$lib/components/goals/CategoryCard.svelte";
    import ReadOnlyBanner from "$lib/components/goals/ReadOnlyBanner.svelte";
    import AssigneePicker from "$lib/components/goals/AssigneePicker.svelte";
    import RequestChangeModal from "$lib/components/goals/RequestChangeModal.svelte";
    import CommentPopover from "$lib/components/goals/GoalCommentModal.svelte";
    import { toCsv } from "$lib/utils/export";

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

    let successMsg = $state("");
    let errorMsg = $state("");
    let creatingCategory = $state(false);
    let isAnyInlineEditing = $state(false);
    let newCatName = $state('');
    let newCatDesc = $state('');
    let newCatWeight = $state(0);
    let newCatError = $state('');

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

    function handleSaveCategory(data: { id?: string; name: string; description: string; weight: number }) {
        if (data.id) {
            updateCategory(data.id, { name: data.name, description: data.description, weight: data.weight });
        } else {
            const newCat: GoalCategory = { id: `cat-${Date.now()}`, name: data.name, description: data.description, weight: data.weight };
            addCategory(newCat);
        }
        creatingCategory = false;
        isAnyInlineEditing = false;
        successMsg = data.id ? 'Categoría actualizada correctamente.' : 'Categoría creada correctamente.';
        setTimeout(() => (successMsg = ""), 3000);
    }

    function handleDeleteCategory(catId: string) {
        deleteCategory(catId);
    }

    function handleDeleteGoal(goalId: string) {
        deleteGoal(goalId);
    }

    function handleSaveGoal(data: { id?: string; categoryId: string; name: string; description: string; unit: GoalUnit; weight: number; targetValue: number; linkedKpiIds: string[] }) {
        if (data.id) {
            updateGoal(data.id, { name: data.name, description: data.description, unit: data.unit, weight: data.weight, targetValue: data.targetValue });
            const currentLinked = getKpisForGoal(data.id).map(k => k.id);
            const toAdd = data.linkedKpiIds.filter(id => !currentLinked.includes(id));
            const toRemove = currentLinked.filter(id => !data.linkedKpiIds.includes(id));
            for (const kpiId of toAdd) linkKpiToGoal(data.id, kpiId);
            for (const kpiId of toRemove) unlinkKpiFromGoal(data.id, kpiId);
        } else {
            const newGoal: Goal = { id: `goal-${Date.now()}`, name: data.name, description: data.description, categoryId: data.categoryId, weight: data.weight, unit: data.unit, targetValue: data.targetValue };
            addGoal(newGoal);
            for (const kpiId of data.linkedKpiIds) linkKpiToGoal(newGoal.id, kpiId);
        }
        isAnyInlineEditing = false;
        successMsg = data.id ? 'Meta actualizada correctamente.' : 'Meta creada correctamente.';
        setTimeout(() => (successMsg = ""), 3000);
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

    function handleExportCsv() {
        const rows: Record<string, string | number | null>[] = [];
        for (const cat of categories) {
            const catGoals = getGoalsByCategory(cat.id);
            for (const goal of catGoals) {
                const kpis = getKpisForGoal(goal.id);
                const kpiNames = kpis.map((k) => k.name).join(", ");
                rows.push({
                    Categoría: cat.name,
                    "Peso categoría %": cat.weight,
                    Meta: goal.name,
                    Descripción: goal.description,
                    Unidad: goal.unit,
                    "Peso meta %": goal.weight,
                    "Valor objetivo": goal.targetValue,
                    KPIs: kpiNames || "",
                });
            }
        }
        toCsv(rows, "asignacion-anual.csv");
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
            <button
                class="btn btn-outline btn-sm"
                disabled={categories.length === 0}
                onclick={handleExportCsv}
            >
                Exportar CSV
            </button>
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
                    onSaveCategory={handleSaveCategory}
                    onDeleteCategory={handleDeleteCategory}
                    onSaveGoal={handleSaveGoal}
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
                    {allKpis}
                    bind:isAnyInlineEditing
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

    <!-- Nueva categoría inline form (editor only, not in avance or cierre mode) -->
    {#if mode === "editor" && phase !== "medio-anio" && phase !== "fin-anio"}
        <div class="pt-2">
            {#if creatingCategory}
                <div class="w-full border border-base-300 rounded-lg p-4 bg-base-200/50">
                    <form onsubmit={(e) => { e.preventDefault(); const err = validateCategory({ name: newCatName, description: newCatDesc, weight: newCatWeight }); if (err) { newCatError = err; return; } handleSaveCategory({ name: newCatName, description: newCatDesc, weight: newCatWeight }); }}>
                        {#if newCatError}<div class="alert alert-error text-sm mb-3" role="alert"><span>{newCatError}</span></div>{/if}
                        <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
                            <div class="form-control">
                                <label class="label" for="new-cat-name"><span class="label-text text-xs">Nombre</span></label>
                                <input id="new-cat-name" type="text" class="input input-bordered input-sm w-full" bind:value={newCatName} placeholder="Nombre de la categoría" required />
                            </div>
                            <div class="form-control">
                                <label class="label" for="new-cat-desc"><span class="label-text text-xs">Descripción</span></label>
                                <textarea id="new-cat-desc" class="textarea textarea-bordered textarea-sm w-full" rows={1} bind:value={newCatDesc} placeholder="Descripción" required></textarea>
                            </div>
                            <div class="form-control">
                                <label class="label" for="new-cat-weight"><span class="label-text text-xs">Peso (%)</span></label>
                                <input id="new-cat-weight" type="number" class="input input-bordered input-sm w-full" bind:value={newCatWeight} min={0} max={100} step={0.1} placeholder="0" required />
                            </div>
                        </div>
                        <div class="flex justify-end gap-2 mt-3">
                            <button type="button" class="btn btn-ghost btn-sm" onclick={() => { creatingCategory = false; isAnyInlineEditing = false; }}>
                                Cancelar
                            </button>
                            <button type="submit" class="btn btn-primary btn-sm">
                                <Check class="w-4 h-4" /> Guardar categoría
                            </button>
                        </div>
                    </form>
                </div>
            {:else}
                <div class="flex justify-center">
                    <button class="btn btn-outline btn-primary" disabled={isAnyInlineEditing} onclick={() => { creatingCategory = true; isAnyInlineEditing = true; newCatName = ''; newCatDesc = ''; newCatWeight = 0; newCatError = ''; }}>
                        <Plus class="w-4 h-4" /> Nueva categoría
                    </button>
                </div>
            {/if}
        </div>
    {/if}
</div>

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
