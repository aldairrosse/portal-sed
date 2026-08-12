<script lang="ts">
    import { getProfile } from '$lib/stores/devContext.svelte';
    import { goto } from '$app/navigation';
    import { onMount } from 'svelte';
    import { Target, Plus, ChevronDown, ChevronUp, Globe, Users } from '@lucide/svelte';
    import WeightIndicator from '$lib/components/goals/WeightIndicator.svelte';

    const profile = $derived(getProfile());
    
    // Redirect if not RH
    onMount(() => {
        if (profile !== 'rh') {
            goto('/');
        }
    });

    // Mock data for now
    let qualitativeGoals = $state([
        { id: '1', name: 'Mejorar clima laboral', weight: 20, targetValue: 80, unit: 'porcentaje' },
        { id: '2', name: 'Desarrollar liderazgo', weight: 15, targetValue: 70, unit: 'porcentaje' },
    ]);
    
    let quantitativeGoals = $state([
        { id: '3', name: 'Incrementar ventas', weight: 25, targetValue: 500000, unit: 'moneda' },
        { id: '4', name: 'Reducir costos', weight: 10, targetValue: 15, unit: 'porcentaje' },
    ]);

    let showQualitative = $state(true);
    let showQuantitative = $state(true);

    const qualitativeSum = $derived(
        qualitativeGoals.reduce((sum, g) => sum + g.weight, 0)
    );
    const quantitativeSum = $derived(
        quantitativeGoals.reduce((sum, g) => sum + g.weight, 0)
    );
    const totalSum = $derived(qualitativeSum + quantitativeSum);
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

    <!-- Weight summary -->
    <div class="bg-base-200 rounded-lg p-4 mb-6">
        <div class="flex items-center justify-between mb-2">
            <span class="text-sm font-medium">Ponderación total</span>
            <span class="text-sm {totalSum === 100 ? 'text-success' : 'text-warning'}">
                {totalSum}%
            </span>
        </div>
        <WeightIndicator value={totalSum} max={100} />
    </div>

    <!-- Qualitative section -->
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
                                    <p class="text-sm text-base-content/60">Target: {goal.targetValue} {goal.unit === 'porcentaje' ? '%' : ''}</p>
                                </div>
                                <span class="badge badge-primary">{goal.weight}%</span>
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

    <!-- Quantitative section -->
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
                                    <p class="text-sm text-base-content/60">Target: {goal.targetValue} {goal.unit === 'porcentaje' ? '%' : goal.unit === 'moneda' ? 'MXN' : ''}</p>
                                </div>
                                <span class="badge badge-secondary">{goal.weight}%</span>
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

    <!-- Actions -->
    <div class="flex justify-end gap-2 mt-6">
        <button class="btn btn-ghost">Cancelar</button>
        <button class="btn btn-primary" disabled={totalSum !== 100}>
            Guardar objetivos
        </button>
    </div>
</div>
