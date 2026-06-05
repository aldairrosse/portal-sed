<script lang="ts">
	import ComparisonTable from './ComparisonTable.svelte';
	import {
		getPillars,
		getCompetenciesByPillar,
		getCompetencyAcceptanceLevelsByProfile,
		getLevelDefinitions
	} from '$lib/stores/competencyStore.svelte';
	import { getCompetencyRatings } from '$lib/stores/evaluationStore.svelte';
	import { getNodeById } from '$lib/stores/orgHierarchyStore.svelte';
	import type { CompetencyRating } from '$lib/types/evaluation-result';
	import type { EvaluationProfile } from '$lib/types/evaluation';

	interface Props {
		employeeId: string;
		employeeName?: string;
	}

	let { employeeId, employeeName = '' }: Props = $props();

	// ─── Derived data ──────────────────────────────────────────────────────

	const employeeNode = $derived(getNodeById(employeeId));
	const profileId = $derived((employeeNode?.profileId ?? 'colaborador') as EvaluationProfile);
	const displayName = $derived(employeeName || employeeNode?.name || 'Empleado');

	const pillars = $derived(getPillars());
	const levelDefinitions = $derived(getLevelDefinitions());
	const allAcceptance = $derived(getCompetencyAcceptanceLevelsByProfile(profileId));

	/** Build acceptance levels lookup: competencyId → level */
	const acceptanceLevels: Record<string, number> = $derived.by(() => {
		const map: Record<string, number> = {};
		for (const entry of allAcceptance) {
			map[entry.competencyId] = entry.level;
		}
		return map;
	});

	const ratings = $derived<CompetencyRating[]>(getCompetencyRatings(employeeId));
</script>

<div class="flex flex-col gap-6">
	{#each pillars as pillar (pillar.id)}
		{@const pillarCompetencies = getCompetenciesByPillar(pillar.id)}
		{#if pillarCompetencies.length > 0}
			<section>
				<h3 class="text-sm font-semibold text-base-content/60 uppercase tracking-wider mb-3">
					{pillar.name}
				</h3>
				<ComparisonTable
					ratings={ratings}
					competencies={pillarCompetencies}
					acceptanceLevels={acceptanceLevels}
					{levelDefinitions}
					showRhColumn={true}
				/>
			</section>
		{/if}
	{/each}

	{#if ratings.length === 0}
		<div class="text-center py-8">
			<p class="text-sm text-base-content/40">No hay evaluaciones registradas para {displayName}.</p>
		</div>
	{/if}
</div>
