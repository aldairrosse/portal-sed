<script lang="ts">
	import AcceptanceLevelEditor from '$lib/components/competency/AcceptanceLevelEditor.svelte';
	import PageSkeleton from '$lib/components/ui/PageSkeleton.svelte';

	let loading = $state(true);

	$effect(() => {
		const t = setTimeout(() => (loading = false), 300);
		return () => clearTimeout(t);
	});
</script>

<svelte:head>
	<title>Niveles de aceptación — SED</title>
</svelte:head>

<div class="max-w-4xl mx-auto">
	<div class="mb-6">
		<h1 class="text-2xl font-bold text-base-content">Niveles de aceptación</h1>
		<p class="text-base-content/50 text-sm mt-1">
			Asigna el nivel de aceptación (1 al 5) para cada competencia según el perfil. Las definiciones de nivel son globales.
		</p>
	</div>

	{#if loading}
		<PageSkeleton rows={5} />
	{:else}
		<AcceptanceLevelEditor />
	{/if}
</div>
