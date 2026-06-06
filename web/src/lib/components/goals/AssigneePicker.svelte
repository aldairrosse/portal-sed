<script lang="ts">
	import CustomSelect from '$lib/components/ui/CustomSelect.svelte';
	import type { EmployeeAssignment } from '$lib/types/goal';

	interface Props {
		assignments: EmployeeAssignment[];
		selectedEmployeeId: string;
		onSelect: (employeeId: string) => void;
		currentUserId: string;
	}

	let { assignments, selectedEmployeeId, onSelect, currentUserId }: Props = $props();

	const options = $derived(
		assignments.map((a) => ({
			value: a.employeeId,
			label: a.employeeId === currentUserId ? `${a.employeeName} (yo)` : a.employeeName
		}))
	);
</script>

<div class="flex items-center gap-2">
	<span class="text-xs font-semibold text-base-content/60">Empleado</span>
	<CustomSelect
		{options}
		value={selectedEmployeeId}
		onChange={onSelect}
		placeholder="Seleccionar empleado"
		ariaLabel="Seleccionar empleado"
	/>
</div>
