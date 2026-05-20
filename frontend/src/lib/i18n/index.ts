import { writable, derived } from 'svelte/store';
import { browser } from '$app/environment';
import { fr } from './fr';
import { en } from './en';

export type Locale = 'fr' | 'en';

const STORAGE_KEY = 'trp_locale';

export const locale = writable<Locale>('fr');

export function initLocale() {
	if (!browser) return;
	const stored = localStorage.getItem(STORAGE_KEY) as Locale | null;
	if (stored === 'en' || stored === 'fr') locale.set(stored);
}

export function setLocale(l: Locale) {
	locale.set(l);
	if (browser) localStorage.setItem(STORAGE_KEY, l);
}

const dicts: Record<Locale, Record<string, string>> = { fr, en };

export const t = derived(locale, ($locale) => {
	const dict = dicts[$locale];
	return (key: string, vars?: Record<string, string>): string => {
		let str = dict[key] ?? key;
		if (vars) {
			for (const [k, v] of Object.entries(vars)) str = str.replace(`{${k}}`, v);
		}
		return str;
	};
});
