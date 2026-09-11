<script lang="ts">
	import type {
		Pillar,
		Competency,
		LevelDefinition,
	} from '$lib/types/competency';
	import type { CompetencyRating } from '$lib/types/evaluation-result';
	import ScaleRatingSelector from './ScaleRatingSelector.svelte';
	import PageSkeleton from '$lib/components/ui/PageSkeleton.svelte';
	import ErrorState from '$lib/components/ui/ErrorState.svelte';
	import { isAvance, isCierre } from '$lib/types/cycle';
	import { Pencil } from '@lucide/svelte';
	import { tick } from 'svelte';
	import { getCriteriaText } from '$lib/stores/competencyStore.svelte';

	interface Props {
		pillar: Pillar;
		competencies: Competency[];
		ratings: CompetencyRating[];
		levelDefinitions: LevelDefinition[];
		acceptanceLevels: Record<string, number>;
		mode: 'self' | 'manager' | 'rh';
		onRate: (
			competencyId: string,
			level: 1 | 2 | 3 | 4 | 5,
			comment?: string,
			evaluationId?: string,
			phase?: 'avance' | 'cierre',
		) => void;
		onRhRate?: (
			competencyId: string,
			level: 1 | 2 | 3 | 4 | 5,
			comment?: string,
			evaluationId?: string,
			phase?: 'avance' | 'cierre',
			managerComment?: string,
		) => void;
		disabled?: boolean;
		evaluationId?: string;
		phase?: 'avance' | 'cierre';
		showCommentInput?: boolean;
		loading?: boolean;
		error?: string | null;
	}

	let {
		pillar,
		competencies,
		ratings,
		levelDefinitions,
		acceptanceLevels,
		mode,
		onRate,
		onRhRate,
		disabled = false,
		evaluationId,
		phase = 'cierre',
		showCommentInput = true,
		loading = false,
		error = null,
	}: Props = $props();

	const isAvancePhase = $derived(isAvance(phase));
	const isCierrePhase = $derived(isCierre(phase));
	const phaseEditable = $derived(isAvancePhase || isCierrePhase);
	// ponytail: disabled separado por fase (avance=self, cierre=rh), mismo gate editable
	const selfDisabled = $derived(disabled || !phaseEditable);
	const rhDisabled = $derived(disabled || !phaseEditable);

	function getRating(competencyId: string): CompetencyRating | undefined {
		return (ratings ?? []).find((r) => r.competencyId === competencyId);
	}

	// ponytail: borradores locales por competencia — se envían solo con el botón Enviar (sin oninput auto)
	let draftLevels = $state<Record<string, 1 | 2 | 3 | 4 | 5>>({});
	let draftComments = $state<Record<string, string>>({});
	// Edición exclusiva del comentario propio por competencia (solo viewerMode puede editar el suyo).
	let editing = $state<Record<string, 'self' | 'rh' | 'manager'>>({});
	let attempted = $state<Record<string, boolean>>({});

	function isValidLevel(v: unknown): v is 1 | 2 | 3 | 4 | 5 {
		return typeof v === 'number' && v >= 1 && v <= 5;
	}

	function canSend(competencyId: string): boolean {
		return isValidLevel(draftLevels[competencyId]);
	}

	function hasRealComment(text?: string | null): boolean {
		const t = text?.trim() ?? '';
		if (!t) return false;
		return !/^Sin (autoevaluación|evaluación|comentario)/i.test(t);
	}

	function authorLine(
		author?: string | null,
		date?: string | null,
		text?: string | null,
	): string {
		if (!hasRealComment(text)) return '';
		const name = author?.trim() ?? '';
		if (!name) return '';
		if (!date) return name;
		const d = new Date(date);
		if (Number.isNaN(d.getTime())) return name;
		return `${name} · ${d.toLocaleDateString('es-MX')} ${d.toLocaleTimeString('es-MX', { hour: '2-digit', minute: '2-digit' })}`;
	}

	const TEXTAREA_MAX_HEIGHT = 128;

	function autoGrow(node: HTMLTextAreaElement) {
		node.style.height = 'auto';
		const capped = Math.min(node.scrollHeight, TEXTAREA_MAX_HEIGHT);
		node.style.height = `${capped}px`;
		node.style.overflowY =
			node.scrollHeight > TEXTAREA_MAX_HEIGHT ? 'auto' : 'hidden';
	}

	function ownCommentSaved(competencyId: string): string {
		const rating = getRating(competencyId);
		if (mode === 'self') return rating?.selfComment ?? '';
		if (mode === 'rh') return rating?.rhComment ?? '';
		return rating?.managerComment ?? '';
	}

	async function startEdit(competencyId: string) {
		editing[competencyId] = mode;
		draftComments[competencyId] = ownCommentSaved(competencyId);
		const saved = getRating(competencyId);
		const savedLevel = mode === 'self' ? saved?.selfRating : saved?.rhRating;
		if (isValidLevel(savedLevel)) draftLevels[competencyId] = savedLevel;
		delete attempted[competencyId];
		await tick();
		const el = document.getElementById(`comment-${competencyId}`);
		if (el instanceof HTMLTextAreaElement) autoGrow(el);
		el?.focus();
	}

	function cancelEdit(competencyId: string) {
		delete editing[competencyId];
		delete draftLevels[competencyId];
		delete draftComments[competencyId];
		delete attempted[competencyId];
	}

	function draftComment(competencyId: string, fallback: string): string {
		return draftComments[competencyId] ?? fallback;
	}

	function handleRatingChange(competencyId: string, level: 1 | 2 | 3 | 4 | 5) {
		if (!isValidLevel(level)) return;
		draftLevels[competencyId] = level;
	}

	function handleCommentInput(competencyId: string, comment: string) {
		draftComments[competencyId] = comment;
	}

	function tooltipByLevel(
		competencyId: string,
		pillarId: string,
	): Record<number, string> {
		return {
			1: getCriteriaText(competencyId, pillarId, 1),
			2: getCriteriaText(competencyId, pillarId, 2),
			3: getCriteriaText(competencyId, pillarId, 3),
			4: getCriteriaText(competencyId, pillarId, 4),
			5: getCriteriaText(competencyId, pillarId, 5),
		};
	}

	function handleSend(competencyId: string) {
		if (!canSend(competencyId)) {
			attempted[competencyId] = true;
			return;
		}
		const rating = getRating(competencyId);
		const level = draftLevels[competencyId];
		if (mode === 'manager') {
			// Jefe writes to the shared rh endpoint: rating goes to rhRating,
			// comment goes to managerComment (rhComment preserved as-is).
			const managerComment =
				draftComments[competencyId] ?? rating?.managerComment ?? '';
			onRhRate?.(
				competencyId,
				level,
				rating?.rhComment,
				evaluationId,
				phase,
				managerComment,
			);
		} else {
			const fallbackComment =
				mode === 'rh' ? (rating?.rhComment ?? '') : (rating?.selfComment ?? '');
			const comment = draftComments[competencyId] ?? fallbackComment;
			if (mode === 'rh') {
				onRhRate?.(competencyId, level, comment, evaluationId, phase);
			} else {
				onRate(competencyId, level, comment, evaluationId, phase);
			}
		}
		delete draftLevels[competencyId];
		delete draftComments[competencyId];
		delete editing[competencyId];
		delete attempted[competencyId];
	}
</script>

{#if loading}
	<PageSkeleton variant="card" rows={1} />
{:else if error}
	<ErrorState message={error} />
{:else}
	<div class="card bg-base-100 border border-base-300">
		<div class="card-body px-0">
			<h3 class="text-base font-semibold text-base-content mb-1">
				{pillar.name}
			</h3>
			<p class="text-xs text-base-content/50 mb-4">{pillar.description}</p>

			<div class="flex flex-col gap-5">
				{#each competencies as competency (competency.id)}
					{@const rating = getRating(competency.id)}
					{@const acceptanceLevel = acceptanceLevels[competency.id]}
					{@const ownSavedComment =
						mode === 'self'
							? rating?.selfComment
							: mode === 'rh'
								? rating?.rhComment
								: rating?.managerComment}
					{@const ownHasComment = hasRealComment(ownSavedComment)}
					{@const isEditingOwn = editing[competency.id] === mode}
					{@const ownHasRating =
						mode === 'self'
							? rating?.selfRating != null
							: rating?.rhRating != null}
					{@const ownReadMode = ownHasRating && !isEditingOwn}
					{@const ownDisabled = mode === 'self' ? selfDisabled : rhDisabled}
					{@const canEditOwn = !ownDisabled}
					{@const ownLabel =
						mode === 'self'
							? 'Autoevaluación'
							: mode === 'rh'
								? 'Evaluación de RH'
								: 'Evaluación de Jefe'}
					{@const ownPlaceholder =
						mode === 'self'
							? 'Escribe tu autoevaluación...'
							: mode === 'rh'
								? 'Escribe evaluación de RH...'
								: 'Escribe evaluación de Jefe...'}
					{@const ownEmpty =
						mode === 'self'
							? 'Sin autoevaluación'
							: mode === 'rh'
								? 'Sin evaluación de RH'
								: 'Sin evaluación de Jefe'}
					{@const otherManagerAuthor =
						rating?.managerCommentAuthor ?? rating?.authorName}
					{@const otherManagerDate =
						rating?.managerCommentCreatedAt ?? rating?.commentCreatedAt}
					{@const ownAuthor =
						mode === 'manager'
							? (rating?.managerCommentAuthor ?? rating?.authorName)
							: rating?.authorName}
					{@const ownDate =
						mode === 'manager'
							? (rating?.managerCommentCreatedAt ?? rating?.commentCreatedAt)
							: rating?.commentCreatedAt}
					<div class="pt-3 first:pt-0">
						<div class="flex items-center gap-3">
							<!-- Name + description (left) -->
							<div class="flex-1 min-w-0">
								<p class="text-sm font-semibold text-base-content">
									{competency.name}
								</p>
								<p class="text-xs text-base-content/50">
									{competency.description}
								</p>
							</div>
							<!-- Rating selector (right): shared 1-5 input for self, and shared for rh/manager -->
							<div class="shrink-0">
								{#if mode === 'self'}
									<ScaleRatingSelector
										value={draftLevels[competency.id] ??
											(ownReadMode ? rating?.selfRating : undefined)}
										{acceptanceLevel}
										disabled={selfDisabled}
										{levelDefinitions}
										tooltips={tooltipByLevel(
											competency.id,
											competency.pillarId ?? pillar.id,
										)}
										readOnlySingle={ownReadMode}
										required={!ownReadMode && canEditOwn}
										invalid={!ownReadMode &&
											canEditOwn &&
											!canSend(competency.id)}
										attempted={!!attempted[competency.id]}
										onChange={(level) =>
											handleRatingChange(competency.id, level)}
									/>
								{:else}
									<ScaleRatingSelector
										value={draftLevels[competency.id] ??
											(ownReadMode ? rating?.rhRating : undefined)}
										{acceptanceLevel}
										disabled={rhDisabled}
										{levelDefinitions}
										tooltips={tooltipByLevel(
											competency.id,
											competency.pillarId ?? pillar.id,
										)}
										readOnlySingle={ownReadMode}
										required={!ownReadMode && canEditOwn}
										invalid={!ownReadMode &&
											canEditOwn &&
											!canSend(competency.id)}
										attempted={!!attempted[competency.id]}
										onChange={(level) =>
											handleRatingChange(competency.id, level)}
									/>
								{/if}
							</div>
							{#if (mode === 'rh' || mode === 'manager') && rating && rating.selfRating != null}
								<span
									class="badge badge-ghost badge-sm shrink-0"
									title="Autoevaluación"
								>
									Auto: {rating?.selfRating ?? '—'}
								</span>
							{:else if mode === 'manager' && !rating?.selfRating}
								<span
									class="badge badge-ghost badge-outline badge-sm shrink-0"
									title="Autoevaluación">Sin autoevaluación</span
								>
							{/if}
						</div>

						<!-- Otros comentarios (lectura): todos visibles, solo el propio es editable -->
						{#if mode !== 'self' && hasRealComment(rating?.selfComment)}
							<p class="text-xs text-base-content/50 mt-2">
								<span class="font-semibold">Autoevaluación:</span>
								<span class="italic">"{rating?.selfComment}"</span>
							</p>
							{#if authorLine(rating?.authorName, rating?.commentCreatedAt, rating?.selfComment)}
								<p class="text-[11px] text-base-content/40 mt-1">
									{authorLine(
										rating?.authorName,
										rating?.commentCreatedAt,
										rating?.selfComment,
									)}
								</p>
							{/if}
						{/if}
						{#if mode !== 'rh' && hasRealComment(rating?.rhComment)}
							<p class="text-xs text-base-content/50 mt-2">
								<span class="font-semibold">Evaluación de RH:</span>
								<span class="italic">"{rating?.rhComment}"</span>
							</p>
							{#if authorLine(rating?.authorName, rating?.commentCreatedAt, rating?.rhComment)}
								<p class="text-[11px] text-base-content/40 mt-1">
									{authorLine(
										rating?.authorName,
										rating?.commentCreatedAt,
										rating?.rhComment,
									)}
								</p>
							{/if}
						{/if}
						{#if mode !== 'manager' && hasRealComment(rating?.managerComment)}
							<p class="text-xs text-base-content/50 mt-2">
								<span class="font-semibold">Evaluación de Jefe:</span>
								<span class="italic">"{rating?.managerComment}"</span>
							</p>
							{#if authorLine(otherManagerAuthor, otherManagerDate, rating?.managerComment)}
								<p class="text-[11px] text-base-content/40 mt-1">
									{authorLine(
										otherManagerAuthor,
										otherManagerDate,
										rating?.managerComment,
									)}
								</p>
							{/if}
						{/if}

						<!-- Comentario propio: lectura con pencil, edición exclusiva -->
						{#if showCommentInput && canEditOwn && !ownReadMode}
							<label
								class="label px-0 py-1 mt-2"
								for={`comment-${competency.id}`}
							>
								<span class="label-text text-xs">{ownLabel}</span>
							</label>
							<textarea
								id={`comment-${competency.id}`}
								class="textarea textarea-bordered textarea-xs w-full mt-2 resize-y overflow-hidden max-h-32 !min-h-[2.25rem] h-auto !pt-2 !pb-1 focus:ring-2 focus:ring-primary/40 focus:outline-none"
								use:autoGrow
								placeholder={ownPlaceholder}
								value={draftComment(competency.id, ownSavedComment ?? '')}
								oninput={(e) => {
									handleCommentInput(
										competency.id,
										(e.target as HTMLTextAreaElement).value,
									);
									autoGrow(e.target as HTMLTextAreaElement);
								}}
								disabled={ownDisabled}
								rows="1"
								aria-label={ownLabel}
							></textarea>
							{#if authorLine(ownAuthor, ownDate, ownSavedComment)}
								<p class="text-[11px] text-base-content/40 mt-1">
									{authorLine(ownAuthor, ownDate, ownSavedComment)}
								</p>
							{/if}
							<div class="mt-2 flex gap-2">
								<button
									class="btn btn-primary btn-xs"
									onclick={() => handleSend(competency.id)}
									disabled={ownDisabled || !canSend(competency.id)}
									aria-label="Enviar evaluación de {competency.name}"
								>
									Enviar
								</button>
								{#if ownHasRating}
									<button
										class="btn btn-ghost btn-xs"
										onclick={() => cancelEdit(competency.id)}
										aria-label="Cancelar edición de {competency.name}"
									>
										Cancelar
									</button>
								{/if}
							</div>
						{:else if showCommentInput || ownHasComment}
							<div class="mt-2">
								<span class="text-xs text-base-content/50">{ownLabel}:</span>
								<p class="text-sm text-base-content/70 bg-base-200 rounded p-2">
									{ownHasComment
										? ownSavedComment
										: ownHasRating
											? 'Sin comentario'
											: ownEmpty}
									{#if ownReadMode && canEditOwn}
										<button
											class="btn btn-ghost btn-xs ml-1 align-middle"
											onclick={() => startEdit(competency.id)}
											aria-label="Editar {ownLabel.toLowerCase()} de {competency.name}"
										>
											<Pencil size={14} />
										</button>
									{/if}
								</p>
								{#if authorLine(ownAuthor, ownDate, ownSavedComment)}
									<p class="text-[11px] text-base-content/40 mt-1">
										{authorLine(ownAuthor, ownDate, ownSavedComment)}
									</p>
								{/if}
							</div>
						{/if}
					</div>
				{/each}
			</div>
		</div>
	</div>
{/if}
