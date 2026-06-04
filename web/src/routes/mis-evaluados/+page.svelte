<script lang="ts">
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import EvaluationStatusBadge from '$lib/components/evaluation/EvaluationStatusBadge.svelte';
	import ComparisonTable from '$lib/components/evaluation/ComparisonTable.svelte';
	import GoalClosureCard from '$lib/components/evaluation/GoalClosureCard.svelte';
	import AssigneePicker from '$lib/components/goals/AssigneePicker.svelte';
	import { getPhase, getProfile } from '$lib/stores/devContext.svelte';
	import { getPillars, getCompetenciesByPillar, getLevelDefinitions, getCompetencyAcceptanceLevel } from '$lib/stores/competencyStore.svelte';
	import {
		getCompetencyRatings,
		getGoalClosures,
		getEvaluationStatus,
		addManagerComment
	} from '$lib/stores/evaluationStore.svelte';
	import { MANAGER_MAP } from '$lib/types/goal';
	import type { EvaluationProfile } from '$lib/types/evaluation';
	import { getAssignments, getAssignmentsByProfile, getGoalsByCategory, getCategories, getKpisForGoal } from '$lib/stores/goalsStore.svelte';
	import { EVALUATION_PROFILES } from '$lib/types/evaluation';

	const phase = $derived(getPhase());
	const profile = $derived(getProfile());
	const pillars = $derived(getPillars());
	const levelDefinitions = $derived(getLevelDefinitions());
	const categories = $derived(getCategories());
	const allAssignments = $derived(getAssignments());

	// Build inverse MANAGER_MAP to find direct reports for this profile
	const inverseManagerMap = $derived.by(() => {
		const map: Record<string, EvaluationProfile[]> = {};
		for (const [subordinate, manager] of Object.entries(MANAGER_MAP)) {
			if (!map[manager]) map[manager] = [];
			map[manager].push(subordinate as EvaluationProfile);
		}
		return map;
	});

	const subordinateProfiles = $derived(inverseManagerMap[profile] ?? []);

	const subordinateAssignments = $derived(
		subordinateProfiles.flatMap((prof) => getAssignmentsByProfile(prof))
	);

	let selectedEmployeeId = $state('');

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

	function handleManagerComment(goalId: string, comment: string) {
		addManagerComment(selectedEmployeeId, goalId, comment);
	}

	const allCompIds = $derived(allCompetencies.map((c) => c.id));
	const acceptanceLevels = $derived(buildAcceptanceLevels(allCompIds));

	const status = $derived(
		selectedEmployeeId
			? getEvaluationStatus(selectedEmployeeId, allCompetencies.length, [])
			: 'pending'
	);
</script>

<svelte:head>
	<title>Mis evaluados — SED</title>
</svelte:head>

{#if phase !== 'fin-anio'}
	<EmptyState
		title="Mis evaluados"
		message="Evaluación no disponible hasta fin de año."
		actionLabel="Volver al inicio"
		actionHref="/"
	/>
{:else}
	<div class="flex flex-col gap-6">
		<!-- Header -->
		<div>
			<h1 class="text-2xl font-bold text-base-content">Mis evaluados</h1>
			<p class="text-sm text-base-content/50 mt-1">
				Revisión de evaluaciones de tu equipo
			</p>
		</div>

		<!-- Employee picker (direct reports only) -->
		<div class="flex items-end gap-4 flex-wrap">
			<AssigneePicker
				assignments={subordinateAssignments}
				selectedEmployeeId={selectedEmployeeId}
				onSelect={(id) => (selectedEmployeeId = id)}
			/>
			{#if selectedEmployeeId}
				<EvaluationStatusBadge {status} />
			{/if}
		</div>

		{#if selectedEmployeeId}
			<!-- Section 1: Comparison table (read-only, no RH column) -->
			<section>
				<h2 class="text-lg font-semibold text-base-content mb-4">Resumen de competencias</h2>
				<ComparisonTable
					{ratings}
					competencies={allCompetencies}
					{acceptanceLevels}
					{levelDefinitions}
					showRhColumn={true}
				/>
			</section>

			<!-- Section 2: Read-only competency view -->
			<section>
				<h2 class="text-lg font-semibold text-base-content mb-4">Competencias</h2>
				{#each pillars as pillar (pillar.id)}
					{@const competencies = getCompetencies(pillar.id)}
					<div class="card bg-base-100 border border-base-300 mb-4">
						<div class="card-body px-4 py-4">
							<h3 class="text-base font-semibold text-base-content mb-4">{pillar.name}</h3>
							<div class="flex flex-col gap-3">
								{#each competencies as competency (competency.id)}
									{@const rating = ratings.find((r) => r.competencyId === competency.id)}
									<div class="border-t border-base-200 pt-3 first:border-t-0 first:pt-0">
										<div class="flex items-start justify-between">
											<div>
												<p class="text-sm font-medium text-base-content">{competency.name}</p>
												<p class="text-xs text-base-content/50">{competency.description}</p>
											</div>
											<div class="flex items-center gap-2 shrink-0">
												{#if rating?.selfRating}
													<span class="badge badge-sm" title="Autoevaluación">
														Auto: {rating.selfRating}
													</span>
												{:else}
													<span class="text-xs text-base-content/30">Sin autoevaluación</span>
												{/if}
												{#if rating?.rhRating}
													<span class="badge badge-sm" title="Evaluación RH">
														RH: {rating.rhRating}
													</span>
												{/if}
											</div>
										</div>
										{#if rating?.selfComment}
											<p class="text-xs text-base-content/50 mt-1 italic">
												"{rating.selfComment}"
											</p>
										{/if}
									</div>
								{/each}
							</div>
						</div>
					</div>
				{/each}
			</section>

			<!-- Section 3: Goal closures with manager comments -->
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
											mode="manager"
											onManagerComment={handleManagerComment}
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
