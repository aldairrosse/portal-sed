export type EvaluationStatus = 'pending' | 'in-progress' | 'completed';

export interface CompetencyRating {
	id: string;
	employeeId: string;
	competencyId: string;
	selfRating?: 1 | 2 | 3 | 4 | 5;
	selfComment?: string;
	rhRating?: 1 | 2 | 3 | 4 | 5;
	rhComment?: string;
	managerComment?: string | null;
	managerCommentAuthor?: string;
	managerCommentCreatedAt?: string;
	acceptanceLevel?: number;
	// F3: author/date stub — populated when the API returns comment metadata.
	authorName?: string;
	commentCreatedAt?: string;
}

export interface GoalClosure {
	id: string;
	employeeId: string;
	goalId: string;
	finalProgress: number;
	selfAssessment?: string;
	rhAssessment?: string;
	managerComment?: string;
	// F2: author/date per comment (row-level timestamps from API).
	selfAssessmentAuthor?: string;
	selfAssessmentCreatedAt?: string;
	rhAssessmentAuthor?: string;
	rhAssessmentCreatedAt?: string;
	managerCommentAuthor?: string;
	managerCommentCreatedAt?: string;
	closedAt?: string;
}
