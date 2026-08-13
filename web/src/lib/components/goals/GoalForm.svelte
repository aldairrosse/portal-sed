<script lang="ts">
    import { Check } from '@lucide/svelte';
    import type { GoalUnit, KPI } from '$lib/types/goal';
    import { validateGoal, UNIT_OPTIONS } from './goalValidation';
    import CustomSelect from '$lib/components/ui/CustomSelect.svelte';

    interface Props {
        mode: 'create' | 'edit' | 'proposal' | 'readonly';
        categoryId: string;
        goalId?: string;
        initialName?: string;
        initialDescription?: string;
        initialUnit?: GoalUnit;
        initialWeight?: number;
        initialTarget?: number;
        initialDirection?: 'ascendente' | 'descendente';
        initialBaseline?: number;
        initialKpiIds?: string[];
        allKpis: KPI[];
        submitLabel?: string;
        error?: string;
        onsave: (data: {
            name: string;
            description: string;
            unit: GoalUnit;
            weight: number;
            targetValue: number;
            direction: 'ascendente' | 'descendente';
            baselineValue?: number;
            kpiIds: string[];
        }) => void | Promise<void>;
        oncancel: () => void;
    }

    let {
        mode = 'edit',
        categoryId,
        goalId,
        initialName = '',
        initialDescription = '',
        initialUnit = 'porcentaje' as GoalUnit,
        initialWeight = 0,
        initialTarget = 0,
        initialDirection = 'ascendente' as 'ascendente' | 'descendente',
        initialBaseline,
        initialKpiIds = [],
        allKpis,
        submitLabel = 'Guardar',
        error: externalError = '',
        onsave,
        oncancel,
    }: Props = $props();

    // svelte-ignore state_referenced_locally (intentional: form state seeded once from prop)
    let name = $state(initialName);
    // svelte-ignore state_referenced_locally (intentional: form state seeded once from prop)
    let description = $state(initialDescription);
    // svelte-ignore state_referenced_locally (intentional: form state seeded once from prop)
    let unit = $state(initialUnit);
    // svelte-ignore state_referenced_locally (intentional: form state seeded once from prop)
    let weight = $state(initialWeight);
    // svelte-ignore state_referenced_locally (intentional: form state seeded once from prop)
    let target = $state(initialTarget);
    // svelte-ignore state_referenced_locally (intentional: form state seeded once from prop)
    let direction = $state(initialDirection);
    // svelte-ignore state_referenced_locally (intentional: form state seeded once from prop)
    let baseline = $state<number | undefined>(initialBaseline);
    // svelte-ignore state_referenced_locally (intentional: form state seeded once from prop)
    let kpiIds = $state<string[]>([...initialKpiIds]);
    let localError = $state('');

    let error = $derived(localError || externalError);

    function handleToggleKpi(kpiId: string) {
        kpiIds = kpiIds.includes(kpiId)
            ? kpiIds.filter(id => id !== kpiId)
            : [...kpiIds, kpiId];
    }

    async function handleSave() {
        const err = validateGoal({
            name, description, weight,
            targetValue: target,
            direction, baselineValue: baseline,
            categoryId, goalId,
        });
        if (err) { localError = err; return; }
        localError = '';
        try {
            await onsave({
                name: name.trim(),
                description: description.trim(),
                unit, weight,
                targetValue: target,
                direction,
                baselineValue: direction === 'descendente' ? baseline : undefined,
                kpiIds,
            });
        } catch (e) {
            localError = e instanceof Error ? e.message : 'Error al guardar';
        }
    }
</script>

<div class="border border-base-300 rounded-lg p-4 bg-base-200/50 w-full">
    {#if error}
        <div class="alert alert-error text-sm mb-3" role="alert"><span>{error}</span></div>
    {/if}
    <div class="grid grid-cols-1 md:grid-cols-2 gap-3 mb-3">
        <div class="form-control">
            <label class="label" for="goal-name"><span class="label-text text-xs">Nombre de meta u objetivo</span></label>
            <input id="goal-name" type="text" class="input input-bordered input-sm w-full"
                bind:value={name} placeholder="Nombre" required readonly={mode === 'readonly'} />
        </div>
        <div class="form-control">
            <label class="label" for="goal-desc"><span class="label-text text-xs">Descripci&oacute;n a realizar</span></label>
            <textarea id="goal-desc" class="textarea textarea-bordered textarea-sm w-full"
                rows={1} bind:value={description} placeholder="Descripci&oacute;n" required
                readonly={mode === 'readonly'}></textarea>
        </div>
        <div class="form-control">
            <label class="label" for="goal-weight"><span class="label-text text-xs">Ponderaci&oacute;n (%)</span></label>
            <input id="goal-weight" type="number" class="input input-bordered input-sm w-full"
                bind:value={weight} min={0} max={100} step={0.1} required
                readonly={mode === 'readonly'} />
        </div>
        <div class="form-control">
            <label class="label" for="goal-unit"><span class="label-text text-xs">Unidad de medida</span></label>
            {#if mode === 'readonly'}
                <input class="input input-bordered input-sm w-full" value={unit} readonly />
            {:else}
                <CustomSelect
                    options={UNIT_OPTIONS}
                    value={unit}
                    onChange={(v) => { unit = v as GoalUnit; }}
                    ariaLabel="Unidad"
                />
            {/if}
        </div>
        <div class="form-control">
            <p class="label"><span class="label-text text-xs">Direcci&oacute;n del objetivo</span></p>
            <div class="flex gap-4 pt-1">
                <label class="flex items-center gap-1.5 cursor-pointer">
                    <input type="radio" class="radio radio-primary radio-xs"
                        disabled={mode === 'readonly'}
                        name="goal-dir" value="ascendente"
                        checked={direction === 'ascendente'}
                        onchange={() => { direction = 'ascendente'; baseline = undefined; }} />
                    <span class="text-xs">Ascendente (↑)</span>
                </label>
                <label class="flex items-center gap-1.5 cursor-pointer">
                    <input type="radio" class="radio radio-primary radio-xs"
                        disabled={mode === 'readonly'}
                        name="goal-dir" value="descendente"
                        checked={direction === 'descendente'}
                        onchange={() => direction = 'descendente'} />
                    <span class="text-xs">Descendente (↓)</span>
                </label>
            </div>
        </div>
        <div class="form-control">
            <label class="label" for="goal-target"><span class="label-text text-xs">Valor objetivo a alcanzar</span></label>
            <input id="goal-target" type="number" class="input input-bordered input-sm w-full"
                bind:value={target} min={0} step={0.01} required
                readonly={mode === 'readonly'} />
        </div>
        {#if direction === 'descendente'}
            <div class="form-control">
                <label class="label" for="goal-baseline"><span class="label-text text-xs">Punto de partida actual</span></label>
                <input id="goal-baseline" type="number" class="input input-bordered input-sm w-full"
                    bind:value={baseline} min={0} step={0.01} required
                    readonly={mode === 'readonly'} />
            </div>
        {/if}
    </div>
    {#if allKpis.length > 0}
        <div class="form-control mb-3">
            <span class="label"><span class="label-text text-xs">Indicadores clave (KPI)</span></span>
            <div class="flex flex-wrap gap-2">
                {#each allKpis as kpi (kpi.id)}
                    <label class="flex items-center gap-1.5 cursor-pointer px-2 py-1 rounded border border-base-300 hover:bg-base-200/50 text-xs">
                        <input type="checkbox" class="checkbox checkbox-xs checkbox-primary"
                            checked={kpiIds.includes(kpi.id)}
                            disabled={mode === 'readonly'}
                            onchange={() => handleToggleKpi(kpi.id)} />
                        <span>{kpi.name}</span>
                    </label>
                {/each}
            </div>
        </div>
    {/if}
    {#if mode !== 'readonly'}
        <div class="flex justify-end gap-2">
            <button class="btn btn-ghost btn-sm" onclick={oncancel}>Cancelar</button>
            <button class="btn btn-primary btn-sm" onclick={handleSave}>
                <Check class="w-4 h-4" /> {submitLabel}
            </button>
        </div>
    {/if}
</div>
