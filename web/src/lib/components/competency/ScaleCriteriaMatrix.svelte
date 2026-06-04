<script lang="ts">
	import { Star } from '@lucide/svelte';
	import { getPillars, getCompetencies, getScaleCriteriaForCell, getLevelDefinitions } from '$lib/stores/competencyStore.svelte';

	const PILLAR_COLORS = ['primary', 'secondary', 'accent', 'info', 'success', 'warning'] as const;

	interface Props {
		onEditCell: (competencyId: string, pillarId: string, competencyName: string, pillarName: string) => void;
	}

	let { onEditCell }: Props = $props();

	const pillars = $derived(getPillars());
	const competencies = $derived(getCompetencies());
	const levelDefs = $derived(getLevelDefinitions());

	const levels = [1, 2, 3, 4, 5] as const;

	function getLevelLabel(level: number): string {
		return levelDefs.find((d) => d.level === level)?.label ?? 'Nivel ' + level;
	}

	function getDescriptions(competencyId: string, pillarId: string, level: number): string[] {
		const criteria = getScaleCriteriaForCell(competencyId, pillarId);
		return criteria.filter((c) => c.level === level).map((c) => c.description).filter(Boolean);
	}

	function pillarBadgeClass(index: number): string {
		const color = PILLAR_COLORS[index % PILLAR_COLORS.length];
		return `badge-${color}`;
	}
</script>

{#each pillars as pillar, i (pillar.id)}
	{@const pillarCompetencies = competencies.filter((c) => c.pillarId === pillar.id)}
	<div class="mb-8">
		<div class="flex items-center gap-2 mb-3">
			<div class="badge {pillarBadgeClass(i)} text-sm px-3 py-2">{pillar.name}</div>
		</div>

		<div class="overflow-x-auto rounded-box border border-base-300">
			<table class="table table-zebra" aria-label="Criterios de escala - {pillar.name}">
				<thead>
				<tr>
						<th class="w-48 min-w-[12rem]">Competencia</th>
						{#each levels as level}
							<th class="min-w-[10rem] text-center">
								<span class="inline-flex items-center gap-1">
									<Star class="w-3 h-3" strokeWidth={2} />
									N{level} - {getLevelLabel(level)}
								</span>
							</th>
						{/each}
					</tr>
				</thead>
				<tbody>
					{#each pillarCompetencies as competency (competency.id)}
						<tr>
							<td class="font-medium text-sm">
								<span class="text-base-content">{competency.name}</span>
							</td>
							{#each levels as level}
								{@const texts = getDescriptions(competency.id, pillar.id, level)}
								<td class="p-1.5 align-top">
									<button
										class="w-full min-h-[4rem] rounded-lg border border-base-300 p-2 text-left transition-colors hover:border-primary hover:bg-primary/5 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary"
										onclick={() => onEditCell(competency.id, pillar.id, competency.name, pillar.name)}
										aria-label="Editar criterios de {competency.name} nivel {level} en {pillar.name}"
									>
										{#if texts.length > 0}
											<div class="space-y-1">
												{#each texts as text}
													<span class="text-xs text-base-content/60 leading-tight line-clamp-3 block">{text}</span>
												{/each}
											</div>
										{:else}
											<span class="text-xs italic text-base-content/30">Sin definir</span>
										{/if}
									</button>
								</td>
							{/each}
						</tr>
					{/each}
				</tbody>
			</table>
		</div>

		{#if pillarCompetencies.length === 0}
			<div class="text-center py-8 text-base-content/50 text-sm">
				Este pilar no tiene competencias asignadas.
			</div>
		{/if}
	</div>
{/each}
