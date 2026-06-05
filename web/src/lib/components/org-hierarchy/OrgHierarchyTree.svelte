<script lang="ts">
	import type { OrgNode } from '$lib/types/org-hierarchy';
	import TreeNode from './TreeNode.svelte';

	interface Props {
		node: OrgNode;
		onNodeSelect?: (node: OrgNode) => void;
		selectedNodeId?: string;
		maxDepth?: number;
		/** IDs of nodes that start expanded (default: root only) */
		initialExpandedIds?: string[];
	}

	let {
		node,
		onNodeSelect = () => {},
		selectedNodeId = '',
		maxDepth = 99,
		initialExpandedIds = [node.id]
	}: Props = $props();
</script>

<ul class="menu bg-base-100 w-full text-sm p-0">
	<TreeNode
		{node}
		{onNodeSelect}
		{selectedNodeId}
		{maxDepth}
		depth={0}
		initialExpanded={initialExpandedIds.includes(node.id)}
		{initialExpandedIds}
	/>
</ul>
