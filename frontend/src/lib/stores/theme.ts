import { writable } from 'svelte/store';
import { browser } from '$app/environment';

/** Supported UI theme values. */
export type Theme = 'dark' | 'light';

const STORAGE_KEY = 'trp_theme';

export const theme = writable<Theme>('dark');

/**
 * Applies the given theme to the document root.
 * @param t theme to apply
 */
function apply(t: Theme) {
	if (!browser) return;
	document.documentElement.dataset.theme = t;
}

/**
 * Restores the persisted theme from storage and applies it.
 */
export function initTheme() {
	if (!browser) return;
	const stored = localStorage.getItem(STORAGE_KEY) as Theme | null;
	const initial: Theme = stored === 'light' || stored === 'dark' ? stored : 'dark';
	theme.set(initial);
	apply(initial);
}

/**
 * Sets the active theme, applies it, and persists it to storage.
 * @param t theme to set
 */
export function setTheme(t: Theme) {
	theme.set(t);
	apply(t);
	if (browser) localStorage.setItem(STORAGE_KEY, t);
}

/**
 * Switches between dark and light themes and persists the new value.
 */
export function toggleTheme() {
	let next: Theme = 'dark';
	theme.update((t) => {
		next = t === 'dark' ? 'light' : 'dark';
		return next;
	});
	apply(next);
	if (browser) localStorage.setItem(STORAGE_KEY, next);
}
