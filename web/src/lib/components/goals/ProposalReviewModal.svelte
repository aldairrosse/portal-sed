<script lang="ts">
    import GoalForm from './GoalForm.svelte';
    import type { GoalProposal, KPI } from '$lib/types/goal';

    interface Props {
        open: boolean;
        proposal: GoalProposal;
        allKpis: KPI[];
        onclose: () => void;
        onaccept: () => void | Promise<void>;
        onreject: () => void | Promise<void>;
    }

    let { open, proposal, allKpis, onclose, onaccept, onreject }: Props = $props();
    let dialogEl: HTMLDialogElement | undefined = $state();

    $effect(() => {
        if (!dialogEl) return;
        if (open) {
            dialogEl.showModal();
        } else {
            dialogEl.close();
        }
    });
</script>

<dialog bind:this={dialogEl} class="modal" onclose={onclose}>
    <div class="modal-box">
        <form method="dialog">
            <button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2">✕</button>
        </form>
        <h3 class="text-lg font-bold">Propuesta de cambio</h3>
        <div class="py-4 pb-0">
            <div class="mb-4 text-sm text-base-content/60">
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
                <button class="btn btn-sm btn-ghost" onclick={onreject}>Rechazar</button>
                <button class="btn btn-sm btn-success" onclick={onaccept}>Aceptar propuesta</button>
            </div>
        </div>
    </div>
    <form method="dialog" class="modal-backdrop">
        <button>close</button>
    </form>
</dialog>
