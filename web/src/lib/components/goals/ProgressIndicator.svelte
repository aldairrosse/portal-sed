<script lang="ts">
	interface Props {
		value: number;
		max?: number;
		label?: string;
		color?: 'primary' | 'success' | 'warning' | 'error';
	}

	let { value, max = 100, label, color }: Props = $props();

	let percentage = $derived(max > 0 ? Math.min((value / max) * 100, 100) : 0);

	let resolvedColor = $derived(
		color ?? (percentage < 40 ? 'error' : percentage < 80 ? 'warning' : 'success')
	);

	let progressClass = $derived(`progress progress-${resolvedColor}`);
	let badgeClass = $derived(`badge badge-sm badge-${resolvedColor}`);
	let displayValue = $derived(Math.round(percentage));
</script>

<div class="flex items-center gap-2">
	<progress class="{progressClass} w-24" value={percentage} max="100"></progress>
	<span class="{badgeClass}">{displayValue}%</span>
	{#if label}
		<span class="text-xs text-base-content/60">{label}</span>
	{/if}
</div>
