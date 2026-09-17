<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import CustomSelect from './CustomSelect.svelte';
	import {
		createEmployeePickerStore,
		type EmployeePickerStore,
	} from '$lib/stores/employeePickerStore.svelte';

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
		picker,
	}: Props = $props();

	const store = $derived(picker ?? createEmployeePickerStore());
	const owned = $derived(!picker);

	const options = $derived(store.getOptions());
	const hasMore = $derived(store.hasMoreEmployees());
	const loadingMore = $derived(store.isLoadingMore());
	const loading = $derived(store.isLoading());
	const error = $derived(store.getError());

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
	{options}
	{value}
	{onChange}
	{placeholder}
	{ariaLabel}
	class={className}
	searchable
	onSearch={store.search}
	loadMore={store.loadMore}
	allLoaded={!hasMore}
	{loadingMore}
	{loading}
	{error}
/>
