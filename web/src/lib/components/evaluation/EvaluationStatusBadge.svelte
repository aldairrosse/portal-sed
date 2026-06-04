<script lang="ts">
	import type { EvaluationStatus } from '$lib/types/evaluation-result';

	interface Props {
		status: EvaluationStatus;
	}

	let { status }: Props = $props();

	const classMap: Record<EvaluationStatus, string> = {
		pending: 'badge-ghost',
		'in-progress': 'badge-warning',
		completed: 'badge-success'
	};

	const labelMap: Record<EvaluationStatus, string> = {
		pending: 'Pendiente',
		'in-progress': 'En progreso',
		completed: 'Completada'
	};
</script>

<span class="badge badge-sm {classMap[status]}" aria-label="Estado: {labelMap[status]}">
	{labelMap[status]}
</span>
