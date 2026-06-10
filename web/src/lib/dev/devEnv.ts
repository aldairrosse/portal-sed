export function isDev(): boolean {
	return import.meta.env.DEV || import.meta.env.VITE_SHOW_DEV_TOOLBAR === 'true';
}
