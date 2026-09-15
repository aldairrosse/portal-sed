import { baseURL } from './client';

export interface EvaluationExportRow {
	employeeId: string;
	employeeNumber: string;
	employeeName: string;
	goalProgress: number;
	selfAvg: number;
	rhAvg: number;
	rating: number;
	status: string;
}

export interface EvaluationExportResponse {
	data: EvaluationExportRow[];
	meta: { cycleId: string; phase: string; total: number };
}

export async function fetchExport(params: {
	cycleId?: string | null;
	phase?: string | null;
	q?: string;
}): Promise<EvaluationExportResponse> {
	const qs = new URLSearchParams();
	if (params.cycleId) qs.set('cycle_id', params.cycleId);
	if (params.phase) qs.set('phase', params.phase);
	if (params.q) qs.set('q', params.q);
	const res = await fetch(
		`${baseURL}/evaluations/export${qs.size ? `?${qs}` : ''}`,
		{ credentials: 'include' },
	);
	if (!res.ok) throw new Error('No se pudo exportar evaluaciones');
	return (await res.json()) as EvaluationExportResponse;
}
