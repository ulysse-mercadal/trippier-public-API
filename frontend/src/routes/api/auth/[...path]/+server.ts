import { env } from '$env/dynamic/private';
import type { RequestHandler } from './$types';

/**
 * Forwards an incoming request to the auth API and relays its response.
 * @param request the incoming request to forward
 * @param params route parameters, including the wildcard path
 * @returns the proxied response from the auth API
 */
const proxy: RequestHandler = async ({ request, params }) => {
	const upstream = new URL(`${env.AUTH_API_URL}/v1/${params.path}`);
	upstream.search = new URL(request.url).search;

	const headers = new Headers();
	for (const name of ['content-type', 'authorization', 'accept', 'accept-language']) {
		const value = request.headers.get(name);
		if (value) headers.set(name, value);
	}
	const body = ['GET', 'HEAD'].includes(request.method) ? undefined : await request.text();

	const res = await fetch(upstream.toString(), { method: request.method, headers, body, redirect: 'manual' });

	const resHeaders = new Headers();
	const ct = res.headers.get('content-type');
	if (ct) resHeaders.set('content-type', ct);
	const location = res.headers.get('location');
	if (location) resHeaders.set('location', location);

	return new Response(res.status >= 300 && res.status < 400 ? null : res.body, {
		status: res.status,
		headers: resHeaders,
	});
};

export const GET = proxy;
export const POST = proxy;
export const DELETE = proxy;
export const PUT = proxy;
export const PATCH = proxy;
