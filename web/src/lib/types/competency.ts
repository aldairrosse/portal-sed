import type { EvaluationProfile } from './evaluation';

/**
 * A pillar (pilar) groups related competencies.
 * Examples: Liderazgo, Técnico, Comportamental.
 */
export interface Pillar {
	id: string;
	name: string;
	description: string;
}

/**
 * A competency belongs to exactly one pillar.
 */
export interface Competency {
	id: string;
	name: string;
	description: string;
	pillarId: string;
}

/**
 * Scale criterion describes what a specific level (1-5) means
 * for a given competency within a pillar.
 */
export interface ScaleCriterion {
	id: string;
	competencyId: string;
	pillarId: string;
	level: 1 | 2 | 3 | 4 | 5;
	description: string;
}

/**
 * Acceptance level defines the label and description
 * for each level (1-5) per evaluation profile.
 *
 * @deprecated Use LevelDefinition + CompetencyAcceptanceLevel instead.
 * Level definitions are now global, and competency acceptance
 * levels are assigned per competency per profile.
 */
export interface AcceptanceLevel {
	profileId: EvaluationProfile;
	level: 1 | 2 | 3 | 4 | 5;
	label: string;
	description: string;
}

/**
 * Global level definition — same label and description across all profiles.
 */
export interface LevelDefinition {
	level: 1 | 2 | 3 | 4 | 5;
	label: string;
	description: string;
}

/**
 * Per-competency, per-profile acceptance level assignment.
 */
export interface CompetencyAcceptanceLevel {
	competencyId: string;
	profileId: EvaluationProfile;
	level: 1 | 2 | 3 | 4 | 5;
}
