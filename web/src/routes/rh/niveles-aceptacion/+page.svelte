<script lang="ts">
	import AcceptanceLevelEditor from '$lib/components/competency/AcceptanceLevelEditor.svelte';
	import PageSkeleton from '$lib/components/ui/PageSkeleton.svelte';
	import { load, isLoading } from '$lib/stores/competencyStore.svelte';
	import { FileText } from '@lucide/svelte';

	const loading = $derived(isLoading());
	let loaded = $state(false);

	$effect(() => {
		load().then(() => { loaded = true; });
	});
</script>

<svelte:head>
	<title>Niveles de aceptación — SED</title>
</svelte:head>

<div class="max-w-4xl mx-auto">
	<div class="mb-6">
		<h1 class="text-2xl font-bold text-base-content flex items-center gap-2">
			<FileText class="w-6 h-6" />
			Niveles de aceptación
		</h1>
		<p class="text-base-content/50 text-sm mt-1">
			Asigna el nivel de aceptación (1 al 5) para cada competencia según el perfil. Las definiciones de nivel son globales.
		</p>
	</div>

	{#if !loaded && loading}
		<PageSkeleton rows={5} />
	{:else}
		<AcceptanceLevelEditor />
	{/if}
</div>
