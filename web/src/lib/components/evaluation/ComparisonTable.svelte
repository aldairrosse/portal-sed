<script lang="ts">
	import type { CompetencyRating } from '$lib/types/evaluation-result';
	import type { Competency, LevelDefinition } from '$lib/types/competency';

	interface Props {
		ratings: CompetencyRating[];
		competencies: Competency[];
		acceptanceLevels: Record<string, number>;
		levelDefinitions: LevelDefinition[];
		showRhColumn?: boolean;
	}

	let {
		ratings,
		competencies,
		acceptanceLevels,
		showRhColumn = true,
	}: Props = $props();

	function getRating(competencyId: string): CompetencyRating | undefined {
		return ratings.find((r) => r.competencyId === competencyId);
	}

	function weightedScore(
		selfRating: number | undefined,
		rhRating: number | undefined,
	): number | undefined {
		if (selfRating !== undefined && rhRating !== undefined)
			return selfRating * 0.2 + rhRating * 0.8;
		return rhRating ?? selfRating;
	}

	function getGapClass(
		selfRating: number | undefined,
		rhRating: number | undefined,
		acceptanceLevel: number,
	): string {
		const weighted = weightedScore(selfRating, rhRating);
		if (weighted === undefined) return 'badge-ghost';

		const meetsExpected = weighted >= acceptanceLevel;
		const rhDiff =
			rhRating !== undefined && selfRating !== undefined
				? Math.abs(rhRating - selfRating)
				: 0;

		if (!meetsExpected) return 'badge-error';
		if (showRhColumn && rhRating !== undefined && rhDiff >= 2)
			return 'badge-warning';
		return 'badge-success';
	}

	function getGapLabel(
		selfRating: number | undefined,
		rhRating: number | undefined,
		acceptanceLevel: number,
	): string {
		const weighted = weightedScore(selfRating, rhRating);
		if (weighted === undefined) return '—';

		const rounded = Math.round(weighted * 10) / 10;
		const diff = Math.round((weighted - acceptanceLevel) * 10) / 10;
		const rhDiff =
			rhRating !== undefined && selfRating !== undefined
				? Math.abs(rhRating - selfRating)
				: 0;

		if (diff < 0) return `${diff} (por debajo)`;
		if (showRhColumn && rhRating !== undefined && rhDiff >= 2)
			return `brecha RH ${rhDiff}`;
		if (diff === 0) return '0 (cumple)';
		return `+${diff} (supera · ${rounded})`;
	}
</script>

<div class="w-full max-w-full overflow-x-auto">
	<table class="table table-sm w-full" aria-label="Comparación de evaluaciones">
		<thead>
			<tr>
				<th class="text-xs font-semibold text-base-content/60">Competencia</th>
				<th class="text-xs font-semibold text-base-content/60 text-center"
					>Autoevaluación</th
				>
				{#if showRhColumn}
					<th class="text-xs font-semibold text-base-content/60 text-center"
						>Evaluación</th
					>
				{/if}
				<th class="text-xs font-semibold text-base-content/60 text-center"
					>Nivel esperado</th
				>
				<th class="text-xs font-semibold text-base-content/60 text-center"
					>Brecha</th
				>
			</tr>
		</thead>
		<tbody>
			{#each competencies as competency (competency.id)}
				{@const rating = getRating(competency.id)}
				{@const selfVal = rating?.selfRating}
				{@const rhVal = rating?.rhRating}
				{@const acceptance = acceptanceLevels[competency.id] ?? 3}
				{@const gapClass = getGapClass(selfVal, rhVal, acceptance)}
				{@const gapLabel = getGapLabel(selfVal, rhVal, acceptance)}
				<tr>
					<td class="text-sm font-medium text-base-content"
						>{competency.name}</td
					>
					<td class="text-center">
						{#if selfVal}
							<span class="badge badge-sm">{selfVal}</span>
						{:else}
							<span class="text-base-content/30 text-xs">—</span>
						{/if}
					</td>
					{#if showRhColumn}
						<td class="text-center">
							{#if rhVal}
								<span class="badge badge-sm">{rhVal}</span>
							{:else}
								<span class="text-base-content/30 text-xs">—</span>
							{/if}
						</td>
					{/if}
					<td class="text-center">
						<span class="badge badge-ghost badge-sm">{acceptance}</span>
					</td>
					<td class="text-center">
						<span class="badge badge-sm {gapClass}">{gapLabel}</span>
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
</div>
