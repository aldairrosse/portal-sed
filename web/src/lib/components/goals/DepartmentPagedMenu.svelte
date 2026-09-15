<script lang="ts">
	import type { OrgNode } from '$lib/types/org-hierarchy';
	import DepartmentPagedMenu from './DepartmentPagedMenu.svelte';

	interface Props {
		nodes: OrgNode[];
		selectedId: string;
		groupName: string;
		onchange: (id: string) => void;
		isChildren?: boolean;
		openIds?: Set<string>;
		menuMaxHeightClass?: string;
	}

	let {
		nodes,
		selectedId,
		groupName,
		onchange,
		isChildren = false,
		openIds,
		menuMaxHeightClass = 'max-h-64',
	}: Props = $props();

	function findPath(
		nodes: OrgNode[],
		target: string,
		acc: string[] = [],
	): string[] | null {
		for (const n of nodes) {
			if (n.id === target) return [...acc, n.id];
			if (n.children?.length) {
				const p = findPath(n.children, target, [...acc, n.id]);
				if (p) return p;
			}
		}
		return null;
	}

	let path = $derived(selectedId ? (findPath(nodes, selectedId) ?? []) : []);
	let ancestors = $derived(new Set(path.slice(0, -1)));
	let effectiveOpen = $derived(openIds ?? ancestors);
</script>

<ul
	class="{!isChildren
		? 'menu menu-sm menu-paged menu-vertical '
		: ''}w-full max-w-full overflow-hidden"
>
	{#each nodes as node (node.id)}
		{#if (node.children ?? []).length > 0}
			<li class="min-w-0 w-full max-w-full overflow-hidden">
				<details
					class="flex-1 min-w-0 w-full max-w-full overflow-hidden"
					open={effectiveOpen.has(node.id)}
				>
					<summary
						aria-label="Atrás"
						class="min-w-0 w-full max-w-full overflow-hidden"
					>
						<span class="flex min-w-0 w-full flex-1 items-center gap-2">
							<input
								type="radio"
								class="radio radio-sm shrink-0"
								name={groupName}
								value={node.id}
								checked={selectedId === node.id}
								onclick={(e) => e.stopPropagation()}
								onchange={() => onchange(node.id)}
								aria-label={node.name}
							/>
							<span class="min-w-0 flex-1 truncate max-w-full">{node.name}</span
							>
						</span>
					</summary>
					<DepartmentPagedMenu
						nodes={node.children ?? []}
						{selectedId}
						{groupName}
						{onchange}
						{menuMaxHeightClass}
						isChildren={true}
						openIds={effectiveOpen}
					/>
				</details>
			</li>
		{:else}
			<li class="min-w-0 w-full max-w-full overflow-hidden">
				<button
					type="button"
					class="flex min-w-0 w-full max-w-full items-center gap-2 overflow-hidden text-left"
					onclick={() => onchange(node.id)}
				>
					<input
						type="radio"
						class="radio radio-sm shrink-0"
						name={groupName}
						value={node.id}
						checked={selectedId === node.id}
						onclick={(e) => {
							e.stopPropagation();
							onchange(node.id);
						}}
						onchange={() => onchange(node.id)}
						aria-label={node.name}
					/>
					<span class="min-w-0 flex-1 truncate max-w-full">{node.name}</span>
				</button>
			</li>
		{/if}
	{/each}
</ul>
