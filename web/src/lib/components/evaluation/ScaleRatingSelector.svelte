<script lang="ts">
	import type { LevelDefinition } from '$lib/types/competency';

	interface Props {
		value?: 1 | 2 | 3 | 4 | 5;
		acceptanceLevel?: number;
		disabled?: boolean;
		levelDefinitions: LevelDefinition[];
		onChange: (level: 1 | 2 | 3 | 4 | 5) => void;
	}

	let { value, acceptanceLevel, disabled = false, levelDefinitions, onChange }: Props = $props();

	const levels: (1 | 2 | 3 | 4 | 5)[] = [1, 2, 3, 4, 5];
</script>

<div class="flex flex-col gap-1.5">
	<div class="flex flex-wrap gap-1" role="radiogroup" aria-label="Nivel de calificación">
		{#each levels as level (level)}
			{@const def = levelDefinitions.find((ld) => ld.level === level)}
			<button
				class="btn btn-sm min-w-[3.5rem] {value === level
					? 'btn-primary'
					: 'btn-ghost border border-base-300'} {acceptanceLevel === level
					? 'ring-2 ring-info/40'
					: ''}"
				role="radio"
				aria-checked={value === level}
				aria-label="{level} - {def?.label ?? ''}"
				disabled={disabled}
				onclick={() => onChange(level)}
			>
				<span class="font-bold text-xs">{level}</span>
			</button>
		{/each}
	</div>
	<div class="flex flex-wrap gap-x-4 gap-y-0.5 text-[11px] text-base-content/50">
		{#each levels as level (level)}
			{@const def = levelDefinitions.find((ld) => ld.level === level)}
			<span class="min-w-[3.5rem] text-center">{def?.label ?? ''}</span>
		{/each}
	</div>
	{#if acceptanceLevel}
		<span class="text-xs text-info/70 font-medium">
			Mínimo esperado: {acceptanceLevel} - {levelDefinitions.find((ld) => ld.level === acceptanceLevel)?.label ?? ''}
		</span>
	{/if}
</div>
