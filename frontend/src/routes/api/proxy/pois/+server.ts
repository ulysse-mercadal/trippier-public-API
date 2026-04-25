import { env } from '$env/dynamic/private';
import { buildInternalAuth } from '$lib/server/internal-auth';
import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ url }) => {
	const base = env.FRONTEND_POI_API_URL;
	if (!base) return json({ error: 'FRONTEND_POI_API_URL not configured' }, { status: 503 });

	const upstream = new URL(`${base}/pois/search`);
	upstream.search = url.search;

	try {
		const res = await fetch(upstream.toString(), {
			headers: { 'X-Internal-Auth': buildInternalAuth(env.INTERNAL_SECRET ?? '') },
		});
		const body = await res.text();
		return new Response(body, {
			status: res.status,
			headers: { 'content-type': res.headers.get('content-type') ?? 'application/json' },
		});
	} catch {
		return json({ error: 'poi-api unreachable' }, { status: 503 });
	}
};
