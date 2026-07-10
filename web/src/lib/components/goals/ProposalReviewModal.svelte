<script lang="ts">
    import GoalForm from './GoalForm.svelte';
    import { X } from '@lucide/svelte';
    import type { GoalProposal, KPI } from '$lib/types/goal';

    interface Props {
        open: boolean;
        proposal: GoalProposal;
        allKpis: KPI[];
        onclose: () => void;
        onaccept: () => void | Promise<void>;
        onreject: () => void | Promise<void>;
    }

    let {
        open, proposal, allKpis,
        onclose, onaccept, onreject,
    }: Props = $props();
</script>

{#if open}
    <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
    <div class="fixed inset-0 z-50 flex items-start justify-center pt-16 bg-black/50" onclick={onclose}>
        <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
        <div class="bg-base-100 rounded-2xl shadow-2xl border border-base-300 w-full max-w-2xl mx-4 max-h-[80vh] overflow-y-auto" onclick={(e) => e.stopPropagation()}>
            <div class="flex items-center justify-between p-4 border-b border-base-300">
                <h3 class="text-lg font-semibold">Propuesta de cambio</h3>
                <button class="btn btn-ghost btn-square btn-sm" onclick={onclose} aria-label="Cerrar">
                    <X class="w-4 h-4" />
                </button>
            </div>
            <div class="p-4">
                <div class="mb-4 text-sm text-base-content/60">
                    <p>Propuesta por: <span class="font-medium text-base-content">{proposal.requestedBy}</span></p>
                    <p>Estado: <span class="badge badge-sm badge-warning">Pendiente</span></p>
                </div>
                <GoalForm
                    mode="readonly"
                    categoryId={proposal.goalId}
                    initialName={proposal.name}
                    initialDescription={proposal.description}
                    initialUnit={proposal.unit}
                    initialWeight={proposal.weight}
                    initialTarget={proposal.targetValue}
                    initialDirection={proposal.direction}
                    initialBaseline={proposal.baselineValue}
                    initialKpiIds={proposal.kpiIds}
                    {allKpis}
                    onsave={async () => {}}
                    oncancel={onclose}
                />
                <div class="flex justify-end gap-2 mt-4">
                    <button class="btn btn-ghost" onclick={onreject}>Rechazar</button>
                    <button class="btn btn-success" onclick={onaccept}>Aceptar propuesta</button>
                </div>
            </div>
        </div>
    </div>
{/if}
