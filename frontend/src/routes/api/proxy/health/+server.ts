import { env } from '$env/dynamic/private';
import { proxyToService } from '$lib/server/proxy';
import type { RequestHandler } from './$types';

/**
 * Proxies GET requests to the POI service health endpoint.
 * @returns The proxied health check response.
 */
export const GET: RequestHandler = () =>
	proxyToService(env.FRONTEND_POI_API_URL, '/health');
