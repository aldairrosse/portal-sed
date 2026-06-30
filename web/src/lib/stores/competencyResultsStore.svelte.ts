import { client } from '$lib/api/client';

// ─── Types ──────────────────────────────────────────────────────────────────────

interface CompetencyResultItem {
	id: string;
	name: string;
	profileName: string;
	selfRatingAvg?: number | null;
	rhRatingAvg?: number | null;
	status: string;
}

interface CompetencyResultsResponse {
	data: CompetencyResultItem[];
	meta: {
		hasMore: boolean;
		total: number;
		offset: number;
		limit: number;
	};
}

// ─── Reactive state ──────────────────────────────────────────────────────────────

let items = $state<CompetencyResultItem[]>([]);
let loading = $state(false);
let error = $state<string | null>(null);
let hasMore = $state(false);
let hasPrev = $state(false);
let apiTotal = $state(0);
let scopeFilter = $state<'all' | 'team'>('all');

// ─── Reactive getters ─────────────────────────────────────────────────────────

export function getItems(): CompetencyResultItem[] { return items; }
export function isLoading(): boolean { return loading; }
export function getError(): string | null { return error; }
export function hasMoreItems(): boolean { return hasMore; }
export function hasPrevItems(): boolean { return hasPrev; }
export function getCurrentPage(): number { return currentPage; }
export function getTotalCount(): number { return apiTotal; }
export function getScope(): 'all' | 'team' { return scopeFilter; }

// ─── Internal pagination state ───────────────────────────────────────────────────

let currentPage = $state(0);
const PAGE_SIZE = 50;

/** Current active search query (if any). */
let currentQ: string | undefined;

/** Cycle ID set via init(). */
let cycleId: string | undefined;

// ─── Error helpers ───────────────────────────────────────────────────────────────

function apiErrorMessage(payload: unknown, fallback: string): string {
	const msg = (payload as { error?: { message?: string } })?.error?.message;
	return msg ?? fallback;
}

// ─── Init ────────────────────────────────────────────────────────────────────────

export function init(id: string): void {
	cycleId = id;
	load();
}

// ─── Load ────────────────────────────────────────────────────────────────────────

export async function load(): Promise<void> {
	if (!cycleId) return;

	loading = true;
	error = null;

	try {
		const offset = currentPage * PAGE_SIZE;
		const res = await (client as any).GET('/evaluations/competency-results', {
			params: {
				query: { cycle_id: cycleId, q: currentQ, scope: scopeFilter, offset, limit: PAGE_SIZE }
			}
		});

		if (res.error) {
			throw new Error(apiErrorMessage(res.error, 'Error al cargar resultados de competencias'));
		}

		const body = res.data as CompetencyResultsResponse;

		items = body.data ?? [];
		hasMore = body.meta?.hasMore ?? false;
		hasPrev = currentPage > 0;
		apiTotal = body.meta?.total ?? 0;
	} catch (e) {
		error = e instanceof Error ? e.message : 'Error desconocido al cargar resultados de competencias';
		items = [];
	} finally {
		loading = false;
	}
}

// ─── Pagination ──────────────────────────────────────────────────────────────────

export async function next(): Promise<void> {
	if (!hasMore) return;
	currentPage++;
	await load();
}

export async function prev(): Promise<void> {
	if (currentPage <= 0) return;
	currentPage--;
	await load();
}

// ─── Search with debounce ────────────────────────────────────────────────────────

let debounceTimer: ReturnType<typeof setTimeout> | undefined;

export function search(query: string): void {
	if (debounceTimer) clearTimeout(debounceTimer);

	debounceTimer = setTimeout(() => {
		currentPage = 0;
		currentQ = query || undefined;
		load();
	}, 300);
}

// ─── Scope filter ────────────────────────────────────────────────────────────────

export function setScope(s: 'all' | 'team'): void {
	scopeFilter = s;
	currentPage = 0;
	load();
}
