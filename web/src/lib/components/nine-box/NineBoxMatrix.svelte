<script lang="ts">
	import { computeQuadrant } from '$lib/stores/nineBoxStore.svelte';
	import type { NineBoxEntry, NineBoxScale, NineBoxQuadrantDef } from '$lib/types/nine-box';

	interface Props {
		entries: NineBoxEntry[];
		quadrantDefs: NineBoxQuadrantDef[];
		onCellClick: (entries: NineBoxEntry[], perf: NineBoxScale, pot: NineBoxScale) => void;
	}

	let { entries, quadrantDefs, onCellClick }: Props = $props();

	const perfValues: NineBoxScale[] = [1, 2, 3, 4, 5, 6, 7, 8, 9];
	const potValues: NineBoxScale[] = [9, 8, 7, 6, 5, 4, 3, 2, 1];

	// ─── Reactive derived data ─────────────────────────────────────────────────

	const entriesByCell = $derived.by(() => {
		// eslint-disable-next-line svelte/prefer-svelte-reactivity
		const map = new Map<string, NineBoxEntry[]>();
		for (const entry of entries) {
			const key = `${entry.performance}-${entry.potential}`;
			const list = map.get(key) ?? [];
			list.push(entry);
			map.set(key, list);
		}
		return map;
	});

	function getCellEntries(perf: NineBoxScale, pot: NineBoxScale): NineBoxEntry[] {
		return entriesByCell.get(`${perf}-${pot}`) ?? [];
	}

	function getCellCount(perf: NineBoxScale, pot: NineBoxScale): number {
		return getCellEntries(perf, pot).length;
	}

	function getCellQuadrantDef(perf: NineBoxScale, pot: NineBoxScale): NineBoxQuadrantDef | undefined {
		const q = computeQuadrant(perf, pot);
		return quadrantDefs.find((d) => d.id === q);
	}

	// ─── Keyboard navigation ───────────────────────────────────────────────────

	let activePerf = $state<NineBoxScale>(5);
	let activePot = $state<NineBoxScale>(5);

	function handleKeydown(e: KeyboardEvent) {
		let perf = activePerf;
		let pot = activePot;
		let moved = false;

		switch (e.key) {
			case 'ArrowUp':
				if (pot < 9) { pot = (pot + 1) as NineBoxScale; moved = true; }
				break;
			case 'ArrowDown':
				if (pot > 1) { pot = (pot - 1) as NineBoxScale; moved = true; }
				break;
			case 'ArrowLeft':
				if (perf > 1) { perf = (perf - 1) as NineBoxScale; moved = true; }
				break;
			case 'ArrowRight':
				if (perf < 9) { perf = (perf + 1) as NineBoxScale; moved = true; }
				break;
		case 'Enter':
		case ' ':
			e.preventDefault();
			{const cellEntries = getCellEntries(activePerf, activePot);
			if (cellEntries.length > 0) onCellClick(cellEntries, activePerf, activePot);}
			return;
			default:
				return;
		}

		if (moved) {
			e.preventDefault();
			activePerf = perf;
			activePot = pot;
			document.getElementById(`nb-cell-${perf}-${pot}`)?.focus();
		}
	}
</script>

<div
	role="grid"
	aria-label="Matriz 9×9 de desempeño y potencial"
	aria-rowcount="9"
	aria-colcount="9"
	class="grid gap-px w-full max-w-[40rem] mx-auto select-none"
	style="grid-template-columns: 2rem 1rem 2.25rem repeat(9, 1fr)"
	tabindex="0"
	onkeydown={handleKeydown}
>
	<!-- Row 1: Corner + empty top cells -->
	<div role="presentation" class="min-h-[1.5rem]"></div>
	<div role="presentation" class="min-h-[1.5rem]"></div>
	<div role="presentation" class="min-h-[1.5rem]"></div>
	{#each perfValues as _perf (_perf)}
		<div role="presentation"></div>
	{/each}

	<!-- Data rows: Potential 9 → 1 -->
	{#each potValues as pot, i (pot)}
		<!-- Potential label (only on first row) -->
		{#if i === 0}
			<div
				role="presentation"
				class="row-span-9 flex items-center justify-center"
			>
				<span
					class="text-[10px] font-medium text-base-content/40 whitespace-nowrap"
					style="writing-mode: vertical-rl; transform: rotate(180deg);"
				>
					Potencial
				</span>
			</div>
		{/if}

		<!-- Potential level labels: Alto (rows 0-2), Medio (3-5), Bajo (6-8) -->
		{#if i === 0}
			<!-- Alto: spans rows 0-2 -->
			<div
				role="presentation"
				class="row-span-3 flex items-center justify-center"
			>
				<span
					class="text-[10px] font-medium text-base-content/50 whitespace-nowrap"
					style="writing-mode: vertical-rl; transform: rotate(180deg);"
				>
					Alto
				</span>
			</div>
		{:else if i === 3}
			<!-- Medio: spans rows 3-5 -->
			<div
				role="presentation"
				class="row-span-3 flex items-center justify-center"
			>
				<span
					class="text-[10px] font-medium text-base-content/50 whitespace-nowrap"
					style="writing-mode: vertical-rl; transform: rotate(180deg);"
				>
					Medio
				</span>
			</div>
		{:else if i === 6}
			<!-- Bajo: spans rows 6-8 -->
			<div
				role="presentation"
				class="row-span-3 flex items-center justify-center"
			>
				<span
					class="text-[10px] font-medium text-base-content/50 whitespace-nowrap"
					style="writing-mode: vertical-rl; transform: rotate(180deg);"
				>
					Bajo
				</span>
			</div>
		{/if}

		<!-- Row header -->
		<div role="rowheader" class="text-right text-[11px] font-medium text-base-content/50 leading-none pr-1 self-center">
			{pot}
		</div>

		<!-- 9 cells per row -->
		{#each perfValues as perf (perf)}
			{@const count = getCellCount(perf, pot)}
			{@const cellEntries = getCellEntries(perf, pot)}
			{@const qDef = getCellQuadrantDef(perf, pot)}
			{@const isActive = activePerf === perf && activePot === pot}
			{@const cellId = `nb-cell-${perf}-${pot}`}
			<button
				type="button"
				id={cellId}
				role="gridcell"
				tabindex={isActive ? 0 : -1}
				aria-rowindex={10 - pot}
				aria-colindex={perf}
				aria-label="Desempeño {perf}, Potencial {pot}, {count} empleados"
				class="relative flex items-center justify-center rounded-sm transition-all duration-100 cursor-pointer hover:brightness-95 active:scale-95 {qDef?.colorClass ?? 'bg-base-200'} {isActive ? 'ring-2 ring-primary ring-offset-1' : ''} min-h-[2.5rem]"
				onclick={() => {
					if (cellEntries.length > 0) onCellClick(cellEntries, perf, pot);
				}}
				onfocus={() => {
					activePerf = perf;
					activePot = pot;
				}}
			>
				{#if count > 0}
					<span class="text-sm font-bold text-base-content/70">{count}</span>
				{/if}
			</button>
		{/each}
	{/each}

	<!-- Bottom: Column numbers + zones + label -->
	<div role="presentation" class="min-h-[1.5rem]"></div>
	<div role="presentation" class="min-h-[1.5rem]"></div>
	<div role="presentation" class="min-h-[1.5rem]"></div>
	{#each perfValues as perf (perf)}
		<div role="columnheader" class="text-center text-[11px] font-medium text-base-content/50 leading-none self-start pt-1">
			{perf}
		</div>
	{/each}

	<!-- Performance zones: Bajo / Medio / Alto -->
	<div role="presentation" class="min-h-[1.5rem]"></div>
	<div role="presentation" class="min-h-[1.5rem]"></div>
	<div role="presentation" class="min-h-[1.5rem]"></div>
	<div role="presentation" class="col-span-3 text-center text-[10px] font-medium text-base-content/50 pt-0.5">
		Bajo
	</div>
	<div role="presentation" class="col-span-3 text-center text-[10px] font-medium text-base-content/50 pt-0.5">
		Medio
	</div>
	<div role="presentation" class="col-span-3 text-center text-[10px] font-medium text-base-content/50 pt-0.5">
		Alto
	</div>

	<!-- Bottom row: Performance axis label -->
	<div role="presentation" class="min-h-[1.5rem]"></div>
	<div role="presentation" class="min-h-[1.5rem]"></div>
	<div role="presentation" class="min-h-[1.5rem]"></div>
	<div role="presentation" class="col-span-9 text-center text-[10px] text-base-content/40 pt-1">
		Desempeño
	</div>
</div>
