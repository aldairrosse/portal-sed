<script lang="ts">
	let { title = 'Sin acceso', message = 'No tienes permisos para ver esta página.' }: {
		title?: string;
		message?: string;
	} = $props();
</script>

<div class="flex flex-col items-center justify-center py-16 text-center">
	<h1 class="text-6xl font-bold text-error mb-2">403</h1>
	<h2 class="text-2xl font-semibold text-base-content mb-4">{title}</h2>
	<p class="text-base-content/60 mb-6">{message}</p>
	<a href="/" class="btn btn-primary">Volver al inicio</a>
</div>
