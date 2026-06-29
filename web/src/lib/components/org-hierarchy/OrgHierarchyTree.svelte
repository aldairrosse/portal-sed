<script lang="ts">
	import type { OrgNode } from '$lib/types/org-hierarchy';
	import TreeNode from './TreeNode.svelte';
	import PageSkeleton from '$lib/components/ui/PageSkeleton.svelte';
	import ErrorState from '$lib/components/ui/ErrorState.svelte';

	interface Props {
		viewType?: 'users' | 'departments';
		node?: OrgNode | null;
		onNodeSelect?: (node: OrgNode) => void;
		selectedNodeId?: string;
		maxDepth?: number;
		/** IDs of nodes that start expanded (default: root only) */
		initialExpandedIds?: string[];
		/** Show loading skeleton instead of tree */
		loading?: boolean;
		/** Show error state with optional retry */
		error?: string | null;
		/** Called when the retry button is clicked */
		onretry?: () => void;
	}

	let {
		node = null,
		onNodeSelect = () => {},
		selectedNodeId = '',
		maxDepth = 99,
		initialExpandedIds = [],
		loading = false,
		error = null,
		viewType = 'users',
		onretry
	}: Props = $props();
</script>

{#if loading}
	<div class="p-4">
		<PageSkeleton variant="card" rows={4} />
	</div>
{:else if error}
	<ErrorState message={error} {onretry} />
{:else if node?.children}
	<ul class="menu bg-base-100 w-full text-sm p-0">
		<TreeNode
			{node}
			{onNodeSelect}
			{selectedNodeId}
			{maxDepth}
			depth={0}
			{viewType}
			initialExpanded={initialExpandedIds.includes(node.id)}
			{initialExpandedIds}
		/>
	</ul>
{:else if node}
	<div class="p-4 text-center text-sm text-base-content/40">Sin datos</div>
{/if}
