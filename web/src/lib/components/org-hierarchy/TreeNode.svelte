<script lang="ts">
	import type { OrgNode } from '$lib/types/org-hierarchy';
	import TreeNode from './TreeNode.svelte';

	interface EmployeeLeaf {
		id: string;
		firstName: string;
		lastName: string;
		jobTitle?: string;
		profileDescription?: string;
	}

	interface Props {
		node: OrgNode;
		onNodeSelect?: (node: OrgNode) => void;
		selectedNodeId?: string;
		maxDepth?: number;
		depth: number;
		initialExpanded?: boolean;
		initialExpandedIds?: string[];
		viewType?: 'users' | 'departments';
		/** Employees to show as leaf nodes under each nodeId */
		employeeLeaves?: Record<string, EmployeeLeaf[]>;
		/** Called when an employee leaf is clicked */
		onEmployeeSelect?: (emp: EmployeeLeaf) => void;
		/** Currently selected employee leaf id for highlighting */
		selectedEmployeeId?: string;
	}

	let {
		node,
		onNodeSelect = () => {},
		selectedNodeId = '',
		maxDepth = 99,
		depth,
		initialExpanded = false,
		initialExpandedIds = [],
		viewType = 'users',
		employeeLeaves = {},
		onEmployeeSelect = () => {},
		selectedEmployeeId = ''
	}: Props = $props();

	const children = $derived(node.children ?? []);
	const isSelected = $derived(selectedNodeId === node.id);
	// ponytail: filter head employee from leaves — already visible as node header
	const employeeChildren = $derived(
		(employeeLeaves[node.id] ?? []).filter(
			(e) => !node.headEmployee || e.id !== node.headEmployee.id,
		),
	);
	// Track if this node was ever expandable or clicked — prevents button→details swap
	let wasExpandable = $derived(children.length > 0);

	const hasEmployees = $derived((node.employeeCount ?? 0) > 0);
	const isExpandable = $derived.by(() => {
		if (depth >= maxDepth) return false;
		// ponytail: departments view — only child departments matter, not employee leaves
		if (viewType === 'departments') return wasExpandable;
		return wasExpandable || hasEmployees || employeeChildren.length > 0;
	});

	$effect(() => {
		if (children.length > 0) wasExpandable = true;
	});

	// Local state to persist open/close across re-renders
	let isOpen = $derived(initialExpanded)

	function handleSummaryClick(_e: MouseEvent) {
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
					<span class="truncate font-medium flex-grow">
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
						{employeeLeaves}
						{onEmployeeSelect}
						{selectedEmployeeId}
					/>
				{/each}
				<!-- Employee leaf nodes -->
				{#each employeeChildren as emp (emp.id)}
					<li>
						<button
							type="button"
							class="flex items-center gap-2 w-full text-left cursor-pointer"
							class:menu-active={selectedEmployeeId === emp.id}
							onclick={(e) => {
								e.stopPropagation();
								onEmployeeSelect(emp);
							}}
						>
						<span class="text-sm font-medium">
								{emp.firstName} {emp.lastName}
							</span>
							{#if emp.jobTitle}
								<span class="truncate text-xs text-base-content/40 ml-auto">
									{emp.jobTitle}
								</span>
							{/if}
						</button>
					</li>
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
			<span class="truncate font-medium flex-grow">
				{nodeTitle(node)}
			</span>
			<span class="truncate text-xs text-base-content/40">
				{nodeSubtitle(node)}
			</span>
		</button>
	{/if}
</li>
