import { env } from '$env/dynamic/private';
import { proxyToService } from '$lib/server/proxy';
import type { RequestHandler } from './$types';

export const POST: RequestHandler = async ({ request }) => {
	const body = await request.text();
	return proxyToService(env.FRONTEND_ITINERARY_API_URL, '/itinerary/generate', {
		method: 'POST',
		body,
		contentType: 'application/json',
	});
};
