<script lang="ts">
	import { Eye } from '@lucide/svelte';

	interface Props {
		employeeName: string;
	}

	let { employeeName }: Props = $props();
</script>

<div class="alert alert-warning" role="status">
	<Eye class="w-4 h-4 shrink-0" aria-hidden="true" />
	<span>
		Estás viendo las metas de <strong>{employeeName}</strong>. Solo puedes solicitar cambios.
	</span>
</div>
