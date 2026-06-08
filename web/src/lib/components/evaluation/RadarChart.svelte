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
	const activeDatasets = $derived<ChartDataset<'radar'>[]>(buildDatasets());

	function buildDatasets(): ChartDataset<'radar'>[] {
		const datasets: ChartDataset<'radar'>[] = [];
		if (hasSelf) {
			datasets.push({
				label: 'Autoevaluación',
				data: allCompetencies.map((c) => c.selfRating),
				backgroundColor: 'rgba(13, 148, 136, 0.15)',
				borderColor: 'rgb(13, 148, 136)',
				pointBackgroundColor: 'rgb(13, 148, 136)',
				borderWidth: 2,
				pointRadius: 3
			});
		}
		if (hasRh) {
			datasets.push({
				label: 'RH',
				data: allCompetencies.map((c) => c.rhRating),
				backgroundColor: 'rgba(245, 158, 11, 0.15)',
				borderColor: 'rgb(245, 158, 11)',
				pointBackgroundColor: 'rgb(245, 158, 11)',
				borderWidth: 2,
				pointRadius: 3
			});
		}
		return datasets;
	}

	onMount(() => {
		if (!browser) return;

		chart = new Chart(canvas, {
			type: 'radar',
			data: {
				labels: [],
				datasets: []
			},
			options: {
				responsive: true,
				maintainAspectRatio: true,
				scales: {
					r: {
						min: 1,
						max: 5,
						ticks: {
							stepSize: 1,
							callback: (v) => `${v}`
						},
						pointLabels: {
							font: { size: 11 },
							color: '#6b7280'
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
	});

	onDestroy(() => {
		chart?.destroy();
		chart = undefined;
	});

	$effect(() => {
		const _ = pillarGroups;
		if (!chart || !browser) return;

		const all = _.flatMap((g) => g.competencies);
		chart.data.labels = all.map((c) => c.competencyName);

		chart.data.datasets = [];
		if (all.some((c) => c.selfRating !== null)) {
			chart.data.datasets.push({
				label: 'Autoevaluación',
				data: all.map((c) => c.selfRating),
				backgroundColor: 'rgba(13, 148, 136, 0.15)',
				borderColor: 'rgb(13, 148, 136)',
				pointBackgroundColor: 'rgb(13, 148, 136)',
				borderWidth: 2,
				pointRadius: 3
			});
		}
		if (all.some((c) => c.rhRating !== null)) {
			chart.data.datasets.push({
				label: 'RH',
				data: all.map((c) => c.rhRating),
				backgroundColor: 'rgba(245, 158, 11, 0.15)',
				borderColor: 'rgb(245, 158, 11)',
				pointBackgroundColor: 'rgb(245, 158, 11)',
				borderWidth: 2,
				pointRadius: 3
			});
		}

		chart.update('none');
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
		<canvas bind:this={canvas}></canvas>
	</div>

	{#if activeDatasets.length > 0}
		<div class="flex justify-center gap-6 mt-4 text-sm" aria-label="Leyenda del radar">
			{#if hasSelf}
				<span class="flex items-center gap-2">
					<span class="w-3 h-3 rounded-full bg-teal-600"></span>
					Autoevaluación
				</span>
			{/if}
			{#if hasRh}
				<span class="flex items-center gap-2">
					<span class="w-3 h-3 rounded-full bg-amber-500"></span>
					RH
				</span>
			{/if}
		</div>
	{/if}
{/if}
