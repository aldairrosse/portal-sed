<script lang="ts">
    import {
        Pencil,
        Trash2,
        MessageCircle,
        TrendingDown,
        TrendingUp,
        Eye,
    } from "@lucide/svelte";
    import type { Goal, GoalUnit, KPI, CyclePhase } from "$lib/types/goal";
    import { deltaIndicator, formatDelta } from "$lib/utils/scoring";
    import KpiBadge from "./KpiBadge.svelte";
    import ProgressIndicator from "./ProgressIndicator.svelte";
    import GoalForm from "./GoalForm.svelte";
    import ProposalReviewModal from "./ProposalReviewModal.svelte";

    interface Props {
        goal: Goal;
        kpis: KPI[];
        onSaveGoal: (data: {
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
        }) => void;
        onDeleteGoal: (goalId: string) => void;
        mode?: "editor" | "reader";
        onSaveProposal?: (goalId: string, data: {
            name: string;
            description: string;
            unit: GoalUnit;
            weight: number;
            targetValue: number;
            direction: "ascendente" | "descendente";
            baselineValue?: number;
            kpiIds: string[];
        }) => void | Promise<void>;
        onAcceptProposal?: (goalId: string, proposalId: string) => void | Promise<void>;
        onRejectProposal?: (goalId: string, proposalId: string) => void | Promise<void>;
        phase?: CyclePhase;
        canEditProgress?: boolean;
        canComment?: boolean;
        canDelete?: boolean;
        canClose?: boolean;
        allKpis: KPI[];
        editingGoalId: string | null;
        onEditingChange: (id: string | null) => void;
        onUpdateProgress?: (goalId: string, progress: number) => void;
        onOpenComments?: (goal: Goal) => void;
    }

    let {
        goal,
        kpis,
        onSaveGoal,
        onDeleteGoal,
        mode = "editor",
        onSaveProposal,
        onAcceptProposal,
        onRejectProposal,
        phase = "inicio-anio",
        canEditProgress = false,
        canComment = false,
        canDelete = true,
        canClose = false,
        allKpis,
        editingGoalId,
        onEditingChange,
        onUpdateProgress,
        onOpenComments,
    }: Props = $props();

    // svelte-ignore state_referenced_locally (intentional: form state seeded once from prop)
    // eslint-disable-next-line svelte/prefer-writable-derived
    let progressValue = $state(goal.progress ?? 0);

    $effect(() => {
        progressValue = goal.progress ?? 0;
    });

    const unitLabels: Record<string, string> = {
        porcentaje: "%",
        moneda: "$",
        numero: "#",
        binario: "",
    };

    // ─── Inline edit state ────────────────────────────────────────────────

    let isEditing = $derived(editingGoalId === goal.id);
    let isProposing = $state(false);
    let showReviewProposal = $state(false);

    function handleProgressInput(e: Event) {
        const val = parseFloat((e.target as HTMLInputElement).value);
        if (!isNaN(val)) {
            progressValue = val;
            onUpdateProgress?.(goal.id, val);
        }
    }

    // ─── Delta indicator ───────────────────────────────────────────────────

    let delta = $derived.by(() => {
        if (goal.progress === undefined || goal.progress === null) return null;
        return deltaIndicator(
            goal.progress,
            goal.targetValue,
            goal.baselineValue,
            goal.direction,
        );
    });

    let deltaBadgeClass = $derived.by(() => {
        if (!delta || delta.type === "neutral") return "";
        if (delta.type === "positive") return "badge-success";
        return goal.direction === "descendente"
            ? "badge-error"
            : "badge-warning";
    });

    let deltaLabel = $derived.by(() => {
        if (!delta || delta.type === "neutral") return "";
        const formatted = formatDelta(delta.value, delta.type);
        if (delta.type === "positive" && goal.direction === "descendente")
            return formatted + "%";
        return formatted;
    });
</script>

<tr>
    {#if isEditing}
        <td colspan={phase !== "fin-anio" ? 5 : 4} class="p-0">
            <div class="m-2">
                <GoalForm
                    mode="edit"
                    categoryId={goal.categoryId}
                    goalId={goal.id}
                    initialName={goal.name}
                    initialDescription={goal.description}
                    initialUnit={goal.unit}
                    initialWeight={goal.weight}
                    initialTarget={goal.targetValue}
                    initialDirection={goal.direction}
                    initialBaseline={goal.baselineValue}
                    initialKpiIds={kpis.map((k) => k.id)}
                    {allKpis}
                    submitLabel="Guardar meta"
                    onsave={async (data) => {
                        await onSaveGoal({
                            id: goal.id,
                            categoryId: goal.categoryId,
                            name: data.name,
                            description: data.description,
                            unit: data.unit,
                            weight: data.weight,
                            targetValue: data.targetValue,
                            direction: data.direction,
                            baselineValue: data.baselineValue,
                            linkedKpiIds: data.kpiIds,
                        });
                        onEditingChange(null);
                    }}
                    oncancel={() => onEditingChange(null)}
                />
            </div>
        </td>
    {:else if isProposing}
        <td colspan={phase !== "fin-anio" ? 5 : 4} class="p-0">
            <div class="m-2">
                <GoalForm
                    mode="proposal"
                    categoryId={goal.categoryId}
                    goalId={goal.id}
                    initialName={goal.name}
                    initialDescription={goal.description}
                    initialUnit={goal.unit}
                    initialWeight={goal.weight}
                    initialTarget={goal.targetValue}
                    initialDirection={goal.direction}
                    initialBaseline={goal.baselineValue}
                    initialKpiIds={kpis.map((k) => k.id)}
                    {allKpis}
                    submitLabel="Proponer cambio"
                    onsave={async (data) => {
                        await onSaveProposal?.(goal.id, data);
                        isProposing = false;
                    }}
                    oncancel={() => { isProposing = false; }}
                />
            </div>
        </td>
    {:else}
        <td class="font-medium text-sm">{goal.name}</td>
        <td class="text-sm text-base-content/70 flex items-center gap-1">
            {#if goal.direction === "ascendente"}
                <TrendingUp class="w-3 h-3" />
            {:else}
                <TrendingDown class="w-3 h-3" />
            {/if}
            {#if goal.unit === "binario"}
                {goal.targetValue ? "Sí" : "No"}
            {:else}
            {goal.targetValue}{unitLabels[goal.unit] ?? goal.unit}
            {/if}
        </td>
        <td class="text-sm text-base-content/70">
            {goal.weight}%</td
        >
        <td>
            {#if phase === "fin-anio"}
                <div class="flex items-center gap-2">
                    {#if canClose}
                        <input
                            type="number"
                            class="input input-bordered input-xs w-20"
                            value={progressValue}
                            min="0"
                            max={goal.unit === "porcentaje"
                                ? 100
                                : goal.targetValue}
                            oninput={handleProgressInput}
                            aria-label="Avance final de {goal.name}"
                        />
                    {/if}
                    <ProgressIndicator
                        value={progressValue}
                        max={100}
                        label="Avance final"
                    />
                    {#if delta && delta.type !== "neutral"}
                        <span class="badge badge-sm {deltaBadgeClass}"
                            >{deltaLabel}</span
                        >
                    {/if}
                </div>
            {:else if phase === "medio-anio"}
                <div class="flex items-center gap-2">
                    <input
                        type="number"
                        class="input input-bordered input-xs w-20"
                        value={progressValue}
                        min="0"
                        max={goal.unit === "porcentaje" ? 100 : undefined}
                        oninput={handleProgressInput}
                        aria-label="Avance de {goal.name}"
                        readonly={!canEditProgress}
                    />
                    {#if goal.unit === "porcentaje"}
                        <ProgressIndicator
                            value={progressValue}
                            max={100}
                            color={progressValue >= goal.targetValue
                                ? "success"
                                : "error"}
                        />
                    {:else}
                        <ProgressIndicator
                            value={progressValue}
                            max={goal.targetValue}
                        />
                    {/if}
                    {#if delta && delta.type !== "neutral"}
                        <span class="badge badge-sm {deltaBadgeClass}"
                            >{deltaLabel}</span
                        >
                    {/if}
                </div>
            {:else if kpis.length > 0}
                <div class="flex flex-wrap gap-1">
                    {#each kpis as kpi (kpi.id)}
                        <KpiBadge {kpi} />
                    {/each}
                </div>
            {:else}
                <span class="text-xs text-base-content/30 italic">Sin KPI</span>
            {/if}
        </td>
        {#if phase !== "fin-anio"}
            <td class="text-right">
                <div class="flex items-center justify-end gap-1">
                    {#if phase === "medio-anio"}
                        {#if canComment}
                            <button
                                class="btn btn-ghost btn-square btn-xs relative"
                                title="Comentarios"
                                onclick={() => onOpenComments?.(goal)}
                                aria-label="Comentarios de {goal.name}"
                            >
                                <MessageCircle class="w-3.5 h-3.5" />
                                {#if (goal.comments?.length ?? 0) > 0}
                                    <span
                                        class="badge badge-xs badge-primary absolute -top-1.5 -right-1.5"
                                        >{goal.comments?.length}</span
                                    >
                                {/if}
                            </button>
                        {/if}
                    {:else if mode === "editor"}
                        <!-- ponytail: show comment button in inicio-anio when there are existing comments from the boss -->
                        {#if (goal.comments?.length ?? 0) > 0}
                            <button
                                class="btn btn-ghost btn-square btn-xs relative"
                                title="Comentarios"
                                onclick={() => onOpenComments?.(goal)}
                                aria-label="Comentarios de {goal.name}"
                            >
                                <MessageCircle class="w-3.5 h-3.5" />
                                <span
                                    class="badge badge-xs badge-primary absolute -top-1.5 -right-1.5"
                                    >{goal.comments?.length}</span
                                >
                            </button>
                        {/if}
                        {#if goal.pendingProposal}
                            <button
                                class="btn btn-ghost btn-xs btn-square"
                                title="Ver propuesta"
                                onclick={() => { showReviewProposal = true; }}
                                aria-label="Ver propuesta de {goal.name}"
                            >
                                <Eye class="w-3.5 h-3.5" />
                            </button>
                        {:else}
                            <button
                                class="btn btn-ghost btn-square btn-xs"
                                title="Editar"
                                onclick={() => onEditingChange(goal.id)}
                                disabled={editingGoalId !== null && editingGoalId !== goal.id}
                                aria-label="Editar {goal.name}"
                            >
                                <Pencil class="w-3.5 h-3.5" />
                            </button>
                        {/if}
                        {#if canDelete}
                            <button
                                class="btn btn-ghost btn-square btn-xs text-error"
                                title="Eliminar"
                                onclick={() => onDeleteGoal(goal.id)}
                                disabled={editingGoalId !== null}
                                aria-label="Eliminar {goal.name}"
                            >
                                <Trash2 class="w-3.5 h-3.5" />
                            </button>
                        {/if}
                    {:else}
                        <button
                            class="btn btn-ghost btn-xs text-warning"
                            title="Proponer cambio"
                            onclick={() => { isProposing = true; }}
                            aria-label="Proponer cambio en {goal.name}"
                        >
                            <Pencil class="w-3.5 h-3.5" />
                            Proponer
                        </button>
                        <button
                            class="btn btn-ghost btn-square btn-xs relative"
                            title="Comentarios"
                            onclick={() => onOpenComments?.(goal)}
                            aria-label="Comentarios de {goal.name}"
                        >
                            <MessageCircle class="w-3.5 h-3.5" />
                            {#if (goal.comments?.length ?? 0) > 0}
                                <span
                                    class="badge badge-xs badge-primary absolute -top-1.5 -right-1.5"
                                    >{goal.comments?.length}</span
                                >
                            {/if}
                        </button>
                    {/if}
                </div>
            </td>
        {/if}
    {/if}
</tr>

{#if showReviewProposal && goal.pendingProposal}
    <ProposalReviewModal
        open={showReviewProposal}
        proposal={goal.pendingProposal}
        {allKpis}
        onclose={() => { showReviewProposal = false; }}
        onaccept={async () => {
            await onAcceptProposal?.(goal.id, goal.pendingProposal!.id);
            showReviewProposal = false;
        }}
        onreject={async () => {
            await onRejectProposal?.(goal.id, goal.pendingProposal!.id);
            showReviewProposal = false;
        }}
    />
{/if}
