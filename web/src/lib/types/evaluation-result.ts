import type { EvaluationProfile } from './evaluation';

export type EvaluationStatus = 'pending' | 'in-progress' | 'completed';

export interface CompetencyRating {
	id: string;
	employeeId: string;
	competencyId: string;
	selfRating?: 1 | 2 | 3 | 4 | 5;
	selfComment?: string;
	rhRating?: 1 | 2 | 3 | 4 | 5;
	rhComment?: string;
}

export interface GoalClosure {
	id: string;
	employeeId: string;
	goalId: string;
	finalProgress: number;
	selfAssessment?: string;
	rhAssessment?: string;
	managerComment?: string;
	closedAt?: string;
}
