<script lang="ts">
	import type { OrgNode } from '$lib/types/org-hierarchy';
	import TreeNode from './TreeNode.svelte';

	interface Props {
		node: OrgNode;
		onNodeSelect?: (node: OrgNode) => void;
		selectedNodeId?: string;
		maxDepth?: number;
		depth: number;
		initialExpanded?: boolean;
		initialExpandedIds?: string[];
	}

	let {
		node,
		onNodeSelect = () => {},
		selectedNodeId = '',
		maxDepth = 99,
		depth,
		initialExpanded = false,
		initialExpandedIds = []
	}: Props = $props();

	const isExpandable = $derived(node.children.length > 0 && depth < maxDepth);
	const isSelected = $derived(selectedNodeId === node.id);

	// Local state to persist open/close across re-renders
	let isOpen = $state(initialExpanded);

	function handleSummaryClick(e: MouseEvent) {
		// Toggle is handled natively by <details>
		// Just select the node
		onNodeSelect(node);
	}
</script>

<li>
	{#if isExpandable}
		<details bind:open={isOpen}>
			<summary
				class="flex items-center gap-2 cursor-pointer flex-grow"
				class:menu-active={isSelected}
				onclick={handleSummaryClick}
			>
				<div class="flex items-center gap-2 w-full text-left">
					<span class="font-medium">{node.name}</span>
					<span class="badge badge-ghost badge-xs capitalize">
						{node.profileId.replace('-', ' ')}
					</span>
				</div>
			</summary>
			<ul>
				{#each node.children as child (child.id)}
					<TreeNode
						node={child}
						{onNodeSelect}
						{selectedNodeId}
						{maxDepth}
						depth={depth + 1}
						initialExpanded={initialExpandedIds.includes(child.id)}
						{initialExpandedIds}
					/>
				{/each}
			</ul>
		</details>
	{:else}
		<button
			type="button"
			class="flex items-center gap-2 w-full text-left cursor-pointer"
			class:menu-active={isSelected}
			onclick={(e) => {
				e.stopPropagation();
				onNodeSelect(node);
			}}
		>
			<span class="font-medium">{node.name}</span>
			<span class="badge badge-ghost badge-xs capitalize">
				{node.profileId.replace('-', ' ')}
			</span>
		</button>
	{/if}
</li>
