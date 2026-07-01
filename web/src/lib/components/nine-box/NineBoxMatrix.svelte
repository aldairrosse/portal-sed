<script lang="ts">
	import { isLoading, getError, reload } from '$lib/stores/nineBoxStore.svelte';
	import type { NineBoxEntry, NineBoxTier, NineBoxQuadrantDef } from '$lib/types/nine-box';
	import PageSkeleton from '$lib/components/ui/PageSkeleton.svelte';
	import ErrorState from '$lib/components/ui/ErrorState.svelte';
    import { SvelteMap } from 'svelte/reactivity';

	interface Props {
		entries: NineBoxEntry[];
		quadrantDefs: NineBoxQuadrantDef[];
		onCellClick: (entries: NineBoxEntry[], perfTier: NineBoxTier, potTier: NineBoxTier) => void;
	}

	let { entries, quadrantDefs, onCellClick }: Props = $props();

	let loading = $derived(isLoading());
	let error = $derived(getError());

	const perfTiers: NineBoxTier[] = [1, 2, 3];
	const potTiers: NineBoxTier[] = [3, 2, 1];

	const PERF_LABELS: Record<number, string> = { 1: 'Bajo', 2: 'Medio', 3: 'Alto' };
	const POT_LABELS: Record<number, string> = { 3: 'Alto', 2: 'Medio', 1: 'Bajo' };

	function getQuadrantNumber(perfTier: number, potTier: number): number {
		return (potTier - 1) * 3 + perfTier;
	}

	const entriesByQuadrant = $derived.by(() => {
		const map = new SvelteMap<number, NineBoxEntry[]>();
		for (const entry of entries) {
			const q = entry.quadrant;
			const list = map.get(q) ?? [];
			list.push(entry);
			map.set(q, list);
		}
		return map;
	});

	function getQuadrantEntries(quadrant: number): NineBoxEntry[] {
		return entriesByQuadrant.get(quadrant) ?? [];
	}

	function getQuadrantDef(quadrant: number): NineBoxQuadrantDef | undefined {
		return quadrantDefs.find((d) => d.quadrant === quadrant);
	}

	// ─── WCAG: Luminance-based text color ─────────────────────────────────────

	/** Relative luminance of an sRGB hex color, per WCAG 2.1 formula. */
	function relativeLuminance(hex: string): number {
		const h = hex.replace('#', '');
		if (h.length < 6) return 0;
		const r = parseInt(h.substring(0, 2), 16) / 255;
		const g = parseInt(h.substring(2, 4), 16) / 255;
		const b = parseInt(h.substring(4, 6), 16) / 255;
		const lin = (c: number) => (c <= 0.04045 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4);
		return 0.2126 * lin(r) + 0.7152 * lin(g) + 0.0722 * lin(b);
	}

	/** Returns 'text-white' or 'text-black' based on WCAG 2.1 contrast. */
	function textColorForBg(hex: string | undefined): string {
		if (!hex) return 'text-white';
		const lum = relativeLuminance(hex);
		// Luminance threshold: light bg → black text, dark bg → white text
		return lum > 0.35 ? 'text-black' : 'text-white';
	}

	/** Badge class matching the text color for contrast. */
	function badgeClassForBg(hex: string | undefined): string {
		if (!hex) return 'badge-ghost text-white border-white/40';
		const lum = relativeLuminance(hex);
		if (lum > 0.35) {
			return 'badge-ghost text-black/70 border-black/30';
		}
		return 'badge-ghost text-white border-white/40';
	}

	// ─── Keyboard navigation (WCAG roving tabindex) ───────────────────────────

	let activePerf = $state<NineBoxTier>(2);
	let activePot = $state<NineBoxTier>(2);

	/** The aria-activedescendant value for the grid container. */
	let activeDescendantId = $derived(`nb-cell-${getQuadrantNumber(activePerf, activePot)}`);

	function handleKeydown(e: KeyboardEvent) {
		let perf = activePerf;
		let pot = activePot;
		let moved = false;

		switch (e.key) {
			case 'ArrowUp':
				if (pot < 3) { pot = (pot + 1) as NineBoxTier; moved = true; }
				break;
			case 'ArrowDown':
				if (pot > 1) { pot = (pot - 1) as NineBoxTier; moved = true; }
				break;
			case 'ArrowLeft':
				if (perf > 1) { perf = (perf - 1) as NineBoxTier; moved = true; }
				break;
			case 'ArrowRight':
				if (perf < 3) { perf = (perf + 1) as NineBoxTier; moved = true; }
				break;
			case 'Enter':
			case ' ':
				e.preventDefault();
				{
					const q = getQuadrantNumber(activePerf, activePot);
					const cellEntries = getQuadrantEntries(q);
					if (cellEntries.length > 0) onCellClick(cellEntries, activePerf, activePot);
				}
				return;
			default:
				return;
		}

		if (moved) {
			e.preventDefault();
			activePerf = perf;
			activePot = pot;
			const q = getQuadrantNumber(perf, pot);
			document.getElementById(`nb-cell-${q}`)?.focus();
		}
	}

	// ─── Screen reader live region ────────────────────────────────────────────

	let announcement = $state('');
	let srTimeout: ReturnType<typeof setTimeout> | undefined;

	function announce(msg: string) {
		clearTimeout(srTimeout);
		announcement = msg;
		srTimeout = setTimeout(() => { announcement = ''; }, 3000);
	}
</script>

{#if loading}
	<PageSkeleton variant="card" rows={3} />
{:else if error}
	<ErrorState
		title="Error al cargar la matriz"
		message={error}
		retryLabel="Reintentar"
		onretry={reload}
	/>
{:else}
	<!-- Screen reader live region for announcements -->
	<div
		aria-live="polite"
		aria-atomic="true"
		class="sr-only"
	>{announcement}</div>

	<div
		role="grid"
		aria-label="Matriz 9-Box 3x3"
		aria-activedescendant={activeDescendantId}
		class="grid gap-1 w-full max-w-[36rem] mx-auto select-none outline-none"
		style="grid-template-columns: 2.5rem 1.5rem repeat(3, 1fr);"
		tabindex="0"
		onkeydown={handleKeydown}
		onfocus={() => announce(`Matriz 3x3 cargada. Use flechas para navegar. Celda activa: ${PERF_LABELS[activePerf]} desempeño, ${POT_LABELS[activePot]} potencial.`)}
	>
		<!-- Data rows: Potential 3 → 1 (Alto → Bajo) -->
		{#each potTiers as pot (pot)}
			{@const potLabel = POT_LABELS[pot]}

			{#if pot === 3}
				<!-- Vertical axis label — only on first row, spans 3 rows -->
				<div
					role="presentation"
					class="row-span-3 flex items-center justify-center"
				>
					<span
						class="text-[10px] font-semibold text-base-content/40 whitespace-nowrap"
						style="writing-mode: vertical-rl; transform: rotate(180deg);"
					>
						Potencial
					</span>
				</div>
			{/if}

			<!-- Row header: Pot label (rotated) -->
			<div
				role="rowheader"
				class="flex items-center justify-center self-center h-full"
			>
				<span
					class="text-[10px] font-semibold text-base-content/40 whitespace-nowrap"
					style="writing-mode: vertical-rl; transform: rotate(180deg);"
				>
					{potLabel}
				</span>
			</div>

			<!-- 3 cells per row -->
			{#each perfTiers as perf (perf)}
				{@const quadrant = getQuadrantNumber(perf, pot)}
				{@const qDef = getQuadrantDef(quadrant)}
				{@const cellEntries = getQuadrantEntries(quadrant)}
				{@const count = cellEntries.length}
				{@const isActive = activePerf === perf && activePot === pot}
				{@const cellId = `nb-cell-${quadrant}`}
				{@const textColor = textColorForBg(qDef?.colorHex)}
				{@const badgeClass = badgeClassForBg(qDef?.colorHex)}
				<button
					type="button"
					id={cellId}
					role="gridcell"
					tabindex={isActive ? 0 : -1}
					aria-label="Cuadrante {quadrant}: {POT_LABELS[pot]} potencial, {PERF_LABELS[perf]} desempeño, {count} empleados"
					class="relative flex flex-col items-center justify-center gap-1 rounded-lg transition-all duration-100 cursor-pointer hover:brightness-95 active:scale-[0.98] min-h-[6rem] p-2 {isActive ? 'ring-2 ring-primary ring-offset-2' : ''}"
					style="background-color: {qDef?.colorHex ?? '#6B7280'};"
					onclick={() => {
						if (count > 0) {
							onCellClick(cellEntries, perf, pot);
						}
					}}
					onfocus={() => {
						activePerf = perf;
						activePot = pot;
						announce(`Celda: ${PERF_LABELS[perf]} desempeño, ${POT_LABELS[pot]} potencial. ${count} empleados.`);
					}}
				>
					<span class="text-xs font-semibold {textColor} drop-shadow-sm text-center leading-tight">
						{qDef?.title ?? ''}
					</span>
					{#if count > 0}
						<span class="badge badge-sm {badgeClass}">{count}</span>
					{/if}
				</button>
			{/each}
		{/each}

		<!-- Bottom: Performance axis labels -->
		<div role="presentation" class="min-h-[1.5rem]"></div>
		<div role="presentation" class="min-h-[1.5rem]"></div>
		<div role="columnheader" class="text-center text-[11px] font-medium text-base-content/50 flex items-center justify-center">
			Bajo
		</div>
		<div role="columnheader" class="text-center text-[11px] font-medium text-base-content/50 flex items-center justify-center">
			Medio
		</div>
		<div role="columnheader" class="text-center text-[11px] font-medium text-base-content/50 flex items-center justify-center">
			Alto
		</div>
		
		<!-- Bottom: Performance axis title -->
		<div role="presentation" class="min-h-[2.5rem]"></div>
		<div role="presentation" class="min-h-[2.5rem]"></div>
		<div role="presentation" class="min-h-[2.5rem]"></div>
		<div role="presentation" class="text-center text-[10px] font-semibold text-base-content/40 pt-1 flex items-center justify-center">
			Desempeño
		</div>
		<div role="presentation" class="min-h-[2.5rem]"></div>
	</div>
{/if}
