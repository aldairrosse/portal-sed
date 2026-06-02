<script lang="ts">
	import { ShieldOff } from '@lucide/svelte';

	interface Props {
		title?: string;
		message?: string;
	}

	let { title = 'Sin permiso', message = 'No tienes acceso a esta sección. Contacta a tu administrador si crees que esto es un error.' }: Props = $props();
</script>

<div class="flex flex-col items-center justify-center py-16 text-center" role="alert">
	<div class="w-16 h-16 rounded-2xl bg-warning/10 flex items-center justify-center mb-5">
		<ShieldOff class="w-8 h-8 text-warning/60" strokeWidth={1.5} />
	</div>
	<h3 class="text-lg font-semibold text-warning">{title}</h3>
	<p class="text-base-content/50 mt-1.5 max-w-sm">{message}</p>
	<a href="/" class="btn btn-outline btn-sm mt-5 px-6">Volver al inicio</a>
</div>