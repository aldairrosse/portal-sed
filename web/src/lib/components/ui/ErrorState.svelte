<script lang="ts">
	import { AlertTriangle } from '@lucide/svelte';

	interface Props {
		title?: string;
		message?: string;
		retryLabel?: string;
	}

	let { title = 'Error de red', message = 'No se pudo conectar con el servidor. Verifica tu conexión e intenta de nuevo.', retryLabel = 'Reintentar' }: Props = $props();

	function handleRetry() {
		location.reload();
	}
</script>

<div class="flex flex-col items-center justify-center py-16 text-center" role="alert">
	<div class="w-16 h-16 rounded-2xl bg-error/8 flex items-center justify-center mb-5">
		<AlertTriangle class="w-8 h-8 text-error/50" strokeWidth={1.5} />
	</div>
	<h3 class="text-lg font-semibold text-error">{title}</h3>
	<p class="text-base-content/50 mt-1.5 max-w-sm">{message}</p>
	<button class="btn btn-error btn-sm mt-5 px-6" onclick={handleRetry}>
		{retryLabel}
	</button>
</div>