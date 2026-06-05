<script lang="ts">
	import type { EmployeeAssignment } from '$lib/types/goal';
	import type { Snippet } from 'svelte';

	interface Props {
		employees: EmployeeAssignment[];
		onSelect: (employeeId: string) => void;
		selectedEmployeeId?: string;
		detail?: Snippet;
	}

	let { employees, onSelect, selectedEmployeeId = '', detail }: Props = $props();

	let searchQuery = $state('');
	let showDropdown = $state(false);

	const filteredEmployees = $derived(
		searchQuery.trim() === ''
			? employees
			: employees.filter((e) =>
					e.employeeName.toLowerCase().includes(searchQuery.toLowerCase())
				)
	);

	const selectedEmployeeName = $derived(
		employees.find((e) => e.employeeId === selectedEmployeeId)?.employeeName ?? ''
	);

	function handleSelect(employeeId: string) {
		onSelect(employeeId);
		searchQuery = selectedEmployeeName;
		showDropdown = false;
	}
</script>

<div class="flex flex-col gap-6">
	<!-- Search input -->
	<div class="relative w-full max-w-sm">
		<label class="label" for="employee-search">
			<span class="label-text">Buscar evaluado</span>
		</label>
		<input
			id="employee-search"
			type="text"
			class="input input-bordered input-sm w-full"
			placeholder="Escribí un nombre..."
			value={searchQuery}
			oninput={(e) => {
				searchQuery = (e.target as HTMLInputElement).value;
				showDropdown = true;
			}}
			onfocus={() => (showDropdown = true)}
			onblur={() => setTimeout(() => (showDropdown = false), 150)}
			aria-label="Buscar evaluado"
			aria-expanded={showDropdown}
			aria-autocomplete="list"
			role="combobox"
			aria-controls="employee-listbox"
		/>
		{#if showDropdown && filteredEmployees.length > 0}
			<ul
				id="employee-listbox"
				class="absolute z-50 mt-1 w-full bg-base-100 border border-base-300 rounded-lg shadow-lg max-h-60 overflow-auto"
				role="listbox"
			>
				{#each filteredEmployees as employee (employee.employeeId)}
					<li>
						<button
							type="button"
							class="w-full text-left px-4 py-2 text-sm hover:bg-base-200 {selectedEmployeeId === employee.employeeId ? 'bg-base-200 font-semibold' : ''}"
							onmousedown={() => handleSelect(employee.employeeId)}
							role="option"
							aria-selected={selectedEmployeeId === employee.employeeId}
						>
							{employee.employeeName}
						</button>
					</li>
				{/each}
			</ul>
		{/if}
	</div>

	{#if selectedEmployeeId}
		{@render detail?.()}
	{:else}
		<p class="text-sm text-base-content/30 italic text-center py-8">
			Selecciona un evaluado para comenzar.
		</p>
	{/if}
</div>
