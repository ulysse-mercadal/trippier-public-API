import { env } from '$env/dynamic/private';
import type { RequestHandler } from './$types';

const proxy: RequestHandler = async ({ request, params }) => {
	const upstream = new URL(`${env.AUTH_API_URL}/${params.path}`);
	upstream.search = new URL(request.url).search; // forward query string (?token=… etc.)

	const headers = new Headers();
	for (const name of ['content-type', 'authorization', 'cookie', 'accept', 'accept-language']) {
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
