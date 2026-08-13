import { client } from '$lib/api/client';

interface EmployeeOption {
	value: string;
	label: string;
}

interface EmployeeListItem {
	id: string;
	firstName: string;
	lastName: string;
}

interface EmployeesBody {
	data?: EmployeeListItem[];
	meta?: { hasMore?: boolean };
}

// ─── Reactive state ──────────────────────────────────────────────────────────────

let items = $state<EmployeeOption[]>([]);
let loading = $state(false);
let loadingMore = $state(false);
let error = $state<string | null>(null);
let hasMore = $state(false);
let allLoaded = $state(false);
let currentQuery = $state('');

let page = $state(0);
const PAGE_SIZE = 50;

// ─── Reactive getters ─────────────────────────────────────────────────────────

export function getEmployeeOptions(): EmployeeOption[] { return items; }
export function isLoading(): boolean { return loading; }
export function isLoadingMore(): boolean { return loadingMore; }
export function hasMoreEmployees(): boolean { return hasMore; }
export function getError(): string | null { return error; }

// ─── Error helpers ───────────────────────────────────────────────────────────────

function apiErrorMessage(payload: unknown, fallback: string): string {
	const msg = (payload as { error?: { message?: string } })?.error?.message;
	return msg ?? fallback;
}

// ─── Fetch ────────────────────────────────────────────────────────────────────────

async function fetchEmployees(query: string | undefined, offset: number): Promise<EmployeesBody> {
	// ponytail: path-level query is `never` in generated types but operation supports q/offset/limit
	const res = await client.GET('/employees', {
		params: { query: { q: query, offset, limit: PAGE_SIZE } }
	}) as { data?: unknown; error?: unknown };

	if (res.error) {
		throw new Error(apiErrorMessage(res.error, 'Error al cargar empleados'));
	}

	return (res.data as EmployeesBody) ?? {};
}

function toOption(e: EmployeeListItem): EmployeeOption {
	return { value: e.id, label: `${e.firstName} ${e.lastName}`.trim() };
}

// ─── Load ────────────────────────────────────────────────────────────────────────

export async function loadFirstPage(query?: string): Promise<void> {
	loading = true;
	error = null;
	page = 0;
	currentQuery = query ?? '';
	items = [];

	try {
		const body = await fetchEmployees(query, 0);
		items = (body.data ?? []).map(toOption);
		hasMore = body.meta?.hasMore ?? false;
		allLoaded = !hasMore && !query;
	} catch (e) {
		error = e instanceof Error ? e.message : 'Error desconocido al cargar empleados';
		items = [];
		hasMore = false;
		allLoaded = false;
	} finally {
		loading = false;
	}
}

export async function loadMore(): Promise<void> {
	if (!hasMore || loadingMore || currentQuery) return;

	loadingMore = true;
	try {
		const nextPage = page + 1;
		const body = await fetchEmployees(undefined, nextPage * PAGE_SIZE);
		const fresh = (body.data ?? []).map(toOption);
		const seen: Record<string, true> = {};
		for (const i of items) seen[i.value] = true;
		items = [...items, ...fresh.filter(i => !seen[i.value])];
		page = nextPage;
		hasMore = body.meta?.hasMore ?? false;
		allLoaded = !hasMore;
	} catch (e) {
		error = e instanceof Error ? e.message : 'Error desconocido al cargar empleados';
	} finally {
		loadingMore = false;
	}
}

// ─── Search with debounce ────────────────────────────────────────────────────────

let debounceTimer: ReturnType<typeof setTimeout> | undefined;

export function search(q: string): void {
	if (debounceTimer) clearTimeout(debounceTimer);

	debounceTimer = setTimeout(() => {
		if (q === '' && allLoaded) {
			currentQuery = '';
			return;
		}
		loadFirstPage(q || undefined);
	}, 300);
}

/** Restore the full in-memory list after closing the picker. */
export function clearSearch(): void {
	search('');
}
