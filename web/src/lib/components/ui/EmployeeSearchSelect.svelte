<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import CustomSelect from './CustomSelect.svelte';
	import { createEmployeePickerStore, type EmployeePickerStore } from '$lib/stores/employeePickerStore.svelte';

	interface Props {
		value: string;
		onChange: (value: string) => void;
		placeholder?: string;
		ariaLabel?: string;
		class?: string;
		picker?: EmployeePickerStore;
	}

	let {
		value,
		onChange,
		placeholder = 'Seleccionar empleado',
		ariaLabel = 'Empleado',
		class: className = '',
		picker
	}: Props = $props();

	const store = picker ?? createEmployeePickerStore();
	const owned = !picker;

	const options = $derived(store.getOptions());
	const hasMore = $derived(store.hasMoreEmployees());
	const loadingMore = $derived(store.isLoadingMore());

	onMount(() => {
		store.loadFirstPage();
		return () => {
			store.clearDebounce();
			if (owned) store.reset();
		};
	});

	onDestroy(() => {
		store.clearDebounce();
		if (owned) store.reset();
	});
</script>

<CustomSelect
	options={options}
	value={value}
	onChange={onChange}
	placeholder={placeholder}
	ariaLabel={ariaLabel}
	class={className}
	searchable
	onSearch={store.search}
	loadMore={store.loadMore}
	allLoaded={!hasMore}
	loadingMore={loadingMore}
/>
