import type {
    Pillar,
    Competency,
    ScaleCriterion,
    LevelDefinition,
    CompetencyAcceptanceLevel,
    AcceptanceLevel,
    Profile,
} from "$lib/types/competency";
import type { EvaluationProfile } from "$lib/types/evaluation";

import { client } from "$lib/api/client";
import { SvelteMap } from "svelte/reactivity";

// ─── Internal data shape ──────────────────────────────────────────────────────

interface StoreData {
    profiles: Profile[];
    pillars: Pillar[];
    competencies: Competency[];
    scaleCriteria: ScaleCriterion[];
    levelDefinitions: LevelDefinition[];
    competencyAcceptanceLevels: CompetencyAcceptanceLevel[];
}

// ─── Profile lookup maps (DB UUID ↔ frontend slug) ──────────────────────────

const uuidToName = new SvelteMap<string, EvaluationProfile>();
const nameToUuid = new SvelteMap<EvaluationProfile, string>();

function resolveProfileName(uuid: string): EvaluationProfile {
    return uuidToName.get(uuid) ?? ("colaborador" as EvaluationProfile);
}

function resolveProfileUuid(name: EvaluationProfile): string {
    return nameToUuid.get(name) ?? "";
}

// ─── Triplete state ───────────────────────────────────────────────────────────

let data = $state<StoreData | null>(null);
let loading = $state(true);
let error = $state<string | null>(null);
let currentType: 'competencias' | 'metas' | undefined = undefined;

/** @returns true while load() is in progress. */
export function isLoading(): boolean {
    return loading;
}

/** @returns the current error message, or null if no error. */
export function getError(): string | null {
    return error;
}

/**
 * Load competency data.
 *
 */
export async function load(type?: 'competencias' | 'metas'): Promise<void> {
   return _doLoad(type);
}

async function _doLoad(type?: 'competencias' | 'metas'): Promise<void> {
    const effectiveType = type ?? 'competencias';
    if (type !== undefined) {
        currentType = type;
    }
    loading = true;
    error = null;

    try {
        const pillarQuery: Record<string, string> = {};
        if (effectiveType) {
            pillarQuery.type = effectiveType;
        }
        // For competencias, include competencies; for metas, skip
        if (effectiveType === 'competencias') {
            pillarQuery.include = 'competencies';
        }

        const [pillarsRes, levelsRes, profilesRes, acceptanceRes] =
            await Promise.all([
                client.GET("/pillars", {
                    params: { query: pillarQuery },
                }),
                client.GET("/levels", {}),
                client.GET("/profiles", {}),
                client.GET("/acceptance-levels", {}),
            ]);

        const apiPillars =
            (
                pillarsRes.data as {
                    data?: Array<{
                        id?: string;
                        name?: string;
                        description?: string;
                        type?: string;
                        updated_at?: string;
                        competencies?: Array<{
                            id?: string;
                            name?: string;
                            description?: string;
                            updated_at?: string;
                        }>;
                    }>;
                }
            )?.data ?? [];

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

        // ── Normalize ──────────────────────────────────────────────────────

        const pillars: Pillar[] = [];
        const competencies: Competency[] = [];

        for (const p of apiPillars) {
            const pillarId = p.id ?? crypto.randomUUID();
            pillars.push({
                id: pillarId,
                name: p.name ?? "",
                description: p.description ?? "",
                type: (p.type === 'metas' ? 'metas' : 'competencias') as 'competencias' | 'metas',
                updatedAt: p.updated_at ?? new Date().toISOString(),
            });
            for (const c of p.competencies ?? []) {
                competencies.push({
                    id: c.id ?? crypto.randomUUID(),
                    name: c.name ?? "",
                    description: c.description ?? "",
                    pillarId,
                    updatedAt: c.updated_at ?? new Date().toISOString(),
                });
            }
        }

        const levelDefinitions: LevelDefinition[] = apiLevels.map((l) => ({
            level: (l.level ?? 1) as 1 | 2 | 3 | 4 | 5,
            label: l.label ?? "",
            description: l.description ?? "",
        }));

        // Normalize profiles and build UUID ↔ name lookup maps
        const apiProfiles = (profilesRes.data ?? []) as Array<{
            id?: string;
            name?: string;
            description?: string;
        }>;

        const profiles: Profile[] = apiProfiles.map((p) => ({
            id: p.id ?? "",
            name: (p.name ?? "colaborador") as EvaluationProfile,
            description: p.description,
        }));

        uuidToName.clear();
        nameToUuid.clear();
        for (const p of profiles) {
            if (p.id && p.name) {
                uuidToName.set(p.id, p.name);
                nameToUuid.set(p.name, p.id);
            }
        }

        const competencyAcceptanceLevels: CompetencyAcceptanceLevel[] =
            apiAcceptanceLevels.map((al) => ({
                competencyId: al.competency_id ?? "",
                profileId: resolveProfileName(al.profile_id ?? ""),
                level: (al.level ?? 3) as 1 | 2 | 3 | 4 | 5,
            }));

        // Phase 2 — fetch scale criteria per competency
        const competencyIds = competencies.map((c) => c.id);
        const scaleCriteriaResults = await Promise.all(
            competencyIds.map((id) =>
                client.GET("/competencies/{id}/scale-criteria", {
                    params: { path: { id } },
                }),
            ),
        );

        const scaleCriteria: ScaleCriterion[] = [];
        for (const res of scaleCriteriaResults) {
            const scRes = res.data as {
                competency_id?: string;
                criteria?: { [key: string]: string[] };
            } | null;
            const competencyId = scRes?.competency_id ?? "";
            const competency = competencies.find((c) => c.id === competencyId);
            const pillarId = competency?.pillarId ?? "";

            if (scRes?.criteria) {
                for (const [levelStr, descriptions] of Object.entries(
                    scRes.criteria,
                )) {
                    const level = parseInt(levelStr, 10) as 1 | 2 | 3 | 4 | 5;
                    if (level < 1 || level > 5) continue;
                    for (const desc of descriptions) {
                        scaleCriteria.push({
                            id: `sc-${competencyId}-${level}-${crypto.randomUUID().slice(0, 8)}`,
                            competencyId,
                            pillarId,
                            level,
                            description: desc,
                        });
                    }
                }
            }
        }

        data = {
            profiles,
            pillars,
            competencies,
            scaleCriteria,
            levelDefinitions,
            competencyAcceptanceLevels,
        };
    } catch (e) {
        error =
            e instanceof Error
                ? e.message
                : "Error desconocido al cargar datos de competencias";
    } finally {
        loading = false;
    }
}

/** Reload with the last-used type filter. */
export function reload(): Promise<void> {
    return load(currentType);
}

// ─── Getters: Profiles ────────────────────────────────────────────────────────

export function getProfiles(): Profile[] {
    return data?.profiles ?? [];
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
    pillarId: string,
): ScaleCriterion[] {
    return (data?.scaleCriteria ?? []).filter(
        (sc) => sc.competencyId === competencyId && sc.pillarId === pillarId,
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
    _profileId: EvaluationProfile,
): AcceptanceLevel[] {
    return [];
}

// ─── Getters: Level Definitions ───────────────────────────────────────────────

export function getLevelDefinitions(): LevelDefinition[] {
    return data?.levelDefinitions ?? [];
}

export function getLevelDefinition(
    level: 1 | 2 | 3 | 4 | 5,
): LevelDefinition | undefined {
    return (data?.levelDefinitions ?? []).find((ld) => ld.level === level);
}

// ─── Getters: Competency Acceptance Levels ────────────────────────────────────

export function getCompetencyAcceptanceLevels(): CompetencyAcceptanceLevel[] {
    return data?.competencyAcceptanceLevels ?? [];
}

export function getCompetencyAcceptanceLevelsByProfile(
    profileId: EvaluationProfile,
): CompetencyAcceptanceLevel[] {
    return (data?.competencyAcceptanceLevels ?? []).filter(
        (cal) => cal.profileId === profileId,
    );
}

export function getCompetencyAcceptanceLevel(
    competencyId: string,
    profileId: EvaluationProfile,
): CompetencyAcceptanceLevel | undefined {
    return (data?.competencyAcceptanceLevels ?? []).find(
        (cal) =>
            cal.competencyId === competencyId && cal.profileId === profileId,
    );
}

// ─── Mutations: Pillars ──────────────────────────────────────────────────────

export async function addPillar(pillar: Pillar): Promise<void> {
    const { error: apiError } = await client.POST("/pillars", {
        body: { name: pillar.name, description: pillar.description, type: pillar.type },
        params: { header: { "Idempotency-Key": crypto.randomUUID() } },
    });
    if (apiError)
        throw new Error(
            (apiError as { error?: { message?: string } })?.error?.message ??
                "Error al crear pilar",
        );
    await reload();
}

export async function updatePillar(
    id: string,
    updates: Partial<Omit<Pillar, "id">>,
): Promise<void> {
    const pillar = data?.pillars.find((p) => p.id === id);
    const { error: apiError } = await client.PUT("/pillars/{id}", {
        params: {
            path: { id },
            header: { "If-Match": pillar?.updatedAt ?? "placeholder" },
        },
        body: {
            name: updates.name ?? "",
            description: updates.description ?? "",
            type: updates.type,
        },
    });
    if (apiError)
        throw new Error(
            (apiError as { error?: { message?: string } })?.error?.message ??
                "Error al actualizar pilar",
        );
    await reload();
}

export async function deletePillar(id: string): Promise<void> {
    const { error: apiError } = await client.DELETE("/pillars/{id}", {
        params: { path: { id }, query: { force: true } },
    });
    if (apiError)
        throw new Error(
            (apiError as { error?: { message?: string } })?.error?.message ??
                "Error al eliminar pilar",
        );
    await reload();
}

// ─── Mutations: Competencies ─────────────────────────────────────────────────

export async function addCompetency(competency: Competency): Promise<void> {
    const { error: apiError } = await client.POST(
        "/pillars/{pillarId}/competencies",
        {
            params: {
                path: { pillarId: competency.pillarId },
                header: { "Idempotency-Key": crypto.randomUUID() },
            },
            body: {
                name: competency.name,
                description: competency.description,
            },
        },
    );
    if (apiError)
        throw new Error(
            (apiError as { error?: { message?: string } })?.error?.message ??
                "Error al crear competencia",
        );
    await reload();
}

export async function updateCompetency(
    id: string,
    updates: Partial<Omit<Competency, "id">>,
): Promise<void> {
    const competency = data?.competencies.find((c) => c.id === id);
    const { error: apiError } = await client.PUT("/competencies/{id}", {
        params: {
            path: { id },
            header: { "If-Match": competency?.updatedAt ?? "placeholder" },
        },
        body: {
            name: updates.name ?? "",
            description: updates.description ?? "",
            pillar_id: updates.pillarId,
        },
    });
    if (apiError)
        throw new Error(
            (apiError as { error?: { message?: string } })?.error?.message ??
                "Error al actualizar competencia",
        );
    await reload();
}

export async function deleteCompetency(id: string): Promise<void> {
    const { error: apiError } = await client.DELETE("/competencies/{id}", {
        params: { path: { id }, query: { force: true } },
    });
    if (apiError)
        throw new Error(
            (apiError as { error?: { message?: string } })?.error?.message ??
                "Error al eliminar competencia",
        );
    await reload();
}

// ─── Mutations: Scale Criteria ───────────────────────────────────────────────

export async function replaceScaleCriteria(
    competencyId: string,
    criteria: { level: number; description: string }[],
): Promise<void> {
    const { data: resData, error: apiError } = await client.POST(
        "/competencies/{id}/scale-criteria",
        {
            params: {
                path: { id: competencyId },
                header: { "Idempotency-Key": crypto.randomUUID() },
            },
            body: { criteria },
        },
    );
    if (apiError)
        throw new Error(
            (apiError as { error?: { message?: string } })?.error?.message ??
                "Error al guardar criterios de escala",
        );

    const res = resData as {
        pillar_id: string;
        competency_id?: string;
        criteria?: { [key: string]: string[] };
    } | null;

    const pillarId =
        res?.pillar_id ??
        data?.scaleCriteria.find((sc) => sc.competencyId === competencyId)
            ?.pillarId ??
        "";

    const newCriteria: ScaleCriterion[] = [];
    if (res?.criteria) {
        for (const [levelStr, descriptions] of Object.entries(res.criteria)) {
            const level = parseInt(levelStr, 10) as 1 | 2 | 3 | 4 | 5;
            if (level < 1 || level > 5) continue;
            for (const desc of descriptions) {
                newCriteria.push({
                    id: `sc-${competencyId}-${level}-${crypto.randomUUID().slice(0, 8)}`,
                    competencyId,
                    pillarId,
                    level,
                    description: desc,
                });
            }
        }
    }

    data = {
        ...data!,
        scaleCriteria: [
            ...(data?.scaleCriteria ?? []).filter(
                (sc) => sc.competencyId !== competencyId,
            ),
            ...newCriteria,
        ],
    };
}

// ─── Mutations: Acceptance Levels (deprecated) ───────────────────────────────

/**
 * @deprecated Use updateLevelDefinition() + setCompetencyAcceptanceLevel() instead.
 */
export async function updateAcceptanceLevel(
    _profileId: EvaluationProfile,
    _level: 1 | 2 | 3 | 4 | 5,
    _updates: Partial<Omit<AcceptanceLevel, "profileId" | "level">>,
): Promise<void> {
    // Deprecated — no-op
}

// ─── Mutations: Level Definitions ────────────────────────────────────────────

export async function updateLevelDefinition(
    level: 1 | 2 | 3 | 4 | 5,
    label: string,
    description: string,
): Promise<void> {
    const { data: putData, error: apiError } = await client.PUT(
        "/levels/{level}",
        {
            params: { path: { level } },
            body: { label, description },
        },
    );
    if (apiError) {
        throw new Error(
            (apiError as { error?: { message?: string } })?.error?.message ??
                "Error al actualizar definición de nivel",
        );
    }

    // Update local state directly from the PUT response — do NOT re-fetch GET /levels
    // because that endpoint routes to a read replica with replication lag, which
    // returns stale data (or empty) right after a PUT to the primary.
    const newDef: LevelDefinition = {
        level,
        label: (putData as { label?: string } | null)?.label ?? label,
        description:
            (putData as { description?: string } | null)?.description ??
            description,
    };
    data = {
        ...data!,
        levelDefinitions: (data?.levelDefinitions ?? []).map((ld) =>
            ld.level === level ? newDef : ld,
        ),
    };
}

// ─── Mutations: Competency Acceptance Levels ─────────────────────────────────

export async function setCompetencyAcceptanceLevel(
    competencyId: string,
    profileId: EvaluationProfile,
    level: 1 | 2 | 3 | 4 | 5,
): Promise<void> {
    const profileUuid = resolveProfileUuid(profileId);
    if (!profileUuid)
        throw new Error(`Profile "${profileId}" not found in database`);
    const { error: apiError } = await client.POST("/acceptance-levels", {
        body: { competency_id: competencyId, profile_id: profileUuid, level },
    });
    if (apiError)
        throw new Error(
            (apiError as { error?: { message?: string } })?.error?.message ??
                "Error al establecer nivel de aceptación",
        );
    await reload();
}

/**
 * Optimistic update: mutates the acceptance-levels array in place (no data
 * reference replacement) so only getters reading competencyAcceptanceLevels
 * re-evaluate. Returns revert/commit handles.
 */
export function setCompetencyAcceptanceLevelOptimistic(
    competencyId: string,
    profileId: EvaluationProfile,
    level: 1 | 2 | 3 | 4 | 5,
): { revert: () => void; commit: () => Promise<void> } {
    const arr = data?.competencyAcceptanceLevels;
    if (!arr) return { revert: () => {}, commit: async () => {} };

    const idx = arr.findIndex(
        (cal) =>
            cal.competencyId === competencyId && cal.profileId === profileId,
    );
    const prev = idx >= 0 ? arr[idx].level : 3;

    // In-place mutation — no new data reference
    if (idx >= 0) {
        arr[idx].level = level;
    } else {
        arr.push({ competencyId, profileId, level });
    }

    return {
        revert: () => {
            if (idx >= 0) {
                arr[idx].level = prev;
            } else {
                const i = arr.findIndex(
                    (cal) =>
                        cal.competencyId === competencyId &&
                        cal.profileId === profileId,
                );
                if (i >= 0) arr.splice(i, 1);
            }
        },
        commit: async () => {
            const profileUuid = resolveProfileUuid(profileId);
            if (!profileUuid)
                throw new Error(`Profile "${profileId}" not found in database`);
            const { error: apiError } = await client.POST(
                "/acceptance-levels",
                {
                    body: {
                        competency_id: competencyId,
                        profile_id: profileUuid,
                        level,
                    },
                },
            );
            if (apiError)
                throw new Error(
                    (apiError as { error?: { message?: string } })?.error
                        ?.message ?? "Error al establecer nivel de aceptación",
                );
        },
    };
}

export async function setCompetencyAcceptanceLevelsForProfile(
    profileId: EvaluationProfile,
    assignments: { competencyId: string; level: 1 | 2 | 3 | 4 | 5 }[],
): Promise<void> {
    const profileUuid = resolveProfileUuid(profileId);
    if (!profileUuid)
        throw new Error(`Profile "${profileId}" not found in database`);
    // Upsert each assignment individually, then reload
    try {
        await Promise.all(
            assignments.map((a) =>
                client.POST("/acceptance-levels", {
                    body: {
                        competency_id: a.competencyId,
                        profile_id: profileUuid,
                        level: a.level,
                    },
                }),
            ),
        );
    } catch {
        throw new Error("Error al establecer niveles de aceptación");
    }
    await reload();
}

// ─── Min Acceptable Level per Profile ────────────────────────────────────────

const DEFAULT_MIN_LEVELS: Record<EvaluationProfile, number> = {
    colaborador: 3,
    jefe: 3,
    vendedor: 3,
    "gerente-tienda": 4,
    divisional: 4,
    regional: 4,
    director: 4,
    "director-general": 4,
    rh: 3,
};

let profileMinLevels = $state<Record<EvaluationProfile, number>>({
    ...DEFAULT_MIN_LEVELS,
});

export function getProfileMinLevels(): Record<EvaluationProfile, number> {
    return profileMinLevels;
}

export function getProfileMinLevel(profileId: EvaluationProfile): number {
    return profileMinLevels[profileId];
}

export function setProfileMinLevel(
    profileId: EvaluationProfile,
    level: number,
): void {
    profileMinLevels = { ...profileMinLevels, [profileId]: level };
}
