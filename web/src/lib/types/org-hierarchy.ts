export interface OrgNode {
	id: string;
	name: string;
	profileId?: string;
	managerId?: string | null;
	headEmployeeId?: string;
	headEmployee?: {
		id: string;
		firstName: string;
		lastName: string;
		jobTitle: string;
	};
	children?: OrgNode[];
}
