import { writable, derived } from 'svelte/store';
import { browser } from '$app/environment';
import { fr } from './fr';
import { en } from './en';

/** Supported locale codes. */
export type Locale = 'fr' | 'en';

const STORAGE_KEY = 'trp_locale';

export const locale = writable<Locale>('fr');

/**
 * Restores the persisted locale from localStorage, if running in the browser.
 */
export function initLocale() {
	if (!browser) return;
	const stored = localStorage.getItem(STORAGE_KEY) as Locale | null;
	if (stored === 'en' || stored === 'fr') locale.set(stored);
}

/**
 * Updates the active locale and persists it to localStorage.
 * @param l locale to activate
 */
export function setLocale(l: Locale) {
	locale.set(l);
	if (browser) localStorage.setItem(STORAGE_KEY, l);
}

const dicts: Record<Locale, Record<string, string>> = { fr, en };

/**
 * Builds the translation function for the current locale.
 * @param $locale currently active locale
 * @returns translation function bound to the current locale's dictionary
 */
export const t = derived(locale, ($locale) => {
	const dict = dicts[$locale];
	/**
	 * Translates a key using the outer closure's dictionary, interpolating variables.
	 * @param key dictionary key to look up
	 * @param vars optional placeholder values to interpolate into the string
	 * @returns translated (and interpolated) string, or the key itself if missing
	 */
	return (key: string, vars?: Record<string, string>): string => {
		let str = dict[key] ?? key;
		if (vars) {
			for (const [k, v] of Object.entries(vars)) str = str.replace(`{${k}}`, v);
		}
		return str;
	};
});
