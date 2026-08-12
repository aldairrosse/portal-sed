<script lang="ts">
    import { getProfile } from '$lib/stores/devContext.svelte';
    import { goto } from '$app/navigation';
    import { onMount } from 'svelte';
    import { Users, Plus, ChevronDown, ChevronUp, Edit, Trash, Loader2 } from '@lucide/svelte';
    import WeightIndicator from '$lib/components/goals/WeightIndicator.svelte';
    import { listSharedGoals, deleteSharedGoal, type SharedGoal } from '$lib/api/sharedGoals';

    const profile = $derived(getProfile());
    const allowedProfiles = ['jefe', 'gerente-tienda', 'divisional', 'regional', 'director'];

    let goals = $state<SharedGoal[]>([]);
    let loading = $state(true);
    let error = $state('');

    onMount(() => {
        if (!allowedProfiles.includes(profile)) {
            goto('/');
            return;
        }
        loadGoals();
    });

    async function loadGoals() {
        try {
            loading = true;
            goals = await listSharedGoals('creator');
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

    let showQualitative = $state(true);
    let showQuantitative = $state(true);

    const qualitativeGoals = $derived(goals.filter(g => g.goal_kind === 'qualitative'));
    const quantitativeGoals = $derived(goals.filter(g => g.goal_kind === 'quantitative'));

    const qualitativeSum = $derived(
        qualitativeGoals.reduce((sum, g) => sum + g.weight, 0)
    );
    const quantitativeSum = $derived(
        quantitativeGoals.reduce((sum, g) => sum + g.weight, 0)
    );
    const totalSum = $derived(qualitativeSum + quantitativeSum);

    function formatTarget(goal: SharedGoal) {
        if (goal.unit === 'porcentaje') return `${goal.target_value}%`;
        if (goal.unit === 'moneda') return `$${goal.target_value.toLocaleString()}`;
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
                <span class="text-sm font-medium">Ponderación total</span>
                <span class="text-sm {totalSum === 100 ? 'text-success' : 'text-warning'}">
                    {totalSum}%
                </span>
            </div>
            <WeightIndicator current={totalSum} label="Ponderación total" />
        </div>

        <!-- Cualitativos -->
        <div class="border border-base-300 rounded-lg mb-4">
            <button
                class="w-full flex items-center justify-between p-4 hover:bg-base-200 transition-colors"
                onclick={() => showQualitative = !showQualitative}
            >
                <div class="flex items-center gap-2">
                    <span class="font-medium">Cualitativos</span>
                    <span class="badge badge-sm">{qualitativeGoals.length}</span>
                </div>
                <div class="flex items-center gap-2">
                    <span class="text-sm text-base-content/60">{qualitativeSum}%</span>
                    {#if showQualitative}
                        <ChevronUp class="w-4 h-4" />
                    {:else}
                        <ChevronDown class="w-4 h-4" />
                    {/if}
                </div>
            </button>

            {#if showQualitative}
                <div class="border-t border-base-300 p-4">
                    {#if qualitativeGoals.length === 0}
                        <p class="text-center text-base-content/60 py-4">No hay metas cualitativas</p>
                    {:else}
                        <div class="space-y-2">
                            {#each qualitativeGoals as goal (goal.id)}
                                <div class="flex items-center justify-between p-3 bg-base-100 rounded-lg">
                                    <div class="flex-1">
                                        <p class="font-medium">{goal.name}</p>
                                        <p class="text-sm text-base-content/60">
                                            Target: {formatTarget(goal)} · {goal.members.length} miembros
                                        </p>
                                    </div>
                                    <div class="flex items-center gap-2">
                                        <span class="badge badge-primary">{goal.weight}%</span>
                                        <button class="btn btn-ghost btn-xs">
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
                    <button class="btn btn-outline btn-sm mt-4 w-full">
                        <Plus class="w-4 h-4" /> Nueva meta cualitativa
                    </button>
                </div>
            {/if}
        </div>

        <!-- Cuantitativos -->
        <div class="border border-base-300 rounded-lg mb-4">
            <button
                class="w-full flex items-center justify-between p-4 hover:bg-base-200 transition-colors"
                onclick={() => showQuantitative = !showQuantitative}
            >
                <div class="flex items-center gap-2">
                    <span class="font-medium">Cuantitativos</span>
                    <span class="badge badge-sm">{quantitativeGoals.length}</span>
                </div>
                <div class="flex items-center gap-2">
                    <span class="text-sm text-base-content/60">{quantitativeSum}%</span>
                    {#if showQuantitative}
                        <ChevronUp class="w-4 h-4" />
                    {:else}
                        <ChevronDown class="w-4 h-4" />
                    {/if}
                </div>
            </button>

            {#if showQuantitative}
                <div class="border-t border-base-300 p-4">
                    {#if quantitativeGoals.length === 0}
                        <p class="text-center text-base-content/60 py-4">No hay metas cuantitativas</p>
                    {:else}
                        <div class="space-y-2">
                            {#each quantitativeGoals as goal (goal.id)}
                                <div class="flex items-center justify-between p-3 bg-base-100 rounded-lg">
                                    <div class="flex-1">
                                        <p class="font-medium">{goal.name}</p>
                                        <p class="text-sm text-base-content/60">
                                            Target: {formatTarget(goal)} · {goal.members.length} miembros
                                        </p>
                                    </div>
                                    <div class="flex items-center gap-2">
                                        <span class="badge badge-secondary">{goal.weight}%</span>
                                        <button class="btn btn-ghost btn-xs">
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
                    <button class="btn btn-outline btn-sm mt-4 w-full">
                        <Plus class="w-4 h-4" /> Nueva meta cuantitativa
                    </button>
                </div>
            {/if}
        </div>

        <div class="flex justify-end gap-2 mt-6">
            <button class="btn btn-ghost">Cancelar</button>
            <button class="btn btn-primary" disabled={totalSum !== 100}>
                Guardar metas compartidas
            </button>
        </div>
    {/if}
</div>
