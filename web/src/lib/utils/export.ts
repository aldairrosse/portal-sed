export function toCsv(rows: Record<string, string | number | null>[], filename: string): void {
	if (rows.length === 0) {
		console.warn(`[export] No data to export for "${filename}"`);
		return;
	}

	const headers = Object.keys(rows[0]);
	const escapeField = (value: unknown): string => {
		const s = value == null ? '' : String(value);
		if (s.includes('"') || s.includes(';') || s.includes('\n') || s.includes('\r')) {
			return `"${s.replace(/"/g, '""')}"`;
		}
		return s;
	};

	const bom = '\uFEFF';
	const headerLine = headers.join(';');
	const dataLines = rows.map((row) => headers.map((h) => escapeField(row[h])).join(';'));
	const csv = bom + headerLine + '\r\n' + dataLines.join('\r\n');

	const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
	const url = URL.createObjectURL(blob);
	const a = document.createElement('a');
	a.href = url;
	a.download = filename.endsWith('.csv') ? filename : `${filename}.csv`;
	document.body.appendChild(a);
	a.click();
	document.body.removeChild(a);
	URL.revokeObjectURL(url);
}
