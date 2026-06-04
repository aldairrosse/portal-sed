<script lang="ts">
	import { Eye } from '@lucide/svelte';
	import type { CyclePhase } from '$lib/types/goal';

	interface Props {
		employeeName: string;
		phase?: CyclePhase;
	}

	let { employeeName, phase = 'inicio-anio' }: Props = $props();
</script>

<div class="alert alert-warning" role="status">
	<Eye class="w-4 h-4 shrink-0" aria-hidden="true" />
	<span>
		{#if phase === 'medio-anio'}
			Estás viendo las metas de <strong>{employeeName}</strong>. Puedes editar avances y agregar comentarios.
		{:else}
			Estás viendo las metas de <strong>{employeeName}</strong>. Solo puedes solicitar cambios.
		{/if}
	</span>
</div>
