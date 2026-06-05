/** 9×9 performance scale (1-9) */
export type NineBoxScale = 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9;

/** Quadrant classification based on performance × potential */
export type NineBoxQuadrant =
	| 'star' // high perf + high potential (top-right)
	| 'growth' // high perf + medium potential
	| 'high-potential' // medium perf + high potential
	| 'core-player' // medium perf + medium potential (center)
	| 'risk' // low perf + high potential
	| 'effective' // low perf + medium potential
	| 'underperformer'; // low perf + low potential (bottom-left)

export interface NineBoxEntry {
	id: string;
	employeeId: string;
	employeeName: string;
	profileId: string;
	performance: NineBoxScale; // X axis (1-9)
	potential: NineBoxScale; // Y axis (1-9)
	quadrant: NineBoxQuadrant; // computed from scores
	cycleId?: string;
}

export interface NineBoxQuadrantDef {
	id: NineBoxQuadrant;
	label: string;
	description: string;
	/** DaisyUI-compatible color class for cell background */
	colorClass: string;
	/** Performance range [min, max] inclusive */
	perfRange: [number, number];
	/** Potential range [min, max] inclusive */
	potRange: [number, number];
}

export interface NineBoxMatrix {
	entries: NineBoxEntry[];
	/** Filter: which employees are in scope for current viewer */
	scopeEmployeeIds: string[];
}

export interface NineBoxScaleDef {
	level: NineBoxScale;
	label: string;
	description: string;
}
