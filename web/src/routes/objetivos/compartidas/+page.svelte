<script lang="ts">
    import { getProfile } from '$lib/stores/devContext.svelte';
    import { goto } from '$app/navigation';
    import { onMount } from 'svelte';
    import { Users, Plus, Edit, Trash, Loader2 } from '@lucide/svelte';
    import ProgressIndicator from '$lib/components/goals/ProgressIndicator.svelte';
    import SharedGoalCreateForm from '$lib/components/goals/SharedGoalCreateForm.svelte';
    import { listSharedGoals, deleteSharedGoal, updateSharedGoal, type SharedGoal } from '$lib/api/sharedGoals';
    import { getActivePhase } from '$lib/api/cycle.svelte';

    const profile = $derived(getProfile());
    const allowedProfiles = ['jefe', 'director', 'director-general'];

    const phase = $derived(getActivePhase() ?? 'inicio-anio');
    const canEditProgress = $derived(phase === 'medio-anio' || phase === 'fin-anio');

    let goals = $state<SharedGoal[]>([]);
    let loading = $state(true);
    let error = $state('');

    onMount(() => {
        loadGoals();
        if (!allowedProfiles.includes(profile)) {
            goto('/');
        }
    });

    async function loadGoals() {
        try {
            loading = true;
            goals = (await listSharedGoals('creator')) ?? [];
        } catch (e) {
            error = e instanceof Error ? e.message : 'Error al cargar metas compartidas';
        } finally {
            loading = false;
        }
    }

    async function handleDelete(goalId: string) {
        if (!confirm('¿Eliminar esta meta compartida?')) return;
        try {
            await deleteSharedGoal(goalId);
            await loadGoals();
        } catch (e) {
            error = e instanceof Error ? e.message : 'Error al eliminar';
        }
    }

    let editGoal = $state<SharedGoal | null>(null);

    const editInitial = $derived(
        editGoal
            ? {
                  name: editGoal.name,
                  description: editGoal.description,
                  unit: editGoal.unit,
                  direction: editGoal.direction as 'ascendente' | 'descendente',
                  weight: editGoal.weight,
                  target_value: editGoal.target_value,
                  baseline_value: editGoal.baseline_value,
              }
            : undefined
    );

    let savingIds = $state<string[]>([]);

    const progressTimers: Record<string, ReturnType<typeof setTimeout>> = {};

    function scheduleProgressSave(goal: SharedGoal) {
        const prev = progressTimers[goal.id];
        if (prev) clearTimeout(prev);
        progressTimers[goal.id] = setTimeout(() => saveProgress(goal.id), 800);
    }

    async function saveProgress(goalId: string) {
        delete progressTimers[goalId];
        const goal = goals.find(g => g.id === goalId);
        if (!goal) return;
        savingIds = [...savingIds, goalId];
        try {
            await updateSharedGoal(goalId, {
                name: goal.name,
                description: goal.description,
                unit: goal.unit,
                direction: goal.direction,
                goal_kind: goal.goal_kind,
                weight: goal.weight,
                target_value: goal.target_value,
                baseline_value: goal.baseline_value,
                current_value: goal.current_value,
            });
        } catch (e) {
            error = e instanceof Error ? e.message : 'Error al guardar avance';
        } finally {
            savingIds = savingIds.filter(id => id !== goalId);
        }
    }

    let showCreate = $state(false);
    let createKind = $state<'qualitative' | 'quantitative'>('qualitative');

    const qualitativeGoals = $derived(goals.filter(g => g.goal_kind === 'qualitative'));
    const quantitativeGoals = $derived(goals.filter(g => g.goal_kind === 'quantitative'));

    const qualitativeSum = $derived(
        qualitativeGoals.reduce((sum, g) => sum + g.weight, 0)
    );
    const quantitativeSum = $derived(
        quantitativeGoals.reduce((sum, g) => sum + g.weight, 0)
    );
    const totalSum = $derived(qualitativeSum + quantitativeSum);

    const weightedProgress = $derived((() => {
        const totalWeight = goals.reduce((s, g) => s + g.weight, 0);
        if (totalWeight === 0) return 0;
        return Math.min(100, Math.max(0, goals.reduce((s, g) => s + progressPercent(g) * g.weight, 0) / totalWeight));
    })());

    function progressPercent(goal: { current_value: number; target_value: number; baseline_value?: number | null; direction: string }): number {
        const current = goal.current_value ?? 0;
        const target = goal.target_value;
        const baseline = goal.baseline_value ?? 0;
        if (goal.direction === 'descendente') {
            if (baseline === target) return 0;
            const pct = ((baseline - current) / (baseline - target)) * 100;
            return Math.min(100, Math.max(0, pct));
        }
        if (target === 0) return 0;
        return Math.min(100, Math.max(0, (current / target) * 100));
    }

    function formatTarget(goal: SharedGoal) {
        if (goal.unit === 'porcentaje') return `${goal.target_value}%`;
        if (goal.unit === 'moneda') return `$${goal.target_value.toLocaleString()}`;
        if (goal.unit === 'binario') return goal.target_value === 1 ? 'Sí' : 'No';
        return `${goal.target_value}`;
    }
</script>

<div class="container mx-auto px-4 py-6 max-w-6xl">
    <div class="flex items-center gap-3 mb-6">
        <div class="p-2 bg-secondary/10 rounded-lg">
            <Users class="w-6 h-6 text-secondary" />
        </div>
        <div>
            <h1 class="text-2xl font-bold">Metas compartidas</h1>
            <p class="text-sm text-base-content/60">Crear y gestionar metas para tu grupo</p>
        </div>
    </div>

    {#if error}
        <div class="alert alert-error mb-4">
            <span>{error}</span>
            <button class="btn btn-ghost btn-xs" onclick={() => error = ''}>×</button>
        </div>
    {/if}

    {#if loading}
        <div class="flex justify-center py-12">
            <Loader2 class="w-8 h-8 animate-spin text-secondary" />
        </div>
    {:else}
        <div class="bg-base-200 rounded-lg p-4 mb-6">
            <div class="flex items-center justify-between mb-2">
                <span class="text-sm font-medium">Progreso global</span>
                <span class="text-sm font-semibold">{Math.round(totalSum)}%</span>
            </div>
            <ProgressIndicator value={weightedProgress} wide />
        </div>

        <!-- Cualitativos -->
        <details open class="collapse collapse-arrow bg-base-100 border border-base-300 rounded-lg mb-4">
            <summary class="collapse-title min-h-0 pl-4 pr-8 py-3 flex items-center justify-between gap-2">
                <div class="flex items-center gap-2">
                    <span class="font-medium">Cualitativos</span>
                    <span class="badge badge-sm">{qualitativeGoals.length}</span>
                </div>
                <div class="flex items-center gap-2">
                    <span class="text-sm text-base-content/60 mr-4">{qualitativeSum}%</span>
                </div>
            </summary>
            <div class="collapse-content px-4">
                {#if qualitativeGoals.length === 0}
                        <p class="text-center text-base-content/60 py-4">No hay metas cualitativas</p>
                    {:else}
                        <div class="space-y-2">
                            {#each qualitativeGoals as goal (goal.id)}
                                <div class="flex items-center justify-between p-3 bg-base-100 rounded-lg">
                                    <div class="flex-1">
                                        <p class="font-medium flex items-center gap-2">
                                            {goal.name}
                                            <span class="badge badge-ghost shrink-0">Peso: {goal.weight}%</span>
                                        </p>
                                        <p class="text-sm text-base-content/60">
                                            Target: {formatTarget(goal)} · {goal.members.length} miembros
                                        </p>
                                    </div>
                                    <div class="flex items-center gap-2">
                                        {#if canEditProgress}
                                            <div class="flex items-center gap-1">
                                                <input
                                                    type="number"
                                                    class="input input-bordered input-xs w-20"
                                                    bind:value={goal.current_value}
                                                    min={0}
                                                    step={0.1}
                                                    oninput={() => scheduleProgressSave(goal)}
                                                    aria-label={`Avance de ${goal.name}`}
                                                />
                                                <span class="text-xs text-base-content/60">{progressPercent(goal).toFixed(1)}%</span>
                                                {#if savingIds.includes(goal.id)}
                                                    <span class="text-xs text-base-content/40">Guardando…</span>
                                                {/if}
                                            </div>
                                        {/if}
                                        <button
                                            type="button"
                                            class="btn btn-ghost btn-xs"
                                            onclick={() => { editGoal = goal; createKind = goal.goal_kind as 'qualitative' | 'quantitative'; showCreate = true; }}
                                        >
                                            <Edit class="w-3 h-3" />
                                        </button>
                                        <button
                                            class="btn btn-ghost btn-xs text-error"
                                            onclick={() => handleDelete(goal.id)}
                                        >
                                            <Trash class="w-3 h-3" />
                                        </button>
                                    </div>
                                </div>
                            {/each}
                        </div>
                    {/if}
                    <button
                        class="btn btn-outline btn-sm mt-4 w-full"
                        onclick={() => { createKind = 'qualitative'; showCreate = true; }}
                    >
                        <Plus class="w-4 h-4" /> Nueva meta cualitativa
                    </button>
            </div>
        </details>

        <!-- Cuantitativos -->
        <details open class="collapse collapse-arrow bg-base-100 border border-base-300 rounded-lg mb-4">
            <summary class="collapse-title min-h-0 pl-4 pr-8 py-3 flex items-center justify-between gap-2">
                <div class="flex items-center gap-2">
                    <span class="font-medium">Cuantitativos</span>
                    <span class="badge badge-sm">{quantitativeGoals.length}</span>
                </div>
                <div class="flex items-center gap-2">
                    <span class="text-sm text-base-content/60 mr-4">{quantitativeSum}%</span>
                </div>
            </summary>
            <div class="collapse-content px-4">
                {#if quantitativeGoals.length === 0}
                        <p class="text-center text-base-content/60 py-4">No hay metas cuantitativas</p>
                    {:else}
                        <div class="space-y-2">
                            {#each quantitativeGoals as goal (goal.id)}
                                <div class="flex items-center justify-between p-3 bg-base-100 rounded-lg">
                                    <div class="flex-1">
                                        <p class="font-medium flex items-center gap-2">
                                            {goal.name}
                                            <span class="badge badge-ghost shrink-0">Peso: {goal.weight}%</span>
                                        </p>
                                        <p class="text-sm text-base-content/60">
                                            Target: {formatTarget(goal)} · {goal.members.length} miembros
                                        </p>
                                    </div>
                                    <div class="flex items-center gap-2">
                                        {#if canEditProgress}
                                            <div class="flex items-center gap-1">
                                                <input
                                                    type="number"
                                                    class="input input-bordered input-xs w-20"
                                                    bind:value={goal.current_value}
                                                    min={0}
                                                    step={0.1}
                                                    oninput={() => scheduleProgressSave(goal)}
                                                    aria-label={`Avance de ${goal.name}`}
                                                />
                                                <span class="text-xs text-base-content/60">{progressPercent(goal).toFixed(1)}%</span>
                                                {#if savingIds.includes(goal.id)}
                                                    <span class="text-xs text-base-content/40">Guardando…</span>
                                                {/if}
                                            </div>
                                        {/if}
                                        <button
                                            type="button"
                                            class="btn btn-ghost btn-xs"
                                            onclick={() => { editGoal = goal; createKind = goal.goal_kind as 'qualitative' | 'quantitative'; showCreate = true; }}
                                        >
                                            <Edit class="w-3 h-3" />
                                        </button>
                                        <button
                                            class="btn btn-ghost btn-xs text-error"
                                            onclick={() => handleDelete(goal.id)}
                                        >
                                            <Trash class="w-3 h-3" />
                                        </button>
                                    </div>
                                </div>
                            {/each}
                        </div>
                    {/if}
                    <button
                        class="btn btn-outline btn-sm mt-4 w-full"
                        onclick={() => { createKind = 'quantitative'; showCreate = true; }}
                    >
                        <Plus class="w-4 h-4" /> Nueva meta cuantitativa
                    </button>
            </div>
        </details>
    {/if}
</div>

<SharedGoalCreateForm
    open={showCreate}
    goalKind={createKind}
    goalId={editGoal?.id}
    initial={editInitial}
    oncancel={() => { showCreate = false; editGoal = null; }}
    onsaved={() => { showCreate = false; editGoal = null; loadGoals(); }}
/>
