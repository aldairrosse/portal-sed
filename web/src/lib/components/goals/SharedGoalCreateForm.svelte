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
        };
        oncancel: () => void;
        onsaved: () => void;
    }

    let { open, goalKind, goalId, initial, oncancel, onsaved }: Props = $props();

    const UNIT_OPTIONS = [
        { value: 'porcentaje', label: 'Porcentaje' },
        { value: 'moneda', label: 'Moneda' },
        { value: 'numero', label: 'Número' },
    ];

    let name = $state('');
    let description = $state('');
    let unit = $state('porcentaje');
    let direction = $state<'ascendente' | 'descendente'>('ascendente');
    let weight = $state(0);
    let targetValue = $state(0);
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
        weight = 0;
        targetValue = 0;
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
            error = '';
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
            next[member.id] = { employee_id: member.id, weight: 0, target_value: 0 };
        }
        memberForm = next;
    }

    function setMemberWeight(memberId: string, value: number) {
        memberForm = { ...memberForm, [memberId]: { ...memberForm[memberId], weight: value || 0 } };
    }

    function setMemberTarget(memberId: string, value: number) {
        memberForm = { ...memberForm, [memberId]: { ...memberForm[memberId], target_value: value || 0 } };
    }

    function validate(): string {
        if (!name.trim()) return 'El nombre de la meta es obligatorio';
        if (!goalId && !groupName.trim()) return 'El nombre del grupo es obligatorio';
        if (weight < 0 || weight > 100) return 'La ponderación debe estar entre 0 y 100';
        if (targetValue <= 0) return 'El valor objetivo debe ser mayor a 0';
        if (!goalId && Object.keys(memberForm).length === 0) return 'Selecciona al menos un miembro del grupo';
        return '';
    }

    async function handleSave() {
        const err = validate();
        if (err) { error = err; return; }
        error = '';
        saving = true;
        try {
            if (goalId) {
                // ponytail: UpdateSharedGoalRequest type lacks weight; backend accepts it
                await updateSharedGoal(goalId, {
                    name: name.trim(),
                    description: description.trim(),
                    unit,
                    direction,
                    goal_kind: goalKind,
                    weight,
                    target_value: targetValue,
                } as UpdateSharedGoalRequest);
            } else {
                const membersPayload: CreateMemberRequest[] = Object.values(memberForm).map((m) => ({
                    employee_id: m.employee_id,
                    weight: m.weight,
                    target_value: m.target_value,
                }));
                const request: CreateSharedGoalRequest = {
                    name: name.trim(),
                    description: description.trim(),
                    unit,
                    direction,
                    goal_kind: goalKind,
                    weight,
                    target_value: targetValue,
                    group_name: groupName.trim(),
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
    <div class="modal-box max-w-2xl">
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
                <label class="label" for="shared-group-name"><span class="label-text text-xs">Nombre del grupo</span></label>
                <input id="shared-group-name" type="text" class="input input-bordered input-sm w-full"
                    bind:value={groupName} placeholder="Grupo directo" required />
            </div>
            <div class="form-control">
                <label class="label" for="shared-desc"><span class="label-text text-xs">Descripción</span></label>
                <textarea id="shared-desc" class="textarea textarea-bordered textarea-sm w-full"
                    rows={1} bind:value={description} placeholder="Descripción"></textarea>
            </div>
            <div class="form-control">
                <label class="label" for="shared-group-desc"><span class="label-text text-xs">Descripción del grupo</span></label>
                <textarea id="shared-group-desc" class="textarea textarea-bordered textarea-sm w-full"
                    rows={1} bind:value={groupDescription} placeholder="Descripción del grupo"></textarea>
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
                            onchange={() => direction = 'ascendente'} />
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
                    bind:value={weight} min={0} max={100} step={0.1} required />
            </div>
            <div class="form-control">
                <label class="label" for="shared-target"><span class="label-text text-xs">Valor objetivo</span></label>
                <input id="shared-target" type="number" class="input input-bordered input-sm w-full"
                    bind:value={targetValue} min={0} step={0.01} required />
            </div>
        </div>

        {#if !goalId}
        <div class="form-control">
            <span class="label"><span class="label-text text-xs">Miembros del grupo</span></span>
            {#if members.length > 0}
                <div class="flex items-center gap-2 px-1 mb-1">
                    <span class="w-4"></span>
                    <span class="label-text text-xs flex-1">Nombre</span>
                    <span class="label-text text-xs w-20">Peso %</span>
                    <span class="label-text text-xs w-24">Objetivo</span>
                </div>
            {/if}
            {#if members.length === 0}
                <p class="text-sm text-base-content/60">No se encontró tu equipo.</p>
            {:else}
                <div class="space-y-2 max-h-64 overflow-y-auto pr-1">
                    {#each members as member (member.id)}
                        <div class="flex items-center gap-2 p-2 border border-base-300 rounded-lg">
                            <label class="flex items-center gap-1.5 cursor-pointer flex-1 min-w-0">
                                <input type="checkbox" class="checkbox checkbox-xs checkbox-primary"
                                    checked={memberForm[member.id] !== undefined}
                                    onchange={() => toggleMember(member)} />
                                <span class="text-xs truncate">{member.firstName} {member.lastName}</span>
                            </label>
                            {#if memberForm[member.id]}
                                <input type="number" class="input input-bordered input-xs w-20"
                                    placeholder="Peso %" min={0} max={100} step={0.1}
                                    value={memberForm[member.id].weight}
                                    oninput={(e) => setMemberWeight(member.id, Number(e.currentTarget.value))} />
                                <input type="number" class="input input-bordered input-xs w-24"
                                    placeholder="Objetivo" min={0} step={0.01}
                                    value={memberForm[member.id].target_value}
                                    oninput={(e) => setMemberTarget(member.id, Number(e.currentTarget.value))} />
                            {/if}
                        </div>
                    {/each}
                </div>
            {/if}
        </div>
        {/if}

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
