<script lang="ts">
	import { Check, Plus, Trash2, Star } from '@lucide/svelte';
	import {
		getPillars,
		getCompetencies,
		getScaleCriteriaForCell,
		getLevelDefinitions,
		updateScaleCriterion,
		addScaleCriterion,
		removeScaleCriterion
	} from '$lib/stores/competencyStore.svelte';
	import type { ScaleCriterion } from '$lib/types/competency';

	const PILLAR_COLORS = ['primary', 'secondary', 'accent', 'info', 'success', 'warning'] as const;

	interface Props {
		isAnyInlineEditing?: boolean;
	}

	let { isAnyInlineEditing = $bindable(false) }: Props = $props();

	const pillars = $derived(getPillars());
	const competencies = $derived(getCompetencies());
	const levelDefs = $derived(getLevelDefinitions());
	const levels = [1, 2, 3, 4, 5] as const;

	// ─── Inline editing state ──────────────────────────────────────────────
	let editingCellKey: string | null = $state(null);
	let editEntries: Array<{ localId: string; serverId: string | null; description: string }> = $state([]);
	let editCellCompetencyId = $state('');
	let editCellPillarId = $state('');
	let editCellLevel = $state<number>(1);
	let tempIdCounter = $state(0);
	let errorMsg = $state('');

	$effect(() => {
		isAnyInlineEditing = editingCellKey !== null;
	});

	function getCellKey(competencyId: string, pillarId: string, level: number): string {
		return `${competencyId}-${pillarId}-${level}`;
	}

	function getLevelLabel(level: number): string {
		return levelDefs.find((d) => d.level === level)?.label ?? 'Nivel ' + level;
	}

	function getDescriptions(competencyId: string, pillarId: string, level: number): string[] {
		return getScaleCriteriaForCell(competencyId, pillarId)
			.filter((c) => c.level === level)
			.map((c) => c.description)
			.filter(Boolean);
	}

	function getLevelCriteria(competencyId: string, pillarId: string, level: number): ScaleCriterion[] {
		return getScaleCriteriaForCell(competencyId, pillarId).filter((c) => c.level === level);
	}

	function pillarBadgeClass(index: number): string {
		return `badge-${PILLAR_COLORS[index % PILLAR_COLORS.length]}`;
	}

	// ─── Inline editing actions ────────────────────────────────────────────

	function startEditing(competencyId: string, pillarId: string, level: number) {
		if (editingCellKey !== null) return; // only one cell at a time
		editingCellKey = getCellKey(competencyId, pillarId, level);
		editCellCompetencyId = competencyId;
		editCellPillarId = pillarId;
		editCellLevel = level;
		const existing = getLevelCriteria(competencyId, pillarId, level);
		editEntries = existing.map((c) => ({
			localId: c.id,
			serverId: c.id,
			description: c.description
		}));
		tempIdCounter = 0;
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

		const existingIds = new Set(getLevelCriteria(editCellCompetencyId, editCellPillarId, editCellLevel).map((c) => c.id));
		const finalIds = new Set<string>();
		const newEntries: Array<Omit<ScaleCriterion, 'id'>> = [];
		const updatePairs: Array<{ id: string; description: string }> = [];

		for (const entry of editEntries) {
			if (entry.serverId) {
				finalIds.add(entry.serverId);
				updatePairs.push({ id: entry.serverId, description: entry.description.trim() });
			} else {
				newEntries.push({
					competencyId: editCellCompetencyId,
					pillarId: editCellPillarId,
					level: editCellLevel as 1 | 2 | 3 | 4 | 5,
					description: entry.description.trim()
				});
			}
		}

		for (const id of existingIds) {
			if (!finalIds.has(id)) {
				removeScaleCriterion(id);
			}
		}

		for (const pair of updatePairs) {
			updateScaleCriterion(pair.id, pair.description);
		}

		for (const nc of newEntries) {
			addScaleCriterion(nc);
		}

		editingCellKey = null;
		editEntries = [];
		errorMsg = '';
	}
</script>

{#each pillars as pillar, i (pillar.id)}
	{@const pillarCompetencies = competencies.filter((c) => c.pillarId === pillar.id)}
	<div class="mb-8">
		<div class="flex items-center gap-2 mb-3">
			<div class="badge {pillarBadgeClass(i)} text-sm px-3 py-2">{pillar.name}</div>
		</div>

		<div class="overflow-x-auto rounded-box border border-base-300">
			<table class="table table-zebra" aria-label="Criterios de escala - {pillar.name}">
				<thead>
				<tr>
						<th class="w-48 min-w-[12rem]">Competencia</th>
								{#each levels as level (level)}
							<th class="min-w-[10rem] text-center">
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
								{@const texts = getDescriptions(competency.id, pillar.id, level)}
								{@const isEditing = editingCellKey === cellKey}
								<td class="p-1.5 align-top">
									{#if isEditing}
										<div class="border border-base-300 rounded-lg p-4 bg-base-200/50">
											{#if errorMsg}
												<div class="alert alert-error text-sm mb-3" role="alert"><span>{errorMsg}</span></div>
											{/if}

											{#each editEntries as entry (entry.localId)}
												<div class="form-control">
													<label class="label" for="criterion-{cellKey}-{entry.localId}">
														<span class="label-text text-xs">Criterio</span>
													</label>
													<div class="flex items-start gap-2">
														<textarea
															id="criterion-{cellKey}-{entry.localId}"
															class="textarea textarea-bordered textarea-sm w-full"
															rows={2}
															bind:value={entry.description}
															required
														></textarea>
														<button
															type="button"
															class="btn btn-ghost btn-square btn-xs text-error flex-shrink-0 mt-1"
															onclick={() => removeEntry(entry.localId)}
															aria-label="Eliminar criterio"
														>
															<Trash2 class="w-4 h-4" />
														</button>
													</div>
												</div>
											{/each}

											<button type="button" class="btn btn-ghost btn-xs mt-2 text-primary" onclick={addEntry}>
												<Plus class="w-3 h-3" /> Agregar criterio
											</button>

											<div class="flex justify-end gap-2 mt-3">
												<button class="btn btn-ghost btn-sm" onclick={cancelEditing}>Cancelar</button>
												<button class="btn btn-primary btn-sm" onclick={saveEditing}>
													<Check class="w-4 h-4" /> Guardar
												</button>
											</div>
										</div>
									{:else}
										<button
											class="w-full min-h-[4rem] rounded-lg border border-base-300 p-2 text-left transition-colors hover:border-primary hover:bg-primary/5 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary {editingCellKey !== null ? 'opacity-40 pointer-events-none' : ''}"
											onclick={() => startEditing(competency.id, pillar.id, level)}
											aria-label="Editar criterios de {competency.name} nivel {level} en {pillar.name}"
										>
											{#if texts.length > 0}
												<div class="space-y-1">
													{#each texts as text (text)}
														<span class="text-xs text-base-content/60 leading-tight line-clamp-3 block">{text}</span>
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
