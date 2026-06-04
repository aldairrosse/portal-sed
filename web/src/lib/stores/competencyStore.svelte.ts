import type {
	Pillar,
	Competency,
	ScaleCriterion,
	AcceptanceLevel,
	LevelDefinition,
	CompetencyAcceptanceLevel
} from '$lib/types/competency';
import type { EvaluationProfile } from '$lib/types/evaluation';

import pillarsData from '$lib/fixtures/competency/pillars.json';
import competenciesData from '$lib/fixtures/competency/competencies.json';
import scaleCriteriaData from '$lib/fixtures/competency/scale-criteria.json';
import levelDefinitionsData from '$lib/fixtures/competency/acceptance-levels.json';
import competencyAcceptanceLevelsData from '$lib/fixtures/competency/competency-acceptance-levels.json';

// ─── State ───────────────────────────────────────────────────────────────────
let pillars = $state<Pillar[]>(structuredClone(pillarsData));
let competencies = $state<Competency[]>(structuredClone(competenciesData));
let scaleCriteria = $state<ScaleCriterion[]>(
	structuredClone(scaleCriteriaData as ScaleCriterion[])
);

/**
 * @deprecated Use levelDefinitions + competencyAcceptanceLevels instead.
 * AcceptanceLevel was per-profile per-level; now level definitions are global
 * and competency acceptance levels are per-competency per-profile.
 * Kept as empty array for backward compatibility of the type and API surface.
 */
let acceptanceLevels = $state<AcceptanceLevel[]>([]);

/** Global level definitions — same labels across all profiles. */
let levelDefinitions = $state<LevelDefinition[]>(
	structuredClone(levelDefinitionsData as LevelDefinition[])
);

/** Per-competency per-profile acceptance level assignments. */
let competencyAcceptanceLevels = $state<CompetencyAcceptanceLevel[]>(
	structuredClone(competencyAcceptanceLevelsData as CompetencyAcceptanceLevel[])
);

// ─── Getters ─────────────────────────────────────────────────────────────────

export function getPillars(): Pillar[] {
	return pillars;
}

export function getCompetencies(): Competency[] {
	return competencies;
}

export function getCompetenciesByPillar(pillarId: string): Competency[] {
	return competencies.filter((c) => c.pillarId === pillarId);
}

export function getScaleCriteria(): ScaleCriterion[] {
	return scaleCriteria;
}

export function getScaleCriteriaForCell(
	competencyId: string,
	pillarId: string
): ScaleCriterion[] {
	return scaleCriteria.filter(
		(sc) => sc.competencyId === competencyId && sc.pillarId === pillarId
	);
}

/**
 * @deprecated Use getLevelDefinitions() + getCompetencyAcceptanceLevels() instead.
 */
export function getAcceptanceLevels(): AcceptanceLevel[] {
	return acceptanceLevels;
}

/**
 * @deprecated Use getCompetencyAcceptanceLevelsByProfile() instead.
 */
export function getAcceptanceLevelsByProfile(
	profileId: EvaluationProfile
): AcceptanceLevel[] {
	return acceptanceLevels.filter((al) => al.profileId === profileId);
}

// ─── Getters: Level Definitions ──────────────────────────────────────────────

export function getLevelDefinitions(): LevelDefinition[] {
	return levelDefinitions;
}

export function getLevelDefinition(level: 1 | 2 | 3 | 4 | 5): LevelDefinition | undefined {
	return levelDefinitions.find((ld) => ld.level === level);
}

// ─── Getters: Competency Acceptance Levels ────────────────────────────────────

export function getCompetencyAcceptanceLevels(): CompetencyAcceptanceLevel[] {
	return competencyAcceptanceLevels;
}

export function getCompetencyAcceptanceLevelsByProfile(
	profileId: EvaluationProfile
): CompetencyAcceptanceLevel[] {
	return competencyAcceptanceLevels.filter((cal) => cal.profileId === profileId);
}

export function getCompetencyAcceptanceLevel(
	competencyId: string,
	profileId: EvaluationProfile
): CompetencyAcceptanceLevel | undefined {
	return competencyAcceptanceLevels.find(
		(cal) => cal.competencyId === competencyId && cal.profileId === profileId
	);
}

// ─── Mutations: Pillars ──────────────────────────────────────────────────────

export function addPillar(pillar: Pillar): void {
	pillars = [...pillars, pillar];
}

export function updatePillar(id: string, updates: Partial<Omit<Pillar, 'id'>>): void {
	pillars = pillars.map((p) => (p.id === id ? { ...p, ...updates } : p));
}

export function deletePillar(id: string): void {
	pillars = pillars.filter((p) => p.id !== id);
	// Cascade: remove competencies of this pillar
	competencies = competencies.filter((c) => c.pillarId !== id);
	// Cascade: remove scale criteria of this pillar
	scaleCriteria = scaleCriteria.filter((sc) => sc.pillarId !== id);
}

// ─── Mutations: Competencies ─────────────────────────────────────────────────

export function addCompetency(competency: Competency): void {
	competencies = [...competencies, competency];
}

export function updateCompetency(id: string, updates: Partial<Omit<Competency, 'id'>>): void {
	competencies = competencies.map((c) => (c.id === id ? { ...c, ...updates } : c));
}

export function deleteCompetency(id: string): void {
	competencies = competencies.filter((c) => c.id !== id);
	// Cascade: remove scale criteria for this competency
	scaleCriteria = scaleCriteria.filter((sc) => sc.competencyId !== id);
}

// ─── Mutations: Scale Criteria ───────────────────────────────────────────────

export function updateScaleCriterion(id: string, description: string): void {
	scaleCriteria = scaleCriteria.map((sc) =>
		sc.id === id ? { ...sc, description } : sc
	);
}

export function addScaleCriterion(
	criterion: Omit<ScaleCriterion, 'id'>
): void {
	const id = `sc-${criterion.competencyId}-${criterion.pillarId}-${criterion.level}-${crypto.randomUUID().slice(0, 8)}`;
	scaleCriteria = [...scaleCriteria, { id, ...criterion }];
}

export function removeScaleCriterion(id: string): void {
	scaleCriteria = scaleCriteria.filter((sc) => sc.id !== id);
}

// ─── Mutations: Acceptance Levels (deprecated) ───────────────────────────────

/**
 * @deprecated Use updateLevelDefinition() + setCompetencyAcceptanceLevel() instead.
 */
export function updateAcceptanceLevel(
	profileId: EvaluationProfile,
	level: 1 | 2 | 3 | 4 | 5,
	updates: Partial<Omit<AcceptanceLevel, 'profileId' | 'level'>>
): void {
	acceptanceLevels = acceptanceLevels.map((al) =>
		al.profileId === profileId && al.level === level ? { ...al, ...updates } : al
	);
}

// ─── Mutations: Level Definitions ────────────────────────────────────────────

export function updateLevelDefinition(
	level: 1 | 2 | 3 | 4 | 5,
	label: string,
	description: string
): void {
	levelDefinitions = levelDefinitions.map((ld) =>
		ld.level === level ? { ...ld, label, description } : ld
	);
}

// ─── Mutations: Competency Acceptance Levels ─────────────────────────────────

export function setCompetencyAcceptanceLevel(
	competencyId: string,
	profileId: EvaluationProfile,
	level: 1 | 2 | 3 | 4 | 5
): void {
	const existing = competencyAcceptanceLevels.find(
		(cal) => cal.competencyId === competencyId && cal.profileId === profileId
	);
	if (existing) {
		competencyAcceptanceLevels = competencyAcceptanceLevels.map((cal) =>
			cal.competencyId === competencyId && cal.profileId === profileId
				? { ...cal, level }
				: cal
		);
	} else {
		competencyAcceptanceLevels = [
			...competencyAcceptanceLevels,
			{ competencyId, profileId, level }
		];
	}
}

export function setCompetencyAcceptanceLevelsForProfile(
	profileId: EvaluationProfile,
	assignments: { competencyId: string; level: 1 | 2 | 3 | 4 | 5 }[]
): void {
	// Remove existing entries for this profile
	const filtered = competencyAcceptanceLevels.filter(
		(cal) => cal.profileId !== profileId
	);
	// Add updated entries
	const updated = assignments.map((a) => ({
		competencyId: a.competencyId,
		profileId,
		level: a.level
	}));
	competencyAcceptanceLevels = [...filtered, ...updated];
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
	rh: 3
};

let profileMinLevels = $state<Record<EvaluationProfile, number>>({ ...DEFAULT_MIN_LEVELS });

export function getProfileMinLevels(): Record<EvaluationProfile, number> {
	return profileMinLevels;
}

export function getProfileMinLevel(profileId: EvaluationProfile): number {
	return profileMinLevels[profileId];
}

export function setProfileMinLevel(profileId: EvaluationProfile, level: number): void {
	profileMinLevels = { ...profileMinLevels, [profileId]: level };
}
