<script lang="ts">
    import { goto } from "$app/navigation";
    import { Save, Plus, Library, MessageSquare, Check, FileDown, Target } from "@lucide/svelte";
    import type {
        Goal,
        GoalCategory,
        GoalUnit,
        GoalComment,
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
		getWeightedScore,
		getCategoryProgressAverage,
		storeState,
		load,
	} from "$lib/stores/goalsStore.svelte";
    import { getSession } from "$lib/api/session.svelte";
    import { client } from "$lib/api/client";
    import { load as loadOrgHierarchy, getDescendants, getRoot } from "$lib/stores/orgHierarchyStore.svelte";
    import WeightIndicator from "$lib/components/goals/WeightIndicator.svelte";
    import ProgressIndicator from "$lib/components/goals/ProgressIndicator.svelte";
	import { validateCategory } from "$lib/components/goals/goalValidation";
    import CategoryCard from "$lib/components/goals/CategoryCard.svelte";
    import ReadOnlyBanner from "$lib/components/goals/ReadOnlyBanner.svelte";
    import AssigneePicker from "$lib/components/goals/AssigneePicker.svelte";
    import RequestChangeModal from "$lib/components/goals/RequestChangeModal.svelte";
    import CommentPopover from "$lib/components/goals/GoalCommentModal.svelte";
    import PageSkeleton from "$lib/components/ui/PageSkeleton.svelte";
    import ErrorState from "$lib/components/ui/ErrorState.svelte";
    import EmptyState from "$lib/components/ui/EmptyState.svelte";
    import { toCsv } from "$lib/utils/export";
    import * as notifications from "$lib/stores/notifications.svelte";

    // ─── Load data ────────────────────────────────────────────────────────────

    $effect(() => { load(); });
    $effect(() => { loadOrgHierarchy(); });

    // ─── Mode detection ──────────────────────────────────────────────────────
    // ponytail: derive own assignment by employeeId (not profileId) so the
    // logged-in colaborador always gets their own assignment, not someone
    // else's with the same role. Boss detection uses org-node headEmployeeId.

    const session = $derived(getSession());
    const viewerProfile = $derived(session.user?.profileId ?? 'colaborador');
    const viewerEmployeeId = $derived(session.user?.employeeId ?? '');

    const allAssignments = $derived(getAssignments());

    // Own assignment = assignment where employeeId matches logged-in user
    const ownAssignment = $derived(
        allAssignments.find((a) => a.employeeId === viewerEmployeeId),
    );

    // Boss detection: user is a boss if they are head_employee_id of any org node.
    // A boss can see subordinates' assignments in reader mode.
    const isBoss = $derived(isUserHeadOfAnyNode(viewerEmployeeId));

    // ─── Team members (boss only) ──────────────────────────────────────────
    // ponytail: call /team endpoint to get all employees in the boss's org
    // nodes, then merge with existing assignments from goalsStore. Employees
    // without an assignment get a stub entry so the picker always has data.

    type TeamMember = { id: string; firstName: string; lastName: string; orgNodeId: string };
    let teamMembers = $state<TeamMember[]>([]);

    $effect(() => {
        if (!viewerEmployeeId || !isBoss) return;
        client.GET('/employees/{empId}/team', {
            params: { path: { empId: viewerEmployeeId } }
        }).then((res) => {
            const data = (res.data as { data?: Array<TeamMember> })?.data;
            if (data) teamMembers = data;
        });
    });

    // Build available assignments: own + all team members (with or without assignment)
    const availableAssignments = $derived.by(() => {
        const result: EmployeeAssignment[] = [];
        const seen = new Set<string>();

        // ponytail: always place the logged-in user first, regardless of
        // whether ownAssignment has loaded yet. If it has, use it;
        // otherwise create a stub so the picker always shows "(yo)" at top.
        const own = ownAssignment ?? {
            id: `stub-${viewerEmployeeId}`,
            employeeId: viewerEmployeeId,
            employeeName: session.user?.name ?? '',
            profileId: viewerProfile,
            managerId: null,
            goalIds: [],
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
        };
        result.push(own);
        seen.add(viewerEmployeeId);

        // Team members from /team endpoint (skip viewer, already first)
        for (const member of teamMembers) {
            if (seen.has(member.id)) continue;
            seen.add(member.id);
            const existing = getAssignmentByEmployee(member.id);
            if (existing) {
                result.push(existing);
            } else {
                result.push({
                    id: `stub-${member.id}`,
                    employeeId: member.id,
                    employeeName: `${member.firstName} ${member.lastName}`,
                    profileId: 'colaborador',
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

    // Reset selected employee when assignment context changes
    $effect(() => {
        if (!selectedEmployeeId) {
            selectedEmployeeId = viewerEmployeeId;
        }
    });

    const targetAssignment = $derived(
        availableAssignments.find((a) => a.employeeId === selectedEmployeeId),
    );

    const targetEmployeeName = $derived(targetAssignment?.employeeName ?? "");

    // Mode: editor when viewing own assignment or no assignment yet,
    // reader only when viewing a subordinate's assignment.
    const mode = $derived<"editor" | "reader">(
        !targetAssignment || targetAssignment.employeeId === viewerEmployeeId ? "editor" : "reader",
    );

    // ─── Org hierarchy helpers ───────────────────────────────────────────────

    function isUserHeadOfAnyNode(employeeId: string): boolean {
        const root = getRoot();
        if (!root || !employeeId) return false;
        return searchHeadEmployee(root, employeeId);
    }

    function searchHeadEmployee(node: import("$lib/types/org-hierarchy").OrgNode, employeeId: string): boolean {
        if (node.headEmployeeId === employeeId) return true;
        for (const child of node.children ?? []) {
            if (searchHeadEmployee(child, employeeId)) return true;
        }
        return false;
    }

    function getSubordinateEmployeeIds(bossEmployeeId: string): string[] {
        const root = getRoot();
        if (!root) return [];
        const ids = new Set<string>();
        collectBossNodeEmployeeIds(root, bossEmployeeId, ids);
        return [...ids];
    }

    function collectBossNodeEmployeeIds(
        node: import("$lib/types/org-hierarchy").OrgNode,
        bossEmployeeId: string,
        out: Set<string>,
    ): void {
        if (node.headEmployeeId === bossEmployeeId) {
            // Collect all descendant org nodes' head employees
            for (const desc of getDescendants(node.id)) {
                if (desc.headEmployeeId) out.add(desc.headEmployeeId);
            }
        }
        for (const child of node.children ?? []) {
            collectBossNodeEmployeeIds(child, bossEmployeeId, out);
        }
    }

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

    function startCreateCategory() {
        creatingCategory = true;
        isAnyInlineEditing = true;
        newCatName = '';
        newCatDesc = '';
        newCatWeight = 0;
        newCatError = '';
    }

    async function handleSaveCategory(data: { id?: string; name: string; description: string; weight: number }) {
        try {
            if (data.id) {
                await updateCategory(data.id, { name: data.name, description: data.description, weight: data.weight });
            } else {
                const newCat: GoalCategory = { id: `cat-${Date.now()}`, name: data.name, description: data.description, weight: data.weight };
                await addCategory(newCat);
            }
            creatingCategory = false;
            isAnyInlineEditing = false;
            notifications.success(data.id ? 'Categoría actualizada correctamente.' : 'Categoría creada correctamente.');
        } catch (e) {
            newCatError = e instanceof Error ? e.message : 'Error al guardar categoría';
        }
    }

    async function handleDeleteCategory(catId: string) {
        try {
            await deleteCategory(catId);
        } catch (e) {
            // silent
        }
    }

    async function handleDeleteGoal(goalId: string) {
        try {
            await deleteGoal(goalId);
        } catch (e) {
            // silent
        }
    }

    async function handleSaveGoal(data: { id?: string; categoryId: string; name: string; description: string; unit: GoalUnit; weight: number; targetValue: number; direction: 'ascendente' | 'descendente'; baselineValue?: number; linkedKpiIds: string[] }) {
        try {
            if (data.id) {
                await updateGoal(data.id, { name: data.name, description: data.description, unit: data.unit, weight: data.weight, targetValue: data.targetValue, direction: data.direction, baselineValue: data.baselineValue });
                const currentLinked = getKpisForGoal(data.id).map(k => k.id);
                const toAdd = data.linkedKpiIds.filter(id => !currentLinked.includes(id));
                const toRemove = currentLinked.filter(id => !data.linkedKpiIds.includes(id));
                for (const kpiId of toAdd) await linkKpiToGoal(data.id, kpiId);
                for (const kpiId of toRemove) await unlinkKpiFromGoal(data.id, kpiId);
            } else {
                const newGoal: Goal = { id: `goal-${Date.now()}`, name: data.name, description: data.description, categoryId: data.categoryId, weight: data.weight, unit: data.unit, targetValue: data.targetValue, direction: data.direction, baselineValue: data.baselineValue };
                await addGoal(newGoal);
                for (const kpiId of data.linkedKpiIds) await linkKpiToGoal(newGoal.id, kpiId);
            }
            isAnyInlineEditing = false;
            notifications.success(data.id ? 'Meta actualizada correctamente.' : 'Meta creada correctamente.');
        } catch (e) {
            throw e;
        }
    }

    async function handleSaveAssignment() {
        if (!targetAssignment) return;
        try {
            if (targetAssignment.id.startsWith('stub-')) {
                await addAssignment({
                    id: '',
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
            notifications.error(e instanceof Error ? e.message : 'Error al guardar asignación');
        }
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

    async function handleUpdateProgress(goalId: string, progress: number) {
        try {
            await updateGoalProgress(goalId, progress);
        } catch (e) {
            // Progress update failed — the UI will revert on next reload
        }
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
                <h1 class="text-2xl font-bold text-base-content flex items-center gap-2">
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
                    class="btn btn-primary btn-sm"
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

        <!-- Weighted score display (only when there's progress data) -->
        {#if phase === "medio-anio" || phase === "fin-anio"}
            {@const score = getWeightedScore()}
            <div class="mt-3 pt-3 border-t border-base-300">
                <div class="flex items-center justify-between">
                    <span class="text-sm font-semibold text-base-content">Puntaje ponderado</span>
                    <span class="text-2xl font-bold font-mono text-base-content">
                        {Math.round(score)}
                        <span class="text-base font-normal text-base-content/50">/100</span>
                    </span>
                </div>
                <details class="mt-2">
                    <summary class="text-xs text-base-content/50 cursor-pointer hover:text-base-content/80 select-none">
                        Desglose por categoría
                    </summary>
                    <div class="mt-2 space-y-1.5">
                        {#each categories as cat (cat.id)}
                            {@const catProgress = getCategoryProgressAverage(cat.id)}
                            <div class="flex items-center justify-between text-xs">
                                <span class="text-base-content/70">{cat.name} ({cat.weight}%)</span>
                                <span class="font-mono text-base-content">{Math.round(catProgress)}%</span>
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
    {:else if (!creatingCategory && mode === "editor" && phase !== "medio-anio" && phase !== "fin-anio")}
        <EmptyState
            title="Sin categorías"
            message="No hay categorías registradas. Cree la primera categoría para comenzar."
            actionLabel="Nueva categoría"
            onaction={startCreateCategory}
        />
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
            {:else if (categories.length > 0)}
                <div class="flex justify-center">
                    <button class="btn btn-outline btn-primary" disabled={isAnyInlineEditing} onclick={startCreateCategory}>
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
