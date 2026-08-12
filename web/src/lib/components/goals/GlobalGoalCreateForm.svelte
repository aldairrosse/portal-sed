<script lang="ts">
    import { Loader2, Plus, Trash2 } from '@lucide/svelte';
    import CustomSelect from '$lib/components/ui/CustomSelect.svelte';
    import { createGlobalGoal, updateGlobalGoal, type CreateGlobalGoalRequest } from '$lib/api/globalGoals';
    import { load, getRoot, getAllLeafIds, getNodeById, getScopeIds } from '$lib/stores/orgHierarchyStore.svelte';
    import type { OrgNode } from '$lib/types/org-hierarchy';

    interface Props {
        open: boolean;
        goalKind: 'qualitative' | 'quantitative';
        goalId?: string;
        initial?: {
            name: string;
            description: string;
            unit: string;
            direction: 'ascendente' | 'descendente';
            weight: number;
            target_value: number;
        };
        oncancel: () => void;
        onsaved: () => void;
    }

    let { open, goalKind, goalId, initial, oncancel, onsaved }: Props = $props();

    const UNIT_OPTIONS = [
        { value: 'porcentaje', label: 'Porcentaje (%)' },
        { value: 'moneda', label: 'Moneda ($)' },
        { value: 'numero', label: 'Número' },
    ];

    const RULE_TYPE_OPTIONS = [
        { value: 'department', label: 'Por departamento' },
        { value: 'min_direct_reports', label: 'Mínimo de reportes directos' },
    ];

    const root = $derived(getRoot());

    const employeeOptions = $derived(
        (root?.id ? getAllLeafIds(root.id) : [])
            .map(id => getNodeById(id))
            .filter((n): n is OrgNode => !!n?.headEmployee)
            .map(n => ({ value: n.headEmployee!.id, label: `${n.headEmployee!.firstName} ${n.headEmployee!.lastName}` }))
    );

    const departmentOptions = $derived(
        (root?.id ? getScopeIds(root.id) : [])
            .map(id => getNodeById(id))
            .filter((n): n is OrgNode => !!n && ((n.children ?? []).length > 0 || !!n.headEmployee))
            .map(n => ({ value: n.id, label: n.name }))
    );

    interface AssignmentRow {
        employeeId: string;
        employeeName: string;
        weight: number;
        targetValue: number;
    }

    interface RuleRow {
        ruleType: 'department' | 'min_direct_reports';
        departmentId: string;
        minDirectReports: number;
        defaultWeight: number;
    }

    let name = $state('');
    let description = $state('');
    let unit = $state('porcentaje');
    let direction = $state<'ascendente' | 'descendente'>('ascendente');
    let weight = $state(0);
    let targetValue = $state(0);
    let assignments = $state<AssignmentRow[]>([]);
    let rules = $state<RuleRow[]>([]);
    let employeeSelect = $state('');
    let error = $state('');
    let saving = $state(false);

    $effect(() => {
        if (open) {
            if (initial) {
                name = initial.name;
                description = initial.description;
                unit = initial.unit;
                direction = initial.direction;
                weight = initial.weight;
                targetValue = initial.target_value;
            } else {
                name = '';
                description = '';
                unit = 'porcentaje';
                direction = 'ascendente';
                weight = 0;
                targetValue = 0;
            }
            assignments = [];
            rules = [];
            employeeSelect = '';
            error = '';
            saving = false;
            load();
        }
    });

    function addAssignment() {
        const emp = employeeOptions.find(o => o.value === employeeSelect) ?? employeeOptions[0];
        if (!emp) return;
        if (assignments.some(a => a.employeeId === emp.value)) return;
        assignments = [...assignments, { employeeId: emp.value, employeeName: emp.label, weight: 0, targetValue: 0 }];
    }

    function removeAssignment(index: number) {
        assignments = assignments.filter((_, i) => i !== index);
    }

    function addRule() {
        rules = [...rules, { ruleType: 'department', departmentId: departmentOptions[0]?.value ?? '', minDirectReports: 0, defaultWeight: 0 }];
    }

    function removeRule(index: number) {
        rules = rules.filter((_, i) => i !== index);
    }

    function changeRuleType(index: number, value: string) {
        const ruleType = value === 'min_direct_reports' ? 'min_direct_reports' : 'department';
        rules = rules.map((r, i) => i === index ? { ...r, ruleType, departmentId: '', minDirectReports: 0 } : r);
    }

    function validate(): string {
        if (!name.trim()) return 'El nombre es obligatorio';
        if (!(weight >= 0 && weight <= 100)) return 'La ponderación debe estar entre 0 y 100';
        if (!(targetValue > 0)) return 'El valor objetivo debe ser mayor a 0';
        if (!goalId && assignments.length === 0 && rules.length === 0) return 'Selecciona al menos un empleado o una regla';
        for (const a of assignments) {
            if (!(a.weight >= 0 && a.weight <= 100)) return 'Ponderación de empleado inválida (0-100)';
            if (!(a.targetValue > 0)) return 'El valor objetivo del empleado debe ser mayor a 0';
        }
        for (const r of rules) {
            if (!(r.defaultWeight >= 0 && r.defaultWeight <= 100)) return 'Ponderación por defecto inválida (0-100)';
            if (r.ruleType === 'department' && !r.departmentId) return 'Selecciona un departamento';
            if (r.ruleType === 'min_direct_reports' && !(r.minDirectReports > 0)) return 'El mínimo de reportes directos debe ser mayor a 0';
        }
        return '';
    }

    async function handleSave() {
        const err = validate();
        if (err) { error = err; return; }
        error = '';
        saving = true;
        try {
            if (goalId) {
                await updateGlobalGoal(goalId, {
                    name: name.trim(),
                    description: description.trim(),
                    unit,
                    direction,
                    goal_kind: goalKind,
                    weight,
                    target_value: targetValue,
                });
            } else {
                const request: CreateGlobalGoalRequest = {
                    name: name.trim(),
                    description: description.trim(),
                    unit,
                    direction,
                    goal_kind: goalKind,
                    weight,
                    target_value: targetValue,
                    assignments: assignments.length > 0
                        ? assignments.map(a => ({
                            employee_id: a.employeeId,
                            weight: a.weight,
                            target_value: a.targetValue,
                        }))
                        : undefined,
                    rules: rules.length > 0
                        ? rules.map(r => ({
                            rule_type: r.ruleType,
                            ...(r.ruleType === 'department' ? { department_id: r.departmentId } : { min_direct_reports: r.minDirectReports }),
                            default_weight: r.defaultWeight,
                        }))
                        : undefined,
                };
                await createGlobalGoal(request);
            }
            onsaved();
        } catch (e) {
            error = e instanceof Error ? e.message : 'Error al guardar';
        } finally {
            saving = false;
        }
    }
</script>

<dialog
    class="modal"
    open={open}
    aria-modal="true"
    onclick={(e) => { if (e.target === e.currentTarget) oncancel(); }}
>
    <div class="modal-box max-w-2xl">
        <h3 class="font-bold text-lg mb-4">
            {goalId ? 'Editar' : 'Nueva'} meta {goalKind === 'qualitative' ? 'cualitativa' : 'cuantitativa'}
        </h3>

        {#if error}
            <div class="alert alert-error text-sm mb-3" role="alert"><span>{error}</span></div>
        {/if}

        <div class="grid grid-cols-1 md:grid-cols-2 gap-3 mb-3">
            <div class="form-control">
                <label class="label" for="global-goal-name"><span class="label-text text-xs">Nombre de la meta</span></label>
                <input id="global-goal-name" type="text" class="input input-bordered input-sm w-full"
                    bind:value={name} placeholder="Nombre" required />
            </div>
            <div class="form-control">
                <label class="label" for="global-goal-desc"><span class="label-text text-xs">Descripción</span></label>
                <textarea id="global-goal-desc" class="textarea textarea-bordered textarea-sm w-full"
                    rows={1} bind:value={description} placeholder="Descripción"></textarea>
            </div>
            <div class="form-control">
                <span class="label"><span class="label-text text-xs">Unidad de medida</span></span>
                <CustomSelect
                    options={UNIT_OPTIONS}
                    value={unit}
                    onChange={(v) => { unit = v; }}
                    ariaLabel="Unidad"
                />
            </div>
            <div class="form-control">
                <p class="label"><span class="label-text text-xs">Dirección del objetivo</span></p>
                <div class="flex gap-4 pt-1">
                    <label class="flex items-center gap-1.5 cursor-pointer">
                        <input type="radio" class="radio radio-primary radio-xs"
                            name="global-goal-dir" value="ascendente"
                            checked={direction === 'ascendente'}
                            onchange={() => direction = 'ascendente'} />
                        <span class="text-xs">Ascendente (↑)</span>
                    </label>
                    <label class="flex items-center gap-1.5 cursor-pointer">
                        <input type="radio" class="radio radio-primary radio-xs"
                            name="global-goal-dir" value="descendente"
                            checked={direction === 'descendente'}
                            onchange={() => direction = 'descendente'} />
                        <span class="text-xs">Descendente (↓)</span>
                    </label>
                </div>
            </div>
            <div class="form-control">
                <label class="label" for="global-goal-weight"><span class="label-text text-xs">Ponderación (%)</span></label>
                <input id="global-goal-weight" type="number" class="input input-bordered input-sm w-full"
                    bind:value={weight} min={0} max={100} step={0.1} required />
            </div>
            <div class="form-control">
                <label class="label" for="global-goal-target"><span class="label-text text-xs">Valor objetivo a alcanzar</span></label>
                <input id="global-goal-target" type="number" class="input input-bordered input-sm w-full"
                    bind:value={targetValue} min={0} step={0.01} required />
            </div>
        </div>

        {#if !goalId}
        <!-- Asignación a empleados -->
        <div class="border border-base-300 rounded-lg p-3 mb-3">
            <div class="flex items-center justify-between mb-2">
                <span class="label-text text-xs font-semibold">Asignar a empleados</span>
                <div class="flex items-center gap-2">
                    <CustomSelect
                        options={employeeOptions}
                        value={employeeSelect}
                        onChange={(v) => { employeeSelect = v; }}
                        ariaLabel="Empleado"
                    />
                    <button class="btn btn-outline btn-sm" onclick={addAssignment} type="button">
                        <Plus class="w-4 h-4" /> Agregar
                    </button>
                </div>
            </div>
            {#if employeeOptions.length === 0}
                <p class="text-xs text-base-content/60">No hay empleados disponibles en el árbol organizacional.</p>
            {/if}
            {#if assignments.length > 0}
                <div class="space-y-2">
                    {#each assignments as a, i (a.employeeId)}
                        <div class="flex items-center gap-2">
                            <span class="text-sm flex-1 truncate">{a.employeeName}</span>
                            <div class="form-control w-24">
                                <input type="number" class="input input-bordered input-sm w-full"
                                    bind:value={a.weight} min={0} max={100} step={0.1}
                                    aria-label={`Ponderación de ${a.employeeName}`} placeholder="Peso %" />
                            </div>
                            <div class="form-control w-28">
                                <input type="number" class="input input-bordered input-sm w-full"
                                    bind:value={a.targetValue} min={0} step={0.01}
                                    aria-label={`Valor objetivo de ${a.employeeName}`} placeholder="Objetivo" />
                            </div>
                            <button class="btn btn-ghost btn-xs text-error" onclick={() => removeAssignment(i)} type="button">
                                <Trash2 class="w-4 h-4" />
                            </button>
                        </div>
                    {/each}
                </div>
            {/if}
        </div>

        <!-- Reglas de asignación -->
        <div class="border border-base-300 rounded-lg p-3 mb-3">
            <div class="flex items-center justify-between mb-2">
                <span class="label-text text-xs font-semibold">Reglas de asignación (opcional)</span>
                <button class="btn btn-outline btn-sm" onclick={addRule} type="button">
                    <Plus class="w-4 h-4" /> Agregar regla
                </button>
            </div>
            {#if rules.length > 0}
                <div class="space-y-2">
                    {#each rules as r, i (i)}
                        <div class="flex items-center gap-2">
                            <CustomSelect
                                options={RULE_TYPE_OPTIONS}
                                value={r.ruleType}
                                onChange={(v) => changeRuleType(i, v)}
                                ariaLabel="Tipo de regla"
                            />
                            {#if r.ruleType === 'department'}
                                <CustomSelect
                                    options={departmentOptions}
                                    value={r.departmentId}
                                    onChange={(v) => { r.departmentId = v; }}
                                    ariaLabel="Departamento"
                                />
                            {:else}
                                <div class="form-control w-32">
                                    <input type="number" class="input input-bordered input-sm w-full"
                                        bind:value={r.minDirectReports} min={1} step={1}
                                        aria-label="Mínimo de reportes directos" placeholder="Min. reportes" />
                                </div>
                            {/if}
                            <div class="form-control w-24">
                                <input type="number" class="input input-bordered input-sm w-full"
                                    bind:value={r.defaultWeight} min={0} max={100} step={0.1}
                                    aria-label="Ponderación por defecto" placeholder="Peso %" />
                            </div>
                            <button class="btn btn-ghost btn-xs text-error" onclick={() => removeRule(i)} type="button">
                                <Trash2 class="w-4 h-4" />
                            </button>
                        </div>
                    {/each}
                </div>
            {/if}
        </div>
        {/if}

        <div class="modal-action">
            <button class="btn btn-ghost btn-sm" onclick={oncancel} disabled={saving} type="button">Cancelar</button>
            <button class="btn btn-primary btn-sm" onclick={handleSave} disabled={saving} type="button">
                {#if saving}
                    <Loader2 class="w-4 h-4 animate-spin" />
                {/if}
                Guardar
            </button>
        </div>
    </div>
</dialog>
