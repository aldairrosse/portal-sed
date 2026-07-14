<script lang="ts">
    import { goto } from "$app/navigation";
    import {
        Save,
        Plus,
        Library,
        MessageSquare,
        FileDown,
        Target,
        Trash,
        MessageCircle,
    } from "@lucide/svelte";
    import type {
        Goal,
        GoalCategory,
        GoalUnit,
        EmployeeAssignment,
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
        getAssignments,
        getAssignmentByEmployee,
        getCyclePhase,
        getGoalPermissions,
        updateGoalProgress,
        addAssignment,
        addGoalComment,
        deleteGoalComment,
        getGoalComments,
        addCategoryComment,
        deleteCategoryComment,
        getCategoryComments,
        loadAllGoalComments,
        getWeightedScore,
        getCategoryProgressAverage,
        storeState,
        load,
        reload,
        loadForEmployee,
        createGoalProposal,
        acceptGoalProposal,
        rejectGoalProposal,
        getAssignmentComments,
        loadAssignmentComments,
        addAssignmentComment,
        deleteAssignmentComment,
    } from "$lib/stores/goalsStore.svelte";
    import { getSession } from "$lib/api/session.svelte";
    import {
        load as loadOrgHierarchy,
        getRoot,
    } from "$lib/stores/orgHierarchyStore.svelte";
    import { loadTeam, getTeamMembers } from "$lib/stores/teamStore.svelte";
    import { load as loadPillars, getPillars } from "$lib/stores/competencyStore.svelte";
    import WeightIndicator from "$lib/components/goals/WeightIndicator.svelte";
    import ProgressIndicator from "$lib/components/goals/ProgressIndicator.svelte";
    import CategoryCard from "$lib/components/goals/CategoryCard.svelte";
    import ReadOnlyBanner from "$lib/components/goals/ReadOnlyBanner.svelte";
    import AssigneePicker from "$lib/components/goals/AssigneePicker.svelte";
    import RequestChangeModal from "$lib/components/goals/RequestChangeModal.svelte";
    import CommentPopover from "$lib/components/goals/GoalCommentModal.svelte";
    import PageSkeleton from "$lib/components/ui/PageSkeleton.svelte";
    import ErrorState from "$lib/components/ui/ErrorState.svelte";
    import EmptyState from "$lib/components/ui/EmptyState.svelte";
    import CategoryCreateForm from "$lib/components/goals/CategoryCreateForm.svelte";
    import ConfirmDialog from "$lib/components/ui/ConfirmDialog.svelte";
    import ExportCsvModal from "$lib/components/goals/ExportCsvModal.svelte";
    import { toCsv } from "$lib/utils/export";
    import * as notifications from "$lib/stores/notifications.svelte";
    import { SvelteSet } from "svelte/reactivity";

    // ─── Load data ────────────────────────────────────────────────────────────

    $effect(() => {
        load();
    });
    $effect(() => {
        loadOrgHierarchy();
    });
    $effect(() => {
        loadPillars('metas');
    });

    const pillarOptions = $derived(
        getPillars().map(p => ({ value: p.id, label: p.name }))
    );

    // ─── Mode detection ──────────────────────────────────────────────────────

    const session = $derived(getSession());
    const viewerProfile = $derived(session.user?.profileId ?? "colaborador");
    const viewerEmployeeId = $derived(session.user?.employeeId ?? "");

    const allAssignments = $derived(getAssignments());

    const ownAssignment = $derived(
        allAssignments.find((a) => a.employeeId === viewerEmployeeId),
    );

    const isBoss = $derived(isUserHeadOfAnyNode(viewerEmployeeId));

    // ─── Team members (boss only) ──────────────────────────────────────────

    $effect(() => {
        if (!viewerEmployeeId || !isBoss) return;
        loadTeam(viewerEmployeeId);
    });

    const teamMembers = $derived(getTeamMembers());

    const availableAssignments = $derived.by(() => {
        const result: EmployeeAssignment[] = [];
        const seen = new SvelteSet<string>();

        const own = ownAssignment ?? {
            id: `stub-${viewerEmployeeId}`,
            employeeId: viewerEmployeeId,
            employeeName: session.user?.name ?? "",
            profileId: viewerProfile,
            managerId: null,
            goalIds: [],
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
        };
        result.push({
            ...own,
            employeeName: own.employeeName || session.user?.name || "",
        });
        seen.add(viewerEmployeeId);

        for (const member of teamMembers) {
            if (seen.has(member.id)) continue;
            seen.add(member.id);
            const memberName = `${member.firstName} ${member.lastName}`;
            const existing = getAssignmentByEmployee(member.id);
            if (existing) {
                result.push({
                    ...existing,
                    employeeName: existing.employeeName || memberName,
                    employeeNumber: member.employeeNumber,
                });
            } else {
                result.push({
                    id: `stub-${member.id}`,
                    employeeId: member.id,
                    employeeName: memberName,
                    employeeNumber: member.employeeNumber,
                    profileId: "colaborador",
                    managerId: null,
                    goalIds: [],
                    createdAt: new Date().toISOString(),
                    updatedAt: new Date().toISOString(),
                });
            }
        }

        return result;
    });

    const showAssigneePicker = $derived(isBoss);

    let selectedEmployeeId = $state("");

    $effect(() => {
        if (!selectedEmployeeId) {
            selectedEmployeeId = viewerEmployeeId;
        }
    });

    const targetAssignment = $derived(
        availableAssignments.find((a) => a.employeeId === selectedEmployeeId),
    );

    const targetEmployeeName = $derived(targetAssignment?.employeeName ?? "");

    const mode = $derived<"editor" | "reader">(
        !targetAssignment || targetAssignment.employeeId === viewerEmployeeId
            ? "editor"
            : "reader",
    );

    // ─── Org hierarchy helpers ───────────────────────────────────────────────

    function isUserHeadOfAnyNode(employeeId: string): boolean {
        const root = getRoot();
        if (!root || !employeeId) return false;
        return searchHeadEmployee(root, employeeId);
    }

    function searchHeadEmployee(
        node: import("$lib/types/org-hierarchy").OrgNode,
        employeeId: string,
    ): boolean {
        if (node.headEmployeeId === employeeId) return true;
        for (const child of node.children ?? []) {
            if (searchHeadEmployee(child, employeeId)) return true;
        }
        return false;
    }

    // ─── Cycle phase & permissions ──────────────────────────────────────────

    const phase = $derived(getCyclePhase());
    const isOwner = $derived(mode === "editor");
    const permissions = $derived(getGoalPermissions(viewerProfile, isOwner));

    // ─── Comment modal state ────────────────────────────────────────────────

    let commentGoal: Goal | null = $state(null);
    let showCommentModal = $state(false);

    function openComments(goal: Goal) {
        commentGoal = goal;
        showCommentModal = true;
    }

    function handleAddComment(goalId: string, content: string) {
        addGoalComment(goalId, viewerEmployeeId, session.user?.name ?? "", content);
    }

    function handleDeleteComment(goalId: string, commentId: string) {
        deleteGoalComment(goalId, commentId);
    }

    // ─── Category comment modal state ──────────────────────────────────────

    let commentCategory: { id: string; name: string } | null = $state(null);
    let showCategoryCommentModal = $state(false);

    function openCategoryComments(category: GoalCategory) {
        commentCategory = category;
        showCategoryCommentModal = true;
    }

    function handleAddCategoryComment(catId: string, content: string) {
        addCategoryComment(catId, viewerEmployeeId, session.user?.name ?? "", content);
    }

    function handleDeleteCategoryComment(catId: string, commentId: string) {
        deleteCategoryComment(catId, commentId);
    }

    // ─── Assignment comment modal state ────────────────────────────────────

    let showAssignmentCommentModal = $state(false);
    let assignmentCommentTarget: EmployeeAssignment | null = $state(null);

    $effect(() => {
        const t = targetAssignment;
        if (t?.id && !t.id.startsWith('stub-')) {
            loadAssignmentComments(t.id);
        }
    });

    function openAssignmentComments() {
        if (!targetAssignment) return;
        assignmentCommentTarget = targetAssignment;
        loadAssignmentComments(targetAssignment.id);
        showAssignmentCommentModal = true;
    }

    function handleAddAssignmentComment(assignmentId: string, content: string) {
        addAssignmentComment(assignmentId, viewerEmployeeId, session.user?.name ?? "", content);
    }

    function handleDeleteAssignmentComment(assignmentId: string, commentId: string) {
        deleteAssignmentComment(assignmentId, commentId);
    }

    // ─── Existing page state ─────────────────────────────────────────────────

    let creatingCategory = $state(false);
    let isAnyInlineEditing = $state(false);

    const categories = $derived(getCategories());
    const goals = $derived(getGoals());
    const allKpis = $derived(getKpis());
    const globalSum = $derived(
        categories.reduce((sum, c) => sum + c.weight, 0),
    );
    const valid = $derived(isAssignmentValid());

    // ─── Request change modal state ─────────────────────────────────────────

    let showRequestModal = $state(false);
    let showDeleteAllConfirm = $state(false);
    let deletingAll = $state(false);
    let requestEntityType: ChangeRequest["entityType"] = $state("goal");
    let requestEntityId = $state("");
    let requestEntityName = $state("");

    const requestComments = $derived.by(() => {
        if (requestEntityType === "goal") return getGoalComments(requestEntityId);
        if (requestEntityType === "category") return getCategoryComments(requestEntityId);
        return [];
    });

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
        if (employeeId === viewerEmployeeId) {
            load();
        } else {
            loadForEmployee(employeeId);
        }
    }

    function handleExportCsv() {
        if(!isBoss) {
            exportCurrent();
            return;
        }
        showExportModal = true;
    }

    function startCreateCategory() {
        creatingCategory = true;
        isAnyInlineEditing = true;
    }

    async function handleSaveCategory(data: {
        id?: string;
        name: string;
        description: string;
        weight: number;
        pillarId?: string;
    }) {
        try {
            if (data.id) {
                await updateCategory(data.id, {
                    name: data.name,
                    description: data.description,
                    weight: data.weight,
                    pillarId: data.pillarId,
                });
            } else {
                const newCat: GoalCategory = {
                    id: `cat-${Date.now()}`,
                    name: data.name,
                    description: data.description,
                    weight: data.weight,
                    pillarId: data.pillarId,
                };
                await addCategory(newCat);
            }
            creatingCategory = false;
            isAnyInlineEditing = false;
            notifications.success(
                data.id
                    ? "Categoría actualizada correctamente."
                    : "Categoría creada correctamente.",
            );
        } catch (e) {
            notifications.error(
                e instanceof Error ? e.message : "Error al guardar categoría",
            );
        }
    }

    async function handleDeleteCategory(catId: string) {
        try {
            await deleteCategory(catId);
        } catch (e) {
            console.error("Error deleting category:", e);
        }
    }

    async function handleDeleteGoal(goalId: string) {
        try {
            await deleteGoal(goalId);
        } catch (e) {
            console.error("Error deleting goal:", e);
        }
    }

    async function handleSaveGoal(data: {
        id?: string;
        categoryId: string;
        name: string;
        description: string;
        unit: GoalUnit;
        weight: number;
        targetValue: number;
        direction: "ascendente" | "descendente";
        baselineValue?: number;
        linkedKpiIds: string[];
    }) {
        try {
            if (data.id) {
                await updateGoal(data.id, {
                    name: data.name,
                    description: data.description,
                    unit: data.unit,
                    weight: data.weight,
                    targetValue: data.targetValue,
                    direction: data.direction,
                    baselineValue: data.baselineValue,
                    version: goals.find((g) => g.id === data.id)?.version ?? 0,
                });
                const currentLinked = getKpisForGoal(data.id).map((k) => k.id);
                const toAdd = data.linkedKpiIds.filter(
                    (id) => !currentLinked.includes(id),
                );
                const toRemove = currentLinked.filter(
                    (id) => !data.linkedKpiIds.includes(id),
                );
                for (const kpiId of toAdd) await linkKpiToGoal(data.id, kpiId);
                for (const kpiId of toRemove)
                    await unlinkKpiFromGoal(data.id, kpiId);
            } else {
                const newGoal: Goal = {
                    id: `goal-${Date.now()}`,
                    name: data.name,
                    description: data.description,
                    categoryId: data.categoryId,
                    weight: data.weight,
                    unit: data.unit,
                    targetValue: data.targetValue,
                    direction: data.direction,
                    baselineValue: data.baselineValue,
                    version: 1,
                };
                const createdId = await addGoal(newGoal);
                for (const kpiId of data.linkedKpiIds)
                    await linkKpiToGoal(createdId, kpiId);
            }
            isAnyInlineEditing = false;
            notifications.success(
                data.id
                    ? "Meta actualizada correctamente."
                    : "Meta creada correctamente.",
            );
        } catch (e) {
            console.error("Error saving goal:", e);
            notifications.error(
                e instanceof Error ? e.message : "Error al guardar meta",
            );
        }
    }

    async function handleDeleteAssignment() {
        showDeleteAllConfirm = true;
    }

    async function confirmDeleteAll() {
        if (deletingAll) return;
        deletingAll = true;
        try {
            for (const cat of categories) {
                await deleteCategory(cat.id, { skipReload: true });
            }
            await reload();
            notifications.success("Asignación borrada correctamente.");
        } catch (e) {
            notifications.error(
                e instanceof Error ? e.message : "Error al borrar la asignación",
            );
        } finally {
            deletingAll = false;
            showDeleteAllConfirm = false;
        }
    }

    async function handleSaveAssignment() {
        if (!targetAssignment) return;
        try {
            if (targetAssignment.id.startsWith("stub-")) {
                await addAssignment({
                    id: "",
                    employeeId: targetAssignment.employeeId,
                    employeeName: targetAssignment.employeeName,
                    profileId: targetAssignment.profileId,
                    managerId: null,
                    goalIds: [],
                    createdAt: new Date().toISOString(),
                    updatedAt: new Date().toISOString(),
                });
            }
            notifications.success("Asignación guardada correctamente.");
        } catch (e) {
            notifications.error(
                e instanceof Error ? e.message : "Error al guardar asignación",
            );
        }
    }

    function handleRequestChangeCategory(category: GoalCategory) {
        openRequestModal("category", category.id, category.name);
    }

    function handleRequestAddComment(entityId: string, content: string) {
        if (requestEntityType === "goal") {
            addGoalComment(entityId, viewerEmployeeId, session.user?.name ?? "", content);
        } else if (requestEntityType === "category") {
            addCategoryComment(entityId, viewerEmployeeId, session.user?.name ?? "", content);
        }
    }

    async function handleUpdateProgress(goalId: string, progress: number) {
        try {
            await updateGoalProgress(goalId, progress);
        } catch (e) {
            console.error("Error updating progress:", e);
        }
    }

    async function handleSaveProposal(goalId: string, data: Parameters<typeof createGoalProposal>[1]) {
        await createGoalProposal(goalId, data);
    }

    async function handleAcceptProposal(goalId: string, proposalId: string) {
        await acceptGoalProposal(goalId, proposalId, viewerEmployeeId);
    }

    async function handleRejectProposal(goalId: string, proposalId: string) {
        await rejectGoalProposal(goalId, proposalId);
    }

    // ─── Export CSV modal ────────────────────────────────────────────────────────

    let showExportModal = $state(false);

    function buildRows(
        assignment: EmployeeAssignment,
    ): Record<string, string | number | null>[] {
        const rows: Record<string, string | number | null>[] = [];
        const cats = getCategories();
        for (const cat of cats) {
            const catGoals = getGoalsByCategory(cat.id);
            for (const goal of catGoals) {
                const kpis = getKpisForGoal(goal.id);
                rows.push({
                    "No. Empleado": assignment.employeeNumber ?? assignment.employeeId,
                    Empleado: assignment.employeeName,
                    Categoría: cat.name,
                    "Peso categoría %": cat.weight,
                    Meta: goal.name,
                    Descripción: goal.description,
                    Unidad: goal.unit,
                    "Peso meta %": goal.weight,
                    "Valor objetivo": formatTarget(goal),
                    KPIs: kpis.map((k) => k.name).join(", ") || "",
                });
            }
        }
        return rows;
    }

    function formatTarget(g: Goal) {
        return g.unit === "porcentaje" ? `${g.targetValue}%` : g.targetValue;
    }

    async function exportCurrent() {
        showExportModal = false;
        toCsv(buildRows(targetAssignment!), "asignacion-anual.csv");
    }

    async function exportAll() {
        showExportModal = false;
        const allRows: Record<string, string | number | null>[] = [];
        const originalId = selectedEmployeeId;

        for (const a of availableAssignments) {
            await loadForEmployee(a.employeeId);
            allRows.push(...buildRows(a));
        }

        await loadForEmployee(originalId);
        toCsv(allRows, "asignacion-anual-todos.csv");
    }
</script>

<svelte:head>
    <title>Asignación anual — SED</title>
</svelte:head>

{#if storeState.loading}
    <PageSkeleton variant="category" rows={3} />
{:else if storeState.error}
    <ErrorState message={storeState.error} onretry={load} />
{:else}
    <div class="space-y-6 max-w-full min-w-0">
        <!-- Page header -->
        <div>
            <div class="flex items-start justify-between gap-4">
                <div>
                    <h1
                        class="text-2xl font-bold text-base-content flex items-center gap-2"
                    >
                        <Target class="w-6 h-6" />
                        {phase === "medio-anio"
                            ? "Avance de metas"
                            : phase === "fin-anio"
                              ? "Evaluación anual"
                              : "Asignación anual"}
                    </h1>
                    <p class="text-sm text-base-content/50 mt-1">
                        {phase === "medio-anio"
                            ? "Registre el avance de sus metas y agregue comentarios."
                            : "Defina las categorías y metas para el período de evaluación."}
                    </p>
                </div>
                {#if phase === "inicio-anio"}
                    <button
                        class="btn btn-ghost btn-sm"
                        onclick={() => goto("/objetivos/asignacion/biblioteca")}
                        aria-label="Biblioteca de KPI"
                    >
                        <Library class="w-4 h-4" />
                        Biblioteca de KPI
                    </button>
                {/if}
            </div>
            <div class="flex items-center gap-2 mt-3 flex-wrap">
                {#if showAssigneePicker}
                    <AssigneePicker
                        assignments={availableAssignments}
                        {selectedEmployeeId}
                        onSelect={handleAssigneeSelect}
                        currentUserId={viewerEmployeeId}
                    />
                {/if}
                <button
                    class="btn btn-outline btn-sm"
                    disabled={categories.length === 0}
                    onclick={handleExportCsv}
                >
                    <FileDown class="w-4 h-4" />
                    Exportar CSV
                </button>
                {#if mode === "editor" && phase !== "medio-anio" && phase !== "fin-anio"}
                    <div class="flex-1"></div>
                    <button
                        class="btn btn-error btn-sm btn-ghost"
                        disabled={deletingAll}
                        onclick={handleDeleteAssignment}
                    >
                        <Trash class="w-4 h-4" />
                        Borrar todo
                    </button>
                    <button
                        class="btn btn-primary btn-sm"
                        disabled={!valid}
                        onclick={handleSaveAssignment}
                    >
                        <Save class="w-4 h-4" />
                        Enviar asignación
                    </button>
                {:else if mode === "reader"}
                    {@const count = assignmentCommentTarget?.id ? getAssignmentComments(assignmentCommentTarget.id).length : 0}
                    <button
                        class="btn btn-primary btn-sm ml-auto"
                        onclick={openAssignmentComments}
                        aria-label="Comentar en asignación"
                    >
                        <MessageCircle class="w-4 h-4" />
                        Comentar
                        {#if count > 0}
                            <span class="badge badge-sm">{count}</span>
                        {/if}
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
                                      : ((g.progress ?? 0) /
                                            (g.targetValue || 1)) *
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
                        La suma de pesos debe ser 100% tanto a nivel global como
                        en cada categoría.
                    </p>
                {/if}
            {/if}

            <!-- Weighted score display (only when there's progress data) -->
            {#if phase === "medio-anio" || phase === "fin-anio"}
                {@const score = getWeightedScore()}
                <div class="mt-3 pt-3 border-t border-base-300">
                    <div class="flex items-center justify-between">
                        <span class="text-sm font-semibold text-base-content"
                            >Puntaje ponderado</span
                        >
                        <span
                            class="text-2xl font-bold font-mono text-base-content"
                        >
                            {Math.round(score)}
                            <span
                                class="text-base font-normal text-base-content/50"
                                >/100</span
                            >
                        </span>
                    </div>
                    <details class="mt-2">
                        <summary
                            class="text-xs text-base-content/50 cursor-pointer hover:text-base-content/80 select-none"
                        >
                            Desglose por categoría
                        </summary>
                        <div class="mt-2 space-y-1.5">
                            {#each categories as cat (cat.id)}
                                {@const catProgress =
                                    getCategoryProgressAverage(cat.id)}
                                <div
                                    class="flex items-center justify-between text-xs"
                                >
                                    <span class="text-base-content/70"
                                        >{cat.name} ({cat.weight}%)</span
                                    >
                                    <span class="font-mono text-base-content"
                                        >{Math.round(catProgress)}%</span
                                    >
                                </div>
                            {/each}
                        </div>
                    </details>
                </div>
            {/if}
        </div>

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
                        pillars={pillarOptions}
                        onRequestChangeCategory={handleRequestChangeCategory}
                        onSaveProposal={handleSaveProposal}
                        onAcceptProposal={handleAcceptProposal}
                        onRejectProposal={handleRejectProposal}
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
                        onOpenCategoryComments={openCategoryComments}
                    />
                {/each}
            </div>
        {:else if !creatingCategory && mode === "editor" && phase !== "medio-anio" && phase !== "fin-anio"}
            <EmptyState
                title="Sin categorías"
                message="No hay categorías registradas. Cree la primera categoría para comenzar."
                actionLabel="Nueva categoría"
                onaction={startCreateCategory}
            />
        {/if}

        <!-- Nueva categoría inline form -->
        {#if mode === "editor" && phase !== "medio-anio" && phase !== "fin-anio"}
            <div class="pt-2">
                {#if creatingCategory}
                    <CategoryCreateForm
                        mode="create"
                        pillars={pillarOptions}
                        onSave={(data) => handleSaveCategory(data)}
                        onCancel={() => {
                            creatingCategory = false;
                            isAnyInlineEditing = false;
                        }}
                    />
                {:else if categories.length > 0}
                    <div class="flex justify-center">
                        <button
                            class="btn btn-outline btn-primary"
                            disabled={isAnyInlineEditing}
                            onclick={startCreateCategory}
                        >
                            <Plus class="w-4 h-4" /> Nueva categoría
                        </button>
                    </div>
                {/if}
            </div>
        {/if}
    </div>
{/if}

{#if targetAssignment}
    <RequestChangeModal
        open={showRequestModal}
        entityType={requestEntityType}
        entityId={requestEntityId}
        entityName={requestEntityName}
        requestedBy={viewerEmployeeId}
        onClose={closeRequestModal}
        comments={requestComments}
        onAddComment={handleRequestAddComment}
        currentUserId={viewerEmployeeId}
    />
{/if}

{#if showCommentModal && commentGoal}
    <CommentPopover
        open={showCommentModal}
        goal={commentGoal}
        comments={getGoalComments(commentGoal.id)}
        onAdd={handleAddComment}
        onDelete={handleDeleteComment}
        onClose={() => (showCommentModal = false)}
        currentUserId={viewerEmployeeId}
    />
{/if}

{#if showCategoryCommentModal && commentCategory}
    <CommentPopover
        open={showCategoryCommentModal}
        goal={null}
        comments={getCategoryComments(commentCategory.id)}
        onAdd={handleAddCategoryComment}
        onDelete={handleDeleteCategoryComment}
        onClose={() => (showCategoryCommentModal = false)}
        currentUserId={viewerEmployeeId}
        category={commentCategory}
    />
{/if}

{#if showAssignmentCommentModal && assignmentCommentTarget}
    <CommentPopover
        open={showAssignmentCommentModal}
        goal={null}
        comments={getAssignmentComments(assignmentCommentTarget.id)}
        onAdd={handleAddAssignmentComment}
        onDelete={handleDeleteAssignmentComment}
        onClose={() => (showAssignmentCommentModal = false)}
        currentUserId={viewerEmployeeId}
        assignment={{ id: assignmentCommentTarget.id, name: assignmentCommentTarget.employeeName || assignmentCommentTarget.employeeId }}
    />
{/if}

<ExportCsvModal
    open={showExportModal}
    {targetAssignment}
    {availableAssignments}
    {isBoss}
    onExportCurrent={exportCurrent}
    onExportAll={exportAll}
    onClose={() => (showExportModal = false)}
/>

<ConfirmDialog
    open={showDeleteAllConfirm}
    variant="error"
    title="Borrar toda la asignación"
    message="Se eliminarán todas las categorías y sus metas. Esta acción no se puede deshacer."
    confirmLabel="Sí, borrar todo"
    cancelLabel="Cancelar"
    onconfirm={confirmDeleteAll}
    disabled={deletingAll}
    oncancel={() => (showDeleteAllConfirm = false)}
/>
