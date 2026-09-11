<script lang="ts">
    import type { LevelDefinition } from "$lib/types/competency";

    interface Props {
        value?: 1 | 2 | 3 | 4 | 5;
        acceptanceLevel?: number;
        disabled?: boolean;
        levelDefinitions: LevelDefinition[];
        onChange: (level: 1 | 2 | 3 | 4 | 5) => void;
        readOnlySingle?: boolean;
        required?: boolean;
        invalid?: boolean;
        attempted?: boolean;
        tooltips?: Record<number, string> | string[];
    }

    let {
        value,
        acceptanceLevel,
        disabled = false,
        levelDefinitions,
        onChange,
        readOnlySingle = false,
        required = false,
        invalid = false,
        attempted = false,
        tooltips,
    }: Props = $props();

    const levels: (1 | 2 | 3 | 4 | 5)[] = [1, 2, 3, 4, 5];
    const singleLevel = $derived(readOnlySingle && value != null ? value : null);
    const showError = $derived(invalid && attempted);

    function tooltipFor(level: 1 | 2 | 3 | 4 | 5): string {
        if (!tooltips) return '';
        return Array.isArray(tooltips) ? (tooltips[level - 1] ?? '') : (tooltips[level] ?? '');
    }
</script>

<div class="flex items-center gap-2">
    {#if singleLevel == null}
    <span class="text-xs shrink-0 {showError ? 'text-error' : 'text-base-content/50'}" role={showError ? 'alert' : undefined}>Selecciona 1-5</span>
    {/if}
    <div
        class="flex gap-4"
        role="radiogroup"
        aria-label="Nivel de calificación"
        aria-required={required || undefined}
        aria-invalid={invalid || undefined}
    >
        {#if singleLevel != null}
            {@const def = levelDefinitions.find((ld) => ld.level === singleLevel)}
            {@const singleTip = tooltipFor(singleLevel)}
            <div class="flex flex-col items-center gap-0.5 {singleTip ? 'tooltip tooltip-top tooltip-end' : ''}">
                {#if singleTip}<div class="tooltip-content max-w-[280px] whitespace-pre-wrap h-auto text-left text-wrap p-2 text-xs">{singleTip}</div>{/if}
                <button
                    class="btn btn-sm min-w-[2.75rem] justify-center btn-primary cursor-default {acceptanceLevel ===
                    singleLevel
                        ? 'ring-2 ring-info/40'
                        : ''}"
                    role="radio"
                    aria-checked={true}
                    aria-label="{singleLevel} - {def?.label ?? ''}"
                >
                    <span class="font-bold text-xs">{singleLevel} </span>
                </button>
                <span class="text-xs text-base-content/50 text-center">
                    {def?.label}
                </span>
            </div>
        {:else}
        {#each levels as level (level)}
            {@const def = levelDefinitions.find((ld) => ld.level === level)}
            {@const tip = tooltipFor(level)}

            <div class="flex flex-col items-center gap-0.5 {tip ? 'tooltip tooltip-top tooltip-end' : ''}">
                {#if tip}<div class="tooltip-content max-w-[280px] whitespace-pre-wrap h-auto text-left text-wrap p-2 text-xs">{tip}</div>{/if}
                <button
                    class="btn btn-sm min-w-[2.75rem] justify-center {value ===
                    level
                        ? 'btn-primary'
                        : 'btn-ghost border border-base-300'} {acceptanceLevel ===
                    level
                        ? 'ring-2 ring-info/40'
                        : ''} {showError && value !== level ? 'border-error' : ''}"
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
            </div>
        {/each}
        {/if}
    </div>
</div>
