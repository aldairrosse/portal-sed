import { client } from '$lib/api/client';

export type AssignmentStatus = 'no_iniciado' | 'borrador' | 'enviada';

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
	assignmentStatus: AssignmentStatus;
	assignmentId?: string;
}

function normalizeAssignmentStatus(raw: unknown): AssignmentStatus {
	return raw === 'enviada' || raw === 'borrador' || raw === 'no_iniciado' ? raw : 'no_iniciado';
}

export function getAssignmentStatus(employeeId: string): AssignmentStatus {
	const item = items.find((x) => x.id === employeeId);
	return item?.assignmentStatus ?? 'no_iniciado';
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

/** Employee ID for the current evaluator, set by init(). */
let employeeId = '';

// ─── Error helpers ───────────────────────────────────────────────────────────────

function apiErrorMessage(payload: unknown, fallback: string): string {
	const msg = (payload as { error?: { message?: string } })?.error?.message;
	return msg ?? fallback;
}

// ─── Init ─────────────────────────────────────────────────────────────────────────

export function init(empId: string): void {
	employeeId = empId;
}

// ─── Load ────────────────────────────────────────────────────────────────────────

export async function load(): Promise<void> {
	loading = true;
	error = null;

	try {
		const offset = currentPage * PAGE_SIZE;
		// ponytail: cast needed until PR 1 updates the schema to include query params
		const res = await (client as unknown as { GET(url: string, init: unknown): Promise<{ data?: unknown; error?: unknown }> }).GET('/employees/{empId}/evaluatees', {
			params: {
				path: { empId: employeeId },
				query: { q: currentQ, offset, limit: PAGE_SIZE }
			}
		});

		if (res.error) {
			throw new Error(apiErrorMessage(res.error, 'Error al cargar evaluados'));
		}

		const body = res.data as {
			data?: Array<Record<string, unknown>>;
			meta?: { hasMore?: boolean; total?: number };
		};

		items = (body.data ?? []).map((raw) => {
			const r = raw as Record<string, unknown> & EmployeeListItemExtended & { assignment_status?: unknown; assignmentStatus?: unknown; assignment_id?: unknown; assignmentId?: unknown };
			return {
				...r,
				assignmentStatus: normalizeAssignmentStatus(r.assignmentStatus ?? r.assignment_status),
				assignmentId: (r.assignmentId ?? r.assignment_id ?? undefined) as string | undefined
			} as EmployeeListItemExtended;
		});
		hasMore = body.meta?.hasMore ?? false;
		hasPrev = currentPage > 0;
		apiTotal = body.meta?.total ?? 0;
	} catch (e) {
		error = e instanceof Error ? e.message : 'Error desconocido al cargar evaluados';
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
