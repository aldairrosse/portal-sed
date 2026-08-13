import { writable } from 'svelte/store';
import { browser } from '$app/environment';

export type Theme = 'light' | 'dark';

const STORAGE_KEY = 'sed-theme';

function applyTheme(value: Theme) {
	document.documentElement.dataset.theme = value === 'dark' ? 'sed-dark' : 'sed';
}

function createThemeStore() {
	const { subscribe, set } = writable<Theme>('light');

	return {
		subscribe,
		set: (value: Theme) => {
			set(value);
			if (browser) {
				localStorage.setItem(STORAGE_KEY, value);
				applyTheme(value);
			}
		}
	};
}

export const theme = createThemeStore();

/** Restore persisted theme (or default light) before first render. */
export function initTheme(): void {
	if (!browser) return;
	const stored = localStorage.getItem(STORAGE_KEY);
	theme.set(stored === 'dark' ? 'dark' : 'light');
}
