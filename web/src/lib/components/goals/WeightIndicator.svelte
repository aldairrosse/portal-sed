<script lang="ts">
	interface Props {
		current: number;
		label: string;
	}

	let { current, label }: Props = $props();

	const EPSILON = 0.01;

	const isExact = $derived(Math.abs(current - 100) <= EPSILON);
	const isOver = $derived(current > 100 + EPSILON);
	const clamped = $derived(Math.min(100, Math.max(0, current)));

	const barColor = $derived(
		isExact ? 'bg-success/60' : isOver ? 'bg-error/60' : 'bg-warning/60'
	);
	const badgeColor = $derived(
		isExact ? 'badge-success' : isOver ? 'badge-error' : 'badge-warning'
	);
</script>

<div class="flex items-center gap-3" role="group" aria-label={label}>
	<div class="flex-1">
		<div
			class="progress h-3 rounded-full bg-base-300"
			role="progressbar"
			aria-valuenow={current}
			aria-valuemin={0}
			aria-valuemax={100}
			aria-label="{label}: {current}%"
		>
			<div
				class="h-full rounded-full transition-all duration-300 {barColor}"
				style="width: {clamped}%"
			></div>
		</div>
	</div>
	<span
		class="badge font-mono text-sm font-semibold transition-colors duration-300 {badgeColor}"
	>
		{current.toFixed(1)}%
	</span>
</div>
