<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import type { Chart } from 'chart.js';
	import type { ApiCyclePhase } from '$lib/types/cycle';
	import type { HomeGroupProgress } from '$lib/utils/homeProgress';

	let {
		personal,
		global,
		shared,
		total,
		phase,
	}: {
		personal?: HomeGroupProgress;
		global?: HomeGroupProgress;
		shared?: HomeGroupProgress;
		total?: number;
		phase?: ApiCyclePhase;
	} = $props();

	let chart = $state<Chart | undefined>(undefined);
	let canvasEl: HTMLCanvasElement | undefined;

	let colorPersonal = $state('');
	let colorGlobal = $state('');
	let colorShared = $state('');
	let colorGray = $state('');

	let themeObserver: MutationObserver | undefined;
	let themeMq: MediaQueryList | undefined;

	// Donut por aporte: segmento = valor grupo. Gris = restante 100 - total.
	// Si total llega stale (0/NaN al inicio con store vacío) pero los grupos
	// ya tienen valor, se usa la suma de grupos como fallback.
	const groupsSum = $derived(
		(personal?.value ?? 0) + (global?.value ?? 0) + (shared?.value ?? 0),
	);
	const cycleTotal = $derived.by(() => {
		const n = Number(total);
		if (Number.isFinite(n) && n > 0) return Math.min(100, Math.max(0, n));
		const s = Number(groupsSum);
		if (Number.isFinite(s) && s > 0) return Math.min(100, Math.max(0, s));
		return 0;
	});
	const centerInt = $derived(Math.round(cycleTotal));

	const personalPct = $derived(personal?.value ?? 0);
	const globalPct = $derived(global?.value ?? 0);
	const sharedPct = $derived(shared?.value ?? 0);
	const grayPct = $derived(Math.max(0, Math.min(100, 100 - cycleTotal)));

	// Segment order matches the legend: personal, global, shared, sin progreso.
	const segments = $derived([personalPct, globalPct, sharedPct, grayPct]);

	const hasAnyGoals = $derived(
		(personal?.total ?? 0) + (global?.total ?? 0) + (shared?.total ?? 0) > 0,
	);
	const centerLabel = $derived(
		phase === 'cierre' ? 'Cierre final' : 'Avance general',
	);

	function resolveColors() {
		const styles = getComputedStyle(document.documentElement);
		// Personales en cyan fijo para diferenciar de globales (primary azul/índigo).
		colorPersonal = '#06b6d4';
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
		// Leer segments/colores antes del early-return: si el return va
		// primero, el efecto nunca se suscribe a segments y el donut se
		// queda en gris 100% tras loadCycle/loadGoals.
		const data = segments;
		const colors = [colorPersonal, colorGlobal, colorShared, colorGray];
		if (!chart) return;
		const dataset = chart.data.datasets[0];
		dataset.data = [...data];
		dataset.backgroundColor = [...colors];
		chart.update();
	});
</script>

<div class="grid grid-cols-2 gap-3 items-center">
	<div class="relative h-28">
		<canvas bind:this={canvasEl} aria-label="Progreso por tipo de meta"
		></canvas>
		<div
			class="pointer-events-none absolute inset-0 flex flex-col items-center justify-center"
		>
			<span class="text-2xl font-bold leading-none text-base-content"
				>{centerInt}%
			</span>
			<span class="mt-0.5 text-[8px] text-base-content/70">{centerLabel}</span>
		</div>
	</div>
	<div class="space-y-2">
		{#if personal && personal.total > 0}
			<div class="flex items-start gap-2">
				<span class="badge bg-cyan-500 border-transparent text-white badge-xs shrink-0"></span>
				<div class="flex-1">
					<div class="flex items-center justify-between gap-2">
						<p class="text-xs font-bold text-base-content">
							{personal.evaluated} de {personal.total}
						</p>
						<span class="text-xs text-gray-900"
							>{Math.round(personal.value)}%</span
						>
					</div>
					<p class="text-xs text-base-content/70">metas personales evaluadas</p>
				</div>
			</div>
		{/if}
		{#if global && global.total > 0}
			<div class="flex items-start gap-2">
				<span class="badge badge-primary badge-xs shrink-0"></span>
				<div class="flex-1">
					<div class="flex items-center justify-between gap-2">
						<p class="text-xs font-bold text-base-content">
							{global.evaluated} de {global.total}
						</p>
						<span class="text-xs text-gray-900"
							>{Math.round(global.value)}%</span
						>
					</div>
					<p class="text-xs text-base-content/70">metas globales evaluadas</p>
				</div>
			</div>
		{/if}
		{#if shared && shared.total > 0}
			<div class="flex items-start gap-2">
				<span class="badge badge-warning badge-xs shrink-0"></span>
				<div class="flex-1">
					<div class="flex items-center justify-between gap-2">
						<p class="text-xs font-bold text-base-content">
							{shared.evaluated} de {shared.total}
						</p>
						<span class="text-xs text-gray-900"
							>{Math.round(shared.value)}%</span
						>
					</div>
					<p class="text-xs text-base-content/70">metas compartidas evaluadas</p>
				</div>
			</div>
		{/if}
		{#if !hasAnyGoals}
			<div class="flex items-start gap-2">
				<span
					class="badge bg-base-300 text-base-content border-transparent badge-xs shrink-0"
				></span>
				<div>
					<p class="text-xs font-bold text-base-content">Sin metas</p>
				</div>
			</div>
		{:else if grayPct === 100}
			<div class="flex items-start gap-2">
				<span
					class="badge bg-base-300 text-base-content border-transparent badge-xs shrink-0"
				></span>
				<div>
					<p class="text-xs font-bold text-base-content">Sin progreso</p>
				</div>
			</div>
		{/if}
	</div>
</div>
