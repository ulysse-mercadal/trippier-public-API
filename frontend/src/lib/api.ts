import type { ApiKeyWithUsage, CreateKeyResult, User } from '$lib/types';

const BASE = '/api/auth';

class ApiError extends Error {
	constructor(
		public readonly status: number,
		message: string,
	) {
		super(message);
	}
}

async function request<T>(path: string, token?: string, init?: RequestInit): Promise<T> {
	const headers: Record<string, string> = {
		'Content-Type': 'application/json',
		...(init?.headers as Record<string, string>),
	};
	if (token) headers['Authorization'] = `Bearer ${token}`;

	const res = await fetch(`${BASE}${path}`, { ...init, headers });
	const body = await res.json().catch(() => ({}));

	if (!res.ok) {
		throw new ApiError(res.status, (body as { error?: string }).error ?? res.statusText);
	}
	return body as T;
}

export async function register(email: string, password: string): Promise<void> {
	await request('/register', undefined, {
		method: 'POST',
		body: JSON.stringify({ email, password }),
	});
}

export async function login(email: string, password: string): Promise<string> {
	const { token } = await request<{ token: string }>('/login', undefined, {
		method: 'POST',
		body: JSON.stringify({ email, password }),
	});
	return token;
}

export async function getMe(token: string): Promise<User> {
	return request<User>('/me', token);
}

export async function listKeys(token: string): Promise<ApiKeyWithUsage[]> {
	const { keys } = await request<{ keys: ApiKeyWithUsage[] }>('/api-keys', token);
	return keys ?? [];
}

export async function createKey(token: string, name: string): Promise<CreateKeyResult> {
	return request<CreateKeyResult>('/api-keys', token, {
		method: 'POST',
		body: JSON.stringify({ name }),
	});
}

export async function revokeKey(token: string, id: string): Promise<void> {
	await request(`/api-keys/${id}`, token, { method: 'DELETE' });
}

export { ApiError };
