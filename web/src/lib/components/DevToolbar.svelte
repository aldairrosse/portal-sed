<script lang="ts">
	import { EVALUATION_PROFILES, CYCLE_PHASES, PROFILE_LABELS, PHASE_LABELS, type EvaluationProfile, type CyclePhase } from '$lib/types/evaluation';
	import { setProfile, setPhase, getProfile, getPhase } from '$lib/stores/devContext.svelte';
	import { Code } from '@lucide/svelte';

	const profiles = EVALUATION_PROFILES;
	const phases = CYCLE_PHASES;

	const currentProfile = $derived(getProfile());
	const currentPhase = $derived(getPhase());

	function handleProfileChange(e: Event) {
		const target = e.target as HTMLSelectElement;
		setProfile(target.value as EvaluationProfile);
	}

	function handlePhaseChange(e: Event) {
		const target = e.target as HTMLSelectElement;
		setPhase(target.value as CyclePhase);
	}
</script>

<div class="bg-warning/5 border-t border-warning/10 px-4 py-3">
	<div class="flex flex-wrap gap-4 items-center justify-center text-sm">
		<div class="flex items-center gap-2">
			<Code class="w-4 h-4 text-warning/60" />
			<span class="font-medium text-warning/80 text-xs uppercase tracking-wider">Dev</span>
		</div>

		<label class="flex items-center gap-2">
			<span class="text-base-content/40 text-xs">Perfil:</span>
			<select
				class="select select-sm"
				value={currentProfile}
				onchange={handleProfileChange}
				aria-label="Seleccionar perfil de evaluación"
			>
				{#each profiles as p (p)}
					<option value={p}>{PROFILE_LABELS[p]}</option>
				{/each}
			</select>
		</label>

		<label class="flex items-center gap-2">
			<span class="text-base-content/40 text-xs">Fase:</span>
			<select
				class="select select-sm"
				value={currentPhase}
				onchange={handlePhaseChange}
				aria-label="Seleccionar fase del ciclo"
			>
				{#each phases as f (f)}
					<option value={f}>{PHASE_LABELS[f]}</option>
				{/each}
			</select>
		</label>
	</div>
</div>