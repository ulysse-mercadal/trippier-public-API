import { writable } from 'svelte/store';
import type { User } from '$lib/types';

const JWT_KEY = 'trp_jwt';

function createAuthStore() {
	const { subscribe, set } = writable<{ token: string; user: User | null }>({
		token: '',
		user: null,
	});

	return {
		subscribe,
		init(token: string, user: User) {
			set({ token, user });
		},
		logout() {
			if (typeof localStorage !== 'undefined') {
				localStorage.removeItem(JWT_KEY);
			}
			set({ token: '', user: null });
		},
		getStoredToken(): string {
			if (typeof localStorage === 'undefined') return '';
			return localStorage.getItem(JWT_KEY) ?? '';
		},
		storeToken(token: string) {
			if (typeof localStorage !== 'undefined') {
				localStorage.setItem(JWT_KEY, token);
			}
		},
	};
}

export const auth = createAuthStore();
