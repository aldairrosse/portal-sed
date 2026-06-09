<script lang="ts">
	import ComparisonTable from './ComparisonTable.svelte';
	import RadarChart from './RadarChart.svelte';
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
	import type { RadarPillarGroup } from '$lib/types/radar-chart';

	interface Props {
		employeeId: string;
		employeeName?: string;
		activeTab: 'radar' | 'table';
	}

	let { employeeId, employeeName = '', activeTab }: Props = $props();

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

	const pillarGroups: RadarPillarGroup[] = $derived.by(() => {
		return pillars
			.map((pillar) => {
				const comps = getCompetenciesByPillar(pillar.id);
				const competencies = comps.map((c) => {
					const r = ratings.find((r) => r.competencyId === c.id);
					return {
						competencyId: c.id,
						competencyName: c.name,
						selfRating: r?.selfRating ?? null,
						rhRating: r?.rhRating ?? null
					};
				});
				return { pillarId: pillar.id, pillarName: pillar.name, competencies };
			})
			.filter((g) => g.competencies.length > 0);
	});
</script>

<div class="flex flex-col gap-6">
	{#if activeTab === 'table'}
		<div id="panel-table" role="tabpanel" aria-labelledby="view-table" class="flex flex-col gap-10">
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
	{:else}
		<div id="panel-radar" role="tabpanel" aria-labelledby="view-radar">
			<RadarChart pillarGroups={pillarGroups} employeeName={displayName} />
		</div>
	{/if}
</div>
