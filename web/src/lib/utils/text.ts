/** Capitaliza la primera letra de cada palabra, limpia espacios y reemplaza guiones por espacios. */
export function titleCase(value: string): string {
	return value
		.trim()
		.replace(/-/g, ' ')
		.replace(/\s+/g, ' ')
		.replace(/\b\w/g, (c) => c.toUpperCase());
}
