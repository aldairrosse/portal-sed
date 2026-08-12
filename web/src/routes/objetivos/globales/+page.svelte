<script lang="ts">
    import { getProfile } from '$lib/stores/devContext.svelte';
    import { goto } from '$app/navigation';
    import { onMount } from 'svelte';
    import { Target, Plus, ChevronDown, ChevronUp, Globe, Users, Loader2 } from '@lucide/svelte';
    import WeightIndicator from '$lib/components/goals/WeightIndicator.svelte';
    import { listGlobalGoals, deleteGlobalGoal, executeRules, updateGlobalGoal, type GlobalGoal } from '$lib/api/globalGoals';
    import GlobalGoalCreateForm from '$lib/components/goals/GlobalGoalCreateForm.svelte';

    const profile = $derived(getProfile());

    let goals = $state<GlobalGoal[]>([]);
    let loading = $state(true);
    let error = $state('');
    let showCreate = $state(false);
    let createKind = $state<'qualitative' | 'quantitative'>('qualitative');
    let editGoal = $state<GlobalGoal | null>(null);
    let notice = $state('');
    let runningRules = $state('');
    let savingIds = $state<string[]>([]);

    onMount(() => {
        if (profile !== 'rh') {
            goto('/');
            return;
        }
        loadGoals();
    });

    async function loadGoals() {
        try {
            loading = true;
            goals = await listGlobalGoals();
        } catch (e) {
            error = e instanceof Error ? e.message : 'Error al cargar objetivos';
        } finally {
            loading = false;
        }
    }

    async function handleDelete(goalId: string) {
        if (!confirm('¿Eliminar este objetivo global?')) return;
        try {
            await deleteGlobalGoal(goalId);
            await loadGoals();
        } catch (e) {
            error = e instanceof Error ? e.message : 'Error al eliminar';
        }
    }

    function showNotice(msg: string) {
        notice = msg;
        setTimeout(() => { if (notice === msg) notice = ''; }, 5000);
    }

    function openEdit(goal: GlobalGoal) {
        editGoal = goal;
        createKind = goal.goal_kind === 'quantitative' ? 'quantitative' : 'qualitative';
        showCreate = true;
    }

    async function handleExecuteRules(goalId: string) {
        runningRules = goalId;
        try {
            const res = await executeRules(goalId);
            showNotice(`Se asignó la meta a ${res.assignments_created} empleados`);
        } catch (e) {
            error = e instanceof Error ? e.message : 'Error al ejecutar asignación';
        } finally {
            runningRules = '';
        }
    }

    const weightTimers = new Map<string, ReturnType<typeof setTimeout>>();

    function scheduleWeightSave(goal: GlobalGoal) {
        const prev = weightTimers.get(goal.id);
        if (prev) clearTimeout(prev);
        weightTimers.set(goal.id, setTimeout(() => saveWeight(goal.id), 800));
    }

    async function saveWeight(goalId: string) {
        weightTimers.delete(goalId);
        const goal = goals.find(g => g.id === goalId);
        if (!goal) return;
        savingIds = [...savingIds, goalId];
        try {
            await updateGlobalGoal(goalId, {
                name: goal.name,
                description: goal.description,
                unit: goal.unit,
                direction: goal.direction,
                goal_kind: goal.goal_kind,
                weight: goal.weight,
                target_value: goal.target_value,
            });
        } catch (e) {
            error = e instanceof Error ? e.message : 'Error al guardar ponderación';
        } finally {
            savingIds = savingIds.filter(id => id !== goalId);
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

    function formatTarget(goal: GlobalGoal) {
        if (goal.unit === 'porcentaje') return `${goal.target_value}%`;
        if (goal.unit === 'moneda') return `$${goal.target_value.toLocaleString()}`;
        return `${goal.target_value}`;
    }
</script>

<div class="container mx-auto px-4 py-6 max-w-6xl">
    <div class="flex items-center gap-3 mb-6">
        <div class="p-2 bg-primary/10 rounded-lg">
            <Target class="w-6 h-6 text-primary" />
        </div>
        <div>
            <h1 class="text-2xl font-bold">Objetivos globales</h1>
            <p class="text-sm text-base-content/60">Crear y asignar metas globales desde RH</p>
        </div>
    </div>

    {#if error}
        <div class="alert alert-error mb-4">
            <span>{error}</span>
            <button class="btn btn-ghost btn-xs" onclick={() => error = ''}>×</button>
        </div>
    {/if}

    {#if notice}
        <div class="alert alert-success mb-4">
            <span>{notice}</span>
        </div>
    {/if}

    {#if loading}
        <div class="flex justify-center py-12">
            <Loader2 class="w-8 h-8 animate-spin text-primary" />
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
                    <Globe class="w-5 h-5 text-primary" />
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
                                    <div>
                                        <p class="font-medium">{goal.name}</p>
                                        <p class="text-sm text-base-content/60">Target: {formatTarget(goal)}</p>
                                    </div>
                                    <div class="flex items-center gap-2">
                                        <div class="flex items-center gap-1">
                                            <input
                                                type="number"
                                                class="input input-bordered input-xs w-20"
                                                bind:value={goal.weight}
                                                oninput={() => scheduleWeightSave(goal)}
                                                min={0}
                                                max={100}
                                                step={0.1}
                                                aria-label={`Ponderación de ${goal.name}`}
                                            />
                                            <span class="text-xs text-base-content/60">%</span>
                                            {#if savingIds.includes(goal.id)}
                                                <span class="text-xs text-base-content/40">Guardando…</span>
                                            {/if}
                                        </div>
                                        {#if goal.rules.length > 0}
                                            <button
                                                class="btn btn-outline btn-xs"
                                                disabled={runningRules !== ''}
                                                onclick={() => handleExecuteRules(goal.id)}
                                            >
                                                {#if runningRules === goal.id}
                                                    <Loader2 class="w-3 h-3 animate-spin" />
                                                    Ejecutando…
                                                {:else}
                                                    Ejecutar asignación
                                                {/if}
                                            </button>
                                        {/if}
                                        <button
                                            class="btn btn-ghost btn-xs"
                                            onclick={() => openEdit(goal)}
                                        >
                                            Editar
                                        </button>
                                        <button
                                            class="btn btn-ghost btn-xs text-error"
                                            onclick={() => handleDelete(goal.id)}
                                        >
                                            Eliminar
                                        </button>
                                    </div>
                                </div>
                            {/each}
                        </div>
                    {/if}
                    <button class="btn btn-outline btn-sm mt-4 w-full" onclick={() => { createKind = 'qualitative'; editGoal = null; showCreate = true; }}>
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
                    <Users class="w-5 h-5 text-secondary" />
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
                                    <div>
                                        <p class="font-medium">{goal.name}</p>
                                        <p class="text-sm text-base-content/60">Target: {formatTarget(goal)}</p>
                                    </div>
                                    <div class="flex items-center gap-2">
                                        <div class="flex items-center gap-1">
                                            <input
                                                type="number"
                                                class="input input-bordered input-xs w-20"
                                                bind:value={goal.weight}
                                                oninput={() => scheduleWeightSave(goal)}
                                                min={0}
                                                max={100}
                                                step={0.1}
                                                aria-label={`Ponderación de ${goal.name}`}
                                            />
                                            <span class="text-xs text-base-content/60">%</span>
                                            {#if savingIds.includes(goal.id)}
                                                <span class="text-xs text-base-content/40">Guardando…</span>
                                            {/if}
                                        </div>
                                        {#if goal.rules.length > 0}
                                            <button
                                                class="btn btn-outline btn-xs"
                                                disabled={runningRules !== ''}
                                                onclick={() => handleExecuteRules(goal.id)}
                                            >
                                                {#if runningRules === goal.id}
                                                    <Loader2 class="w-3 h-3 animate-spin" />
                                                    Ejecutando…
                                                {:else}
                                                    Ejecutar asignación
                                                {/if}
                                            </button>
                                        {/if}
                                        <button
                                            class="btn btn-ghost btn-xs"
                                            onclick={() => openEdit(goal)}
                                        >
                                            Editar
                                        </button>
                                        <button
                                            class="btn btn-ghost btn-xs text-error"
                                            onclick={() => handleDelete(goal.id)}
                                        >
                                            Eliminar
                                        </button>
                                    </div>
                                </div>
                            {/each}
                        </div>
                    {/if}
                    <button class="btn btn-outline btn-sm mt-4 w-full" onclick={() => { createKind = 'quantitative'; editGoal = null; showCreate = true; }}>
                        <Plus class="w-4 h-4" /> Nueva meta cuantitativa
                    </button>
                </div>
            {/if}
        </div>
    {/if}

    <GlobalGoalCreateForm
        open={showCreate}
        goalKind={createKind}
        goalId={editGoal?.id}
        initial={editGoal ? {
            name: editGoal.name,
            description: editGoal.description,
            unit: editGoal.unit,
            direction: editGoal.direction as 'ascendente' | 'descendente',
            weight: editGoal.weight,
            target_value: editGoal.target_value,
        } : undefined}
        oncancel={() => { showCreate = false; editGoal = null; }}
        onsaved={() => { showCreate = false; editGoal = null; loadGoals(); }}
    />
</div>
