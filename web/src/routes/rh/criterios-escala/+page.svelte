<script lang="ts">
	import { Edit3 } from '@lucide/svelte';
	import ScaleCriteriaMatrix from '$lib/components/competency/ScaleCriteriaMatrix.svelte';
	import LevelDefinitionModal from '$lib/components/competency/LevelDefinitionModal.svelte';
	import PageSkeleton from '$lib/components/ui/PageSkeleton.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import { getCompetencies, getPillars } from '$lib/stores/competencyStore.svelte';

	const pillars = $derived(getPillars());
	const competencies = $derived(getCompetencies());

	let loading = $state(true);
	let showLevelDefModal = $state(false);
	let successMsg = $state('');
	let isAnyInlineEditing = $state(false);

	let pillarCount = $derived(pillars.length);
	let competencyCount = $derived(competencies.length);

	$effect(() => {
		const t = setTimeout(() => (loading = false), 300);
		return () => clearTimeout(t);
	});

	$effect(() => {
		if (successMsg) {
			const t = setTimeout(() => (successMsg = ''), 3000);
			return () => clearTimeout(t);
		}
	});
</script>

<svelte:head>
	<title>Criterios de escala — SED</title>
</svelte:head>

<div class="max-w-6xl mx-auto">
	<div class="flex items-start justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-base-content">Criterios de escala</h1>
			<p class="text-base-content/50 text-sm mt-1">
				Define los criterios de evaluación del nivel 1 al 5 para cada competencia en todos los pilares.
			</p>
		</div>
		<button
			class="btn btn-ghost btn-sm"
			disabled={isAnyInlineEditing}
			onclick={() => (showLevelDefModal = true)}
			aria-label="Editar definiciones de nivel"
		>
			<Edit3 class="w-4 h-4" />
			Editar definiciones de nivel
		</button>
	</div>

	{#if successMsg}
		<div class="alert alert-success mb-4 text-sm" role="status">
			<span>{successMsg}</span>
		</div>
	{/if}

	{#if loading}
		<PageSkeleton rows={6} />
	{:else if pillarCount === 0 || competencyCount === 0}
		<EmptyState
			title="Sin datos"
			message="Se requieren pilares y competencias para mostrar la matriz de criterios de escala."
		/>
	{:else}
		<ScaleCriteriaMatrix bind:isAnyInlineEditing />
	{/if}
</div>

<LevelDefinitionModal open={showLevelDefModal} onClose={() => (showLevelDefModal = false)} />
