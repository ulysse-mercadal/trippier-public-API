import { env } from '$env/dynamic/private';
import { buildInternalAuth } from '$lib/server/internal-auth';
import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

export const POST: RequestHandler = async ({ request }) => {
	const base = env.FRONTEND_ITINERARY_API_URL;
	if (!base) return json({ error: 'FRONTEND_ITINERARY_API_URL not configured' }, { status: 503 });

	const body = await request.text();

	try {
		const res = await fetch(`${base}/itinerary`, {
			method: 'POST',
			headers: {
				'content-type': 'application/json',
				'X-Internal-Auth': buildInternalAuth(env.INTERNAL_SECRET ?? ''),
			},
			body,
		});
		const resBody = await res.text();
		return new Response(resBody, {
			status: res.status,
			headers: { 'content-type': res.headers.get('content-type') ?? 'application/json' },
		});
	} catch {
		return json({ error: 'itinerary-api unreachable' }, { status: 503 });
	}
};
