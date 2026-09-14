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
	selfAvg?: number | null;
	self_avg?: number | null;
	rhAvg?: number | null;
	rh_avg?: number | null;
	// Status por fase activa (backend; null => fallback cliente).
	evaluationStatus?: string | null;
	evaluation_status?: string | null;
	metasStatusFase?: string | null;
	metas_status_fase?: string | null;
	phase?: string | null;
	phaseKind?: string | null;
	phase_kind?: string | null;
	goalTotal?: number | null;
	goal_total?: number | null;
	goalDone?: number | null;
	goal_done?: number | null;
	evaluatedCount?: number | null;
	evaluated_count?: number | null;
	totalCompetencies?: number | null;
	total_competencies?: number | null;
}

// ─── Reactive state ──────────────────────────────────────────────────────────────

let items = $state<EmployeeListItemExtended[]>([]);
let loading = $state(false);
let error = $state<string | null>(null);
let hasMore = $state(false);
let hasPrev = $state(false);
let apiTotal = $state(0);
let completedCount = $state<number | null>(null);

// ─── Reactive getters ─────────────────────────────────────────────────────────

export function getItems(): EmployeeListItemExtended[] {
	return items;
}
export function isLoading(): boolean {
	return loading;
}
export function getError(): string | null {
	return error;
}
export function hasMoreItems(): boolean {
	return hasMore;
}
export function hasPrevItems(): boolean {
	return hasPrev;
}
export function getCurrentPage(): number {
	return currentPage;
}
export function getTotalCount(): number {
	return apiTotal;
}
export function getCompletedCount(): number | null {
	return completedCount;
}
export function getSearchQuery(): string {
	return currentQ ?? '';
}

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

function toAvg(raw: unknown): number | null {
	return typeof raw === 'number' ? raw : null;
}

function normalizeEvalStatus(raw: unknown): string | null {
	return raw === 'pending' || raw === 'in-progress' || raw === 'completed'
		? raw
		: null;
}

function toIntOrNull(raw: unknown): number | null {
	return typeof raw === 'number' && Number.isInteger(raw) ? raw : null;
}

// ─── Load ────────────────────────────────────────────────────────────────────────

export async function load(): Promise<void> {
	loading = true;
	error = null;

	try {
		const offset = currentPage * PAGE_SIZE;
		// ponytail: path-level query is `never` in generated types but operation supports q/offset/limit
		const res = (await client.GET('/employees', {
			params: { query: { q: currentQ, offset, limit: PAGE_SIZE } },
		})) as { data?: unknown; error?: unknown };

		if (res.error) {
			throw new Error(apiErrorMessage(res.error, 'Error al cargar empleados'));
		}

		const body = res.data as {
			data?: Array<Record<string, unknown>>;
			meta?: {
				hasMore?: boolean;
				total?: number;
				completedCount?: number;
				completed_count?: number;
			};
		};

		items = (body.data ?? []).map((raw) => {
			const r = raw as Record<string, unknown> & EmployeeListItemExtended;
			return {
				...r,
				selfAvg: toAvg(r.selfAvg ?? r.self_avg),
				rhAvg: toAvg(r.rhAvg ?? r.rh_avg),
				evaluationStatus: normalizeEvalStatus(
					r.evaluationStatus ?? r.evaluation_status,
				),
				metasStatusFase:
					typeof (r.metasStatusFase ?? r.metas_status_fase) === 'string'
						? (r.metasStatusFase ?? r.metas_status_fase) as string
						: null,
				phase: typeof r.phase === 'string' ? r.phase : null,
				phaseKind:
					typeof (r.phaseKind ?? r.phase_kind) === 'string'
						? (r.phaseKind ?? r.phase_kind) as string
						: null,
				goalTotal: toIntOrNull(r.goalTotal ?? r.goal_total),
				goalDone: toIntOrNull(r.goalDone ?? r.goal_done),
				evaluatedCount: toIntOrNull(r.evaluatedCount ?? r.evaluated_count),
				totalCompetencies: toIntOrNull(
					r.totalCompetencies ?? r.total_competencies,
				),
			} as EmployeeListItemExtended;
		});
		hasMore = body.meta?.hasMore ?? false;
		hasPrev = currentPage > 0;
		apiTotal = body.meta?.total ?? 0;
		const rawCompleted =
			body.meta?.completedCount ?? body.meta?.completed_count;
		completedCount =
			typeof rawCompleted === 'number' && Number.isInteger(rawCompleted)
				? rawCompleted
				: null;
	} catch (e) {
		error =
			e instanceof Error ? e.message : 'Error desconocido al cargar empleados';
		items = [];
		completedCount = null;
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
