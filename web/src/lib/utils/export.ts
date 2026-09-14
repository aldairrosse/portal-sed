import * as XLSX from 'xlsx';

export function toXlsx(
	rows: Record<string, string | number | null>[],
	filename: string,
	sheetName = 'Datos',
): void {
	if (rows.length === 0) {
		console.warn(`[export] No data to export for "${filename}"`);
		return;
	}
	const ws = XLSX.utils.json_to_sheet(rows);
	// Auto width por header/contenido (min 10, max 40)
	const headers = Object.keys(rows[0]);
	const colWidths = headers.map((h) => {
		const maxLen = Math.max(
			h.length,
			...rows.map((r) => String(r[h] ?? '').length),
		);
		return { wch: Math.min(40, Math.max(10, maxLen + 2)) };
	});
	ws['!cols'] = colWidths;

	const wb = XLSX.utils.book_new();
	XLSX.utils.book_append_sheet(wb, ws, sheetName);
	const name = filename.endsWith('.xlsx') ? filename : `${filename}.xlsx`;
	XLSX.writeFile(wb, name);
}

/** @deprecated usar toXlsx */
export const toCsv = toXlsx;
