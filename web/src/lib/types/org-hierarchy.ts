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
		profileName?: string;
		profileDescription?: string;
	};
	employeeCount?: number;
	children?: OrgNode[];
}
