import { writable } from 'svelte/store';
import { browser } from '$app/environment';

export type Theme = 'dark' | 'light';

const STORAGE_KEY = 'trp_theme';

export const theme = writable<Theme>('dark');

function apply(t: Theme) {
	if (!browser) return;
	document.documentElement.dataset.theme = t;
}

export function initTheme() {
	if (!browser) return;
	const stored = localStorage.getItem(STORAGE_KEY) as Theme | null;
	const initial: Theme = stored === 'light' || stored === 'dark' ? stored : 'dark';
	theme.set(initial);
	apply(initial);
}

export function setTheme(t: Theme) {
	theme.set(t);
	apply(t);
	if (browser) localStorage.setItem(STORAGE_KEY, t);
}

export function toggleTheme() {
	let next: Theme = 'dark';
	theme.update((t) => {
		next = t === 'dark' ? 'light' : 'dark';
		return next;
	});
	apply(next);
	if (browser) localStorage.setItem(STORAGE_KEY, next);
}
