import { env } from '$env/dynamic/private';
import { proxyToService } from '$lib/server/proxy';
import type { RequestHandler } from './$types';

/**
 * Forwards an itinerary generation request to the itinerary backend service.
 * @param request - incoming HTTP request with the itinerary generation payload
 * @returns proxied response from the itinerary service
 */
export const POST: RequestHandler = async ({ request }) => {
	const body = await request.text();
	return proxyToService(env.FRONTEND_ITINERARY_API_URL, '/v1/itinerary/generate', {
		method: 'POST',
		body,
		contentType: 'application/json',
	});
};
