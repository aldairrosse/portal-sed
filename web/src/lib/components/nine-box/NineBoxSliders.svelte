<script lang="ts">
	import { getQuadrantDefs, computeQuadrant } from '$lib/stores/nineBoxStore.svelte';
	import type { NineBoxScale } from '$lib/types/nine-box';

	interface Props {
		performance: NineBoxScale;
		potential: NineBoxScale;
		onScoreChange: (perf: NineBoxScale, pot: NineBoxScale) => void;
		disabled?: boolean;
	}

	let { performance, potential, onScoreChange, disabled = false }: Props = $props();

	const quadrantDefs = $derived(getQuadrantDefs());
	const quadrantLabel = $derived.by(() => {
		const q = computeQuadrant(performance, potential);
		return quadrantDefs.find((d) => d.id === q)?.label ?? q;
	});

	function handlePerfChange(e: Event) {
		const target = e.target as HTMLInputElement;
		onScoreChange(Number(target.value) as NineBoxScale, potential);
	}

	function handlePotChange(e: Event) {
		const target = e.target as HTMLInputElement;
		onScoreChange(performance, Number(target.value) as NineBoxScale);
	}
</script>

<div class="flex flex-col gap-5 p-4 bg-base-200 rounded-xl">
	<h4 class="text-sm font-semibold text-base-content/70">Ajustar scores</h4>

	<!-- Performance slider -->
	<div class="flex flex-col gap-1.5">
		<div class="flex items-center justify-between">
			<label for="nb-slider-perf" class="text-xs font-medium text-base-content/60">Desempeño</label>
			<span class="text-sm font-bold tabular-nums text-base-content">{performance}</span>
		</div>
		<input
			id="nb-slider-perf"
			type="range"
			min="1"
			max="9"
			step="1"
			bind:value={performance}
			disabled={disabled}
			class="range range-sm range-primary w-full"
			aria-label="Desempeño: {performance} de 9"
			oninput={handlePerfChange}
		/>
		<div class="flex justify-between text-[10px] text-base-content/30 px-0.5">
			<span>Muy bajo</span>
			<span>Moderado</span>
			<span>Excepcional</span>
		</div>
	</div>

	<!-- Potential slider -->
	<div class="flex flex-col gap-1.5">
		<div class="flex items-center justify-between">
			<label for="nb-slider-pot" class="text-xs font-medium text-base-content/60">Potencial</label>
			<span class="text-sm font-bold tabular-nums text-base-content">{potential}</span>
		</div>
		<input
			id="nb-slider-pot"
			type="range"
			min="1"
			max="9"
			step="1"
			bind:value={potential}
			disabled={disabled}
			class="range range-sm range-secondary w-full"
			aria-label="Potencial: {potential} de 9"
			oninput={handlePotChange}
		/>
		<div class="flex justify-between text-[10px] text-base-content/30 px-0.5">
			<span>Muy bajo</span>
			<span>Moderado</span>
			<span>Excepcional</span>
		</div>
	</div>

	<!-- Current quadrant preview -->
	<div class="text-center mt-1">
		<span class="text-xs text-base-content/40">Cuadrante actual: </span>
		<span class="badge badge-sm font-medium">{quadrantLabel}</span>
	</div>
</div>
