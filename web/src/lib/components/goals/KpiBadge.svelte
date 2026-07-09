<script lang="ts">
	import type { KPI } from '$lib/types/goal';

	interface Props {
		kpi: KPI;
	}

	let { kpi }: Props = $props();

	let hasData = $derived(kpi.currentValue !== undefined && kpi.currentValue !== null);
	let hasTarget = $derived(kpi.targetValue !== undefined && kpi.targetValue !== null && kpi.targetValue > 0);

	let pct = $derived.by(() => {
		if (!hasData || !hasTarget) return undefined;
		if (kpi.direction === 'ascendente') {
			return Math.min(Math.max((kpi.currentValue! / kpi.targetValue!) * 100, 0), 100);
		}
		// descendente: inverse ratio (closer to zero = better)
		return Math.min(Math.max((1 - kpi.currentValue! / kpi.targetValue!) * 100, 0), 100);
	});

	let delta = $derived.by(() => {
		if (!hasData || !hasTarget) return null;
		const val = kpi.direction === 'ascendente'
			? kpi.currentValue! - kpi.targetValue!
			: kpi.targetValue! - kpi.currentValue!;
		return {
			value: val,
			type: val > 0 ? 'positive' : val < 0 ? 'negative' : 'neutral'
		} as const;
	});
</script>

<span class="badge badge-outline badge-sm gap-1 whitespace-nowrap max-w-[16rem]" title={kpi.description}>
	{kpi.name}
	{#if hasData}
		<span class="text-base-content/40">|</span>
		<span class="font-mono font-semibold">{kpi.currentValue}</span>
		{#if pct !== undefined}
			<span class="badge badge-ghost badge-xs">{Math.round(pct)}%</span>
		{/if}
		{#if delta && delta.type !== 'neutral'}
			<span class="badge badge-xs {delta.type === 'positive' ? 'badge-success' : 'badge-error'}">
				{delta.type === 'positive' ? '+' : ''}{delta.value}
			</span>
		{/if}
	{/if}
</span>
