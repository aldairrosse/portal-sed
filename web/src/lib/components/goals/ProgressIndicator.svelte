<script lang="ts">
	interface Props {
		value: number;
		max?: number;
		label?: string;
		color?: 'primary' | 'success' | 'warning' | 'error';
		wide?: boolean;
		decimals?: number;
	}

	let { value, max = 100, label, color, wide, decimals = 0 }: Props = $props();

	let rawPercentage = $derived(max > 0 ? (value / max) * 100 : 0);
	let percentage = $derived(Math.min(rawPercentage, 100));

	let resolvedColor = $derived(
		color ?? (percentage < 40 ? 'error' : percentage < 80 ? 'warning' : 'success')
	);

	let progressClass = $derived(
		`progress opacity-70 ${wide ? 'flex-1' : 'w-24'} ` +
		(resolvedColor === 'error' ? 'progress-error' : resolvedColor === 'warning' ? 'progress-warning' : resolvedColor === 'success' ? 'progress-success' : 'progress-primary')
	);
	let badgeClass = $derived(
		`badge badge-sm ` +
		(resolvedColor === 'error' ? 'badge-error' : resolvedColor === 'warning' ? 'badge-warning' : resolvedColor === 'success' ? 'badge-success' : 'badge-primary')
	);
	// label shows total even when >100, bar is capped to 100
	let displayValue = $derived(decimals > 0 ? rawPercentage.toFixed(decimals) : String(Math.round(rawPercentage)));
</script>

<div class="flex items-center gap-2">
	<progress class={progressClass} value={percentage} max="100"></progress>
	<span class="{badgeClass}">{displayValue}%</span>
	{#if label}
		<span class="text-xs text-base-content/60">{label}</span>
	{/if}
</div>
