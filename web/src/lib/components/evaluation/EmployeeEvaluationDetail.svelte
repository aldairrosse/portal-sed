<script lang="ts">
	import CompetencyRatingCard from './CompetencyRatingCard.svelte';
	import GoalClosureCard from './GoalClosureCard.svelte';
	import ComparisonTable from './ComparisonTable.svelte';
	import EvaluationStatusBadge from './EvaluationStatusBadge.svelte';
	import { getPhase } from '$lib/stores/devContext.svelte';
	import {
		getPillars,
		getCompetenciesByPillar,
		getLevelDefinitions,
		getCompetencyAcceptanceLevel,
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
	} from '$lib/stores/evaluationStore.svelte';
	import {
		getGoalsByCategory,
		getCategories,
		getKpisForGoal,
		getAssignmentByEmployee,
	} from '$lib/stores/goalsStore.svelte';

	interface Props {
		employeeId: string;
		viewerMode: 'self' | 'manager' | 'rh';
		showBreadcrumb?: boolean;
		onBack?: () => void;
	}

	let {
		employeeId,
		viewerMode,
		showBreadcrumb = false,
		onBack,
	}: Props = $props();

	const phase = $derived(getPhase());
	const isFinAnio = $derived(phase === 'fin-anio');
	const pillars = $derived(getPillars());
	const levelDefinitions = $derived(getLevelDefinitions());
	const categories = $derived(getCategories());
	const ratings = $derived(getCompetencyRatings(employeeId));
	const closures = $derived(getGoalClosures(employeeId));
	const assignment = $derived(getAssignmentByEmployee(employeeId));

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

	const status = $derived(
		employeeId ? getEvaluationStatus(employeeId, allCompetencies.length, []) : 'pending'
	);

	const disabled = $derived(!isFinAnio);
	const showCommentInput = $derived(isFinAnio);

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
	});

	function getCompetencies(pillarId: string) {
		return getCompetenciesByPillar(pillarId);
	}

	function handleSelfRate(competencyId: string, level: 1 | 2 | 3 | 4 | 5, comment?: string) {
		rateCompetency(employeeId, competencyId, level, comment);
	}

	function handleRhRate(competencyId: string, level: 1 | 2 | 3 | 4 | 5, comment?: string) {
		rhRateCompetency(employeeId, competencyId, level, comment);
	}

	function handleCloseGoal(goalId: string, finalProgress: number, selfAssessment: string) {
		closeGoal(employeeId, goalId, finalProgress, selfAssessment);
	}

	function handleRhAssessGoal(goalId: string, rhAssessment: string) {
		rhAssessGoal(employeeId, goalId, rhAssessment);
	}

	function handleManagerComment(goalId: string, comment: string) {
		addManagerComment(employeeId, goalId, comment);
	}
</script>

<div class="flex flex-col gap-6">
	{#if showBreadcrumb}
		<nav aria-label="Breadcrumb">
			<ol class="flex items-center gap-2 text-sm">
				<li>
					<button
						type="button"
						class="text-primary hover:underline"
						onclick={onBack}
						aria-label="Volver a la lista"
					>
						← Volver a la lista
					</button>
				</li>
				<li class="text-base-content/30">/</li>
				<li class="font-semibold text-base-content">
					{assignment?.employeeName ?? 'Evaluado'}
				</li>
			</ol>
		</nav>
	{/if}

	<!-- Header -->
	<div class="flex items-center gap-2 flex-wrap">
		<h2 class="text-lg font-semibold text-base-content">
			{assignment?.employeeName ?? 'Evaluación'}
		</h2>
		<EvaluationStatusBadge {status} />
	</div>

	<!-- Tabs -->
	<div role="tablist" class="tabs tabs-bordered">
		{#each tabs as tab (tab)}
			<button
				role="tab"
				class="tab {currentTab === tab ? 'tab-active' : ''}"
				onclick={() => (currentTab = tab)}
				aria-selected={currentTab === tab}
			>
				{tab === 'resumen' ? 'Resumen' : tab === 'metas' ? 'Metas' : 'Competencias'}
			</button>
		{/each}
	</div>

	<!-- Tab: Resumen -->
	{#if currentTab === 'resumen'}
		<section>
			<h3 class="text-lg font-semibold text-base-content mb-4">Resumen de competencias</h3>
			<ComparisonTable
				{ratings}
				competencies={allCompetencies}
				{acceptanceLevels}
				{levelDefinitions}
				showRhColumn={isFinAnio}
			/>
		</section>
	{/if}

	<!-- Tab: Metas -->
	{#if currentTab === 'metas'}
		<section>
			<h3 class="text-lg font-semibold text-base-content mb-4">
				{viewerMode === 'self' ? 'Mis metas' : 'Cierre de metas'}
			</h3>
			{#if categories.length === 0}
				<p class="text-sm text-base-content/30 italic">No hay categorías de metas configuradas.</p>
			{:else}
				<div class="flex flex-col gap-6">
					{#each categories as category (category.id)}
						{@const goals = getGoalsByCategory(category.id)}
						{#if goals.length > 0}
							<div class="card bg-base-100 border border-base-300">
								<div class="card-body">
									<h4 class="text-base font-semibold text-base-content mb-3">{category.name}</h4>
									<div class="flex flex-col gap-4">
										{#each goals as goal (goal.id)}
											{@const kpis = getKpisForGoal(goal.id)}
											{@const closure = closures.find((c) => c.goalId === goal.id)}
											{#if viewerMode === 'self'}
												<GoalClosureCard
													{goal}
													{kpis}
													{closure}
													mode="self"
													canEdit={isFinAnio}
													showSelfAssessment={isFinAnio}
													onSaveClosure={handleCloseGoal}
												/>
											{:else if viewerMode === 'manager'}
												<GoalClosureCard
													{goal}
													{kpis}
													{closure}
													mode="manager"
													canEdit={isFinAnio}
													{employeeId}
													onManagerComment={handleManagerComment}
												/>
											{:else if viewerMode === 'rh'}
												<GoalClosureCard
													{goal}
													{kpis}
													{closure}
													mode="rh"
													onRhAssessGoal={handleRhAssessGoal}
												/>
											{/if}
										{/each}
									</div>
								</div>
							</div>
						{/if}
					{/each}
				</div>
			{/if}
		</section>
	{/if}

	<!-- Tab: Competencias -->
	{#if currentTab === 'competencias'}
		<section>
			<h3 class="text-lg font-semibold text-base-content mb-4">
				{viewerMode === 'self' ? 'Mis competencias' : 'Competencias'}
			</h3>
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
							{showCommentInput}
							onRate={handleSelfRate}
							onRhRate={handleRhRate}
						/>
					{/each}
				</div>
			{/if}
		</section>
	{/if}
</div>
