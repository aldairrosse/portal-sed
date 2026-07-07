<script lang="ts">
	import { Pencil, Trash2, MessageSquare, MessageCircle, Check } from '@lucide/svelte';
	import type { Goal, GoalUnit, KPI, CyclePhase } from '$lib/types/goal';
	import { validateGoal, UNIT_OPTIONS } from './goalValidation';
	import { deltaIndicator, formatDelta } from '$lib/utils/scoring';
	import KpiBadge from './KpiBadge.svelte';
	import ProgressIndicator from './ProgressIndicator.svelte';
	import CustomSelect from '$lib/components/ui/CustomSelect.svelte';

	interface Props {
		goal: Goal;
		kpis: KPI[];
		onSaveGoal: (data: { id?: string; categoryId: string; name: string; description: string; unit: GoalUnit; weight: number; targetValue: number; direction: 'ascendente' | 'descendente'; baselineValue?: number; linkedKpiIds: string[] }) => void;
		onDeleteGoal: (goalId: string) => void;
		mode?: 'editor' | 'reader';
		onRequestChange?: (goal: Goal) => void;
		phase?: CyclePhase;
		canEditProgress?: boolean;
		canComment?: boolean;
		canDelete?: boolean;
		canClose?: boolean;
		allKpis: KPI[];
		editingGoalId: string | null;
		onEditingChange: (id: string | null) => void;
		onUpdateProgress?: (goalId: string, progress: number) => void;
		onOpenComments?: (goal: Goal) => void;
	}

	let {
		goal,
		kpis,
		onSaveGoal,
		onDeleteGoal,
		mode = 'editor',
		onRequestChange,
		phase = 'inicio-anio',
		canEditProgress = false,
		canComment = false,
		canDelete = true,
		canClose = false,
		allKpis,
		editingGoalId,
		onEditingChange,
		onUpdateProgress,
		onOpenComments
	}: Props = $props();

	// eslint-disable-next-line svelte/prefer-writable-derived
	let progressValue = $state(goal.progress ?? 0);

	$effect(() => {
		progressValue = goal.progress ?? 0;
	});

	const unitLabels: Record<string, string> = {
		porcentaje: '%',
		moneda: '$',
		numero: '#',
		binario: 'Sí/No'
	};

	// ─── Inline edit state ────────────────────────────────────────────────

	let isEditing = $state(false);
	let editName = $state('');
	let editDesc = $state('');
	let editUnit = $state<GoalUnit>('porcentaje');
	let editWeight = $state(0);
	let editTarget = $state(0);
	let editDirection = $state<'ascendente' | 'descendente'>('ascendente');
	let editBaseline = $state<number | undefined>(undefined);
	let editKpiIds = $state<string[]>([]);
	let editError = $state('');

	function handleStartEdit() {
		editName = goal.name;
		editDesc = goal.description;
		editUnit = goal.unit;
		editWeight = goal.weight;
		editTarget = goal.targetValue;
		editDirection = goal.direction;
		editBaseline = goal.baselineValue;
		editKpiIds = kpis.map(k => k.id);
		editError = '';
		isEditing = true;
		onEditingChange(goal.id);
	}

	function handleCancelEdit() {
		isEditing = false;
		editError = '';
		onEditingChange(null);
	}

	async function handleSaveEdit() {
		const err = validateGoal({ name: editName, description: editDesc, weight: editWeight, targetValue: editTarget, direction: editDirection, baselineValue: editDirection === 'descendente' ? editBaseline : undefined, categoryId: goal.categoryId, goalId: goal.id });
		if (err) { editError = err; return; }
		try {
			await onSaveGoal({ id: goal.id, categoryId: goal.categoryId, name: editName.trim(), description: editDesc.trim(), unit: editUnit, weight: editWeight, targetValue: editTarget, direction: editDirection, baselineValue: editDirection === 'descendente' ? editBaseline : undefined, linkedKpiIds: editKpiIds });
			isEditing = false;
			editError = '';
			onEditingChange(null);
		} catch (e) {
			editError = e instanceof Error ? e.message : 'Error al guardar meta';
		}
	}

	function handleToggleEditKpi(kpiId: string) {
		editKpiIds = editKpiIds.includes(kpiId) ? editKpiIds.filter(id => id !== kpiId) : [...editKpiIds, kpiId];
	}

	function handleProgressInput(e: Event) {
		const val = parseFloat((e.target as HTMLInputElement).value);
		if (!isNaN(val)) {
			progressValue = val;
			onUpdateProgress?.(goal.id, val);
		}
	}

	// ─── Delta indicator ───────────────────────────────────────────────────

	let delta = $derived.by(() => {
		if (goal.progress === undefined || goal.progress === null) return null;
		return deltaIndicator(goal.progress, goal.targetValue, goal.baselineValue, goal.direction);
	});

	let deltaBadgeClass = $derived.by(() => {
		if (!delta || delta.type === 'neutral') return '';
		if (delta.type === 'positive') return 'badge-success';
		return goal.direction === 'descendente' ? 'badge-error' : 'badge-warning';
	});

	let deltaLabel = $derived.by(() => {
		if (!delta || delta.type === 'neutral') return '';
		const formatted = formatDelta(delta.value, delta.type);
		if (delta.type === 'positive' && goal.direction === 'descendente') return formatted + '%';
		return formatted;
	});
</script>

<tr>
	{#if isEditing}
		<td colspan={phase !== 'fin-anio' ? 5 : 4} class="p-0">
			<div class="border border-base-300 rounded-lg p-4 m-2 bg-base-200/50">
				{#if editError}<div class="alert alert-error text-sm mb-3" role="alert"><span>{editError}</span></div>{/if}
				<div class="grid grid-cols-1 md:grid-cols-2 gap-3 mb-3">
					<div class="form-control">
						<label class="label" for="edit-goal-name-{goal.id}"><span class="label-text text-xs">Nombre</span></label>
						<input id="edit-goal-name-{goal.id}" type="text" class="input input-bordered input-sm w-full" bind:value={editName} placeholder="Nombre" required />
					</div>
					<div class="form-control">
						<label class="label" for="edit-goal-desc-{goal.id}"><span class="label-text text-xs">Descripción</span></label>
						<textarea id="edit-goal-desc-{goal.id}" class="textarea textarea-bordered textarea-sm w-full" rows={1} bind:value={editDesc} placeholder="Descripción" required></textarea>
					</div>
					<div class="form-control">
						<label class="label" for="edit-goal-unit-{goal.id}"><span class="label-text text-xs">Unidad</span></label>
						<CustomSelect
							options={UNIT_OPTIONS}
							value={editUnit}
							onChange={(v) => { editUnit = v as GoalUnit; }}
							ariaLabel="Unidad"
						/>
					</div>
					<div class="form-control">
						<label class="label"><span class="label-text text-xs">Dirección</span></label>
						<div class="flex gap-4 pt-1">
							<label class="flex items-center gap-1.5 cursor-pointer">
								<input type="radio" class="radio radio-primary radio-xs" name="edit-goal-dir-{goal.id}" value="ascendente" checked={editDirection === 'ascendente'} onchange={() => { editDirection = 'ascendente'; if (editBaseline !== undefined) editBaseline = undefined; }} />
								<span class="text-xs">Ascendente (↑)</span>
							</label>
							<label class="flex items-center gap-1.5 cursor-pointer">
								<input type="radio" class="radio radio-primary radio-xs" name="edit-goal-dir-{goal.id}" value="descendente" checked={editDirection === 'descendente'} onchange={() => editDirection = 'descendente'} />
								<span class="text-xs">Descendente (↓)</span>
							</label>
						</div>
					</div>
					<div class="form-control">
						<label class="label" for="edit-goal-weight-{goal.id}"><span class="label-text text-xs">Peso (%)</span></label>
						<input id="edit-goal-weight-{goal.id}" type="number" class="input input-bordered input-sm w-full" bind:value={editWeight} min={0} max={100} step={0.1} required />
					</div>
					<div class="form-control">
						<label class="label" for="edit-goal-target-{goal.id}"><span class="label-text text-xs">Valor objetivo</span></label>
						<input id="edit-goal-target-{goal.id}" type="number" class="input input-bordered input-sm w-full" bind:value={editTarget} min={0} step={0.01} required />
					</div>
					{#if editDirection === 'descendente'}
						<div class="form-control">
							<label class="label" for="edit-goal-baseline-{goal.id}"><span class="label-text text-xs">Punto de partida</span></label>
							<input id="edit-goal-baseline-{goal.id}" type="number" class="input input-bordered input-sm w-full" bind:value={editBaseline} min={0} step={0.01} required />
						</div>
					{/if}
				</div>
				{#if allKpis && allKpis.length > 0}
					<div class="form-control mb-3">
						<label class="label" for="edit-goal-kpi-{goal.id}"><span class="label-text text-xs">Indicadores clave (KPI)</span></label>
						<div id="edit-goal-kpi-{goal.id}" class="flex flex-wrap gap-2">
							{#each allKpis as kpi (kpi.id)}
								<label class="flex items-center gap-1.5 cursor-pointer px-2 py-1 rounded border border-base-300 hover:bg-base-200/50 text-xs">
									<input type="checkbox" class="checkbox checkbox-xs checkbox-primary" checked={editKpiIds.includes(kpi.id)} onchange={() => handleToggleEditKpi(kpi.id)} />
									<span>{kpi.name}</span>
								</label>
							{/each}
						</div>
					</div>
				{/if}
				<div class="flex justify-end gap-2">
					<button class="btn btn-ghost btn-sm" onclick={handleCancelEdit}>Cancelar</button>
					<button class="btn btn-primary btn-sm" onclick={handleSaveEdit}><Check class="w-4 h-4" /> Guardar meta</button>
				</div>
			</div>
		</td>
	{:else}
		<td class="font-medium text-sm">{goal.name}</td>
		<td class="text-sm text-base-content/70">
			{goal.targetValue}{unitLabels[goal.unit] ?? goal.unit}
		</td>
		<td class="text-sm text-base-content/70">{goal.weight}%</td>
		<td>
			{#if phase === 'fin-anio'}
				<div class="flex items-center gap-2">
					{#if canClose}
						<input
							type="number"
							class="input input-bordered input-xs w-20"
							value={progressValue}
							min="0"
							max={goal.unit === 'porcentaje' ? 100 : goal.targetValue}
							oninput={handleProgressInput}
							aria-label="Avance final de {goal.name}"
						/>
					{/if}
					<ProgressIndicator value={progressValue} max={100} label="Avance final" />
					{#if delta && delta.type !== 'neutral'}
						<span class="badge badge-sm {deltaBadgeClass}">{deltaLabel}</span>
					{/if}
				</div>
			{:else if phase === 'medio-anio'}
				<div class="flex items-center gap-2">
					<input
						type="number"
						class="input input-bordered input-xs w-20"
						value={progressValue}
						min="0"
						max={goal.unit === 'porcentaje' ? 100 : undefined}
						oninput={handleProgressInput}
						aria-label="Avance de {goal.name}"
						readonly={!canEditProgress}
					/>
					{#if goal.unit === 'porcentaje'}
						<ProgressIndicator
							value={progressValue}
							max={100}
							color={progressValue >= goal.targetValue ? 'success' : 'error'}
						/>
					{:else}
						<ProgressIndicator value={progressValue} max={goal.targetValue} />
					{/if}
					{#if delta && delta.type !== 'neutral'}
						<span class="badge badge-sm {deltaBadgeClass}">{deltaLabel}</span>
					{/if}
				</div>
			{:else if kpis.length > 0}
				<div class="flex flex-wrap gap-1">
					{#each kpis as kpi (kpi.id)}
						<KpiBadge {kpi} />
					{/each}
				</div>
			{:else}
				<span class="text-xs text-base-content/30 italic">Sin KPI</span>
			{/if}
		</td>
		{#if phase !== 'fin-anio'}
			<td class="text-right">
				<div class="flex items-center justify-end gap-1">
					{#if phase === 'medio-anio'}
						{#if canComment}
							<button
								class="btn btn-ghost btn-square btn-xs relative"
								title="Comentarios"
								onclick={() => onOpenComments?.(goal)}
								aria-label="Comentarios de {goal.name}"
							>
								<MessageCircle class="w-3.5 h-3.5" />
								{#if (goal.comments?.length ?? 0) > 0}
									<span class="badge badge-xs badge-primary absolute -top-1.5 -right-1.5">{goal.comments?.length}</span>
								{/if}
							</button>
						{/if}
					{:else if mode === 'editor'}
						<!-- ponytail: show comment button in inicio-anio when there are existing comments from the boss -->
						{#if (goal.comments?.length ?? 0) > 0}
							<button
								class="btn btn-ghost btn-square btn-xs relative"
								title="Comentarios"
								onclick={() => onOpenComments?.(goal)}
								aria-label="Comentarios de {goal.name}"
							>
								<MessageCircle class="w-3.5 h-3.5" />
								<span class="badge badge-xs badge-primary absolute -top-1.5 -right-1.5">{goal.comments?.length}</span>
							</button>
						{/if}
						<button
							class="btn btn-ghost btn-square btn-xs"
							title="Editar"
							onclick={handleStartEdit}
							disabled={editingGoalId !== null && editingGoalId !== goal.id}
							aria-label="Editar {goal.name}"
						>
							<Pencil class="w-3.5 h-3.5" />
						</button>
						{#if canDelete}
							<button
								class="btn btn-ghost btn-square btn-xs text-error"
								title="Eliminar"
								onclick={() => onDeleteGoal(goal.id)}
								disabled={editingGoalId !== null}
								aria-label="Eliminar {goal.name}"
							>
								<Trash2 class="w-3.5 h-3.5" />
							</button>
						{/if}
					{:else if onRequestChange}
						<!-- ponytail: show comment button for boss in reader mode when there are existing comments -->
						{#if (goal.comments?.length ?? 0) > 0}
							<button
								class="btn btn-ghost btn-square btn-xs relative"
								title="Comentarios"
								onclick={() => onOpenComments?.(goal)}
								aria-label="Comentarios de {goal.name}"
							>
								<MessageCircle class="w-3.5 h-3.5" />
								<span class="badge badge-xs badge-primary absolute -top-1.5 -right-1.5">{goal.comments?.length}</span>
							</button>
						{/if}
						<button
							class="btn btn-ghost btn-xs text-warning"
							title="Solicitar cambio"
							onclick={() => onRequestChange(goal)}
							aria-label="Solicitar cambio en {goal.name}"
						>
							<MessageSquare class="w-3.5 h-3.5" />
							Solicitar cambio
						</button>
					{/if}
				</div>
			</td>
		{/if}
	{/if}
</tr>
