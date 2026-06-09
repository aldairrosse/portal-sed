<script lang="ts">
	import type { Goal, KPI } from '$lib/types/goal';
	import type { GoalClosure } from '$lib/types/evaluation-result';
	import ProgressIndicator from '$lib/components/goals/ProgressIndicator.svelte';
	import KpiBadge from '$lib/components/goals/KpiBadge.svelte';

	interface Props {
		goal: Goal;
		kpis: KPI[];
		closure?: GoalClosure;
		mode: 'self' | 'rh' | 'manager';
		canEdit?: boolean;
		showSelfAssessment?: boolean;
		employeeId?: string;
		onSaveClosure?: (goalId: string, finalProgress: number, selfAssessment: string) => void;
		onRhAssessGoal?: (goalId: string, rhAssessment: string) => void;
		onManagerComment?: (goalId: string, comment: string) => void;
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
		onManagerComment
	}: Props = $props();

	const progressId = `progress-${goal.id}`;
	const selfAssessmentId = `self-assessment-${goal.id}`;
	const rhAssessmentId = `rh-assessment-${goal.id}`;
	const managerCommentId = `manager-comment-${goal.id}`;

	let progressValue = $state(closure?.finalProgress ?? goal.progress ?? 0);
	let selfAssessmentValue = $state(closure?.selfAssessment ?? '');
	let rhAssessmentValue = $state(closure?.rhAssessment ?? '');
	let managerCommentValue = $state(closure?.managerComment ?? '');
	let saved = $state(!!closure?.closedAt);

	const unitLabels: Record<string, string> = {
		porcentaje: '%',
		moneda: '$',
		numero: '#',
		binario: 'Sí/No'
	};

	function handleSaveClosure() {
		onSaveClosure?.(goal.id, progressValue, selfAssessmentValue);
		saved = true;
	}

	function handleRhAssessment() {
		onRhAssessGoal?.(goal.id, rhAssessmentValue);
	}

	function handleManagerComment() {
		onManagerComment?.(goal.id, managerCommentValue);
	}

	function handleProgressInput(e: Event) {
		progressValue = parseFloat((e.target as HTMLInputElement).value) || 0;
	}
</script>

<div class="card bg-base-100 border border-base-300">
	<div class="card-body px-0 py-4">
		<!-- Goal header -->
		<div class="flex items-start justify-between gap-2 mb-3">
			<div class="flex-1 min-w-0">
				<div class="flex items-center gap-2 flex-wrap">
					<h4 class="text-sm font-semibold text-base-content">{goal.name}</h4>
					{#if saved}
						<span class="badge badge-ghost badge-xs">Cerrada</span>
					{/if}
				</div>
				<p class="text-xs text-base-content/50 mt-0.5">
					Valor objetivo: {goal.targetValue}{unitLabels[goal.unit] ?? goal.unit} · Peso: {goal.weight}%
				</p>
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
			{#if mode === 'self' && canEdit && !saved}
				<label class="label px-0 py-1" for={progressId}>
					<span class="label-text text-xs">Avance final</span>
				</label>
				<div class="flex items-center gap-2">
					<input
						id={progressId}
						type="number"
						class="input input-bordered input-sm w-24"
						value={progressValue}
						min="0"
						max={goal.unit === 'porcentaje' ? 100 : goal.targetValue}
						oninput={handleProgressInput}
						aria-label="Avance final de {goal.name}"
					/>
					<ProgressIndicator
						value={progressValue}
						max={goal.unit === 'porcentaje' ? 100 : goal.targetValue}
					/>
				</div>
			{:else}
				<div class="flex items-center gap-2">
					<span class="text-xs text-base-content/60">Avance final:</span>
					<ProgressIndicator
						value={progressValue}
						max={goal.unit === 'porcentaje' ? 100 : goal.targetValue}
					/>
				</div>
			{/if}
		</div>

		<!-- Self assessment (editable in self mode, read-only in rh/manager mode) -->
		{#if mode === 'self' && showSelfAssessment}
			<div class="mb-2">
				<label class="label px-0 py-1" for={selfAssessmentId}>
					<span class="label-text text-xs">Autoevaluación</span>
				</label>
				{#if canEdit && !saved}
					<textarea
						id={selfAssessmentId}
						class="textarea textarea-bordered textarea-xs w-full"
						value={selfAssessmentValue}
						oninput={(e) => (selfAssessmentValue = (e.target as HTMLTextAreaElement).value)}
						rows="2"
						placeholder="Describe tu evaluación de cierre para esta meta"
						aria-label="Autoevaluación de {goal.name}"
					></textarea>
					<div class="mt-2">
						<button
							class="btn btn-primary btn-xs"
							onclick={handleSaveClosure}
							aria-label="Guardar cierre de {goal.name}"
						>
							Guardar cierre
						</button>
					</div>
				{:else}
					<p class="text-sm text-base-content/70 bg-base-200 rounded p-2">
						{closure?.selfAssessment ?? 'Sin autoevaluación'}
					</p>
				{/if}
			</div>
		{/if}

		<!-- RH assessment -->
		{#if mode === 'rh'}
			<div class="mb-2">
				<label class="label px-0 py-1" for={rhAssessmentId}>
					<span class="label-text text-xs">Evaluación RH</span>
				</label>
				{#if closure?.selfAssessment}
					<div class="mb-2">
						<span class="text-xs text-base-content/50">Autoevaluación:</span>
						<p class="text-sm text-base-content/70 bg-base-200 rounded p-2">{closure.selfAssessment}</p>
					</div>
				{/if}
				<textarea
					id={rhAssessmentId}
					class="textarea textarea-bordered textarea-xs w-full"
					value={rhAssessmentValue}
					oninput={(e) => (rhAssessmentValue = (e.target as HTMLTextAreaElement).value)}
					rows="2"
					placeholder="Evaluación RH de cierre"
					aria-label="Evaluación RH de {goal.name}"
				></textarea>
				<div class="mt-2">
					<button
						class="btn btn-primary btn-xs"
						onclick={handleRhAssessment}
						aria-label="Guardar evaluación RH de {goal.name}"
					>
						Guardar evaluación RH
					</button>
				</div>
			</div>
		{/if}

		<!-- Manager comment -->
		{#if mode === 'manager'}
			<div class="mb-2">
				{#if closure?.selfAssessment}
					<div class="mb-2">
						<span class="text-xs text-base-content/50">Autoevaluación:</span>
						<p class="text-sm text-base-content/70 bg-base-200 rounded p-2">{closure.selfAssessment}</p>
					</div>
				{/if}
				{#if closure?.rhAssessment}
					<div class="mb-2">
						<span class="text-xs text-base-content/50">Evaluación RH:</span>
						<p class="text-sm text-base-content/70 bg-base-200 rounded p-2">{closure.rhAssessment}</p>
					</div>
				{/if}
				{#if canEdit}
					<label class="label px-0 py-1" for={managerCommentId}>
						<span class="label-text text-xs">Comentario del jefe</span>
					</label>
					<textarea
						id={managerCommentId}
						class="textarea textarea-bordered textarea-xs w-full"
						value={managerCommentValue}
						oninput={(e) => (managerCommentValue = (e.target as HTMLTextAreaElement).value)}
						rows="2"
						placeholder="Agrega tu comentario como jefe"
						aria-label="Comentario del jefe para {goal.name}"
					></textarea>
					<div class="mt-2">
						<button
							class="btn btn-primary btn-xs"
							onclick={handleManagerComment}
							aria-label="Guardar comentario para {goal.name}"
						>
							Guardar comentario
						</button>
					</div>
				{:else if closure?.managerComment}
					<div class="mt-2">
						<span class="text-xs text-base-content/50">Comentario del jefe:</span>
						<p class="text-sm text-base-content/70 bg-base-200 rounded p-2">{closure.managerComment}</p>
					</div>
				{/if}
			</div>
		{/if}
	</div>
</div>
