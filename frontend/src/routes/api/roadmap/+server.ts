import { json, error } from '@sveltejs/kit';
import { readFileSync, writeFileSync } from 'fs';
import { resolve } from 'path';
import { env } from '$env/dynamic/private';
import type { RequestHandler } from './$types';
import type { RoadmapData } from '$lib/data/roadmap';

const DATA_PATH    = resolve('data/roadmap.json');
const ADMIN_EMAILS = (env.ADMIN_EMAILS ?? '').split(',').map(e => e.trim()).filter(Boolean);
const AUTH_API_URL = env.AUTH_API_URL ?? 'http://localhost:8080';

/**
 * Loads the roadmap data from the on-disk JSON file.
 * @returns Parsed roadmap data.
 */
function readData(): RoadmapData {
	return JSON.parse(readFileSync(DATA_PATH, 'utf-8'));
}

/**
 * Persists roadmap data to the on-disk JSON file.
 * @param data Roadmap data to write.
 */
function writeData(data: RoadmapData): void {
	writeFileSync(DATA_PATH, JSON.stringify(data, null, 2), 'utf-8');
}

/**
 * Verifies a bearer token against the auth API and checks admin membership.
 * @param token Bearer token to verify.
 * @returns Admin email if authorized, otherwise null.
 */
async function getAdminEmail(token: string): Promise<string | null> {
	try {
		const res = await fetch(`${AUTH_API_URL}/v1/me`, {
			headers: { Authorization: `Bearer ${token}` },
		});
		if (!res.ok) return null;
		const { email } = await res.json();
		return ADMIN_EMAILS.includes(email) ? email : null;
	} catch {
		return null;
	}
}

/**
 * Handles GET requests by returning the current roadmap data.
 * @returns JSON response with roadmap data.
 */
export const GET: RequestHandler = () => {
	try {
		return json(readData());
	} catch {
		throw error(500, 'Failed to read roadmap data');
	}
};

/**
 * Handles PUT requests by validating admin auth and overwriting roadmap data.
 * @param request Incoming request, expected to hold auth header and JSON body.
 * @returns JSON response confirming the write.
 */
export const PUT: RequestHandler = async ({ request }) => {
	const auth  = request.headers.get('Authorization') ?? '';
	const token = auth.startsWith('Bearer ') ? auth.slice(7) : '';
	if (!token) throw error(401, 'Missing token');

	const admin = await getAdminEmail(token);
	if (!admin) throw error(403, 'Forbidden — admin only');

	let body: RoadmapData;
	try {
		body = await request.json();
	} catch {
		throw error(400, 'Invalid JSON body');
	}

	if (!Array.isArray(body?.columns)) throw error(400, 'Invalid roadmap shape');

	try {
		writeData(body);
		return json({ ok: true });
	} catch {
		throw error(500, 'Failed to write roadmap data');
	}
};
