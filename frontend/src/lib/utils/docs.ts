import type { DocPage, Param } from '$lib/data/docs';

/**
 * Maps an HTTP method to its accent color.
 * @param m HTTP method name.
 * @returns CSS color value for the method.
 */
export function methodColor(m: string): string {
	return ({ GET: 'var(--accent)', POST: '#e4b07a', DELETE: '#e78f8f', PUT: '#c89be0' } as Record<string, string>)[m] ?? 'var(--text)';
}

const SAMPLES: Record<string, string> = {
	lat: '48.8566', lng: '2.3522', radius: '5000',
	date: new Date().toISOString().slice(0, 10),
	mode: 'radius', types: 'see,eat', lang: 'fr',
	limit: '20', offset: '0', min_score: '0.3',
	district: 'Vieux-Lyon', polygon: '45.76,4.83;45.77,4.84;45.76,4.85',
	providers: 'osm,wikidata',
};

/** Params pre-filled by default in the "Try it" console; the rest start empty. */
const DEFAULT_KEYS = new Set(['lat', 'lng', 'radius', 'date']);

/**
 * Returns a representative sample value used as a param's input placeholder.
 * @param p a route parameter.
 * @returns sample string value for the param.
 */
export function sampleValue(p: Param): string {
	return SAMPLES[p.name] ?? 'value';
}

/**
 * Returns the initial value shown in the console for a param — a sample for
 * the primary geo params, empty otherwise.
 * @param p a route parameter.
 * @returns initial input value for the param.
 */
export function initialValue(p: Param): string {
	return DEFAULT_KEYS.has(p.name) ? (SAMPLES[p.name] ?? '') : '';
}

/**
 * Returns the query params to send: non-path params whose value is non-empty.
 * @param route the current route.
 * @param values current name→value map from the console inputs.
 * @returns filtered list of active query params.
 */
export function activeQueryParams(route: DocPage, values: Record<string, string>): Param[] {
	return (route.params ?? []).filter(p => p.in !== 'path' && (values[p.name] ?? '').trim() !== '');
}

/**
 * Returns route.path with any `:name` path params substituted from values.
 * @param route the current route.
 * @param values current name→value map.
 * @returns resolved path with substituted params.
 */
export function resolvePath(route: DocPage, values: Record<string, string>): string {
	let path = route.path ?? '';
	for (const p of route.params ?? []) {
		if (p.in === 'path') {
			const v = (values[p.name] ?? '').trim();
			path = path.replace(`:${p.name}`, v ? encodeURIComponent(v) : `:${p.name}`);
		}
	}
	return path;
}

/**
 * Builds a human-readable "k=v&k=v" query string for snippets (not URL-encoded, for legibility).
 * @param route the current route.
 * @param values current name→value map.
 * @returns joined query string.
 */
function queryString(route: DocPage, values: Record<string, string>): string {
	return activeQueryParams(route, values).map(p => `${p.name}=${values[p.name]}`).join('&');
}

/**
 * Resolves the body to send for a non-GET route: the edited value if present, else the route's example.
 * @param route the current route.
 * @param body edited body text, if any.
 * @returns non-empty JSON body string.
 */
function resolveBody(route: DocPage, body?: string): string {
	const b = body ?? route.body ?? '{}';
	return b.trim() === '' ? '{}' : b;
}

/**
 * Builds a curl snippet reflecting the current param values and body.
 * @param route the current route.
 * @param baseURL API base URL.
 * @param values current name→value map.
 * @param body edited body text, if any.
 * @returns curl command string.
 */
export function buildCurl(route: DocPage, baseURL = 'https://api.poi.trippier.dev', values: Record<string, string> = {}, body?: string): string {
	const path = resolvePath(route, values);
	const query = queryString(route, values);
	const url = baseURL + path + (query && route.method === 'GET' ? `?${query}` : '');
	if (route.method === 'GET') {
		return `curl "${url}" \\\n  -H "X-API-Key: YOUR_API_KEY"`;
	}
	return `curl -X ${route.method} "${url}" \\\n  -H "X-API-Key: YOUR_API_KEY" \\\n  -H "Content-Type: application/json" \\\n  -d '${resolveBody(route, body).replace(/\n\s*/g, ' ')}'`;
}

/**
 * Builds a JavaScript fetch snippet reflecting the current param values and body.
 * @param route the current route.
 * @param baseURL API base URL.
 * @param values current name→value map.
 * @param body edited body text, if any.
 * @returns JavaScript fetch snippet string.
 */
export function buildJs(route: DocPage, baseURL = 'https://api.poi.trippier.dev', values: Record<string, string> = {}, body?: string): string {
	const path = resolvePath(route, values);
	const query = queryString(route, values);
	if (route.method === 'GET') {
		return `const res = await fetch(\n  '${baseURL}${path}${query ? `?${query}` : ''}',\n  { headers: { 'X-API-Key': process.env.TRIPPIER_API_KEY } }\n);\nconst data = await res.json();`;
	}
	return `const res = await fetch('${baseURL}${path}', {\n  method: '${route.method}',\n  headers: {\n    'X-API-Key': process.env.TRIPPIER_API_KEY,\n    'Content-Type': 'application/json',\n  },\n  body: JSON.stringify(${resolveBody(route, body)}),\n});\nconst data = await res.json();`;
}

/**
 * Builds a Python requests snippet reflecting the current param values and body.
 * @param route the current route.
 * @param baseURL API base URL.
 * @param values current name→value map.
 * @param body edited body text, if any.
 * @returns Python requests snippet string.
 */
export function buildPy(route: DocPage, baseURL = 'https://api.poi.trippier.dev', values: Record<string, string> = {}, body?: string): string {
	const path = resolvePath(route, values);
	if (route.method === 'GET') {
		const params = activeQueryParams(route, values)
			.map(p => `"${p.name}": ${isNaN(Number(values[p.name])) ? `"${values[p.name]}"` : values[p.name]}`)
			.join(', ');
		return `import requests, os\n\nres = requests.get(\n    "${baseURL}${path}",\n    params={${params}},\n    headers={"X-API-Key": os.environ["TRIPPIER_API_KEY"]},\n)\ndata = res.json()`;
	}
	const method = (route.method ?? 'POST').toLowerCase();
	return `import requests, os\n\nres = requests.${method}(\n    "${baseURL}${path}",\n    json=${resolveBody(route, body)},\n    headers={\n        "X-API-Key": os.environ["TRIPPIER_API_KEY"],\n        "Content-Type": "application/json",\n    },\n)\ndata = res.json()`;
}
