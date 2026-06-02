<script lang="ts">
	import { page } from '$app/stores';
</script>

<svelte:head>
	<title>Página no encontrada — SED</title>
</svelte:head>

<div class="flex flex-col items-center justify-center py-16 text-center">
	<h1 class="text-6xl font-bold text-primary mb-2">404</h1>
	<h2 class="text-2xl font-semibold text-base-content mb-4">Página no encontrada</h2>
	<p class="text-base-content/60 mb-6">
		La página que buscas no existe o fue movida.
	</p>
	<a href="/" class="btn btn-primary">Volver al inicio</a>
</div>