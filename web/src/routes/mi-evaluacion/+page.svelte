<script lang="ts">
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import CompetencyRatingCard from '$lib/components/evaluation/CompetencyRatingCard.svelte';
	import GoalClosureCard from '$lib/components/evaluation/GoalClosureCard.svelte';
	import { getPhase, getProfile } from '$lib/stores/devContext.svelte';
	import { getPillars, getCompetenciesByPillar, getLevelDefinitions, getCompetencyAcceptanceLevel } from '$lib/stores/competencyStore.svelte';
	import { getCompetencyRatings, rateCompetency, closeGoal, getGoalClosures } from '$lib/stores/evaluationStore.svelte';
	import { getAssignmentsByProfile } from '$lib/stores/goalsStore.svelte';
	import { getGoalsByCategory, getCategories, getKpisForGoal } from '$lib/stores/goalsStore.svelte';

	const phase = $derived(getPhase());
	const profile = $derived(getProfile());
	const assignments = $derived(getAssignmentsByProfile(profile));
	const employeeId = $derived(assignments[0]?.employeeId ?? '');
	const pillars = $derived(getPillars());
	const levelDefinitions = $derived(getLevelDefinitions());
	const ratings = $derived(getCompetencyRatings(employeeId));
	const closures = $derived(getGoalClosures(employeeId));
	const categories = $derived(getCategories());

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

	function handleRate(competencyId: string, level: 1 | 2 | 3 | 4 | 5, comment?: string) {
		rateCompetency(employeeId, competencyId, level, comment);
	}

	function handleCloseGoal(goalId: string, finalProgress: number, selfAssessment: string) {
		closeGoal(employeeId, goalId, finalProgress, selfAssessment);
	}
</script>

<svelte:head>
	<title>Mi evaluación — SED</title>
</svelte:head>

{#if phase !== 'fin-anio'}
	<EmptyState
		title="Mi evaluación"
		message="Evaluación no disponible hasta fin de año."
		actionLabel="Volver al inicio"
		actionHref="/"
	/>
{:else}
	<div class="flex flex-col gap-6">
		<!-- Header -->
		<div>
			<h1 class="text-2xl font-bold text-base-content">Mi evaluación</h1>
			<p class="text-sm text-base-content/50 mt-1">
				Autoevaluación de competencias y cierre de metas
			</p>
		</div>

		<!-- Section 1: Competencias -->
		<section>
			<h2 class="text-lg font-semibold text-base-content mb-4">Competencias</h2>
			{#if pillars.length === 0}
				<p class="text-sm text-base-content/30 italic">No hay pilares configurados.</p>
			{:else}
				<div class="flex flex-col gap-6">
					{#each pillars as pillar (pillar.id)}
						{@const competencies = getCompetencies(pillar.id)}
						{@const compIds = competencies.map((c) => c.id)}
						{@const acceptanceLevels = buildAcceptanceLevels(compIds)}
						<CompetencyRatingCard
							{pillar}
							{competencies}
							{ratings}
							{levelDefinitions}
							{acceptanceLevels}
							mode="self"
							onRate={handleRate}
							disabled={false}
						/>
					{/each}
				</div>
			{/if}
		</section>

		<!-- Section 2: Cierre de metas -->
		<section>
			<h2 class="text-lg font-semibold text-base-content mb-4">Cierre de metas</h2>
			{#if categories.length === 0}
				<p class="text-sm text-base-content/30 italic">No hay categorías de metas configuradas.</p>
			{:else}
				<div class="flex flex-col gap-6">
					{#each categories as category (category.id)}
						{@const goals = getGoalsByCategory(category.id)}
						{#if goals.length > 0}
							<div class="card bg-base-100 border border-base-300">
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
												mode="self"
												canEdit={true}
												onSaveClosure={handleCloseGoal}
											/>
										{/each}
									</div>
								</div>
							</div>
						{/if}
					{/each}
				</div>
			{/if}
		</section>
	</div>
{/if}
