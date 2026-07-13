import { env } from '$env/dynamic/private';
import { json } from '@sveltejs/kit';
import { proxyToService } from '$lib/server/proxy';
import type { RequestHandler } from './$types';

const ALLOWED_SUBPATHS = new Set([
	'search',
	'search/slim',
	'search/custom',
	'search/custom/slim',
	'events',
	'events/slim',
	'events/custom',
	'events/custom/slim',
	'providers',
	'providers/catalog',
	'providers/recommend',
]);

/**
 * Proxies GET requests for POI data to the POI API service.
 * @param url - incoming request URL with query params, including subpath
 * @returns JSON error response if subpath is invalid, otherwise the proxied response
 */
export const GET: RequestHandler = async ({ url }) => {
	const subpath = url.searchParams.get('subpath') ?? 'search';
	if (!ALLOWED_SUBPATHS.has(subpath)) {
		return json({ error: 'invalid subpath' }, { status: 400 });
	}

	const qs = new URLSearchParams(url.searchParams);
	qs.delete('subpath');

	return proxyToService(env.FRONTEND_POI_API_URL, `/v1/pois/${subpath}?${qs}`);
};
