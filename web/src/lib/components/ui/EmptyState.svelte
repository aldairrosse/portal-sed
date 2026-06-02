<script lang="ts">
	import { Inbox } from '@lucide/svelte';

	interface Props {
		title?: string;
		message?: string;
		actionLabel?: string;
		actionHref?: string;
	}

	let { title = 'Sin contenido', message = 'No hay datos para mostrar.', actionLabel, actionHref }: Props = $props();
</script>

<div class="flex flex-col items-center justify-center py-16 text-center" role="status">
	<div class="w-16 h-16 rounded-2xl bg-primary/8 flex items-center justify-center mb-5">
		<Inbox class="w-8 h-8 text-primary/40" strokeWidth={1.5} />
	</div>
	<h3 class="text-lg font-semibold text-base-content">{title}</h3>
	<p class="text-base-content/50 mt-1.5 max-w-sm">{message}</p>
	{#if actionLabel && actionHref}
		<a href={actionHref} class="btn btn-primary btn-sm mt-5 px-6">{actionLabel}</a>
	{/if}
</div>