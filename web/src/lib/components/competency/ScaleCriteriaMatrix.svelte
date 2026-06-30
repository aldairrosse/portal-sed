<script lang="ts">
	import { Plus, Trash2, Star, X, Save } from '@lucide/svelte';
	import {
		getPillars,
		getCompetencies,
		getScaleCriteria,
		getScaleCriteriaForCell,
		getLevelDefinitions,
		replaceScaleCriteria
	} from '$lib/stores/competencyStore.svelte';
	import type { ScaleCriterion } from '$lib/types/competency';

	const PILLAR_COLORS = ['#D97706', '#7C3AED', '#B45309', '#0369A1'] as const;

	interface Props {
		isAnyInlineEditing?: boolean;
	}

	// eslint-disable-next-line no-useless-assignment
	let { isAnyInlineEditing = $bindable() }: Props = $props();

	const pillars = $derived(getPillars());
	const competencies = $derived(getCompetencies());
	const levelDefs = $derived(getLevelDefinitions());
	const levels = [1, 2, 3, 4, 5] as const;

	// ─── Inline editing state ──────────────────────────────────────────────
	let editingCellKey: string | null = $state(null);
	let editEntries: Array<{ localId: string; serverId: string | null; description: string }> = $state([]);
	let editCellCompetencyId = $state('');
	let editCellLevel = $state<number>(1);
	let tempIdCounter = $state(0);
	let errorMsg = $state('');
	let canAdd = $derived(editEntries.length === 0 || editEntries.every((e) => e.description.trim().length > 0));

	$effect(() => {
		isAnyInlineEditing = editingCellKey !== null;
	});

	function getCellKey(competencyId: string, pillarId: string, level: number): string {
		return `${competencyId}-${pillarId}-${level}`;
	}

	function getLevelLabel(level: number): string {
		return levelDefs.find((d) => d.level === level)?.label ?? 'Nivel ' + level;
	}

	function getLevelCriteria(competencyId: string, pillarId: string, level: number): ScaleCriterion[] {
		return getScaleCriteriaForCell(competencyId, pillarId).filter((c) => c.level === level);
	}

	function pillarBadgeStyle(index: number): string {
		return `background-color: ${PILLAR_COLORS[index % PILLAR_COLORS.length]}; color: #fff`;
	}

	// ─── Inline editing actions ────────────────────────────────────────────

	function startEditing(competencyId: string, pillarId: string, level: number) {
		if (editingCellKey !== null) return; // only one cell at a time
		editingCellKey = getCellKey(competencyId, pillarId, level);
		editCellCompetencyId = competencyId;
		editCellLevel = level;
		const existing = getLevelCriteria(competencyId, pillarId, level);
		editEntries = existing.map((c) => ({
			localId: c.id,
			serverId: c.id,
			description: c.description
		}));
		if (editEntries.length === 0) {
			tempIdCounter = 1;
			editEntries = [{ localId: 'new-1', serverId: null, description: '' }];
		} else {
			tempIdCounter = 0;
		}
		errorMsg = '';
	}

	function addEntry() {
		tempIdCounter++;
		editEntries = [...editEntries, { localId: `new-${tempIdCounter}`, serverId: null, description: '' }];
	}

	function removeEntry(localId: string) {
		editEntries = editEntries.filter((e) => e.localId !== localId);
	}

	function cancelEditing() {
		editingCellKey = null;
		editEntries = [];
		errorMsg = '';
	}

	function saveEditing() {
		const hasContent = editEntries.some((e) => e.description.trim().length > 0);
		if (!hasContent) {
			errorMsg = 'Debes tener al menos un criterio con descripción.';
			return;
		}

		// Collect all criteria for this competency, replacing the edited level
		const allCriteria = getScaleCriteria().filter(
			(sc) => sc.competencyId === editCellCompetencyId
		);
		const otherLevelCriteria = allCriteria
			.filter((sc) => sc.level !== editCellLevel)
			.map((sc) => ({ level: sc.level, description: sc.description }));
		const editedCriteria = editEntries
			.filter((e) => e.description.trim().length > 0)
			.map((e) => ({ level: editCellLevel, description: e.description.trim() }));

		replaceScaleCriteria(editCellCompetencyId, [...otherLevelCriteria, ...editedCriteria]);

		editingCellKey = null;
		editEntries = [];
		errorMsg = '';
	}
</script>

{#each pillars as pillar, i (pillar.id)}
	{@const pillarCompetencies = competencies.filter((c) => c.pillarId === pillar.id)}
	<div class="mb-8">
		<div class="flex items-center gap-2 mb-3">
						<div class="badge text-sm px-3 py-2" style={pillarBadgeStyle(i)}>{pillar.name}</div>
		</div>

		<div class="overflow-x-auto rounded-box">
			<table class="table table-zebra" aria-label="Criterios de escala - {pillar.name}">
				<thead>
				<tr>
						<th class="w-48 min-w-[12rem] text-xs tracking-wide font-semibold text-base-content/50">Competencia</th>
								{#each levels as level (level)}
							<th class="min-w-[10rem] text-center text-xs tracking-wide font-semibold text-base-content/50">
								<span class="inline-flex items-center gap-1">
									<Star class="w-3 h-3" strokeWidth={2} />
									N{level} - {getLevelLabel(level)}
								</span>
							</th>
						{/each}
					</tr>
				</thead>
				<tbody>
					{#each pillarCompetencies as competency (competency.id)}
						<tr>
							<td class="font-medium text-sm">
								<span class="text-base-content">{competency.name}</span>
							</td>
						{#each levels as level (level)}
								{@const cellKey = getCellKey(competency.id, pillar.id, level)}
								{@const cellCriteria = getLevelCriteria(competency.id, pillar.id, level)}
								{@const isEditing = editingCellKey === cellKey}
								<td class="p-1.5 align-top">
									{#if isEditing}
										{#if errorMsg}
											<div class="alert alert-error text-sm mb-2" role="alert"><span>{errorMsg}</span></div>
										{/if}

										<div class="grid gap-1">
											{#each editEntries as entry (entry.localId)}
												<div class="grid grid-cols-[1fr_auto] gap-1">
													<textarea
														class="textarea textarea-bordered textarea-sm w-full"
														rows={2}
														bind:value={entry.description}
														required
													></textarea>
													<button
														type="button"
														class="btn btn-ghost btn-square btn-xs text-error h-full"
														onclick={() => removeEntry(entry.localId)}
														aria-label="Eliminar criterio"
													>
														<Trash2 class="w-4 h-4" />
													</button>
												</div>
											{/each}
										</div>

										<div class="grid grid-cols-3 gap-1 mt-1">
											<button type="button" class="btn btn-ghost btn-xs" onclick={addEntry} disabled={!canAdd} title="Agregar criterio">
												<Plus class="w-3 h-3" />
											</button>
											<button class="btn btn-ghost btn-xs" onclick={cancelEditing} title="Cancelar edición">
												<X class="w-3 h-3" />
											</button>
											<button class="btn btn-primary btn-xs" onclick={saveEditing} title="Guardar criterios">
												<Save class="w-3 h-3" />
											</button>
										</div>
									{:else}
										<button
											class="w-full min-h-[4rem] rounded-lg border border-base-300 p-2 text-left transition-colors hover:border-primary hover:bg-primary/5 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary {editingCellKey !== null ? 'opacity-40 pointer-events-none' : ''}"
											onclick={() => startEditing(competency.id, pillar.id, level)}
											aria-label="Editar criterios de {competency.name} nivel {level} en {pillar.name}"
										>
											{#if cellCriteria.length > 0}
												<div class="space-y-1">
													{#each cellCriteria as criterion (criterion.id)}
														<span class="text-xs text-base-content/60 leading-tight line-clamp-3 block">{criterion.description}</span>
													{/each}
												</div>
											{:else}
												<span class="text-xs italic text-base-content/30">Sin definir</span>
											{/if}
										</button>
									{/if}
								</td>
							{/each}
						</tr>
					{/each}
				</tbody>
			</table>
		</div>

		{#if pillarCompetencies.length === 0}
			<div class="text-center py-8 text-base-content/50 text-sm">
				Este pilar no tiene competencias asignadas.
			</div>
		{/if}
	</div>
{/each}
