import { env } from '$env/dynamic/private';
import { json } from '@sveltejs/kit';
import { proxyToService } from '$lib/server/proxy';
import type { RequestHandler } from './$types';

const ALLOWED_SUBPATHS = new Set(['search', 'search/slim', 'events', 'events/slim', 'providers']);

export const GET: RequestHandler = async ({ url }) => {
	const subpath = url.searchParams.get('subpath') ?? 'search';
	if (!ALLOWED_SUBPATHS.has(subpath)) {
		return json({ error: 'invalid subpath' }, { status: 400 });
	}

	const qs = new URLSearchParams(url.searchParams);
	qs.delete('subpath');

	return proxyToService(env.FRONTEND_POI_API_URL, `/pois/${subpath}?${qs}`);
};
