<script lang="ts">
	import type { EmployeeAssignment } from '$lib/types/goal';

	interface Props {
		assignments: EmployeeAssignment[];
		selectedEmployeeId: string;
		onSelect: (employeeId: string) => void;
	}

	let { assignments, selectedEmployeeId, onSelect }: Props = $props();

	function handleChange(e: Event) {
		const target = e.target as HTMLSelectElement;
		onSelect(target.value);
	}
</script>

<div class="form-control">
	<label class="label" for="assignee-picker">
		<span class="label-text">Evaluado</span>
	</label>
	<select
		id="assignee-picker"
		class="select select-bordered select-sm w-full max-w-xs"
		value={selectedEmployeeId}
		onchange={handleChange}
		aria-label="Seleccionar evaluado"
	>
		{#each assignments as assignment (assignment.employeeId)}
			<option value={assignment.employeeId}>
				{assignment.employeeName}
			</option>
		{/each}
	</select>
</div>
