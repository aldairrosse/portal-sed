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

const PAGE_SIZE = 50;

function apiErrorMessage(payload: unknown, fallback: string): string {
	const msg = (payload as { error?: { message?: string } })?.error?.message;
	return msg ?? fallback;
}

async function fetchEmployees(query: string | undefined, offset: number): Promise<EmployeesBody> {
	// ponytail: path-level query is `never` in generated types but operation supports q/offset/limit
	const res = (await client.GET('/employees', {
		params: { query: { q: query, offset, limit: PAGE_SIZE } }
	})) as { data?: unknown; error?: unknown };
	if (res.error) throw new Error(apiErrorMessage(res.error, 'Error al cargar empleados'));
	return (res.data as EmployeesBody) ?? {};
}

async function fetchDevEmployees(query: string | undefined, offset: number): Promise<EmployeesBody> {
	const params = new URLSearchParams();
	if (query) params.set('q', query);
	params.set('offset', String(offset));
	params.set('limit', String(PAGE_SIZE));
	const res = await fetch(`/api/v1/dev/employees?${params.toString()}`, { credentials: 'include' });
	if (!res.ok) throw new Error('Error al cargar empleados');
	const body = (await res.json().catch(() => ({}))) as {
		data?: Array<Record<string, unknown>>;
		employees?: Array<Record<string, unknown>>;
		meta?: { hasMore?: boolean };
	};
	const raw = (body.data ?? body.employees ?? []) as Array<{
		id: string;
		firstName?: string;
		first_name?: string;
		lastName?: string;
		last_name?: string;
	}>;
	const data: EmployeeListItem[] = raw.map((r) => ({
		id: String(r.id),
		firstName: String(r.firstName ?? r.first_name ?? ''),
		lastName: String(r.lastName ?? r.last_name ?? '')
	}));
	const hasMore = body.meta?.hasMore ?? raw.length === PAGE_SIZE;
	return { data, meta: { hasMore } };
}

function toOption(e: EmployeeListItem): EmployeeOption {
	return { value: e.id, label: `${e.firstName} ${e.lastName}`.trim() };
}

class EmployeePickerStoreImpl {
	items = $state<EmployeeOption[]>([]);
	loading = $state(false);
	loadingMore = $state(false);
	error = $state<string | null>(null);
	hasMore = $state(false);
	allLoaded = $state(false);
	currentQuery = $state('');
	page = $state(0);
	debounceTimer: ReturnType<typeof setTimeout> | undefined = undefined;
	useDev = false;
	activeEmployeeId = $state<string | null>(null);
	activeEmployeeLabel = $state<string | null>(null);

	getOptions = (): EmployeeOption[] => {
		const current = this.items;
		if (!this.activeEmployeeId) return current;

		const activeIdx = current.findIndex((o) => o.value === this.activeEmployeeId);
		if (activeIdx !== -1) {
			const active = current[activeIdx];
			const others = current.filter((o) => o.value !== this.activeEmployeeId);
			return [active, ...others];
		}

		if (this.activeEmployeeLabel) {
			return [{ value: this.activeEmployeeId, label: this.activeEmployeeLabel }, ...current];
		}

		return current;
	};
	isLoading = (): boolean => this.loading;
	isLoadingMore = (): boolean => this.loadingMore;
	hasMoreEmployees = (): boolean => this.hasMore;
	getError = (): string | null => this.error;

	setActive = (id: string | null, label?: string): void => {
		this.activeEmployeeId = id;
		this.activeEmployeeLabel = label ?? null;
	};

	clearDebounce = (): void => {
		if (this.debounceTimer) clearTimeout(this.debounceTimer);
		this.debounceTimer = undefined;
	};

	reset = (): void => {
		this.clearDebounce();
		this.items = [];
		this.loading = false;
		this.loadingMore = false;
		this.error = null;
		this.hasMore = false;
		this.allLoaded = false;
		this.currentQuery = '';
		this.page = 0;
		this.activeEmployeeId = null;
	};

	loadFirstPage = async (query?: string): Promise<void> => {
		this.loading = true;
		this.error = null;
		this.page = 0;
		this.currentQuery = query ?? '';
		this.items = [];
		try {
			const fetcher = this.useDev ? fetchDevEmployees : fetchEmployees;
			const body = await fetcher(query, 0);
			this.items = (body.data ?? []).map(toOption);
			this.hasMore = body.meta?.hasMore ?? false;
			this.allLoaded = !this.hasMore && !query;
		} catch (e) {
			this.error = e instanceof Error ? e.message : 'Error desconocido al cargar empleados';
			this.items = [];
			this.hasMore = false;
			this.allLoaded = false;
		} finally {
			this.loading = false;
		}
	};

	loadMore = async (): Promise<void> => {
		if (!this.hasMore || this.loadingMore) return;
		this.loadingMore = true;
		try {
			const nextPage = this.page + 1;
			const fetcher = this.useDev ? fetchDevEmployees : fetchEmployees;
			const body = await fetcher(this.currentQuery || undefined, nextPage * PAGE_SIZE);
			const fresh = (body.data ?? []).map(toOption);
			const seen: Record<string, true> = {};
			for (const i of this.items) seen[i.value] = true;
			this.items = [...this.items, ...fresh.filter((i) => !seen[i.value])];
			this.page = nextPage;
			this.hasMore = body.meta?.hasMore ?? false;
			this.allLoaded = !this.hasMore;
		} catch (e) {
			this.error = e instanceof Error ? e.message : 'Error desconocido al cargar empleados';
		} finally {
			this.loadingMore = false;
		}
	};

	search = (q: string): void => {
		if (this.debounceTimer) clearTimeout(this.debounceTimer);
		this.debounceTimer = setTimeout(() => {
			if (q === '' && this.allLoaded) {
				this.currentQuery = '';
				return;
			}
			this.loadFirstPage(q || undefined);
		}, 300);
	};

	clearSearch = (): void => {
		this.search('');
	};

	destroy = (): void => {
		this.clearDebounce();
	};
}

export function createEmployeePickerStore(opts?: { useDev?: boolean }): EmployeePickerStoreImpl {
	const s = new EmployeePickerStoreImpl();
	if (opts?.useDev) s.useDev = true;
	return s;
}

export type EmployeePickerStore = EmployeePickerStoreImpl;

// ─── Singleton for backward compat (GlobalGoalCreateForm) ────────────────────
const _default = new EmployeePickerStoreImpl();
export const getEmployeeOptions = _default.getOptions;
export const isLoading = _default.isLoading;
export const isLoadingMore = _default.isLoadingMore;
export const hasMoreEmployees = _default.hasMoreEmployees;
export const getError = _default.getError;
export const loadFirstPage = _default.loadFirstPage;
export const loadMore = _default.loadMore;
export const search = _default.search;
export const clearSearch = _default.clearSearch;
export const clearDebounce = _default.clearDebounce;
export const resetStore = _default.reset;
