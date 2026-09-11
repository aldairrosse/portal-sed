<script lang="ts">
	import type { OrgNode } from '$lib/types/org-hierarchy';
	import { ChevronDown } from '@lucide/svelte';
	import DepartmentPagedMenu from './DepartmentPagedMenu.svelte';

	interface Props {
		nodes: OrgNode[];
		selectedId: string;
		groupName: string;
		onchange: (id: string) => void;
		placeholder?: string;
		label?: string;
		maxWidthClass?: string;
	}

	let {
		nodes,
		selectedId,
		groupName,
		onchange,
		placeholder = 'Seleccionar departamento',
		label,
		maxWidthClass = 'max-w-xs',
	}: Props = $props();

	let open = $state(false);

	function findName(list: OrgNode[], id: string): string {
		for (const node of list) {
			if (node.id === id) return node.name;
			const found = findName(node.children ?? [], id);
			if (found) return found;
		}
		return '';
	}

	const selectedLabel = $derived(selectedId ? findName(nodes, selectedId) : '');
</script>

<div class="w-full {maxWidthClass}">
	{#if label}
		<span class="label"><span class="label-text text-xs">{label}</span></span>
	{/if}
	<div class="dropdown dropdown-top w-full {maxWidthClass}">
		<button
			type="button"
			class="input input-bordered input-sm text-left flex items-center justify-between gap-2 cursor-pointer h-8 min-h-0 text-xs truncate w-full {maxWidthClass}"
			onclick={() => (open = !open)}
			aria-expanded={open}
			aria-haspopup="listbox"
		>
			<span class="truncate">{selectedLabel || placeholder}</span>
			<span
				class="shrink-0 transition-transform duration-200 {open
					? 'rotate-180'
					: ''}"
			>
				<ChevronDown class="w-3.5 h-3.5" />
			</span>
		</button>
		{#if open}
			<button
				type="button"
				class="fixed inset-0 z-10 bg-transparent border-0 p-0 cursor-default"
				aria-label="Cerrar menú"
				onclick={() => (open = false)}
				onkeydown={(e) => {
					if (e.key === 'Escape' || e.key === 'Enter' || e.key === ' ')
						open = false;
				}}
			></button>
			<div
				class="dropdown-content bg-base-100 rounded-box shadow-lg mb-2 p-1 w-full max-w-full max-h-none z-100"
			>
				<DepartmentPagedMenu
					{nodes}
					{selectedId}
					{groupName}
					onchange={(id) => {
						onchange(id);
						open = false;
					}}
				/>
			</div>
		{/if}
	</div>
</div>
