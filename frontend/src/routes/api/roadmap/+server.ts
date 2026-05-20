import { json, error } from '@sveltejs/kit';
import { readFileSync, writeFileSync } from 'fs';
import { resolve } from 'path';
import { env } from '$env/dynamic/private';
import type { RequestHandler } from './$types';
import type { RoadmapData } from '$lib/data/roadmap';

const DATA_PATH    = resolve('data/roadmap.json');
const ADMIN_EMAILS = (env.ADMIN_EMAILS ?? '').split(',').map(e => e.trim()).filter(Boolean);
const AUTH_API_URL = env.AUTH_API_URL ?? 'http://localhost:8080';

// Reads roadmap data from disk
function readData(): RoadmapData {
	return JSON.parse(readFileSync(DATA_PATH, 'utf-8'));
}

// Writes roadmap data to disk atomically
function writeData(data: RoadmapData): void {
	writeFileSync(DATA_PATH, JSON.stringify(data, null, 2), 'utf-8');
}

// Verifies JWT and returns admin email or null
async function getAdminEmail(token: string): Promise<string | null> {
	try {
		const res = await fetch(`${AUTH_API_URL}/me`, {
			headers: { Authorization: `Bearer ${token}` },
		});
		if (!res.ok) return null;
		const { email } = await res.json();
		return ADMIN_EMAILS.includes(email) ? email : null;
	} catch {
		return null;
	}
}

export const GET: RequestHandler = () => {
	try {
		return json(readData());
	} catch {
		throw error(500, 'Failed to read roadmap data');
	}
};

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
