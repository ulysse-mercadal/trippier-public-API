import { buildInternalAuth } from '$lib/server/internal-auth';
import { json } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';

/** Optional overrides for a proxied request. */
interface ProxyOptions {
	method?: string;
	body?: string;
	contentType?: string;
}

/**
 * Forwards a request to an internal service, attaching internal auth headers.
 * @param serviceUrl Base URL of the target service; if unset, returns a 503.
 * @param path Path (with leading slash) appended to serviceUrl.
 * @param opts Optional method, body, and content type overrides.
 * @returns The upstream response, or a 503 JSON error if unreachable/unconfigured.
 */
export async function proxyToService(
	serviceUrl: string | undefined,
	path: string,
	opts: ProxyOptions = {}
): Promise<Response> {
	if (!serviceUrl) {
		return json({ error: `${path}: service URL not configured` }, { status: 503 });
	}

	const headers: Record<string, string> = {
		'X-Internal-Auth': buildInternalAuth(env.INTERNAL_SECRET ?? ''),
	};
	if (opts.contentType) headers['content-type'] = opts.contentType;

	try {
		const res = await fetch(`${serviceUrl}${path}`, {
			method: opts.method ?? 'GET',
			headers,
			body: opts.body,
		});
		const body = await res.text();
		return new Response(body, {
			status: res.status,
			headers: { 'content-type': res.headers.get('content-type') ?? 'application/json' },
		});
	} catch {
		return json({ error: 'upstream service unreachable' }, { status: 503 });
	}
}
