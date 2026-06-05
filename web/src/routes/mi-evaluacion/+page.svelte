<script lang="ts">
    import EmployeeEvaluationDetail from "$lib/components/evaluation/EmployeeEvaluationDetail.svelte";
    import { getProfile } from "$lib/stores/devContext.svelte";
    import { getAssignmentsByProfile } from "$lib/stores/goalsStore.svelte";

    const profile = $derived(getProfile());
    const assignments = $derived(getAssignmentsByProfile(profile));
    const employeeId = $derived(assignments[0]?.employeeId ?? "");
</script>

<svelte:head>
    <title>Mi evaluación — SED</title>
</svelte:head>

<div class="flex flex-col gap-6">
    <div>
        <h1 class="text-2xl font-bold text-base-content">Mi evaluación</h1>
        <p class="text-sm text-base-content/50 mt-1">
            Autoevaluación de competencias y cierre de metas
        </p>
    </div>

    {#if employeeId}
        <EmployeeEvaluationDetail {employeeId} viewerMode="self" showBreadcrumb={false} />
    {:else}
        <p class="text-sm text-base-content/30 italic">
            No hay asignación configurada para tu perfil.
        </p>
    {/if}
</div>
