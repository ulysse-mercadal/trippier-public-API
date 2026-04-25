import { afterEach, describe, expect, it, vi } from 'vitest';
import { ApiError, login, register, getMe, listKeys, createKey, revokeKey } from './api';

// Minimal fetch mock: returns a Response-like object.
function mockFetch(status: number, body: unknown) {
	vi.stubGlobal('fetch', async () => ({
		ok: status >= 200 && status < 300,
		status,
		statusText: status === 200 ? 'OK' : 'Error',
		json: async () => body,
	}));
}

afterEach(() => {
	vi.unstubAllGlobals();
});

// ── ApiError ──────────────────────────────────────────────────────────────────

describe('ApiError', () => {
	it('is an instance of Error', () => {
		const e = new ApiError(401, 'Unauthorized');
		expect(e).toBeInstanceOf(Error);
	});

	it('exposes status and message', () => {
		const e = new ApiError(429, 'rate limit exceeded');
		expect(e.status).toBe(429);
		expect(e.message).toBe('rate limit exceeded');
	});
});

// ── login ─────────────────────────────────────────────────────────────────────

describe('login', () => {
	it('returns token on success', async () => {
		mockFetch(200, { token: 'jwt-abc' });
		const token = await login('user@example.com', 'password');
		expect(token).toBe('jwt-abc');
	});

	it('throws ApiError on 401', async () => {
		mockFetch(401, { error: 'invalid credentials' });
		await expect(login('user@example.com', 'wrong')).rejects.toBeInstanceOf(ApiError);
	});

	it('uses the error message from the response body', async () => {
		mockFetch(401, { error: 'email not verified' });
		try {
			await login('user@example.com', 'pw');
			expect.fail('should have thrown');
		} catch (e) {
			expect((e as ApiError).message).toBe('email not verified');
		}
	});

	it('falls back to statusText when body has no error field', async () => {
		vi.stubGlobal('fetch', async () => ({
			ok: false,
			status: 500,
			statusText: 'Internal Server Error',
			json: async () => ({}),
		}));
		try {
			await login('user@example.com', 'pw');
			expect.fail('should have thrown');
		} catch (e) {
			expect((e as ApiError).message).toBe('Internal Server Error');
		}
	});
});

// ── register ──────────────────────────────────────────────────────────────────

describe('register', () => {
	it('resolves without value on success', async () => {
		mockFetch(201, {});
		await expect(register('user@example.com', 'password123')).resolves.toBeUndefined();
	});

	it('throws ApiError on 400', async () => {
		mockFetch(400, { error: 'email already registered' });
		await expect(register('taken@example.com', 'pw')).rejects.toBeInstanceOf(ApiError);
	});
});

// ── getMe ─────────────────────────────────────────────────────────────────────

describe('getMe', () => {
	it('returns user object', async () => {
		const user = { id: 'u1', email: 'user@example.com', verified: true, created_at: '2025-01-01' };
		mockFetch(200, user);
		const result = await getMe('jwt-token');
		expect(result.id).toBe('u1');
		expect(result.email).toBe('user@example.com');
	});

	it('throws ApiError on 401', async () => {
		mockFetch(401, { error: 'unauthorized' });
		await expect(getMe('bad-token')).rejects.toBeInstanceOf(ApiError);
	});
});

// ── listKeys ──────────────────────────────────────────────────────────────────

describe('listKeys', () => {
	it('returns keys array', async () => {
		mockFetch(200, { keys: [{ id: 'k1', name: 'my-app' }] });
		const keys = await listKeys('jwt-token');
		expect(keys).toHaveLength(1);
		expect(keys[0].id).toBe('k1');
	});

	it('returns empty array when keys is null/missing', async () => {
		mockFetch(200, { keys: null });
		const keys = await listKeys('jwt-token');
		expect(keys).toEqual([]);
	});
});

// ── createKey ─────────────────────────────────────────────────────────────────

describe('createKey', () => {
	it('returns CreateKeyResult on success', async () => {
		const result = { key: 'trp_abc123', metadata: { id: 'k2', name: 'prod' } };
		mockFetch(201, result);
		const res = await createKey('jwt-token', 'prod');
		expect(res.key).toBe('trp_abc123');
	});
});

// ── revokeKey ─────────────────────────────────────────────────────────────────

describe('revokeKey', () => {
	it('resolves without value on success', async () => {
		mockFetch(200, {});
		await expect(revokeKey('jwt-token', 'k1')).resolves.toBeUndefined();
	});

	it('throws ApiError on 404', async () => {
		mockFetch(404, { error: 'not found' });
		await expect(revokeKey('jwt-token', 'missing')).rejects.toBeInstanceOf(ApiError);
	});
});
