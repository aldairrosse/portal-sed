<script lang="ts">
    import { Loader2, Plus, Trash2, X } from '@lucide/svelte';
    import { untrack } from 'svelte';
    import CustomSelect from '$lib/components/ui/CustomSelect.svelte';
    import { createGlobalGoal, updateGlobalGoal, type CreateGlobalGoalRequest, type CreateAssignmentRequest, type CreateRuleRequest } from '$lib/api/globalGoals';
    import { client } from '$lib/api/client';
    import { load, getRoot, getNodeById, getScopeIds } from '$lib/stores/orgHierarchyStore.svelte';
    import { loadFirstPage, search, loadMore, getEmployeeOptions, hasMoreEmployees, isLoadingMore } from '$lib/stores/employeePickerStore.svelte';
    import { getProfiles, load as loadCompetencyData } from '$lib/stores/competencyStore.svelte';
    import { PROFILE_LABELS } from '$lib/types/evaluation';
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
            baseline_value?: number;
            rules?: CreateRuleRequest[];
            assignments?: CreateAssignmentRequest[];
        };
        oncancel: () => void;
        onsaved: () => void;
    }

    let { open, goalKind, goalId, initial, oncancel, onsaved }: Props = $props();

    const UNIT_OPTIONS = [
        { value: 'porcentaje', label: 'Porcentaje (%)' },
        { value: 'moneda', label: 'Moneda ($)' },
        { value: 'numero', label: 'Número' },
        { value: 'binario', label: 'Binario (Sí/No)' },
    ];

    const RULE_TYPE_LABELS: Record<'department' | 'min_direct_reports' | 'role', string> = {
        department: 'Departamento',
        min_direct_reports: 'Min. reportes',
        role: 'Rol',
    };

    const root = $derived(getRoot());

    const employeeOptions = $derived(getEmployeeOptions());

    const departmentOptions = $derived(
        (root?.id ? getScopeIds(root.id) : [])
            .map(id => getNodeById(id))
            .filter((n): n is OrgNode => !!n && ((n.children ?? []).length > 0 || !!n.headEmployee))
            .map(n => ({ value: n.id, label: n.name }))
    );

    const profileOptions = $derived(
        getProfiles().map(p => ({ value: p.id, label: PROFILE_LABELS[p.name] ?? p.name }))
    );

    interface AssignmentRow {
        employeeId: string;
        employeeName: string;
        weight: number;
        targetValue: number;
    }

    interface RuleRow {
        ruleType: 'department' | 'min_direct_reports' | 'role';
        departmentId: string;
        minDirectReports: number;
        profileId: string;
        defaultWeight: number;
        defaultTarget: number;
    }

    let name = $state('');
    let description = $state('');
    let unit = $state('porcentaje');
    let direction = $state<'ascendente' | 'descendente'>('ascendente');
    let weight = $state(0);
    let targetValue = $state(0);
    let baselineValue = $state<number | null>(null);
    let assignments = $state<AssignmentRow[]>([]);
    let rules = $state<RuleRow[]>([]);
    let employeeSelect = $state('');
    let error = $state('');
    let saving = $state(false);
    let pickerInitialized = $state(false);
    let loadingNames = $state(false);
    let initializedFor: string | null = null;

    $effect(() => {
        if (!open) {
            // ponytail: reset so re-opening the modal re-runs init once
            initializedFor = null;
            return;
        }
        const key = goalId ?? '';
        if (initializedFor === key) return;
        initializedFor = key;
        {
            const init = initial;
            if (init) {
                name = init.name;
                description = init.description;
                unit = init.unit;
                direction = init.direction;
                weight = init.weight;
                targetValue = init.target_value;
                baselineValue = init?.baseline_value ?? null;
            } else {
                name = '';
                description = '';
                unit = 'porcentaje';
                direction = 'ascendente';
                weight = 0;
                targetValue = 0;
                baselineValue = null;
            }
            assignments = (init?.assignments ?? []).map(a => ({
                employeeId: a.employee_id,
                employeeName: '',
                weight: a.weight,
                targetValue: a.target_value,
            }));
            rules = (init?.rules ?? []).map(r => ({
                ruleType: r.rule_type === 'min_direct_reports' ? 'min_direct_reports' : r.rule_type === 'role' ? 'role' : 'department',
                departmentId: r.department_id ?? '',
                minDirectReports: r.min_direct_reports ?? 0,
                profileId: r.profile_id ?? '',
                defaultWeight: r.default_weight,
                defaultTarget: r.default_target ?? 0,
            }));
            employeeSelect = '';
            error = '';
            saving = false;
            load();
            if (getProfiles().length === 0) {
                loadCompetencyData();
            }
            if (!goalId && !pickerInitialized) {
                pickerInitialized = true;
                loadFirstPage();
            }
            if (goalId && init?.assignments && init.assignments.length > 0) {
                resolveAssignmentNames();
            }
        }
    });

    async function resolveEmployeeName(employeeId: string): Promise<string> {
        const fromPicker = employeeOptions.find(o => o.value === employeeId);
        if (fromPicker) return fromPicker.label;
        const res = await client.GET('/employees/{empId}', {
            params: { path: { empId: employeeId } },
        });
        const data = (res.data as { data?: { firstName?: string; lastName?: string } })?.data;
        return data ? `${data.firstName} ${data.lastName}`.trim() : employeeId;
    }

    async function resolveAssignmentNames() {
        // ponytail: untrack so writing the resolved array below doesn't re-trigger the $effect
        const current = untrack(() => assignments);
        if (!current.some(a => !a.employeeName)) return;
        loadingNames = true;
        try {
            const resolved = await Promise.all(current.map(async a => {
                if (a.employeeName) return a;
                const name = await resolveEmployeeName(a.employeeId);
                return { ...a, employeeName: name || a.employeeId };
            }));
            assignments = resolved;
        } finally {
            loadingNames = false;
        }
    }

    function addAssignment() {
        const emp = employeeOptions.find(o => o.value === employeeSelect);
        if (!emp) return;
        if (assignments.some(a => a.employeeId === emp.value)) return;
        assignments = [...assignments, { employeeId: emp.value, employeeName: emp.label, weight: weight > 0 ? weight : 0, targetValue: targetValue }];
    }

    function removeAssignment(index: number) {
        assignments = assignments.filter((_, i) => i !== index);
    }

    function addDepartmentRule() {
        rules = [...rules, { ruleType: 'department', departmentId: departmentOptions[0]?.value ?? '', minDirectReports: 0, profileId: '', defaultWeight: weight > 0 ? weight : 0, defaultTarget: targetValue }];
    }

    function addRoleRule() {
        rules = [...rules, { ruleType: 'role', departmentId: '', minDirectReports: 0, profileId: '', defaultWeight: weight > 0 ? weight : 0, defaultTarget: targetValue }];
    }

    function addMinReportsRule() {
        rules = [...rules, { ruleType: 'min_direct_reports', departmentId: '', minDirectReports: 0, profileId: '', defaultWeight: weight > 0 ? weight : 0, defaultTarget: targetValue }];
    }

    function removeRule(index: number) {
        rules = rules.filter((_, i) => i !== index);
    }

    function departmentLabel(deptId: string): string {
        if (!deptId) return '';
        return departmentOptions.find(o => o.value === deptId)?.label ?? getNodeById(deptId)?.name ?? '';
    }

    function profileLabel(profileId: string): string {
        if (!profileId) return '';
        return profileOptions.find(o => o.value === profileId)?.label ?? '';
    }

    function validate(): string {
        if (!name.trim()) return 'El nombre es obligatorio';
        if (!(weight >= 0 && weight <= 100)) return 'La ponderación debe estar entre 0 y 100';
        if (targetValue <= 0 && direction !== 'descendente' && unit !== 'binario') return 'El valor objetivo debe ser mayor a 0';
        if (direction === 'descendente') {
            if (baselineValue === null || baselineValue <= targetValue) return 'Para objetivos descendentes, el valor inicial debe ser mayor al objetivo';
        }
        if (!goalId && assignments.length === 0 && rules.length === 0) return 'Selecciona al menos un empleado o una regla';
        for (const a of assignments) {
            if (!(a.weight >= 0 && a.weight <= 100)) return 'Ponderación de empleado inválida (0-100)';
            if (a.targetValue < 0) return 'El valor objetivo del empleado debe ser mayor o igual a 0';
        }
        for (const r of rules) {
            if (!(r.defaultWeight >= 0 && r.defaultWeight <= 100)) return 'Ponderación por defecto inválida (0-100)';
            if (r.defaultTarget < 0) return 'El objetivo por defecto debe ser mayor o igual a 0';
            if (r.ruleType === 'department' && !r.departmentId) return 'Selecciona un departamento';
            if (r.ruleType === 'role' && !r.profileId) return 'Selecciona un perfil';
            if (r.ruleType === 'min_direct_reports' && !(r.minDirectReports > 0)) return 'El mínimo de reportes directos debe ser mayor a 0';
        }
        if (rules.filter(r => r.ruleType === 'min_direct_reports').length > 1) return 'Solo puede haber una regla de mínimo de reportes directos';
        return '';
    }

    async function handleSave() {
        const err = validate();
        if (err) { error = err; return; }
        error = '';
        saving = true;
        try {
            const assignmentsPayload = assignments.length > 0
                ? assignments.map(a => ({
                    employee_id: a.employeeId,
                    weight: a.weight,
                    target_value: a.targetValue,
                }))
                : undefined;
            const rulesPayload = rules.length > 0
                ? rules.map(r => ({
                    rule_type: r.ruleType,
                    ...(r.ruleType === 'department' ? { department_id: r.departmentId } : r.ruleType === 'role' ? { profile_id: r.profileId } : { min_direct_reports: r.minDirectReports }),
                    default_weight: r.defaultWeight,
                    default_target: r.defaultTarget,
                }))
                : undefined;
            if (goalId) {
                await updateGlobalGoal(goalId, {
                    name: name.trim(),
                    description: description.trim(),
                    unit,
                    direction,
                    goal_kind: goalKind,
                    weight,
                    target_value: targetValue,
                    baseline_value: direction === 'descendente' ? (baselineValue ?? undefined) : undefined,
                    assignments: assignmentsPayload,
                    rules: rulesPayload,
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
                    baseline_value: direction === 'descendente' ? (baselineValue ?? undefined) : undefined,
                    assignments: assignmentsPayload,
                    rules: rulesPayload,
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
    <div class="modal-box max-w-2xl max-h-[90vh] overflow-y-auto">
        <div class="flex items-center justify-between mb-4">
            <h3 class="font-bold text-lg">
                {goalId ? 'Editar' : 'Nueva'} meta {goalKind === 'qualitative' ? 'cualitativa' : 'cuantitativa'}
            </h3>
            <button class="btn btn-sm btn-ghost btn-circle" onclick={oncancel} disabled={saving}>
                <X class="w-4 h-4" />
            </button>
        </div>

        {#if error}
            <div class="alert alert-error text-sm mb-3" role="alert"><span>{error}</span></div>
        {/if}

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 mb-3">
            <div class="form-control sm:col-span-2">
                <label class="label" for="global-goal-name"><span class="label-text text-xs">Nombre de la meta</span></label>
                <input id="global-goal-name" type="text" class="input input-bordered input-sm w-full"
                    bind:value={name} placeholder="Nombre" required />
            </div>
            <div class="form-control sm:col-span-2">
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
                    class="w-full"
                />
            </div>
            <div class="form-control">
                <p class="label"><span class="label-text text-xs">Dirección del objetivo</span></p>
                <div class="flex gap-4 pt-1">
                    <label class="flex items-center gap-1.5 cursor-pointer">
                        <input type="radio" class="radio radio-primary radio-xs"
                            name="global-goal-dir" value="ascendente"
                            checked={direction === 'ascendente'}
                            onchange={() => { direction = 'ascendente'; baselineValue = null; }} />
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
            {#if direction === 'descendente'}
                <div class="form-control">
                    <label class="label" for="global-goal-baseline"><span class="label-text text-xs">Valor inicial</span></label>
                    <input id="global-goal-baseline" type="number" class="input input-bordered input-sm w-full"
                        bind:value={baselineValue} min={0} step={0.01} required placeholder="Valor inicial" />
                </div>
            {/if}
        </div>

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
                        class="w-64"
                        searchable
                        onSearch={(q) => search(q)}
                        loadMore={() => loadMore()}
                        allLoaded={!hasMoreEmployees()}
                        loadingMore={isLoadingMore()}
                    />
                    <button class="btn btn-outline btn-sm" onclick={addAssignment} type="button">
                        <Plus class="w-4 h-4" /> Agregar
                    </button>
                </div>
            </div>
            <div class="grid grid-cols-[minmax(0,1fr)_6rem_7rem_2rem] gap-2 mb-1 opacity-60">
                <span class="label-text text-xs">Nombre</span>
                <span class="label-text text-xs">Peso %</span>
                <span class="label-text text-xs">Objetivo</span>
                <span></span>
            </div>
            {#if assignments.length > 0}
                {#if loadingNames}
                    <div class="flex items-center justify-center gap-2 py-4">
                        <Loader2 class="w-4 h-4 animate-spin" />
                        <span class="text-xs text-base-content/60">Cargando nombres…</span>
                    </div>
                {:else}
                    <div class="space-y-2 max-h-64 overflow-y-auto overflow-x-hidden">
                        {#each assignments as a, i (a.employeeId)}
                            <div class="grid grid-cols-[minmax(0,1fr)_6rem_7rem_2rem] gap-2 items-center">
                                <span class="text-sm truncate">{a.employeeName}</span>
                                <div class="form-control">
                                    <input type="number" class="input input-bordered input-sm w-full"
                                        bind:value={a.weight} min={0} max={100} step={0.1}
                                        aria-label={`Ponderación de ${a.employeeName}`} placeholder="Peso %" />
                                </div>
                                <div class="form-control">
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
            {/if}
        </div>

        <!-- Reglas de asignación -->
        <div class="border border-base-300 rounded-lg p-3 mb-3">
            <div class="flex items-center justify-between mb-2">
                <span class="label-text text-xs font-semibold">Reglas de asignación (opcional)</span>
                <div class="flex items-center gap-2">
                    <button class="btn btn-outline btn-xs" onclick={addDepartmentRule} type="button">
                        <Plus class="w-3 h-3" /> Agregar departamento
                    </button>
                    <button class="btn btn-outline btn-xs" onclick={addRoleRule} type="button">
                        <Plus class="w-3 h-3" /> Agregar rol
                    </button>
                    <button class="btn btn-outline btn-xs" onclick={addMinReportsRule} type="button"
                        disabled={rules.some(r => r.ruleType === 'min_direct_reports')}>
                        <Plus class="w-3 h-3" /> Mínimo de reportes
                    </button>
                </div>
            </div>
            {#if rules.length > 0}
            <div class="grid grid-cols-[6rem_minmax(0,1fr)_6rem_7rem_2rem] gap-2 mb-1 opacity-60">
                <span class="label-text text-xs">Tipo</span>
                <span class="label-text text-xs">Detalle</span>
                <span class="label-text text-xs">Peso %</span>
                <span class="label-text text-xs">Objetivo</span>
                <span></span>
            </div>
            {/if}
            {#if rules.length > 0}
                <div class="space-y-2">
                    {#each rules as r, i (i)}
                        <div class="grid grid-cols-[6rem_minmax(0,1fr)_6rem_7rem_2rem] gap-2 items-center">
                            <span class="badge badge-sm badge-outline justify-self-start">{RULE_TYPE_LABELS[r.ruleType]}</span>
                            {#if r.ruleType === 'department'}
                                <CustomSelect
                                    options={departmentOptions}
                                    value={r.departmentId}
                                    onChange={(v) => { r.departmentId = v; }}
                                    ariaLabel="Departamento"
                                    initialLabel={departmentLabel(r.departmentId)}
                                    class="w-full"
                                    searchable
                                />
                            {:else if r.ruleType === 'role'}
                                <CustomSelect
                                    options={profileOptions}
                                    value={r.profileId}
                                    onChange={(v) => { r.profileId = v; }}
                                    ariaLabel="Perfil"
                                    initialLabel={profileLabel(r.profileId)}
                                    class="w-full"
                                />
                            {:else}
                                <div class="form-control">
                                    <input type="number" class="input input-bordered input-sm w-full"
                                        bind:value={r.minDirectReports} min={1} step={1}
                                        aria-label="Mínimo de reportes directos" placeholder="Min. reportes" />
                                </div>
                            {/if}
                            <div class="form-control">
                                <input type="number" class="input input-bordered input-sm w-full"
                                    bind:value={r.defaultWeight} min={0} max={100} step={0.1}
                                    aria-label="Ponderación por defecto" placeholder="Peso %" />
                            </div>
                            <div class="form-control">
                                <input type="number" class="input input-bordered input-sm w-full"
                                    bind:value={r.defaultTarget} min={0} step={0.01}
                                    aria-label="Objetivo por defecto" placeholder="Objetivo" />
                            </div>
                            <button class="btn btn-ghost btn-xs text-error" onclick={() => removeRule(i)} type="button">
                                <Trash2 class="w-4 h-4" />
                            </button>
                        </div>
                    {/each}
                </div>
            {/if}
        </div>

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
