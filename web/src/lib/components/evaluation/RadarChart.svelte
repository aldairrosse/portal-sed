<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { browser } from '$app/environment';
	import {
		Chart,
		RadarController,
		RadialLinearScale,
		PointElement,
		LineElement,
		Tooltip
	} from 'chart.js';
	import type { ChartDataset } from 'chart.js';
	import type { RadarPillarGroup } from '$lib/types/radar-chart';

	Chart.register(RadarController, RadialLinearScale, PointElement, LineElement, Tooltip);

	interface Props {
		pillarGroups: RadarPillarGroup[];
		employeeName?: string;
	}

	let { pillarGroups, employeeName = 'Empleado' }: Props = $props();

	let canvas: HTMLCanvasElement = $state() as HTMLCanvasElement;
	let chart: Chart<'radar'> | undefined;

	const allCompetencies = $derived(pillarGroups.flatMap((g) => g.competencies));
	const hasSelf = $derived(allCompetencies.some((c) => c.selfRating !== null));
	const hasRh = $derived(allCompetencies.some((c) => c.rhRating !== null));

	// ─── Theme-aware colors ─────────────────────────────────────
	function cssVar(name: string, fallback: string): string {
		if (!browser) return fallback;
		return getComputedStyle(document.documentElement).getPropertyValue(name).trim() || fallback;
	}

	function hexToRgb(hex: string): string {
		const n = parseInt(hex.replace('#', ''), 16);
		return `${(n >> 16) & 255}, ${(n >> 8) & 255}, ${n & 255}`;
	}

	function buildColors() {
		const selfHex = cssVar('--color-radar-self', '#0d9488');
		const rhHex = cssVar('--color-radar-rh', '#f59e0b');
		return {
			self: `rgb(${hexToRgb(selfHex)})`,
			selfBg: `rgba(${hexToRgb(selfHex)}, 0.2)`,
			rh: `rgb(${hexToRgb(rhHex)})`,
			rhBg: `rgba(${hexToRgb(rhHex)}, 0.2)`,
			grid: cssVar('--color-chart-grid', 'rgba(0,0,0,0.08)'),
		};
	}

	let colors = $state(buildColors());

	function pushDatasets(target: ChartDataset<'radar'>[]) {
		if (hasSelf) {
			target.push({
				label: 'Autoevaluación',
				data: allCompetencies.map((c) => c.selfRating),
				backgroundColor: colors.selfBg,
				borderColor: colors.self,
				pointBackgroundColor: colors.self,
				borderWidth: 2,
				pointRadius: 3
			});
		}
		if (hasRh) {
			target.push({
				label: 'RH',
				data: allCompetencies.map((c) => c.rhRating),
				backgroundColor: colors.rhBg,
				borderColor: colors.rh,
				pointBackgroundColor: colors.rh,
				borderWidth: 2,
				pointRadius: 3
			});
		}
	}

	function getTextColor(): string {
		return cssVar('--color-base-content', '#334155');
	}

	// ─── Chart lifecycle ─────────────────────────────────────────
	onMount(() => {
		if (!browser) return;

		chart = new Chart(canvas, {
			type: 'radar',
			data: { labels: [], datasets: [] },
			options: {
				responsive: true,
				maintainAspectRatio: true,
				scales: {
					r: {
						min: 1,
						max: 5,
						ticks: {
							stepSize: 1,
							color: getTextColor(),
							backdropColor: 'transparent',
							callback: (v) => `${v}`
						},
						pointLabels: {
							font: { size: 11 },
							color: getTextColor()
						},
						angleLines: {
							color: colors.grid
						},
						grid: {
							color: colors.grid
						}
					}
				},
				plugins: {
					legend: { display: false },
					tooltip: {
						callbacks: {
							label: (ctx) => {
								const label = ctx.dataset.label ?? '';
								return `${label}: ${ctx.parsed.r}`;
							}
						}
					}
				}
			}
		});

		syncChart();
	});

	onDestroy(() => {
		chart?.destroy();
		chart = undefined;
	});

	function syncChart() {
		if (!chart || !browser || !chart.canvas) return;
		chart.data.labels = allCompetencies.map((c) => c.competencyName);
		chart.data.datasets = [];
		pushDatasets(chart.data.datasets);

		const rScale = chart.options.scales?.r;
		if (rScale) {
			const tc = getTextColor();
			rScale.ticks = { ...rScale.ticks, color: tc, backdropColor: 'transparent' };
			rScale.pointLabels = { ...rScale.pointLabels, color: tc };
			rScale.angleLines = { ...rScale.angleLines, color: colors.grid };
			rScale.grid = { ...rScale.grid, color: colors.grid };
		}

		chart.update('none');
	}

	// Re-sync when data or colors change
	$effect(() => {
		const _ = allCompetencies;
		const _c = colors;
		syncChart();
	});

	// Watch theme changes (live update)
	$effect(() => {
		if (!browser) return;

		const mq = window.matchMedia('(prefers-color-scheme: dark)');
		const handler = () => { colors = buildColors(); };
		mq.addEventListener('change', handler);

		const observer = new MutationObserver(handler);
		observer.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] });

		return () => {
			mq.removeEventListener('change', handler);
			observer.disconnect();
		};
	});
</script>

{#if !hasSelf && !hasRh}
	<div class="text-center py-8" role="status" aria-live="polite">
		<p class="text-sm text-base-content/40">
			No hay datos de competencias para {employeeName}.
		</p>
	</div>
{:else}
	<div class="w-full max-w-lg mx-auto" role="img" aria-label="Gráfica radar de competencias de {employeeName}">
		<canvas bind:this={canvas} class="w-full h-full max-h-[400px]"></canvas>
	</div>

	{#if hasSelf || hasRh}
		<div class="flex justify-center gap-6 mt-4 text-sm" aria-label="Leyenda del radar">
			{#if hasSelf}
				<span class="flex items-center gap-2">
					<span
						class="w-3 h-3 rounded-full"
						style="background-color: {colors.self}"
					></span>
					Autoevaluación
				</span>
			{/if}
			{#if hasRh}
				<span class="flex items-center gap-2">
					<span
						class="w-3 h-3 rounded-full"
						style="background-color: {colors.rh}"
					></span>
					RH
				</span>
			{/if}
		</div>
	{/if}
{/if}
