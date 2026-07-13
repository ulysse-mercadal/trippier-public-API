import { writable } from 'svelte/store';
import type { User } from '$lib/types';

const JWT_KEY = 'trp_jwt';

/**
 * Creates the auth store with session state and localStorage-backed token helpers.
 * @returns store object exposing subscribe and auth actions
 */
function createAuthStore() {
	const { subscribe, set } = writable<{ token: string; user: User | null }>({
		token: '',
		user: null,
	});

	return {
		subscribe,
		/**
		 * Sets the active session token and user.
		 * @param token JWT token to store in state
		 * @param user authenticated user object
		 */
		init(token: string, user: User) {
			set({ token, user });
		},
		/**
		 * Clears the stored token and resets session state.
		 */
		logout() {
			if (typeof localStorage !== 'undefined') {
				localStorage.removeItem(JWT_KEY);
			}
			set({ token: '', user: null });
		},
		/**
		 * Reads the persisted JWT from localStorage, if available.
		 * @returns stored token, or empty string if none
		 */
		getStoredToken(): string {
			if (typeof localStorage === 'undefined') return '';
			return localStorage.getItem(JWT_KEY) ?? '';
		},
		/**
		 * Persists a JWT to localStorage.
		 * @param token JWT token to persist
		 */
		storeToken(token: string) {
			if (typeof localStorage !== 'undefined') {
				localStorage.setItem(JWT_KEY, token);
			}
		},
	};
}

export const auth = createAuthStore();
