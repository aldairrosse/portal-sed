<script lang="ts">
    import { FileDown } from "@lucide/svelte";
    import type { EmployeeAssignment } from "$lib/types/goal";

    interface Props {
        open: boolean;
        targetAssignment?: EmployeeAssignment | null;
        availableAssignments: EmployeeAssignment[];
        isBoss: boolean;
        onExportCurrent: () => void;
        onExportAll: () => void;
        onClose: () => void;
    }

    let {
        open,
        targetAssignment,
        availableAssignments,
        isBoss,
        onExportCurrent,
        onExportAll,
        onClose,
    }: Props = $props();

    let dialogEl: HTMLDialogElement | undefined = $state();

    $effect(() => {
        if (!dialogEl) return;
        if (open) {
            dialogEl.showModal();
        } else {
            dialogEl.close();
        }
    });

    function handleCancel() {
        onClose();
    }

    function handleBackdropClick(e: MouseEvent) {
        if (e.target === dialogEl) {
            handleCancel();
        }
    }
</script>

{#if open}
    <dialog
        bind:this={dialogEl}
        class="modal"
        oncancel={handleCancel}
        onclick={handleBackdropClick}
    >
        <div class="modal-box max-w-md">
            <h3 class="font-semibold text-base-content">Exportar CSV</h3>
            <p class="text-sm text-base-content/60 mt-2">
                ¿Qué datos desea exportar?
            </p>
            <div class="mt-4 space-y-2">
                <button
                    class="btn btn-outline w-full justify-start"
                    onclick={onExportCurrent}
                >
                    <FileDown class="w-4 h-4" />
                    {targetAssignment?.employeeName ?? "Empleado actual"}
                </button>
                {#if isBoss}
                    <button
                        class="btn btn-outline w-full justify-start"
                        onclick={onExportAll}
                    >
                        <FileDown class="w-4 h-4" />
                        Todos ({availableAssignments.length} empleados)
                    </button>
                {/if}
            </div>
            <div class="modal-action">
                <button class="btn btn-ghost btn-sm" onclick={handleCancel}>
                    Cancelar
                </button>
            </div>
        </div>
    </dialog>
{/if}
