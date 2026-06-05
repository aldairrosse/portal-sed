export interface OrgNode {
	id: string; // employeeId
	name: string;
	profileId: string; // EvaluationProfile
	managerId: string | null;
	children: OrgNode[];
}
