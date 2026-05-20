import { env } from '$env/dynamic/private';
import { proxyToService } from '$lib/server/proxy';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = () =>
	proxyToService(env.FRONTEND_POI_API_URL, '/health');
