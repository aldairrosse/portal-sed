<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import type { Chart } from 'chart.js';

	let {
		personal,
		global,
		shared,
	}: {
		personal?: { evaluated: number; total: number };
		global?: { evaluated: number; total: number };
		shared?: { evaluated: number; total: number };
	} = $props();

	let chart: Chart | undefined;
	let canvasEl: HTMLCanvasElement | undefined;

	let colorPersonal = $state('');
	let colorGlobal = $state('');
	let colorShared = $state('');
	let colorGray = $state('');

	let themeObserver: MutationObserver | undefined;
	let themeMq: MediaQueryList | undefined;

	const pct = (t: { evaluated: number; total: number } | undefined) =>
		t && t.total > 0 ? Math.round((t.evaluated / t.total) * 100) : 0;

	const personalPct = $derived(pct(personal));
	const globalPct = $derived(pct(global));
	const sharedPct = $derived(pct(shared));
	const grayPct = $derived(
		Math.max(0, 100 - personalPct - globalPct - sharedPct),
	);

	// Segment order matches the legend: personal, global, shared, sin progreso.
	const segments = $derived([personalPct, globalPct, sharedPct, grayPct]);

	function resolveColors() {
		const styles = getComputedStyle(document.documentElement);
		colorPersonal = styles.getPropertyValue('--color-info').trim();
		colorGlobal = styles.getPropertyValue('--color-primary').trim();
		colorShared = styles.getPropertyValue('--color-warning').trim();
		colorGray = styles.getPropertyValue('--color-base-300').trim();
	}

	function handleThemeChange() {
		resolveColors();
		chart?.update();
	}

	onMount(async () => {
		const { Chart: ChartCtor } = await import('chart.js/auto');
		resolveColors();

		chart = new ChartCtor(canvasEl!, {
			type: 'doughnut',
			data: {
				labels: ['personal', 'global', 'shared', 'sin-progreso'],
				datasets: [
					{
						data: segments,
						backgroundColor: [
							colorPersonal,
							colorGlobal,
							colorShared,
							colorGray,
						],
						borderWidth: 0,
					},
				],
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				cutout: '72%',
				plugins: {
					legend: { display: false },
					tooltip: { enabled: false },
				},
			},
		});

		// Live theme updates (mirror RadarChart): DaisyUI toggles data-theme,
		// system dark mode is covered by the prefers-color-scheme listener.
		themeMq = window.matchMedia('(prefers-color-scheme: dark)');
		themeMq.addEventListener('change', handleThemeChange);

		themeObserver = new MutationObserver(handleThemeChange);
		themeObserver.observe(document.documentElement, {
			attributes: true,
			attributeFilter: ['data-theme'],
		});
	});

	onDestroy(() => {
		themeObserver?.disconnect();
		themeObserver = undefined;
		themeMq?.removeEventListener('change', handleThemeChange);
		themeMq = undefined;
		chart?.destroy();
	});

	$effect(() => {
		if (!chart) return;
		const dataset = chart.data.datasets[0];
		dataset.data = segments;
		dataset.backgroundColor = [
			colorPersonal,
			colorGlobal,
			colorShared,
			colorGray,
		];
		chart.update();
	});
</script>

<div class="grid grid-cols-2 gap-3 items-center">
	<div class="h-28">
		<canvas
			bind:this={canvasEl}
			aria-label="Progreso por tipo de meta"
		></canvas>
	</div>
	<div class="space-y-2">
		{#if personal && personal.total > 0}
			<div class="flex items-start gap-2">
				<span
					class="badge badge-info badge-xs mt-0.5 shrink-0"
				></span>
				<div>
					<p class="text-sm font-bold text-base-content">
						{personal.evaluated} de {personal.total}
					</p>
					<p class="text-xs text-base-content/70">
						metas personales evaluadas
					</p>
				</div>
			</div>
		{/if}
		{#if global && global.total > 0}
			<div class="flex items-start gap-2">
				<span
					class="badge badge-primary badge-xs mt-0.5 shrink-0"
				></span>
				<div>
					<p class="text-sm font-bold text-base-content">
						{global.evaluated} de {global.total}
					</p>
					<p class="text-xs text-base-content/70">
						metas globales evaluadas
					</p>
				</div>
			</div>
		{/if}
		{#if shared && shared.total > 0}
			<div class="flex items-start gap-2">
				<span
					class="badge badge-warning badge-xs mt-0.5 shrink-0"
				></span>
				<div>
					<p class="text-sm font-bold text-base-content">
						{shared.evaluated} de {shared.total}
					</p>
					<p class="text-xs text-base-content/70">
						metas compartidas evaluadas
					</p>
				</div>
			</div>
		{/if}
		{#if grayPct > 0}
			<div class="flex items-start gap-2">
				<span
					class="badge bg-base-300 text-base-content border-transparent badge-xs mt-0.5 shrink-0"
				></span>
				<div>
					<p class="text-sm font-bold text-base-content">
						Sin progreso
					</p>
				</div>
			</div>
		{/if}
	</div>
</div>
