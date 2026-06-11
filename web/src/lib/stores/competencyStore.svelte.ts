import type {
	Pillar,
	Competency,
	ScaleCriterion,
	LevelDefinition,
	CompetencyAcceptanceLevel,
	AcceptanceLevel
} from '$lib/types/competency';
import type { EvaluationProfile } from '$lib/types/evaluation';

import pillarsData from '$lib/fixtures/competency/pillars.json';
import competenciesData from '$lib/fixtures/competency/competencies.json';
import scaleCriteriaData from '$lib/fixtures/competency/scale-criteria.json';
import levelDefinitionsData from '$lib/fixtures/competency/acceptance-levels.json';
import competencyAcceptanceLevelsData from '$lib/fixtures/competency/competency-acceptance-levels.json';
import { client } from '$lib/api/client';

// ─── Internal data shape ──────────────────────────────────────────────────────

interface StoreData {
	pillars: Pillar[];
	competencies: Competency[];
	scaleCriteria: ScaleCriterion[];
	levelDefinitions: LevelDefinition[];
	competencyAcceptanceLevels: CompetencyAcceptanceLevel[];
}

// ─── Triplete state ───────────────────────────────────────────────────────────

let data = $state<StoreData | null>(null);
let loading = $state(true);
let error = $state<string | null>(null);

// ─── Helpers ──────────────────────────────────────────────────────────────────

function loadFixtures(): StoreData {
	return {
		pillars: structuredClone(pillarsData as Pillar[]),
		competencies: structuredClone(competenciesData as Competency[]),
		scaleCriteria: structuredClone(scaleCriteriaData as ScaleCriterion[]),
		levelDefinitions: structuredClone(levelDefinitionsData as LevelDefinition[]),
		competencyAcceptanceLevels: structuredClone(
			competencyAcceptanceLevelsData as CompetencyAcceptanceLevel[]
		)
	};
}

/**
 * Load competency data.
 *
 * In DEV without VITE_USE_API: loads from fixture files (structured clone).
 * In production / VITE_USE_API=true: fetches from the real API endpoints.
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
		// Phase 1 — parallel: pillars (with competencies), levels, acceptance levels
		const [pillarsRes, levelsRes, acceptanceRes] = await Promise.all([
			client.GET('/pillars', { params: { query: { include: 'competencies' } } }),
			client.GET('/levels', {}),
			client.GET('/acceptance-levels', {})
		]);

		const apiPillars = (
			(pillarsRes.data as { data?: Array<{ id?: string; name?: string; description?: string }> })
				?.data ?? []
		) as Array<{ id?: string; name?: string; description?: string }>;

		const apiLevels = (levelsRes.data ?? []) as Array<{
			level?: number;
			label?: string;
			description?: string;
		}>;

		const apiAcceptanceLevels = (acceptanceRes.data ?? []) as Array<{
			competency_id?: string;
			profile_id?: string;
			level?: number;
		}>;

		// Phase 2 — fan-out per pillar to get competencies with details
		const pillarIds = apiPillars.map((p) => p.id).filter(Boolean) as string[];
		const detailResults = await Promise.all(
			pillarIds.map((id) =>
				client.GET('/pillars/{id}', {
					params: { path: { id }, query: { include: 'competencies' } }
				})
			)
		);

		// ── Normalize ──────────────────────────────────────────────────────

		const pillars: Pillar[] = apiPillars.map((p) => ({
			id: p.id ?? crypto.randomUUID(),
			name: p.name ?? '',
			description: p.description ?? ''
		}));

		const competencies: Competency[] = [];
		for (const res of detailResults) {
			const detail = res.data as {
				id?: string;
				competencies?: Array<{ id?: string; name?: string; description?: string }>;
			} | null;
			const pillarId = detail?.id ?? '';
			for (const c of detail?.competencies ?? []) {
				competencies.push({
					id: c.id ?? crypto.randomUUID(),
					name: c.name ?? '',
					description: c.description ?? '',
					pillarId
				});
			}
		}

		const levelDefinitions: LevelDefinition[] = apiLevels.map((l) => ({
			level: (l.level ?? 1) as 1 | 2 | 3 | 4 | 5,
			label: l.label ?? '',
			description: l.description ?? ''
		}));

		const competencyAcceptanceLevels: CompetencyAcceptanceLevel[] = apiAcceptanceLevels.map(
			(al) => ({
				competencyId: al.competency_id ?? '',
				profileId: al.profile_id as EvaluationProfile,
				level: (al.level ?? 3) as 1 | 2 | 3 | 4 | 5
			})
		);

		data = {
			pillars,
			competencies,
			scaleCriteria: [],
			levelDefinitions,
			competencyAcceptanceLevels
		};
	} catch (e) {
		error =
			e instanceof Error
				? e.message
				: 'Error desconocido al cargar datos de competencias';
	} finally {
		loading = false;
	}
}

/** Alias for load(). */
export function reload(): Promise<void> {
	return load();
}

// ─── Getters: Pillars ─────────────────────────────────────────────────────────

export function getPillars(): Pillar[] {
	return data?.pillars ?? [];
}

// ─── Getters: Competencies ────────────────────────────────────────────────────

export function getCompetencies(): Competency[] {
	return data?.competencies ?? [];
}

export function getCompetenciesByPillar(pillarId: string): Competency[] {
	return (data?.competencies ?? []).filter((c) => c.pillarId === pillarId);
}

// ─── Getters: Scale Criteria ──────────────────────────────────────────────────

export function getScaleCriteria(): ScaleCriterion[] {
	return data?.scaleCriteria ?? [];
}

export function getScaleCriteriaForCell(
	competencyId: string,
	pillarId: string
): ScaleCriterion[] {
	return (data?.scaleCriteria ?? []).filter(
		(sc) => sc.competencyId === competencyId && sc.pillarId === pillarId
	);
}

// ─── Getters: Acceptance Levels (deprecated) ──────────────────────────────────

/**
 * @deprecated Use getLevelDefinitions() + getCompetencyAcceptanceLevels() instead.
 */
export function getAcceptanceLevels(): AcceptanceLevel[] {
	return [];
}

/**
 * @deprecated Use getCompetencyAcceptanceLevelsByProfile() instead.
 */
export function getAcceptanceLevelsByProfile(
	_profileId: EvaluationProfile
): AcceptanceLevel[] {
	return [];
}

// ─── Getters: Level Definitions ───────────────────────────────────────────────

export function getLevelDefinitions(): LevelDefinition[] {
	return data?.levelDefinitions ?? [];
}

export function getLevelDefinition(
	level: 1 | 2 | 3 | 4 | 5
): LevelDefinition | undefined {
	return (data?.levelDefinitions ?? []).find((ld) => ld.level === level);
}

// ─── Getters: Competency Acceptance Levels ────────────────────────────────────

export function getCompetencyAcceptanceLevels(): CompetencyAcceptanceLevel[] {
	return data?.competencyAcceptanceLevels ?? [];
}

export function getCompetencyAcceptanceLevelsByProfile(
	profileId: EvaluationProfile
): CompetencyAcceptanceLevel[] {
	return (data?.competencyAcceptanceLevels ?? []).filter(
		(cal) => cal.profileId === profileId
	);
}

export function getCompetencyAcceptanceLevel(
	competencyId: string,
	profileId: EvaluationProfile
): CompetencyAcceptanceLevel | undefined {
	return (data?.competencyAcceptanceLevels ?? []).find(
		(cal) => cal.competencyId === competencyId && cal.profileId === profileId
	);
}

// ─── Mutations: Pillars ──────────────────────────────────────────────────────

export async function addPillar(pillar: Pillar): Promise<void> {
	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		data = { ...data!, pillars: [...(data?.pillars ?? []), pillar] };
		return;
	}
	const { error: apiError } = await client.POST('/pillars', {
		body: { name: pillar.name, description: pillar.description },
		params: { header: { 'Idempotency-Key': crypto.randomUUID() } }
	});
	if (apiError)
		throw new Error(
			((apiError as { error?: { message?: string } })?.error?.message) ??
				'Error al crear pilar'
		);
	await reload();
}

export async function updatePillar(
	id: string,
	updates: Partial<Omit<Pillar, 'id'>>
): Promise<void> {
	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		data = {
			...data!,
			pillars: (data?.pillars ?? []).map((p) =>
				p.id === id ? { ...p, ...updates } : p
			)
		};
		return;
	}
	const { error: apiError } = await client.PUT('/pillars/{id}', {
		params: { path: { id }, header: { 'If-Match': 'placeholder' } },
		body: { name: updates.name ?? '', description: updates.description ?? '' }
	});
	if (apiError)
		throw new Error(
			((apiError as { error?: { message?: string } })?.error?.message) ??
				'Error al actualizar pilar'
		);
	await reload();
}

export async function deletePillar(id: string): Promise<void> {
	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		data = {
			...data!,
			pillars: (data?.pillars ?? []).filter((p) => p.id !== id),
			competencies: (data?.competencies ?? []).filter((c) => c.pillarId !== id),
			scaleCriteria: (data?.scaleCriteria ?? []).filter(
				(sc) => sc.pillarId !== id
			)
		};
		return;
	}
	const { error: apiError } = await client.DELETE('/pillars/{id}', {
		params: { path: { id } }
	});
	if (apiError)
		throw new Error(
			((apiError as { error?: { message?: string } })?.error?.message) ??
				'Error al eliminar pilar'
		);
	await reload();
}

// ─── Mutations: Competencies ─────────────────────────────────────────────────

export async function addCompetency(competency: Competency): Promise<void> {
	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		data = {
			...data!,
			competencies: [...(data?.competencies ?? []), competency]
		};
		return;
	}
	const { error: apiError } = await client.POST(
		'/pillars/{pillarId}/competencies',
		{
			params: {
				path: { pillarId: competency.pillarId },
				header: { 'Idempotency-Key': crypto.randomUUID() }
			},
			body: { name: competency.name, description: competency.description }
		}
	);
	if (apiError)
		throw new Error(
			((apiError as { error?: { message?: string } })?.error?.message) ??
				'Error al crear competencia'
		);
	await reload();
}

export async function updateCompetency(
	id: string,
	updates: Partial<Omit<Competency, 'id'>>
): Promise<void> {
	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		data = {
			...data!,
			competencies: (data?.competencies ?? []).map((c) =>
				c.id === id ? { ...c, ...updates } : c
			)
		};
		return;
	}
	const { error: apiError } = await client.PUT('/competencies/{id}', {
		params: { path: { id }, header: { 'If-Match': 'placeholder' } },
		body: {
			name: updates.name ?? '',
			description: updates.description ?? '',
			pillar_id: updates.pillarId
		}
	});
	if (apiError)
		throw new Error(
			((apiError as { error?: { message?: string } })?.error?.message) ??
				'Error al actualizar competencia'
		);
	await reload();
}

export async function deleteCompetency(id: string): Promise<void> {
	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		data = {
			...data!,
			competencies: (data?.competencies ?? []).filter((c) => c.id !== id),
			scaleCriteria: (data?.scaleCriteria ?? []).filter(
				(sc) => sc.competencyId !== id
			)
		};
		return;
	}
	const { error: apiError } = await client.DELETE('/competencies/{id}', {
		params: { path: { id } }
	});
	if (apiError)
		throw new Error(
			((apiError as { error?: { message?: string } })?.error?.message) ??
				'Error al eliminar competencia'
		);
	await reload();
}

// ─── Mutations: Scale Criteria ───────────────────────────────────────────────

export async function updateScaleCriterion(
	id: string,
	description: string
): Promise<void> {
	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		data = {
			...data!,
			scaleCriteria: (data?.scaleCriteria ?? []).map((sc) =>
				sc.id === id ? { ...sc, description } : sc
			)
		};
		return;
	}
	// Local-only for now (API uses bulk replace per competency).
	data = {
		...data!,
		scaleCriteria: (data?.scaleCriteria ?? []).map((sc) =>
			sc.id === id ? { ...sc, description } : sc
		)
	};
}

export async function addScaleCriterion(
	criterion: Omit<ScaleCriterion, 'id'>
): Promise<void> {
	const id = `sc-${criterion.competencyId}-${criterion.pillarId}-${criterion.level}-${crypto.randomUUID().slice(0, 8)}`;
	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		data = {
			...data!,
			scaleCriteria: [...(data?.scaleCriteria ?? []), { id, ...criterion }]
		};
		return;
	}
	// Local-only for now (API uses bulk replace per competency).
	data = {
		...data!,
		scaleCriteria: [...(data?.scaleCriteria ?? []), { id, ...criterion }]
	};
}

export async function removeScaleCriterion(id: string): Promise<void> {
	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		data = {
			...data!,
			scaleCriteria: (data?.scaleCriteria ?? []).filter((sc) => sc.id !== id)
		};
		return;
	}
	// Local-only for now (API uses bulk replace per competency).
	data = {
		...data!,
		scaleCriteria: (data?.scaleCriteria ?? []).filter((sc) => sc.id !== id)
	};
}

// ─── Mutations: Acceptance Levels (deprecated) ───────────────────────────────

/**
 * @deprecated Use updateLevelDefinition() + setCompetencyAcceptanceLevel() instead.
 */
export async function updateAcceptanceLevel(
	_profileId: EvaluationProfile,
	_level: 1 | 2 | 3 | 4 | 5,
	_updates: Partial<Omit<AcceptanceLevel, 'profileId' | 'level'>>
): Promise<void> {
	// Deprecated — no-op
}

// ─── Mutations: Level Definitions ────────────────────────────────────────────

export async function updateLevelDefinition(
	level: 1 | 2 | 3 | 4 | 5,
	label: string,
	description: string
): Promise<void> {
	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		data = {
			...data!,
			levelDefinitions: (data?.levelDefinitions ?? []).map((ld) =>
				ld.level === level ? { ...ld, label, description } : ld
			)
		};
		return;
	}
	// Local-only for now (levels are cacheable / read-only in production).
	data = {
		...data!,
		levelDefinitions: (data?.levelDefinitions ?? []).map((ld) =>
			ld.level === level ? { ...ld, label, description } : ld
		)
	};
}

// ─── Mutations: Competency Acceptance Levels ─────────────────────────────────

export async function setCompetencyAcceptanceLevel(
	competencyId: string,
	profileId: EvaluationProfile,
	level: 1 | 2 | 3 | 4 | 5
): Promise<void> {
	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		const existing = (data?.competencyAcceptanceLevels ?? []).find(
			(cal) => cal.competencyId === competencyId && cal.profileId === profileId
		);
		if (existing) {
			data = {
				...data!,
				competencyAcceptanceLevels: (
					data?.competencyAcceptanceLevels ?? []
				).map((cal) =>
					cal.competencyId === competencyId && cal.profileId === profileId
						? { ...cal, level }
						: cal
				)
			};
		} else {
			data = {
				...data!,
				competencyAcceptanceLevels: [
					...(data?.competencyAcceptanceLevels ?? []),
					{ competencyId, profileId, level }
				]
			};
		}
		return;
	}
	const { error: apiError } = await client.POST('/acceptance-levels', {
		body: { competency_id: competencyId, profile_id: profileId, level }
	});
	if (apiError)
		throw new Error(
			((apiError as { error?: { message?: string } })?.error?.message) ??
				'Error al establecer nivel de aceptación'
		);
	await reload();
}

export async function setCompetencyAcceptanceLevelsForProfile(
	profileId: EvaluationProfile,
	assignments: { competencyId: string; level: 1 | 2 | 3 | 4 | 5 }[]
): Promise<void> {
	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		const filtered = (data?.competencyAcceptanceLevels ?? []).filter(
			(cal) => cal.profileId !== profileId
		);
		const updated = assignments.map((a) => ({
			competencyId: a.competencyId,
			profileId,
			level: a.level
		}));
		data = {
			...data!,
			competencyAcceptanceLevels: [...filtered, ...updated]
		};
		return;
	}
	// Upsert each assignment individually, then reload
	try {
		await Promise.all(
			assignments.map((a) =>
				client.POST('/acceptance-levels', {
					body: {
						competency_id: a.competencyId,
						profile_id: profileId,
						level: a.level
					}
				})
			)
		);
	} catch {
		throw new Error('Error al establecer niveles de aceptación');
	}
	await reload();
}

// ─── Min Acceptable Level per Profile ────────────────────────────────────────

const DEFAULT_MIN_LEVELS: Record<EvaluationProfile, number> = {
	colaborador: 3,
	jefe: 3,
	vendedor: 3,
	'gerente-tienda': 4,
	divisional: 4,
	regional: 4,
	director: 4,
	'director-general': 4,
	rh: 3
};

let profileMinLevels = $state<Record<EvaluationProfile, number>>({
	...DEFAULT_MIN_LEVELS
});

export function getProfileMinLevels(): Record<EvaluationProfile, number> {
	return profileMinLevels;
}

export function getProfileMinLevel(profileId: EvaluationProfile): number {
	return profileMinLevels[profileId];
}

export function setProfileMinLevel(
	profileId: EvaluationProfile,
	level: number
): void {
	profileMinLevels = { ...profileMinLevels, [profileId]: level };
}
