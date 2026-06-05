<script lang="ts">
	import type { Pillar, Competency, LevelDefinition } from '$lib/types/competency';
	import type { CompetencyRating } from '$lib/types/evaluation-result';
	import ScaleRatingSelector from './ScaleRatingSelector.svelte';

	interface Props {
		pillar: Pillar;
		competencies: Competency[];
		ratings: CompetencyRating[];
		levelDefinitions: LevelDefinition[];
		acceptanceLevels: Record<string, number>;
		mode: 'self' | 'rh';
		onRate: (competencyId: string, level: 1 | 2 | 3 | 4 | 5, comment?: string) => void;
		onRhRate?: (competencyId: string, level: 1 | 2 | 3 | 4 | 5, comment?: string) => void;
		disabled?: boolean;
		showCommentInput?: boolean;
	}

	let {
		pillar,
		competencies,
		ratings,
		levelDefinitions,
		acceptanceLevels,
		mode,
		onRate,
		onRhRate,
		disabled = false,
		showCommentInput = true
	}: Props = $props();

	function getRating(competencyId: string): CompetencyRating | undefined {
		return ratings.find((r) => r.competencyId === competencyId);
	}

	function handleRatingChange(competencyId: string, level: 1 | 2 | 3 | 4 | 5) {
		if (mode === 'rh') {
			onRhRate?.(competencyId, level);
		} else {
			onRate(competencyId, level);
		}
	}

	function handleCommentChange(competencyId: string, comment: string) {
		onRate(competencyId, getRating(competencyId)?.selfRating ?? 1, comment);
	}
</script>

<div class="card bg-base-100 border border-base-300">
	<div class="card-body px-4 py-4">
		<h3 class="text-base font-semibold text-base-content mb-1">{pillar.name}</h3>
		<p class="text-xs text-base-content/50 mb-4">{pillar.description}</p>

		<div class="flex flex-col gap-5">
			{#each competencies as competency (competency.id)}
				{@const rating = getRating(competency.id)}
				{@const acceptanceLevel = acceptanceLevels[competency.id]}
				<div class="border-t border-base-200 pt-3 first:border-t-0 first:pt-0">
					<div class="flex items-center gap-3">
						<!-- Name + description (left) -->
						<div class="flex-1 min-w-0">
							<p class="text-sm font-medium text-base-content">{competency.name}</p>
							<p class="text-xs text-base-content/50">{competency.description}</p>
						</div>
						<!-- Rating selector (right) -->
						<div class="shrink-0">
							<ScaleRatingSelector
								value={mode === 'rh' ? rating?.rhRating : rating?.selfRating}
								{acceptanceLevel}
								{disabled}
								{levelDefinitions}
								onChange={(level) => handleRatingChange(competency.id, level)}
							/>
						</div>
						{#if mode === 'rh' && rating?.selfRating}
							<span class="badge badge-ghost badge-sm shrink-0" title="Autoevaluación">
								Auto: {rating.selfRating}
							</span>
						{/if}
					</div>

					{#if showCommentInput}
						<textarea
							class="textarea textarea-bordered textarea-xs w-full mt-2"
							placeholder={mode === 'rh' ? 'Comentario RH (opcional)' : 'Comentario personal (opcional)'}
							value={mode === 'rh' ? (rating?.rhComment ?? '') : (rating?.selfComment ?? '')}
							oninput={(e) => {
								if (mode === 'self') {
									handleCommentChange(competency.id, (e.target as HTMLTextAreaElement).value);
								} else {
									onRhRate?.(competency.id, rating?.rhRating ?? 3, (e.target as HTMLTextAreaElement).value);
								}
							}}
							disabled={disabled}
							rows="2"
							aria-label={mode === 'rh' ? 'Comentario RH' : 'Comentario personal'}
						></textarea>
					{/if}
				</div>
			{/each}
		</div>
	</div>
</div>
