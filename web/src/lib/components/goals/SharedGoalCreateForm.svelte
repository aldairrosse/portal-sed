<script lang="ts">
    import { X, Loader2 } from '@lucide/svelte';
    import CustomSelect from '$lib/components/ui/CustomSelect.svelte';
    import { getSession } from '$lib/api/session.svelte';
    import { loadTeam, getTeamMembers, type TeamMember } from '$lib/stores/teamStore.svelte';
    import { createSharedGoal, updateSharedGoal, type CreateSharedGoalRequest, type CreateMemberRequest, type UpdateSharedGoalRequest } from '$lib/api/sharedGoals';

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
            members?: { employee_id: string; weight: number; target_value: number; baseline_value?: number }[];
        };
        oncancel: () => void;
        onsaved: () => void;
    }

    let { open, goalKind, goalId, initial, oncancel, onsaved }: Props = $props();

    const UNIT_OPTIONS = [
        { value: 'porcentaje', label: 'Porcentaje' },
        { value: 'moneda', label: 'Moneda' },
        { value: 'numero', label: 'Número' },
        { value: 'binario', label: 'Binario (Sí/No)' },
    ];

    let name = $state('');
    let description = $state('');
    let unit = $state('porcentaje');
    let direction = $state<'ascendente' | 'descendente'>('ascendente');
    let weight = $state(1);
    let targetValue = $state(0);
    let baselineValue = $state<number | null>(null);
    let groupName = $state('Grupo directo');
    let groupDescription = $state('');
    let members = $state<TeamMember[]>([]);
    let memberForm = $state<Record<string, { employee_id: string; weight: number; target_value: number }>>({});
    let saving = $state(false);
    let error = $state('');

    function reset() {
        name = '';
        description = '';
        unit = 'porcentaje';
        direction = 'ascendente';
        weight = 1;
        targetValue = 0;
        baselineValue = null;
        groupName = 'Grupo directo';
        groupDescription = '';
        memberForm = {};
        error = '';
    }

    $effect(() => {
        if (!open) return;
        if (initial) {
            name = initial.name;
            description = initial.description;
            unit = initial.unit;
            direction = initial.direction;
            weight = initial.weight;
            targetValue = initial.target_value;
            baselineValue = initial.baseline_value ?? null;
            error = '';
            if (initial.members && initial.members.length > 0) {
                const next: Record<string, { employee_id: string; weight: number; target_value: number }> = {};
                for (const m of initial.members) {
                    next[m.employee_id] = { employee_id: m.employee_id, weight: m.weight, target_value: m.target_value || initial.target_value || 0 };
                }
                memberForm = next;
            } else if (initial.members) {
                memberForm = {};
            }
        } else {
            reset();
        }
        const user = getSession().user;
        if (user?.employeeId) {
            loadTeam(user.employeeId).then(() => {
                members = getTeamMembers();
            });
        }
    });

    function toggleMember(member: TeamMember) {
        const next = { ...memberForm };
        if (next[member.id]) {
            delete next[member.id];
        } else {
            next[member.id] = { employee_id: member.id, weight: weight || 1, target_value: targetValue || 0 };
        }
        memberForm = next;
    }

    function setMemberWeight(memberId: string, value: number) {
        memberForm = { ...memberForm, [memberId]: { ...memberForm[memberId], weight: value || (weight || 1) } };
    }

    function setMemberTarget(memberId: string, value: number) {
        memberForm = { ...memberForm, [memberId]: { ...memberForm[memberId], target_value: value || 0 } };
    }

    function validate(): string {
        if (!name.trim()) return 'El nombre de la meta es obligatorio';
        if (weight < 0.01 || weight > 100) return 'Peso debe ser >0 y ≤100';
        if (targetValue <= 0 && direction !== 'descendente' && unit !== 'binario') return 'El valor objetivo debe ser mayor a 0';
        if (direction === 'descendente') {
            if (baselineValue === null || baselineValue <= targetValue) return 'Para objetivos descendentes, el valor inicial debe ser mayor al objetivo';
        }
        if (Object.keys(memberForm).length === 0) return 'Selecciona al menos un miembro del grupo';
        for (const m of Object.values(memberForm)) {
            if (m.weight < 0.01 || m.weight > 100) return 'Peso debe ser >0 y ≤100 para cada miembro';
            if (m.target_value !== undefined && m.target_value !== null && m.target_value <= 0 && direction !== 'descendente' && unit !== 'binario') return 'El valor objetivo del miembro debe ser mayor a 0';
        }
        return '';
    }

    async function handleSave() {
        if (saving) return;
        const err = validate();
        if (err) { error = err; return; }
        error = '';
        saving = true;
        try {
            if (goalId) {
                const membersPayload: CreateMemberRequest[] = Object.values(memberForm).map((m) => ({
                    employee_id: m.employee_id,
                    weight: m.weight,
                    target_value: m.target_value || targetValue || 0,
                }));
                await updateSharedGoal(goalId, {
                    name: name.trim(),
                    description: description.trim(),
                    unit,
                    direction,
                    goal_kind: goalKind,
                    weight,
                    target_value: targetValue,
                    baseline_value: direction === 'descendente' ? (baselineValue ?? undefined) : undefined,
                    members: membersPayload,
                } as UpdateSharedGoalRequest);
            } else {
                const membersPayload: CreateMemberRequest[] = Object.values(memberForm).map((m) => ({
                    employee_id: m.employee_id,
                    weight: m.weight,
                    target_value: m.target_value || targetValue || 0,
                }));
                const request: CreateSharedGoalRequest = {
                    name: name.trim(),
                    description: description.trim(),
                    unit,
                    direction,
                    goal_kind: goalKind,
                    weight,
                    target_value: targetValue,
                    baseline_value: direction === 'descendente' ? (baselineValue ?? undefined) : undefined,
                    group_name: groupName.trim() || 'Grupo',
                    group_description: groupDescription.trim(),
                    members: membersPayload,
                };
                await createSharedGoal(request);
            }
            onsaved();
        } catch (e) {
            error = e instanceof Error ? e.message : 'Error al guardar la meta compartida';
        } finally {
            saving = false;
        }
    }
</script>

<dialog class="modal" open={open} aria-modal="true">
    <div class="modal-box max-w-2xl max-h-[90vh] overflow-y-auto">
        <div class="flex items-center justify-between mb-4">
            <h3 class="font-bold text-lg">
                {goalId
                    ? `Editar meta ${goalKind === 'qualitative' ? 'cualitativa' : 'cuantitativa'}`
                    : `Nueva meta ${goalKind === 'qualitative' ? 'cualitativa' : 'cuantitativa'}`}
            </h3>
            <button class="btn btn-sm btn-ghost btn-circle" onclick={oncancel} disabled={saving}>
                <X class="w-4 h-4" />
            </button>
        </div>

        {#if error}
            <div class="alert alert-error text-sm mb-3" role="alert"><span>{error}</span></div>
        {/if}

        <div class="grid grid-cols-1 md:grid-cols-2 gap-3 mb-3">
            <div class="form-control">
                <label class="label" for="shared-name"><span class="label-text text-xs">Nombre de la meta</span></label>
                <input id="shared-name" type="text" class="input input-bordered input-sm w-full"
                    bind:value={name} placeholder="Nombre" required />
            </div>
            <div class="form-control">
                <label class="label" for="shared-desc"><span class="label-text text-xs">Descripción</span></label>
                <textarea id="shared-desc" class="textarea textarea-bordered textarea-sm w-full"
                    rows={1} bind:value={description} placeholder="Descripción"></textarea>
            </div>
            <div class="form-control">
                <label class="label" for="shared-unit"><span class="label-text text-xs">Unidad de medida</span></label>
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
                            name="shared-dir" value="ascendente"
                            checked={direction === 'ascendente'}
                            onchange={() => { direction = 'ascendente'; baselineValue = null; }} />
                        <span class="text-xs">Ascendente (↑)</span>
                    </label>
                    <label class="flex items-center gap-1.5 cursor-pointer">
                        <input type="radio" class="radio radio-primary radio-xs"
                            name="shared-dir" value="descendente"
                            checked={direction === 'descendente'}
                            onchange={() => direction = 'descendente'} />
                        <span class="text-xs">Descendente (↓)</span>
                    </label>
                </div>
            </div>
            <div class="form-control">
                <label class="label" for="shared-weight"><span class="label-text text-xs">Ponderación (%)</span></label>
                <input id="shared-weight" type="number" class="input input-bordered input-sm w-full"
                    bind:value={weight} min={0.01} max={100} step={0.1} required />
            </div>
            <div class="form-control">
                <label class="label" for="shared-target"><span class="label-text text-xs">Valor objetivo</span></label>
                <input id="shared-target" type="number" class="input input-bordered input-sm w-full"
                    bind:value={targetValue} min={0} step={0.01} required />
            </div>
            {#if direction === 'descendente'}
                <div class="form-control">
                    <label class="label" for="shared-goal-baseline"><span class="label-text text-xs">Valor inicial</span></label>
                    <input id="shared-goal-baseline" type="number" class="input input-bordered input-sm w-full"
                        bind:value={baselineValue} min={0} step={0.01} required placeholder="Valor inicial" />
                </div>
            {/if}
        </div>

        <div class="form-control">
            <span class="label"><span class="label-text text-xs">Miembros del grupo</span></span>
            {#if members.length > 0}
                <div class="grid grid-cols-[1.25rem_minmax(0,1fr)_6rem_7rem] gap-2 mb-1 opacity-60">
                    <span></span>
                    <span class="label-text text-xs">Nombre</span>
                    <span class="label-text text-xs">Peso %</span>
                    <span class="label-text text-xs">Objetivo</span>
                </div>
            {/if}
            {#if members.length === 0}
                <p class="text-sm text-base-content/60">No se encontró tu equipo.</p>
            {:else}
                <div class="space-y-2 max-h-64 overflow-y-auto pr-1 overflow-x-hidden">
                    {#each members as member (member.id)}
                        <div class="grid grid-cols-[1.25rem_minmax(0,1fr)_6rem_7rem] gap-2 items-center">
                            <input type="checkbox" class="checkbox checkbox-xs checkbox-primary"
                                checked={memberForm[member.id] !== undefined}
                                onchange={() => toggleMember(member)} />
                            <span class="text-xs truncate">{member.firstName} {member.lastName}</span>
                            {#if memberForm[member.id]}
                                <input type="number" class="input input-bordered input-sm w-full"
                                    placeholder="Peso %" min={0.01} max={100} step={0.1}
                                    value={memberForm[member.id].weight}
                                    oninput={(e) => setMemberWeight(member.id, Number(e.currentTarget.value))} />
                                <input type="number" class="input input-bordered input-sm w-full"
                                    placeholder="Objetivo" min={0} step={0.01}
                                    value={memberForm[member.id].target_value}
                                    oninput={(e) => setMemberTarget(member.id, Number(e.currentTarget.value))} />
                            {/if}
                        </div>
                    {/each}
                </div>
            {/if}
        </div>

        <div class="modal-action">
            <button class="btn btn-ghost" onclick={oncancel} disabled={saving}>Cancelar</button>
            <button class="btn btn-primary" onclick={handleSave} disabled={saving}>
                {#if saving}
                    <Loader2 class="w-4 h-4 animate-spin" />
                {/if}
                Guardar
            </button>
        </div>
    </div>
</dialog>
