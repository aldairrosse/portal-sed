<script lang="ts">
    import type { LevelDefinition } from "$lib/types/competency";

    interface Props {
        value?: 1 | 2 | 3 | 4 | 5;
        acceptanceLevel?: number;
        disabled?: boolean;
        levelDefinitions: LevelDefinition[];
        onChange: (level: 1 | 2 | 3 | 4 | 5) => void;
    }

    let {
        value,
        acceptanceLevel,
        disabled = false,
        levelDefinitions,
        onChange,
    }: Props = $props();

    const levels: (1 | 2 | 3 | 4 | 5)[] = [1, 2, 3, 4, 5];
</script>

<div class="flex flex-col gap-1.5">
    <div
        class="flex gap-4"
        role="radiogroup"
        aria-label="Nivel de calificación"
    >
        {#each levels as level (level)}
            {@const def = levelDefinitions.find((ld) => ld.level === level)}

            <diV class="flex flex-col items-center gap-0.5">
                <button
                    class="btn btn-sm min-w-[2.75rem] justify-center {value ===
                    level
                        ? 'btn-primary'
                        : 'btn-ghost border border-base-300'} {acceptanceLevel ===
                    level
                        ? 'ring-2 ring-info/40'
                        : ''}"
                    role="radio"
                    aria-checked={value === level}
                    aria-label="{level} - {def?.label ?? ''}"
                    {disabled}
                    onclick={() => onChange(level)}
                >
                    <span class="font-bold text-xs">{level} </span>
                </button>
                <span class="text-xs text-base-content/50 text-center">
                    {def?.label}
                </span>
            </diV>
        {/each}
    </div>
</div>
