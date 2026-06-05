import type { OrgNode } from '$lib/types/org-hierarchy';

import orgTreeData from '$lib/fixtures/org-hierarchy/org-tree.json';

// ─── State ────────────────────────────────────────────────────────────────────

let tree = $state<OrgNode>(structuredClone(orgTreeData as OrgNode));

// ─── Traversal helpers ────────────────────────────────────────────────────────

function findNode(root: OrgNode, nodeId: string): OrgNode | null {
	if (root.id === nodeId) return root;
	const queue: OrgNode[] = [...root.children];
	while (queue.length > 0) {
		const node = queue.shift()!;
		if (node.id === nodeId) return node;
		queue.push(...node.children);
	}
	return null;
}

function dfsDescendants(node: OrgNode): OrgNode[] {
	const result: OrgNode[] = [];
	const stack = [...node.children];
	while (stack.length > 0) {
		const current = stack.pop()!;
		result.push(current);
		stack.push(...current.children);
	}
	return result;
}

function dfsLeafIds(node: OrgNode): string[] {
	const result: string[] = [];
	const stack = [...node.children];
	while (stack.length > 0) {
		const current = stack.pop()!;
		if (current.children.length === 0) {
			result.push(current.id);
		} else {
			stack.push(...current.children);
		}
	}
	return result;
}

function cloneSubtree(node: OrgNode): OrgNode {
	return {
		id: node.id,
		name: node.name,
		profileId: node.profileId,
		managerId: node.managerId,
		children: node.children.map((child) => cloneSubtree(child))
	};
}

// ─── Getters ──────────────────────────────────────────────────────────────────

export function getRoot(): OrgNode {
	return tree;
}

export function getChildren(nodeId: string): OrgNode[] {
	const node = findNode(tree, nodeId);
	return node ? [...node.children] : [];
}

export function getDescendants(nodeId: string): OrgNode[] {
	const node = findNode(tree, nodeId);
	if (!node) return [];
	return dfsDescendants(node);
}

export function getSubtree(nodeId: string): OrgNode | null {
	const node = findNode(tree, nodeId);
	if (!node) return null;
	return cloneSubtree(node);
}

export function getNodeById(nodeId: string): OrgNode | null {
	return findNode(tree, nodeId);
}

export function getScopeIds(nodeId: string): string[] {
	const node = findNode(tree, nodeId);
	if (!node) return [];
	const descendants = dfsDescendants(node);
	return [node.id, ...descendants.map((n) => n.id)];
}

export function getDepth(nodeId: string): number {
	let depth = 0;
	let currentId: string | null = nodeId;
	while (currentId && currentId !== tree.id) {
		const node = findNode(tree, currentId);
		if (!node || !node.managerId) break;
		depth++;
		currentId = node.managerId;
	}
	return depth;
}

export function getAllLeafIds(nodeId: string): string[] {
	const node = findNode(tree, nodeId);
	if (!node) return [];
	return dfsLeafIds(node);
}

// ─── Mutations ────────────────────────────────────────────────────────────────

export function replaceTree(newTree: OrgNode): void {
	tree = structuredClone(newTree);
}
