<script lang="ts">
	import { page } from '$app/stores';
	import Forbidden from '$lib/components/Forbidden.svelte';

	const status = $derived($page.status);
</script>

<svelte:head>
	<title
		>{status === 403
			? 'Sin acceso — SED'
			: status === 404
				? 'Página no encontrada — SED'
				: `Error ${status} — SED`}</title
	>
</svelte:head>

{#if status === 403}
	<Forbidden />
{:else if status === 404}
	<div class="flex flex-col items-center justify-center py-16 text-center">
		<h1 class="text-6xl font-bold text-primary mb-2">404</h1>
		<h2 class="text-2xl font-semibold text-base-content mb-4">
			Página no encontrada
		</h2>
		<p class="text-base-content/60 mb-6">
			La página que buscas no existe o fue movida.
		</p>
		<a href="/" class="btn btn-primary">Volver al inicio</a>
	</div>
{:else}
	<div class="flex flex-col items-center justify-center py-16 text-center">
		<h1 class="text-6xl font-bold text-primary mb-2">{status}</h1>
		<h2 class="text-2xl font-semibold text-base-content mb-4">
			Algo salió mal
		</h2>
		<p class="text-base-content/60 mb-6">
			Ocurrió un error inesperado (código {status}).
		</p>
		<a href="/" class="btn btn-primary">Volver al inicio</a>
	</div>
{/if}
