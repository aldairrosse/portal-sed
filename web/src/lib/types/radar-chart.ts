export interface RadarCompetencyPoint {
	competencyId: string;
	competencyName: string;
	selfRating: number | null;
	rhRating: number | null;
}

export interface RadarPillarGroup {
	pillarId: string;
	pillarName: string;
	competencies: RadarCompetencyPoint[];
}
