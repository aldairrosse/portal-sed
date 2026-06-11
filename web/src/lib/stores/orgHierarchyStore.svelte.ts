import type { OrgNode } from '$lib/types/org-hierarchy';
import orgTreeData from '$lib/fixtures/org-hierarchy/org-tree.json';
import { client } from '$lib/api/client';

// ─── Internal data shape ──────────────────────────────────────────────────────

interface StoreData {
	root: OrgNode;
}

// ─── Triplete state ───────────────────────────────────────────────────────────

let data = $state<StoreData | null>(null);
let loading = $state(true);
let error = $state<string | null>(null);

// ─── Traversal helpers (pure functions, unchanged) ────────────────────────────

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

// ─── Fixture loader ───────────────────────────────────────────────────────────

function loadFixtures(): StoreData {
	return { root: structuredClone(orgTreeData as OrgNode) };
}

/**
 * Load org tree data.
 *
 * In DEV without VITE_USE_API: loads from fixture file (structured clone).
 * In production / VITE_USE_API=true: fetches from the real API endpoint.
 */
export async function load(): Promise<void> {
	loading = true;
	error = null;

	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		data = loadFixtures();
		loading = false;
		return;
	}

	try {
		// 1. Discover the first corporate tree
		const treesRes = await client.GET('/org-trees' as never, {
			params: { query: { type: 'corporate' as never } }
		});
		if (treesRes.error) {
			throw new Error('Error al cargar árbol organizacional');
		}

		const trees = (treesRes.data as { data?: Array<{ id?: string }> })?.data ?? [];
		if (trees.length === 0) {
			throw new Error('No hay árboles organizacionales disponibles');
		}

		const treeId = trees[0].id;
		if (!treeId) {
			throw new Error('ID de árbol no disponible');
		}

		// 2. Fetch the full nested tree
		const nodesRes = await client.GET('/org-trees/{treeId}/nodes' as never, {
			params: {
				query: { format: 'nested' as never, depth: -1 },
				path: { treeId }
			}
		});
		if (nodesRes.error) {
			throw new Error('Error al cargar nodos del árbol');
		}

		const rootNode = (nodesRes.data as { data?: OrgNode })?.data;
		if (!rootNode) {
			throw new Error('No se recibieron datos del árbol');
		}

		data = { root: structuredClone(rootNode) };
	} catch (e) {
		error = e instanceof Error ? e.message : 'Error desconocido al cargar árbol organizacional';
	} finally {
		loading = false;
	}
}

/** Alias for load(). */
export function reload(): Promise<void> {
	return load();
}

// ─── Getters ──────────────────────────────────────────────────────────────────

export function getRoot(): OrgNode | null {
	return data?.root ?? null;
}

export function getChildren(nodeId: string): OrgNode[] {
	if (!data) return [];
	const node = findNode(data.root, nodeId);
	return node ? [...node.children] : [];
}

export function getDescendants(nodeId: string): OrgNode[] {
	if (!data) return [];
	const node = findNode(data.root, nodeId);
	if (!node) return [];
	return dfsDescendants(node);
}

export function getSubtree(nodeId: string): OrgNode | null {
	if (!data) return null;
	const node = findNode(data.root, nodeId);
	if (!node) return null;
	return cloneSubtree(node);
}

export function getNodeById(nodeId: string): OrgNode | null {
	if (!data) return null;
	return findNode(data.root, nodeId);
}

export function getScopeIds(nodeId: string): string[] {
	if (!data) return [];
	const node = findNode(data.root, nodeId);
	if (!node) return [];
	const descendants = dfsDescendants(node);
	return [node.id, ...descendants.map((n) => n.id)];
}

export function getDepth(nodeId: string): number {
	if (!data) return 0;
	let depth = 0;
	let currentId: string | null = nodeId;
	while (currentId && currentId !== data.root.id) {
		const node = findNode(data.root, currentId);
		if (!node || !node.managerId) break;
		depth++;
		currentId = node.managerId;
	}
	return depth;
}

export function getAllLeafIds(nodeId: string): string[] {
	if (!data) return [];
	const node = findNode(data.root, nodeId);
	if (!node) return [];
	return dfsLeafIds(node);
}

// ─── Loading / error state accessors ──────────────────────────────────────────

export function isLoading(): boolean {
	return loading;
}

export function getError(): string | null {
	return error;
}

// ─── Mutations ────────────────────────────────────────────────────────────────

export function replaceTree(newTree: OrgNode): void {
	data = { root: structuredClone(newTree) };
}
