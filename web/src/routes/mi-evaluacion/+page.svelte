<script lang="ts">
    import { onMount } from 'svelte';
    import EmployeeEvaluationDetail from "$lib/components/evaluation/EmployeeEvaluationDetail.svelte";
    import { getProfile } from "$lib/stores/devContext.svelte";
    import { getAssignmentsByProfile, load as loadGoals, isLoading as goalsLoading } from "$lib/stores/goalsStore.svelte";
    import PageSkeleton from '$lib/components/ui/PageSkeleton.svelte';
    import { ClipboardCheck } from "@lucide/svelte";

    const profile = $derived(getProfile());
    const assignments = $derived(getAssignmentsByProfile(profile));
    const employeeId = $derived(assignments[0]?.employeeId ?? "");
    const loading = $derived(goalsLoading());

    onMount(() => {
        loadGoals();
    });
</script>

<svelte:head>
    <title>Mi evaluación — SED</title>
</svelte:head>

<div class="flex flex-col gap-6">
    <div>
        <h1 class="text-2xl font-bold text-base-content flex items-center gap-2">
            <ClipboardCheck class="w-6 h-6" />
            Mi evaluación
        </h1>
        <p class="text-sm text-base-content/50 mt-1">
            Autoevaluación de competencias y cierre de metas
        </p>
    </div>

    {#if loading}
        <PageSkeleton variant="card" rows={3} />
    {:else if employeeId}
        <EmployeeEvaluationDetail {employeeId} viewerMode="self" showBreadcrumb={false} />
    {:else}
        <div class="flex flex-col items-center justify-center py-16 text-center">
            <div class="w-16 h-16 rounded-2xl bg-base-200 flex items-center justify-center mb-5">
                <ClipboardCheck class="w-8 h-8 text-base-content/30" strokeWidth={1.5} />
            </div>
            <h3 class="text-lg font-semibold text-base-content/70">Sin asignación de metas</h3>
            <p class="text-base-content/40 mt-1.5 max-w-sm">
                Aún no tienes metas asignadas para este ciclo.
                Crea tus categorías y objetivos para empezar.
            </p>
            <a href="/objetivos/asignacion" class="btn btn-primary btn-sm mt-5 px-6">
                Ir a metas
            </a>
        </div>
    {/if}
</div>
