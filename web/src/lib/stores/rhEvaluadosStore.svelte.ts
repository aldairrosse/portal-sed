import { client } from '$lib/api/client';

// ponytail: remove when OpenAPI schema includes profileName/jobTitle
interface EmployeeListItemExtended {
	id: string;
	firstName: string;
	lastName: string;
	email: string;
	employeeNumber: string;
	orgNodeId: string;
	managerId?: string;
	profileId: string;
	profileName: string;
	jobTitle?: string;
	profileDescription?: string;
	isActive: boolean;
}

// ─── Reactive state ──────────────────────────────────────────────────────────────

let items = $state<EmployeeListItemExtended[]>([]);
let loading = $state(false);
let error = $state<string | null>(null);
let hasMore = $state(false);
let hasPrev = $state(false);
let apiTotal = $state(0);

// ─── Reactive getters ─────────────────────────────────────────────────────────

export function getItems(): EmployeeListItemExtended[] { return items; }
export function isLoading(): boolean { return loading; }
export function getError(): string | null { return error; }
export function hasMoreItems(): boolean { return hasMore; }
export function hasPrevItems(): boolean { return hasPrev; }
export function getCurrentPage(): number { return currentPage; }
export function getTotalCount(): number { return apiTotal; }
export function getSearchQuery(): string { return currentQ ?? ''; }

// ─── Internal pagination state ───────────────────────────────────────────────────

let currentPage = $state(0);
const PAGE_SIZE = 50;

/** Current active search query (if any). */
let currentQ: string | undefined;

// ─── Error helpers ───────────────────────────────────────────────────────────────

function apiErrorMessage(payload: unknown, fallback: string): string {
	const msg = (payload as { error?: { message?: string } })?.error?.message;
	return msg ?? fallback;
}

// ─── Load ────────────────────────────────────────────────────────────────────────

export async function load(): Promise<void> {
	loading = true;
	error = null;

	try {
		const offset = currentPage * PAGE_SIZE;
		const res = await client.GET('/employees', {
			params: {
				query: { q: currentQ, offset, limit: PAGE_SIZE }
			}
		});

		if (res.error) {
			throw new Error(apiErrorMessage(res.error, 'Error al cargar empleados'));
		}

		const body = res.data as {
			data?: EmployeeListItemExtended[];
			meta?: { hasMore?: boolean; total?: number };
		};

		items = body.data ?? [];
		hasMore = body.meta?.hasMore ?? false;
		hasPrev = currentPage > 0;
		apiTotal = body.meta?.total ?? 0;
	} catch (e) {
		error = e instanceof Error ? e.message : 'Error desconocido al cargar empleados';
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
