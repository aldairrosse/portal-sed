<script lang="ts">
    import { getProfile } from '$lib/stores/devContext.svelte';
    import { goto } from '$app/navigation';
    import { onMount } from 'svelte';
    import { Target, Plus, Globe, Users, Loader2, TrendingDown, TrendingUp } from '@lucide/svelte';
    import ProgressIndicator from '$lib/components/goals/ProgressIndicator.svelte';
    import { listGlobalGoals, deleteGlobalGoal, updateGlobalGoal, type GlobalGoal } from '$lib/api/globalGoals';
    import { getActivePhase } from '$lib/api/cycle.svelte';
    import GlobalGoalCreateForm from '$lib/components/goals/GlobalGoalCreateForm.svelte';
    import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte';
    import { humanizeError } from '$lib/utils/error';
    import { getCycleWeightConfig, saveCycleWeightConfig } from '$lib/api/weightConfig';

    const profile = $derived(getProfile());

    const phase = $derived(getActivePhase() ?? 'inicio-anio');
    const canEditProgress = $derived(phase === 'medio-anio' || phase === 'fin-anio');

    let goals = $state<GlobalGoal[]>([]);
    let loading = $state(true);
    let error = $state('');
    let showCreate = $state(false);
    let createKind = $state<'qualitative' | 'quantitative'>('qualitative');
    let editGoal = $state<GlobalGoal | null>(null);
    let notice = $state('');
    let savingIds = $state<string[]>([]);
    let confirmGoal = $state<GlobalGoal | null>(null);
    const confirmMessage = $derived(
        confirmGoal
            ? `Se eliminarán las asignaciones (${confirmGoal.assignments.length}) y reglas (${confirmGoal.rules.length}) asociadas y luego la meta. ¿Continuar?`
            : 'Se eliminarán las asignaciones y reglas asociadas y luego la meta. ¿Continuar?'
    );

    let gWeight = $state(0);
    let weightSaving = $state(false);
    let weightError = $state('');
    let weightTimer: ReturnType<typeof setTimeout> | null = null;

    onMount(() => {
        if (profile !== 'rh') {
            goto('/');
            return;
        }
        loadGoals();
        loadWeight();
    });

    async function loadWeight() {
        try {
            const w = await getCycleWeightConfig();
            gWeight = w.g_weight ?? 0;
        } catch {
            // fallback P=100 already
        }
    }
    function onGInput(e: Event) {
        const v = Number((e.target as HTMLInputElement).value);
        gWeight = Math.min(100, Math.max(0, isNaN(v) ? 0 : v));
        if (weightTimer) clearTimeout(weightTimer);
        weightTimer = setTimeout(saveWeight, 600);
    }
    async function saveWeight() {
        if (weightTimer) { clearTimeout(weightTimer); weightTimer = null; }
        weightSaving = true;
        weightError = '';
        try {
            const w = await saveCycleWeightConfig(gWeight);
            gWeight = w.g_weight;
        } catch (e) {
            weightError = humanizeError(e, 'Error al guardar ponderación');
        } finally {
            weightSaving = false;
        }
    }

    async function loadGoals() {
        try {
            loading = true;
            goals = (await listGlobalGoals()) ?? [];
        } catch (e) {
            error = humanizeError(e, 'Error al cargar objetivos');
        } finally {
            loading = false;
        }
    }

    function requestDelete(goal: GlobalGoal) {
        confirmGoal = goal;
    }

    async function confirmDelete() {
        if (!confirmGoal) return;
        const id = confirmGoal.id;
        confirmGoal = null;
        try {
            await deleteGlobalGoal(id);
            await loadGoals();
        } catch (e) {
            error = humanizeError(e, 'Error al eliminar');
        }
    }

    function cancelDelete() {
        confirmGoal = null;
    }

    function openEdit(goal: GlobalGoal) {
        editGoal = goal;
        createKind = goal.goal_kind === 'quantitative' ? 'quantitative' : 'qualitative';
        showCreate = true;
    }

    const progressTimers: Record<string, ReturnType<typeof setTimeout>> = {};

    function scheduleProgressSave(goal: GlobalGoal) {
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
            await updateGlobalGoal(goalId, {
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
            error = humanizeError(e, 'Error al guardar avance');
        } finally {
            savingIds = savingIds.filter(id => id !== goalId);
        }
    }

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

    function formatTarget(goal: GlobalGoal) {
        if (goal.unit === 'porcentaje') return `${goal.target_value}%`;
        if (goal.unit === 'moneda') return `$${goal.target_value.toLocaleString()}`;
        if(goal.unit === 'binario') return goal.target_value === 1 ? 'Sí' : 'No';
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
        <div class="bg-base-100 border border-base-300 rounded-lg p-4 mb-4">
            <h2 class="text-sm font-semibold mb-2">Ponderación institucional</h2>
            <p class="text-xs text-base-content/60 mb-3">RH edita Global; Personal = 100 − Global se calcula automáticamente.</p>
            {#if weightError}<p class="text-xs text-error mb-2">{weightError}</p>{/if}
            <div class="flex items-end gap-4">
                <label class="form-control w-28">
                    <span class="label-text text-xs">Global</span>
                    <input type="number" class="input input-bordered input-sm" min={0} max={100} value={gWeight} oninput={onGInput} aria-label="Global" />
                </label>
                <label class="form-control w-28">
                    <span class="label-text text-xs">Personal</span>
                    <input type="number" class="input input-bordered input-sm" value={100 - gWeight} readonly aria-label="Personal" />
                </label>
                {#if weightSaving}<span class="text-xs text-base-content/50">Guardando…</span>{/if}
            </div>
        </div>
        <div class="bg-base-200 rounded-lg p-4 mb-6">
            <div class="flex items-center justify-between mb-2">
                <span class="text-sm font-medium">Progreso global</span>
                <span class="text-sm font-semibold">{gWeight}%</span>
            </div>
            <ProgressIndicator value={weightedProgress} wide />
        </div>

        <!-- Cualitativos -->
        <details open class="collapse collapse-arrow bg-base-100 border border-base-300 rounded-lg mb-4">
            <summary class="collapse-title min-h-0 pl-4 pr-8 py-3 flex items-center justify-between gap-2">
                <div class="flex items-center gap-2">
                    <Globe class="w-5 h-5 text-primary" />
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
                                    <div>
                                        <p class="font-medium flex items-center gap-2">
                                            {goal.name}
                                            <span class="badge badge-ghost shrink-0">Peso: {goal.weight}%</span>
                                        </p>
                                        <p class="text-sm text-base-content/60 flex items-center gap-1">
                                            {#if goal.direction === "ascendente"}
                                                <TrendingUp class="w-3 h-3" />
                                            {:else}
                                                <TrendingDown class="w-3 h-3" />
                                            {/if}
                                            {formatTarget(goal)}
                                        </p>
                                        <p class="text-xs text-base-content/50">{goal.rules.length} reglas · {goal.assignments.length} asignados</p>
                                    </div>
                                    <div class="flex items-center gap-2">
                                        {#if canEditProgress}
                                            <div class="flex items-center gap-1">
                                                <input
                                                    type="number"
                                                    class="input input-bordered input-xs w-20"
                                                    bind:value={goal.current_value}
                                                    oninput={() => scheduleProgressSave(goal)}
                                                    min={0}
                                                    step={0.1}
                                                    aria-label={`Avance de ${goal.name}`}
                                                />
                                                <span class="text-xs text-base-content/60">{progressPercent(goal).toFixed(1)}%</span>
                                                {#if savingIds.includes(goal.id)}
                                                    <span class="text-xs text-base-content/40">Guardando…</span>
                                                {/if}
                                            </div>
                                        {/if}
                                        <button
                                            class="btn btn-ghost btn-xs"
                                            onclick={() => openEdit(goal)}
                                        >
                                            Editar
                                        </button>
                                        <button
                                            class="btn btn-ghost btn-xs text-error"
                                            onclick={() => requestDelete(goal)}
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
        </details>

        <!-- Cuantitativos -->
        <details open class="collapse collapse-arrow bg-base-100 border border-base-300 rounded-lg mb-4">
            <summary class="collapse-title min-h-0 pl-4 pr-8 py-3 flex items-center justify-between gap-2">
                <div class="flex items-center gap-2">
                    <Users class="w-5 h-5 text-secondary" />
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
                                    <div>
                                        <p class="font-medium flex items-center gap-2">
                                            {goal.name}
                                            <span class="badge badge-ghost shrink-0">Peso: {goal.weight}%</span>
                                        </p>
                                        <p class="text-sm text-base-content/60 flex items-center gap-1">
                                            {#if goal.direction === "ascendente"}
                                                <TrendingUp class="w-3 h-3" />
                                            {:else}
                                                <TrendingDown class="w-3 h-3" />
                                            {/if}
                                            {formatTarget(goal)}
                                        </p>
                                        <p class="text-xs text-base-content/50">{goal.rules.length} reglas · {goal.assignments.length} asignados</p>
                                    </div>
                                    <div class="flex items-center gap-2">
                                        {#if canEditProgress}
                                            <div class="flex items-center gap-1">
                                                <input
                                                    type="number"
                                                    class="input input-bordered input-xs w-20"
                                                    bind:value={goal.current_value}
                                                    oninput={() => scheduleProgressSave(goal)}
                                                    min={0}
                                                    step={0.1}
                                                    aria-label={`Avance de ${goal.name}`}
                                                />
                                                <span class="text-xs text-base-content/60">{progressPercent(goal).toFixed(1)}%</span>
                                                {#if savingIds.includes(goal.id)}
                                                    <span class="text-xs text-base-content/40">Guardando…</span>
                                                {/if}
                                            </div>
                                        {/if}
                                        <button
                                            class="btn btn-ghost btn-xs"
                                            onclick={() => openEdit(goal)}
                                        >
                                            Editar
                                        </button>
                                        <button
                                            class="btn btn-ghost btn-xs text-error"
                                            onclick={() => requestDelete(goal)}
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
        </details>
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
            baseline_value: editGoal.baseline_value,
            rules: editGoal.rules,
            assignments: editGoal.assignments,
        } : undefined}
        oncancel={() => { showCreate = false; editGoal = null; }}
        onsaved={() => { showCreate = false; editGoal = null; loadGoals(); }}
    />

    <ConfirmDialog
        open={confirmGoal !== null}
        title="Eliminar meta global"
        message={confirmMessage}
        variant="error"
        confirmLabel="Eliminar"
        onconfirm={confirmDelete}
        oncancel={cancelDelete}
    />
</div>
