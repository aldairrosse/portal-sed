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
		viewType?: 'users' | 'departments';
	}

	let {
		node,
		onNodeSelect = () => {},
		selectedNodeId = '',
		maxDepth = 99,
		depth,
		initialExpanded = false,
		initialExpandedIds = [],
		viewType = 'users'
	}: Props = $props();

	const children = $derived(node.children ?? []);
	const isExpandable = $derived(children.length > 0 && depth < maxDepth);
	const isSelected = $derived(selectedNodeId === node.id);

	// Local state to persist open/close across re-renders
	let isOpen = $derived(initialExpanded)

	function handleSummaryClick(_e: MouseEvent) {
		// Toggle is handled natively by <details>
		// Just select the node
		onNodeSelect(node);
	}

	function nodeTitle(node: OrgNode) {
		if (viewType === 'users') {
			return node.headEmployee ? `${node.headEmployee.firstName} ${node.headEmployee.lastName}` : node.name;
		}
		return node.name;
	}

	function nodeSubtitle(node: OrgNode) {
		if (viewType === 'users') {
			return node.headEmployee?.jobTitle ?? '';
		}
		return '';
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
					<span class="font-medium flex-grow">
						{nodeTitle(node)}
					</span>
					<span class="truncate text-xs text-base-content/40 pr-2">
						{nodeSubtitle(node)}
					</span>
				</div>
			</summary>
			<ul>
				{#each children as child (child.id)}
					<TreeNode
						{viewType}
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
			<span class="font-medium">
				{nodeTitle(node)}
			</span>
			<span class="truncate text-sm text-base-content/40">
				{nodeSubtitle(node)}
			</span>
		</button>
	{/if}
</li>
