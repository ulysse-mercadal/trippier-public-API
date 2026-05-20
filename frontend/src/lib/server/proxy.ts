import { buildInternalAuth } from '$lib/server/internal-auth';
import { json } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';

interface ProxyOptions {
	method?: string;
	body?: string;
	contentType?: string;
}

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
