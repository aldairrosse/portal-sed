<script lang="ts">
	import type { Goal, KPI } from '$lib/types/goal';
	import type { GoalClosure } from '$lib/types/evaluation-result';
	import { progressPercent } from '$lib/utils/scoring';
	import ProgressIndicator from '$lib/components/goals/ProgressIndicator.svelte';
	import KpiBadge from '$lib/components/goals/KpiBadge.svelte';
	import PageSkeleton from '$lib/components/ui/PageSkeleton.svelte';
	import ErrorState from '$lib/components/ui/ErrorState.svelte';
	import { ArrowDown, ArrowUp, Pencil } from '@lucide/svelte';
	import { tick } from 'svelte';

	interface Props {
		goal: Goal;
		kpis: KPI[];
		closure?: GoalClosure;
		mode: 'self' | 'rh' | 'manager';
		canEdit?: boolean;
		showSelfAssessment?: boolean;
		employeeId?: string;
		phase?: 'avance' | 'cierre';
		onSaveClosure?: (goalId: string, finalProgress: number, selfAssessment: string) => void;
		onRhAssessGoal?: (goalId: string, rhAssessment: string) => void;
		onManagerComment?: (goalId: string, comment: string) => void;
		loading?: boolean;
		error?: string | null;
	}

	let {
		goal,
		kpis,
		closure,
		mode,
		canEdit = false,
		showSelfAssessment = true,
		onSaveClosure,
		onRhAssessGoal,
		onManagerComment,
		loading = false,
		error = null,
		phase = 'cierre',
	}: Props = $props();

	const isAvance = $derived(phase === 'avance');
	const saveLabel = $derived(isAvance ? 'Guardar avance' : 'Guardar cierre');

	const progressId = $derived(`progress-${goal.id}`);
	const selfAssessmentId = $derived(`self-assessment-${goal.id}`);
	const rhAssessmentId = $derived(`rh-assessment-${goal.id}`);
	const managerCommentId = $derived(`manager-comment-${goal.id}`);

	// svelte-ignore state_referenced_locally (intentional: form state seeded once from prop)
	let progressValue = $state(closure?.finalProgress ?? goal.progress ?? 0);
	// svelte-ignore state_referenced_locally (intentional: form state seeded once from prop)
	let selfAssessmentValue = $state(closure?.selfAssessment ?? '');
	// svelte-ignore state_referenced_locally (intentional: form state seeded once from prop)
	let rhAssessmentValue = $state(closure?.rhAssessment ?? '');
	// svelte-ignore state_referenced_locally (intentional: form state seeded once from prop)
	let managerCommentValue = $state(closure?.managerComment ?? '');
	// Hydrate from saved direct snapshot after load/save (not always 5).
	$effect(() => {
		const v: number | undefined = closure?.finalProgress;
		if (v !== undefined) progressValue = v;
	});

	function formatFinal(v: number): string {
		switch (goal.unit) {
			case 'porcentaje':
				return `${v}%`;
			case 'moneda':
				return `$${v.toLocaleString('es-MX')}`;
			case 'binario':
				return v >= 1 ? 'Sí' : 'No';
			default:
				return `${v}`;
		}
	}

	const progressMax = $derived(
		goal.unit === 'porcentaje' ? 100 : goal.unit === 'binario' ? 1 : goal.targetValue
	);
	// F1: bar shows 0–100 percent via shared scoring (descendente uses baseline, ascendente current/target).
	const progressPct = $derived(
		progressPercent(progressValue, goal.targetValue, goal.baselineValue, goal.direction)
	);
	const progressStep = $derived(goal.unit === 'binario' ? '1' : 'any');

	function handleSaveClosure() {
		onSaveClosure?.(goal.id, progressValue, selfAssessmentValue);
		editingSelf = false;
	}

	function handleRhAssessment() {
		onRhAssessGoal?.(goal.id, rhAssessmentValue);
		editingRh = false;
	}

	function handleManagerComment() {
		onManagerComment?.(goal.id, managerCommentValue);
		editingManager = false;
	}

	function handleProgressInput(e: Event) {
		progressValue = parseFloat((e.target as HTMLInputElement).value) || 0;
	}

	function formatAuthorDate(author?: string | null, date?: string | null): string {
		const name = author?.trim() ?? '';
		if (!name) return '';
		if (!date) return name;
		const d = new Date(date);
		if (Number.isNaN(d.getTime())) return name;
		return `${name} · ${d.toLocaleDateString('es-MX')} ${d.toLocaleTimeString('es-MX', { hour: '2-digit', minute: '2-digit' })}`;
	}

	function hasRealComment(text?: string | null): boolean {
		const t = text?.trim() ?? '';
		if (!t) return false;
		return !/^Sin (autoevaluación|evaluación|comentario)/i.test(t);
	}

	function authorLine(author?: string | null, date?: string | null, text?: string | null): string {
		if (!hasRealComment(text)) return '';
		return formatAuthorDate(author, date);
	}

	let editingSelf = $state(false);
	let editingRh = $state(false);
	let editingManager = $state(false);

	const selfHasComment = $derived(hasRealComment(closure?.selfAssessment));
	const rhHasComment = $derived(hasRealComment(closure?.rhAssessment));
	const managerHasComment = $derived(hasRealComment(closure?.managerComment));
	const selfReadMode = $derived(selfHasComment && !editingSelf);
	const rhReadMode = $derived(rhHasComment && !editingRh);
	const managerReadMode = $derived(managerHasComment && !editingManager);

	const TEXTAREA_MAX_HEIGHT = 128;

	function autoGrow(node: HTMLTextAreaElement) {
		node.style.height = 'auto';
		const capped = Math.min(node.scrollHeight, TEXTAREA_MAX_HEIGHT);
		node.style.height = `${capped}px`;
		node.style.overflowY = node.scrollHeight > TEXTAREA_MAX_HEIGHT ? 'auto' : 'hidden';
	}
</script>

{#if loading}
	<PageSkeleton variant="card" rows={1} />
{:else if error}
	<ErrorState message={error} />
{:else}
<div class="card bg-base-100 border border-base-300">
	<div class="card-body px-0 py-4">
		<!-- Goal header -->
		<div class="flex items-start justify-between gap-2 mb-3">
			<div class="flex-1 min-w-0">
				<div class="flex flex-wrap items-center gap-1">
					<h4 class="text-sm font-semibold text-base-content">{goal.name}</h4>
					<span class="badge badge-ghost badge-xs inline-flex items-center gap-1">{#if goal.direction === 'descendente'}<ArrowDown size={12} />{:else}<ArrowUp size={12} />{/if}Objetivo: {formatFinal(goal.targetValue)}</span>
					{#if goal.direction === 'descendente' && goal.baselineValue != null}<span class="badge badge-ghost badge-xs">Inicio: {formatFinal(goal.baselineValue)}</span>{/if}
					<span class="badge badge-ghost badge-xs">Peso: {goal.weight}%</span>
				</div>
			</div>
			{#if kpis.length > 0}
				<div class="flex flex-wrap gap-1 shrink-0">
					{#each kpis as kpi (kpi.id)}
						<KpiBadge {kpi} />
					{/each}
				</div>
			{/if}
		</div>

		<!-- Progress -->
		<div class="mb-3">
			{#if mode === 'self' && canEdit && !selfReadMode}
				<label class="label px-0 py-1" for={progressId}>
					<span class="label-text text-xs">{isAvance ? 'Avance actual' : 'Avance final'}</span>
				</label>
				<div class="flex items-center gap-2">
					<input
						id={progressId}
						type="number"
						class="input input-bordered input-sm w-24"
						value={progressValue}
						min="0"
						max={progressMax}
						step={progressStep}
						oninput={handleProgressInput}
						aria-label="{isAvance ? 'Avance actual' : 'Avance final'} de {goal.name}"
					/>
					<ProgressIndicator
						value={progressPct}
						max={100}
					/>
				</div>
			{:else}
				<div class="flex items-center gap-2">
					<span class="text-xs text-base-content/60">{isAvance ? 'Avance actual:' : 'Avance final:'}</span>
					<span class="text-sm font-semibold" aria-live="polite">{formatFinal(progressValue)}</span>
					<ProgressIndicator
						value={progressPct}
						max={100}
					/>
				</div>
			{/if}
		</div>

		<!-- Self assessment (editable in self mode, read-only in rh/manager mode) -->
		{#if showSelfAssessment}
			<div class="mb-2">
				<label class="label px-0 py-1" for={selfAssessmentId}>
					<span class="label-text text-xs">Autoevaluación</span>
				</label>
		{#if mode === 'self' && canEdit && !selfReadMode}
					<textarea
						id={selfAssessmentId}
						class="textarea textarea-bordered textarea-xs w-full resize-y overflow-hidden max-h-32 !min-h-[2.25rem] h-auto !pt-2 !pb-1 focus:ring-2 focus:ring-primary/40 focus:outline-none"
						use:autoGrow
						value={selfAssessmentValue}
						oninput={(e) => {
							selfAssessmentValue = (e.target as HTMLTextAreaElement).value;
							autoGrow(e.target as HTMLTextAreaElement);
						}}
						rows="1"
						placeholder="Escribe tu autoevaluación..."
						aria-label="Autoevaluación de {goal.name}"
					></textarea>
					{#if authorLine(closure?.selfAssessmentAuthor, closure?.selfAssessmentCreatedAt, closure?.selfAssessment)}
						<p class="text-[11px] text-base-content/40 mt-1">
							{authorLine(closure?.selfAssessmentAuthor, closure?.selfAssessmentCreatedAt, closure?.selfAssessment)}
						</p>
					{/if}
					<div class="mt-2 flex gap-2">
						<button
							class="btn btn-primary btn-xs"
							onclick={handleSaveClosure}
							aria-label="{saveLabel} de {goal.name}"
						>
							{saveLabel}
						</button>
						{#if selfHasComment}
							<button
								class="btn btn-ghost btn-xs"
								onclick={() => {
									editingSelf = false;
									selfAssessmentValue = closure?.selfAssessment ?? '';
								}}
								aria-label="Cancelar edición de autoevaluación de {goal.name}"
							>
								Cancelar
							</button>
						{/if}
					</div>
				{:else}
					<p class="text-sm text-base-content/70 bg-base-200 rounded p-2">
						{closure?.selfAssessment ?? 'Sin autoevaluación'}
						{#if selfReadMode && mode === 'self' && canEdit}
							<button
								class="btn btn-ghost btn-xs ml-1 align-middle"
								onclick={async () => {
									editingSelf = true;
									await tick();
									document.getElementById(selfAssessmentId)?.focus();
								}}
								aria-label="Editar autoevaluación de {goal.name}"
							>
								<Pencil size={14} />
							</button>
						{/if}
					</p>
					{#if authorLine(closure?.selfAssessmentAuthor, closure?.selfAssessmentCreatedAt, closure?.selfAssessment)}
						<p class="text-[11px] text-base-content/40 mt-1">
							{authorLine(closure?.selfAssessmentAuthor, closure?.selfAssessmentCreatedAt, closure?.selfAssessment)}
						</p>
					{/if}
				{/if}
			</div>
		{/if}

		<!-- RH assessment (editable in rh mode, read-only otherwise) -->
		<div class="mb-2">
			<label class="label px-0 py-1" for={rhAssessmentId}>
				<span class="label-text text-xs">Evaluación de RH</span>
			</label>
			{#if mode === 'rh' && !rhReadMode}
				<textarea
					id={rhAssessmentId}
					class="textarea textarea-bordered textarea-xs w-full resize-y overflow-hidden max-h-32 !min-h-[2.25rem] h-auto !pt-2 !pb-1 focus:ring-2 focus:ring-primary/40 focus:outline-none"
					use:autoGrow
					value={rhAssessmentValue}
					oninput={(e) => {
						rhAssessmentValue = (e.target as HTMLTextAreaElement).value;
						autoGrow(e.target as HTMLTextAreaElement);
					}}
					rows="1"
					placeholder="Escribe evaluación de RH..."
					aria-label="Evaluación de RH de {goal.name}"
				></textarea>
				{#if authorLine(closure?.rhAssessmentAuthor, closure?.rhAssessmentCreatedAt, closure?.rhAssessment)}
					<p class="text-[11px] text-base-content/40 mt-1">
						{authorLine(closure?.rhAssessmentAuthor, closure?.rhAssessmentCreatedAt, closure?.rhAssessment)}
					</p>
				{/if}
				<div class="mt-2 flex gap-2">
					<button
						class="btn btn-primary btn-xs"
						onclick={handleRhAssessment}
						aria-label="Guardar evaluación de RH de {goal.name}"
					>
						Guardar evaluación de RH
					</button>
					{#if rhHasComment}
						<button
							class="btn btn-ghost btn-xs"
							onclick={() => {
								editingRh = false;
								rhAssessmentValue = closure?.rhAssessment ?? '';
							}}
							aria-label="Cancelar edición de evaluación de RH de {goal.name}"
						>
							Cancelar
						</button>
					{/if}
				</div>
			{:else}
				<p class="text-sm text-base-content/70 bg-base-200 rounded p-2">
					{closure?.rhAssessment ?? 'Sin evaluación de RH'}
					{#if rhReadMode && mode === 'rh'}
						<button
							class="btn btn-ghost btn-xs ml-1 align-middle"
							onclick={async () => {
								editingRh = true;
								await tick();
								document.getElementById(rhAssessmentId)?.focus();
							}}
							aria-label="Editar evaluación de RH de {goal.name}"
						>
							<Pencil size={14} />
						</button>
					{/if}
				</p>
				{#if authorLine(closure?.rhAssessmentAuthor, closure?.rhAssessmentCreatedAt, closure?.rhAssessment)}
					<p class="text-[11px] text-base-content/40 mt-1">
						{authorLine(closure?.rhAssessmentAuthor, closure?.rhAssessmentCreatedAt, closure?.rhAssessment)}
					</p>
				{/if}
			{/if}
		</div>

		<!-- Manager comment (editable in manager mode, read-only otherwise) -->
		<div class="mb-2">
			{#if mode === 'manager' && canEdit && !managerReadMode}
				<label class="label px-0 py-1" for={managerCommentId}>
					<span class="label-text text-xs">Evaluación de Jefe</span>
				</label>
				<textarea
					id={managerCommentId}
					class="textarea textarea-bordered textarea-xs w-full resize-y overflow-hidden max-h-32 !min-h-[2.25rem] h-auto !pt-2 !pb-1 focus:ring-2 focus:ring-primary/40 focus:outline-none"
					use:autoGrow
					value={managerCommentValue}
					oninput={(e) => {
						managerCommentValue = (e.target as HTMLTextAreaElement).value;
						autoGrow(e.target as HTMLTextAreaElement);
					}}
					rows="1"
					placeholder="Escribe evaluación de Jefe..."
					aria-label="Evaluación de Jefe para {goal.name}"
				></textarea>
				{#if authorLine(closure?.managerCommentAuthor, closure?.managerCommentCreatedAt, closure?.managerComment)}
					<p class="text-[11px] text-base-content/40 mt-1">
						{authorLine(closure?.managerCommentAuthor, closure?.managerCommentCreatedAt, closure?.managerComment)}
					</p>
				{/if}
				<div class="mt-2 flex gap-2">
					<button
						class="btn btn-primary btn-xs"
						onclick={handleManagerComment}
						aria-label="Guardar evaluación de Jefe para {goal.name}"
					>
						Guardar evaluación de Jefe
					</button>
					{#if managerHasComment}
						<button
							class="btn btn-ghost btn-xs"
							onclick={() => {
								editingManager = false;
								managerCommentValue = closure?.managerComment ?? '';
							}}
							aria-label="Cancelar edición de evaluación de Jefe de {goal.name}"
						>
							Cancelar
						</button>
					{/if}
				</div>
			{:else}
				<div class="mt-2">
					<span class="text-xs text-base-content/50">Evaluación de Jefe</span>
					<p class="text-sm text-base-content/70 bg-base-200 rounded p-2">
						{closure?.managerComment ?? 'Sin evaluación de jefe'}
						{#if managerReadMode && mode === 'manager' && canEdit}
							<button
								class="btn btn-ghost btn-xs ml-1 align-middle"
								onclick={async () => {
									editingManager = true;
									await tick();
									document.getElementById(managerCommentId)?.focus();
								}}
								aria-label="Editar evaluación de Jefe de {goal.name}"
							>
								<Pencil size={14} />
							</button>
						{/if}
					</p>
					{#if authorLine(closure?.managerCommentAuthor, closure?.managerCommentCreatedAt, closure?.managerComment)}
						<p class="text-[11px] text-base-content/40 mt-1">
							{authorLine(closure?.managerCommentAuthor, closure?.managerCommentCreatedAt, closure?.managerComment)}
						</p>
					{/if}
				</div>
			{/if}
		</div>
	</div>
</div>
{/if}
