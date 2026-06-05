<script lang="ts">
	import { X } from '@lucide/svelte';
	import {
		getPillars,
		getCompetencies,
		getCompetencyAcceptanceLevel
	} from '$lib/stores/competencyStore.svelte';
	import { EVALUATION_PROFILES, PROFILE_LABELS } from '$lib/types/evaluation';
	import type { EvaluationProfile } from '$lib/types/evaluation';

	const PROFILE_ABBREVIATIONS: Record<EvaluationProfile, string> = {
		colaborador: 'COL',
		jefe: 'JEF',
		vendedor: 'VEN',
		'gerente-tienda': 'GTE',
		divisional: 'DIV',
		regional: 'REG',
		director: 'DIR',
		'director-general': 'DGN',
		rh: 'RH'
	};

	interface Props {
		open: boolean;
		onClose: () => void;
	}

	let { open, onClose }: Props = $props();

	let dialogEl: HTMLDialogElement | undefined = $state();

	const pillars = $derived(getPillars());
	const competencies = $derived(getCompetencies());

	function getCompetenciesByPillar(pillarId: string) {
		return competencies.filter((c) => c.pillarId === pillarId);
	}

	function getLevelForCompetency(competencyId: string, profileId: EvaluationProfile): number {
		const cal = getCompetencyAcceptanceLevel(competencyId, profileId);
		return cal?.level ?? 3;
	}

	$effect(() => {
		if (!dialogEl) return;
		if (open) {
			dialogEl.showModal();
		} else {
			dialogEl.close();
		}
	});

	function handleClose() {
		onClose();
	}

	function handleBackdropClick(e: MouseEvent) {
		if (e.target === dialogEl) {
			handleClose();
		}
	}
</script>

<dialog
	bind:this={dialogEl}
	class="modal"
	class:modal-open={open}
	aria-modal="true"
	aria-labelledby="summary-title"
	onclick={handleBackdropClick}
	onclose={handleClose}
>
	<div class="modal-box max-w-5xl">
		<div class="flex items-center justify-between mb-4">
			<h3 id="summary-title" class="text-lg font-semibold text-base-content">
				Nivel de aceptación por competencia y perfil
			</h3>
			<button
				class="btn btn-ghost btn-square btn-sm"
				onclick={handleClose}
				aria-label="Cerrar"
			>
				<X class="w-4 h-4" />
			</button>
		</div>

		<p class="text-xs text-base-content/50 mb-4">
			Nivel de aceptación asignado (1–5) para cada competencia según el perfil de evaluación.
		</p>

		<div class="overflow-x-auto">
			<table class="table table-zebra text-sm" aria-label="Nivel de aceptación por competencia y perfil">
				<thead>
					<tr>
						<th class="min-w-[12rem]">Competencia</th>
						{#each EVALUATION_PROFILES as profile (profile)}
							<th class="text-center min-w-[4rem]" title={PROFILE_LABELS[profile]}>
								{PROFILE_ABBREVIATIONS[profile]}
							</th>
						{/each}
					</tr>
				</thead>
				<tbody>
					{#each pillars as pillar (pillar.id)}
						{@const pillarComps = getCompetenciesByPillar(pillar.id)}
						{#each pillarComps as competency (competency.id)}
							<tr>
								<td class="font-medium">
									{competency.name}
									<span class="text-xs text-base-content/40 ml-1">({pillar.name})</span>
								</td>
					{#each EVALUATION_PROFILES as profile (profile)}
									{@const level = getLevelForCompetency(competency.id, profile)}
									<td class="text-center">
										<span
											class="inline-flex items-center justify-center w-7 h-7 rounded-full bg-primary/10 text-primary font-bold text-sm"
										>
											{level}
										</span>
									</td>
								{/each}
							</tr>
						{/each}
					{/each}
				</tbody>
			</table>
		</div>

		{#if competencies.length === 0}
			<div class="text-center py-12 text-base-content/50 text-sm">
				No hay competencias registradas.
			</div>
		{/if}

		<div class="modal-action mt-4">
			<button class="btn btn-ghost btn-sm" onclick={handleClose}>Cerrar</button>
		</div>
	</div>
</dialog>
