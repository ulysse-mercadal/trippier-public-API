import type { ApiKeyWithUsage, CreateKeyResult, User } from '$lib/types';

const BASE = '/api/auth';

/** Error thrown when an auth API request fails, carrying the HTTP status. */
class ApiError extends Error {
	/**
	 * Builds an ApiError from a response status and message.
	 * @param status HTTP status code of the failed response
	 * @param message error message
	 */
	constructor(
		public readonly status: number,
		message: string,
	) {
		super(message);
	}
}

/**
 * Sends a request to the auth API and parses the JSON response.
 * @param path endpoint path appended to the base URL
 * @param token optional bearer token for authorization
 * @param init optional fetch request options
 * @returns parsed response body
 */
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

/**
 * Registers a new account and triggers the OTP email.
 * @param email account email address
 * @param password account password
 */
export async function register(email: string, password: string): Promise<void> {
	await request('/register', undefined, {
		method: 'POST',
		body: JSON.stringify({ email, password }),
	});
}

/**
 * Verifies the 6-digit OTP sent by email and logs the user in.
 * @param email account email address
 * @param code 6-digit OTP code
 * @returns JWT token on success
 */
export async function verifyCode(email: string, code: string): Promise<string> {
	const { token } = await request<{ token: string }>('/verify-code', undefined, {
		method: 'POST',
		body: JSON.stringify({ email, code }),
	});
	return token;
}

/**
 * Logs in with email and password.
 * @param email account email address
 * @param password account password
 * @returns JWT token on success
 */
export async function login(email: string, password: string): Promise<string> {
	const { token } = await request<{ token: string }>('/login', undefined, {
		method: 'POST',
		body: JSON.stringify({ email, password }),
	});
	return token;
}

/**
 * Requests a new OTP code to be sent by email.
 * @param email account email address
 */
export async function resendCode(email: string): Promise<void> {
	await request('/resend-code', undefined, {
		method: 'POST',
		body: JSON.stringify({ email }),
	});
}

/**
 * Fetches the current authenticated user's profile.
 * @param token bearer auth token
 * @returns current user
 */
export async function getMe(token: string): Promise<User> {
	return request<User>('/me', token);
}

/**
 * Lists the API keys belonging to the authenticated user, with usage.
 * @param token bearer auth token
 * @returns list of API keys with usage
 */
export async function listKeys(token: string): Promise<ApiKeyWithUsage[]> {
	const { keys } = await request<{ keys: ApiKeyWithUsage[] }>('/api-keys', token);
	return keys ?? [];
}

/**
 * Creates a new API key for the authenticated user.
 * @param token bearer auth token
 * @param name label for the new key
 * @returns created key result
 */
export async function createKey(token: string, name: string): Promise<CreateKeyResult> {
	return request<CreateKeyResult>('/api-keys', token, {
		method: 'POST',
		body: JSON.stringify({ name }),
	});
}

/**
 * Revokes an existing API key.
 * @param token bearer auth token
 * @param id identifier of the key to revoke
 */
export async function revokeKey(token: string, id: string): Promise<void> {
	await request(`/api-keys/${id}`, token, { method: 'DELETE' });
}

export { ApiError };
