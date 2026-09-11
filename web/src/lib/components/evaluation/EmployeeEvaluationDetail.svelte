<script lang="ts">
	import { onMount } from 'svelte';
	import CompetencyRatingCard from './CompetencyRatingCard.svelte';
	import GoalClosureCard from './GoalClosureCard.svelte';
	import ComparisonTable from './ComparisonTable.svelte';
	import EvaluationStatusBadge from './EvaluationStatusBadge.svelte';
	import PageSkeleton from '$lib/components/ui/PageSkeleton.svelte';
	import ErrorState from '$lib/components/ui/ErrorState.svelte';
	import * as notifications from '$lib/stores/notifications.svelte';
	import { getActivePhase } from '$lib/api/cycle.svelte';
	import { isFinAnio as isFinAnioPhase, isMedioAnio as isMedioAnioPhase, isInicioAnio as isInicioAnioPhase } from '$lib/types/cycle';
	import {
		getPillars,
		getCompetenciesByPillar,
		getLevelDefinitions,
		getCompetencyAcceptanceLevel,
		load as loadCompetencies,
	} from '$lib/stores/competencyStore.svelte';
	import {
		getCompetencyRatings,
		rateCompetency,
		closeGoal,
		getGoalClosures,
		getEvaluationStatus,
		rhRateCompetency,
		rhAssessGoal,
		addManagerComment,
		isLoading,
		getError,
		errorWithCode,
		load as loadEvaluations,
	} from '$lib/stores/evaluationStore.svelte';
	import {
		getGoalsByCategory,
		getCategories,
		getKpisForGoal,
		getAssignmentByEmployee,
		isLoading as isGoalsLoading,
		load as loadGoals,
		loadForEmployee as loadGoalsForEmployee,
	} from '$lib/stores/goalsStore.svelte';
	import { Star } from '@lucide/svelte';
	interface Props {
		employeeId: string;
		viewerMode: 'self' | 'manager' | 'rh';
		showBreadcrumb?: boolean;
		employeeName?: string;
		onBack?: () => void;
	}

	let { employeeId, viewerMode, showBreadcrumb = false, employeeName = '', onBack }: Props = $props();

	const backHref = $derived(
		viewerMode === 'self'
			? '/mi-evaluacion'
			: viewerMode === 'rh'
				? '/rh/evaluaciones'
				: '/mis-evaluados'
	);

	const loadingEval = $derived(isLoading());
	const loadingGoals = $derived(isGoalsLoading());
	const errorEval = $derived(getError());

	onMount(() => {
		loadEvaluations(employeeId, viewerMode);
		if (viewerMode === 'self') loadGoals();
		else loadGoalsForEmployee(employeeId);
		loadCompetencies();
	});

	const phase = $derived(getActivePhase() ?? 'inicio-anio');
	const isFinAnio = $derived(isFinAnioPhase(phase));
	const isMedioAnio = $derived(isMedioAnioPhase(phase));
	const isInicioAnio = $derived(isInicioAnioPhase(phase));
	const pillars = $derived(getPillars());
	const levelDefinitions = $derived(getLevelDefinitions());
	const categories = $derived(getCategories());
	const assignment = $derived(getAssignmentByEmployee(employeeId));
	const hasPersonalGoals = $derived(
		categories.some((c) => getGoalsByCategory(c.id).length > 0) ||
			(assignment?.goalIds?.length ?? 0) > 0
	);
	const isPhaseMedioOrFin = $derived(isMedioAnio || isFinAnio);
	const showEmptyMedioFin = $derived(isPhaseMedioOrFin && !hasPersonalGoals);
	const ratings = $derived(getCompetencyRatings(employeeId));
	const closures = $derived(getGoalClosures(employeeId));

	// goalsStore assignments carry employeeName: '' (API has no name field),
	// so fall back to the list-provided name before the generic label.
	const displayName = $derived(
		assignment?.employeeName?.trim() || employeeName?.trim() || 'Evaluado'
	);

	const allCompetencies = $derived(pillars.flatMap((p) => getCompetenciesByPillar(p.id)));
	const allCompIds = $derived(allCompetencies.map((c) => c.id));

	function getAcceptanceLevel(competencyId: string): number | undefined {
		return getCompetencyAcceptanceLevel(competencyId, assignment?.profileId ?? 'colaborador')?.level;
	}

	function buildAcceptanceLevels(competencyIds: string[]): Record<string, number> {
		const result: Record<string, number> = {};
		for (const compId of competencyIds) {
			const level = getAcceptanceLevel(compId);
			if (level) result[compId] = level;
		}
		return result;
	}

	const acceptanceLevels = $derived(buildAcceptanceLevels(allCompIds));

	const phaseKind = $derived<'avance' | 'cierre'>(isMedioAnio ? 'avance' : 'cierre');

	const goalIds = $derived(
		assignment?.goalIds?.length
			? [...assignment.goalIds]
			: categories.flatMap((c) => getGoalsByCategory(c.id).map((g) => g.id))
	);

	const status = $derived(
		employeeId ? getEvaluationStatus(employeeId, allCompetencies.length, goalIds, phaseKind) : 'pending'
	);

	const disabled = $derived(!(isMedioAnio || isFinAnio));
	const canEditInPhase = $derived(isMedioAnio || isFinAnio);
	const showCommentInput = $derived(canEditInPhase);
	const headingLabel = $derived(
		isMedioAnio ? 'Evaluación de avance de medio año' : 'Evaluación de cierre de año'
	);

	const tabs = $derived(
		viewerMode === 'self'
			? (['metas', 'competencias'] as const)
			: (['resumen', 'metas', 'competencias'] as const)
	);

	let currentTab = $state<'metas' | 'competencias' | 'resumen'>('metas');

	// Reset tab when employeeId or viewerMode changes
	$effect(() => {
		const _id = employeeId;
		const _mode = viewerMode;
		currentTab = _mode === 'self' ? 'metas' : 'resumen';
		if (_mode === 'self') loadGoals();
		else loadGoalsForEmployee(_id);
		loadEvaluations(_id, _mode);
	});

	// Unique name for radio group so multiple instances don't conflict
	const radioName = $derived(`eval_tabs_${employeeId}`);

	const sectionLabel = $derived(
		viewerMode === 'manager' ? 'Mis evaluados' : 'Evaluaciones RH'
	);

	function getCompetencies(pillarId: string) {
		return getCompetenciesByPillar(pillarId);
	}

	async function handleSelfRate(competencyId: string, level: 1 | 2 | 3 | 4 | 5, comment?: string) {
		try {
			await rateCompetency(employeeId, competencyId, level, comment);
			notifications.success('Autoevaluación guardada');
		} catch (e) {
			const { message, code } = errorWithCode(e);
			notifications.errorWithCode(message || 'Error al guardar autoevaluación', code);
		}
	}

	async function handleRhRate(competencyId: string, level: 1 | 2 | 3 | 4 | 5, comment?: string, _evaluationId?: string, _phase?: 'avance' | 'cierre', managerComment?: string) {
		try {
			await rhRateCompetency(employeeId, competencyId, level, comment, managerComment);
			notifications.success('Evaluación RH guardada');
		} catch (e) {
			const { message, code } = errorWithCode(e);
			notifications.errorWithCode(message || 'Error al guardar evaluación RH', code);
		}
	}

	async function handleCloseGoal(goalId: string, finalProgress: number, selfAssessment: string) {
		try {
			await closeGoal(employeeId, goalId, finalProgress, selfAssessment);
			notifications.success('Meta guardada');
		} catch (e) {
			const { message, code } = errorWithCode(e);
			notifications.errorWithCode(message || 'Error al cerrar meta', code);
		}
	}

	async function handleRhAssessGoal(goalId: string, rhAssessment: string) {
		try {
			await rhAssessGoal(employeeId, goalId, rhAssessment);
			notifications.success('Evaluación RH guardada');
		} catch (e) {
			const { message, code } = errorWithCode(e);
			notifications.errorWithCode(message || 'Error al guardar evaluación RH', code);
		}
	}

	async function handleManagerComment(goalId: string, comment: string) {
		try {
			await addManagerComment(employeeId, goalId, comment);
			notifications.success('Comentario guardado');
		} catch (e) {
			const { message, code } = errorWithCode(e);
			notifications.errorWithCode(message || 'Error al guardar comentario', code);
		}
	}
</script>

{#if loadingEval}
	<PageSkeleton variant="card" rows={3} />
{:else if errorEval}
	<ErrorState message={errorEval} onretry={loadEvaluations} />
					{:else}
						<div class="flex flex-col gap-6">
	{#if showBreadcrumb}
		<nav aria-label="Breadcrumb">
			<div class="breadcrumbs text-sm">
				<ul>
					<li>
						<a
							href={backHref}
							class="link link-hover"
							aria-label="Volver a {sectionLabel}"
							onclick={(e) => {
								if (onBack) {
									e.preventDefault();
									onBack();
								}
							}}
						>
							{sectionLabel}
						</a>
					</li>
					<li>{displayName}</li>
				</ul>
			</div>
		</nav>
	{/if}

	<!-- Header -->
	<div class="flex items-center gap-3 flex-wrap">
		<div class="avatar placeholder">
			<div class="bg-primary text-primary-content rounded-full w-9 flex items-center justify-center">
				<span class="text-sm font-semibold">
					{(displayName.trim().charAt(0) || '—').toUpperCase()}
				</span>
			</div>
		</div>
		<h2 class="{showBreadcrumb ? 'text-xl' : 'text-lg'} font-semibold text-base-content">
			{displayName}
		</h2>
		<EvaluationStatusBadge {status} />
	</div>
	<p class="text-sm text-base-content/50 -mt-4">{headingLabel}</p>

	<!-- Tabs (tabs-lift) -->
	<div class="tabs tabs-lift">
		{#each tabs as tab (tab)}
			<input
				type="radio"
				name={radioName}
				class="tab"
				aria-label={tab === 'resumen' ? 'Resumen' : tab === 'metas' ? 'Metas' : 'Competencias'}
				checked={currentTab === tab}
				onchange={() => (currentTab = tab)}
			/>
			<section class="tab-content bg-base-100 border-base-300 p-6">
				<!-- Tab: Resumen -->
				{#if tab === 'resumen'}
					<h3 class="text-lg font-semibold text-base-content mb-4">Resumen de competencias</h3>
					<ComparisonTable
						{ratings}
						competencies={allCompetencies}
						{acceptanceLevels}
						{levelDefinitions}
						showRhColumn={canEditInPhase && viewerMode !== 'self'}
					/>
				{/if}

				<!-- Tab: Metas -->
				{#if tab === 'metas'}
					<h3 class="text-lg font-semibold text-base-content mb-4">
						{viewerMode === 'self' ? 'Mis metas' : isMedioAnio ? 'Avance de metas' : 'Cierre de metas'}
					</h3>
					{#if isInicioAnio && viewerMode === 'self' && !hasPersonalGoals}
						<div class="flex flex-col gap-3">
							<p class="text-sm text-base-content/30 italic">No has registrado metas para este ciclo.</p>
							<a href="/objetivos/asignacion" class="btn btn-primary btn-sm w-fit">Ir a asignación</a>
						</div>
					{:else if loadingGoals}
						<p class="text-sm text-base-content/30 italic">Cargando metas…</p>
					{:else if showEmptyMedioFin}
						<p class="text-sm text-base-content/30 italic">No se registraron metas para evaluación.</p>
					{:else if categories.length === 0}
						<p class="text-sm text-base-content/30 italic">No hay categorías de metas configuradas.</p>
					{:else}
						<div class="flex flex-col gap-6">
							{#each categories as category (category.id)}
								{@const goals = getGoalsByCategory(category.id)}
								{#if goals.length > 0}
								<div class="card bg-base-100 border border-base-300">
									<div class="card-body px-0">
											<h4 class="text-base font-semibold text-base-content mb-3">{category.name}</h4>
											<div class="flex flex-col gap-4">
												{#each goals as goal (goal.id)}
											{@const kpis = getKpisForGoal(goal.id)}
												{@const closure = closures.find((c) => c.goalId === goal.id)}
												<!-- F2: single card for every viewerMode; GoalClosureCard shows
													self/rh/manager comments read-only to all, edits stay gated by mode+canEdit. -->
												<GoalClosureCard
													{goal}
													{kpis}
													{closure}
													mode={viewerMode}
													phase={phaseKind}
													canEdit={canEditInPhase}
													showSelfAssessment={canEditInPhase}
													{employeeId}
													onSaveClosure={handleCloseGoal}
													onManagerComment={handleManagerComment}
													onRhAssessGoal={handleRhAssessGoal}
												/>
												{/each}
											</div>
										</div>
									</div>
								{/if}
							{/each}
						</div>
					{/if}
				{/if}

				<!-- Tab: Competencias -->
				{#if tab === 'competencias'}
					<div class="flex items-center justify-between mb-4">
						<h3 class="text-lg font-semibold text-base-content">
							{viewerMode === 'self' ? 'Mis competencias' : 'Competencias'}
						</h3>
						<a href="/evaluacion/9x9/competencias/{employeeId}" class="btn btn-outline btn-sm gap-1.5">
							<Star class="w-4 h-4" />
							Gráfico
						</a>
					</div>
					{#if pillars.length === 0}
						<p class="text-sm text-base-content/30 italic">No hay pilares configurados.</p>
					{:else}
						<div class="flex flex-col gap-6">
							{#each pillars as pillar (pillar.id)}
								{@const competencies = getCompetencies(pillar.id)}
								{@const compIds = competencies.map((c) => c.id)}
								{@const pillarAcceptance = buildAcceptanceLevels(compIds)}
								<CompetencyRatingCard
									{pillar}
									{competencies}
									{ratings}
									{levelDefinitions}
									acceptanceLevels={pillarAcceptance}
									mode={viewerMode}
									{disabled}
									phase={phaseKind}
									{showCommentInput}
									onRate={handleSelfRate}
									onRhRate={handleRhRate}
								/>
							{/each}
						</div>
					{/if}
				{/if}
			</section>
		{/each}
	</div>
</div>
{/if}
