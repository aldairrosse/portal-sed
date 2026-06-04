<script lang="ts">
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import EvaluationStatusBadge from '$lib/components/evaluation/EvaluationStatusBadge.svelte';
	import ComparisonTable from '$lib/components/evaluation/ComparisonTable.svelte';
	import CompetencyRatingCard from '$lib/components/evaluation/CompetencyRatingCard.svelte';
	import GoalClosureCard from '$lib/components/evaluation/GoalClosureCard.svelte';
	import AssigneePicker from '$lib/components/goals/AssigneePicker.svelte';
	import { getPhase, getProfile } from '$lib/stores/devContext.svelte';
	import { getPillars, getCompetenciesByPillar, getLevelDefinitions, getCompetencyAcceptanceLevel } from '$lib/stores/competencyStore.svelte';
	import {
		getCompetencyRatings,
		rhRateCompetency,
		rhAssessGoal,
		getGoalClosures,
		getEvaluationStatus
	} from '$lib/stores/evaluationStore.svelte';
	import { getAssignmentsByProfile } from '$lib/stores/goalsStore.svelte';
	import { getAssignments, getGoalsByCategory, getCategories, getKpisForGoal } from '$lib/stores/goalsStore.svelte';

	const phase = $derived(getPhase());
	const profile = $derived(getProfile());
	const allAssignments = $derived(getAssignments());
	const pillars = $derived(getPillars());
	const levelDefinitions = $derived(getLevelDefinitions());

	let selectedEmployeeId = $state('');

	const categories = $derived(getCategories());
	const allCompetencies = $derived(pillars.flatMap((p) => getCompetenciesByPillar(p.id)));

	const ratings = $derived(getCompetencyRatings(selectedEmployeeId));
	const closures = $derived(getGoalClosures(selectedEmployeeId));

	function getCompetencies(pillarId: string) {
		return getCompetenciesByPillar(pillarId);
	}

	function getAcceptanceLevel(competencyId: string): number | undefined {
		return getCompetencyAcceptanceLevel(competencyId, profile)?.level;
	}

	function buildAcceptanceLevels(competencyIds: string[]): Record<string, number> {
		const result: Record<string, number> = {};
		for (const compId of competencyIds) {
			const level = getAcceptanceLevel(compId);
			if (level) result[compId] = level;
		}
		return result;
	}

	function handleRhRate(competencyId: string, level: 1 | 2 | 3 | 4 | 5, comment?: string) {
		rhRateCompetency(selectedEmployeeId, competencyId, level, comment);
	}

	function handleRhAssessGoal(goalId: string, rhAssessment: string) {
		rhAssessGoal(selectedEmployeeId, goalId, rhAssessment);
	}

	// Build acceptance levels for all competencies
	const allCompIds = $derived(allCompetencies.map((c) => c.id));
	const acceptanceLevels = $derived(buildAcceptanceLevels(allCompIds));

	const status = $derived(
		selectedEmployeeId
			? getEvaluationStatus(selectedEmployeeId, allCompetencies.length, [])
			: 'pending'
	);
</script>

<svelte:head>
	<title>Evaluaciones RH — SED</title>
</svelte:head>

{#if phase !== 'fin-anio'}
	<EmptyState
		title="Evaluaciones RH"
		message="Evaluación no disponible hasta fin de año."
		actionLabel="Volver al inicio"
		actionHref="/"
	/>
{:else}
	<div class="flex flex-col gap-6">
		<!-- Header -->
		<div>
			<h1 class="text-2xl font-bold text-base-content">Evaluaciones RH</h1>
			<p class="text-sm text-base-content/50 mt-1">
				Evaluación formal de competencias y cierre de metas
			</p>
		</div>

		<!-- Employee picker -->
		<div class="flex items-end gap-4 flex-wrap">
			<AssigneePicker
				assignments={allAssignments}
				selectedEmployeeId={selectedEmployeeId}
				onSelect={(id) => (selectedEmployeeId = id)}
			/>
			{#if selectedEmployeeId}
				<EvaluationStatusBadge {status} />
			{/if}
		</div>

		{#if selectedEmployeeId}
			<!-- Comparison table -->
			<section>
				<h2 class="text-lg font-semibold text-base-content mb-4">Comparación</h2>
				<ComparisonTable
					{ratings}
					competencies={allCompetencies}
					{acceptanceLevels}
					{levelDefinitions}
					showRhColumn={true}
				/>
			</section>

			<!-- Section: RH competency ratings -->
			<section>
				<h2 class="text-lg font-semibold text-base-content mb-4">Evaluación de competencias</h2>
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
							mode="rh"
							onRate={handleRhRate}
							onRhRate={handleRhRate}
						/>
					{/each}
				</div>
			</section>

			<!-- Section: Goal closures -->
			<section>
				<h2 class="text-lg font-semibold text-base-content mb-4">Cierre de metas</h2>
				{#each categories as category (category.id)}
					{@const goals = getGoalsByCategory(category.id)}
					{#if goals.length > 0}
						<div class="card bg-base-100 border border-base-300 mb-4">
							<div class="card-body px-4 py-4">
								<h3 class="text-base font-semibold text-base-content mb-3">{category.name}</h3>
								<div class="flex flex-col gap-4">
									{#each goals as goal (goal.id)}
										{@const kpis = getKpisForGoal(goal.id)}
										{@const closure = closures.find((c) => c.goalId === goal.id)}
										<GoalClosureCard
											{goal}
											{kpis}
											{closure}
											mode="rh"
											onRhAssessGoal={handleRhAssessGoal}
										/>
									{/each}
								</div>
							</div>
						</div>
					{/if}
				{/each}
			</section>
		{:else}
			<p class="text-sm text-base-content/30 italic text-center py-8">
				Selecciona un evaluado para comenzar.
			</p>
		{/if}
	</div>
{/if}
