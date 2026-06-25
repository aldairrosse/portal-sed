/** Performance or potential tier (1-3) */
export type NineBoxTier = 1 | 2 | 3;

export interface NineBoxEntry {
	id: string;
	employeeId: string;
	employeeName: string;
	profileId: string;
	performanceTier: NineBoxTier; // X axis (1-3) — derived from goal progress
	potentialTier: NineBoxTier;   // Y axis (1-3) — derived from competency ratings
	quadrant: number;             // 1-9 computed from tiers: (potentialTier - 1) * 3 + performanceTier
	cycleId?: string;
}

export interface NineBoxQuadrantDef {
	quadrant: number;             // 1-9 (fixed: Q1 = Bajo-Bajo, Q5 = Medio-Medio, Q9 = Alto-Alto)
	label: string;                // short name e.g. "Estrella"
	title: string;                // editable by RH
	description: string;          // editable by RH
	colorHex: string;             // hex color e.g. "#22C55E"
	actionRecommendation: string;
}

export interface NineBoxMatrix {
	entries: NineBoxEntry[];
	quadrantDefs: NineBoxQuadrantDef[];
	scopeEmployeeIds: string[];
	cycleId?: string;
	phaseId?: string;
}
