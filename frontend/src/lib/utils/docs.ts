import type { DocPage, Param } from '$lib/data/docs';

// Maps HTTP method to its accent color
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

// Returns a representative sample value for a given parameter
export function sampleValue(p: Param): string {
	return SAMPLES[p.name] ?? 'value';
}

const SNIPPET_KEYS = new Set(['lat', 'lng', 'radius', 'date']);

// Params included in snippet query strings — only the primary geo params + date
function snippetParams(route: DocPage): Param[] {
	return (route.params ?? []).filter(p => p.in !== 'path' && SNIPPET_KEYS.has(p.name));
}

function buildQueryString(route: DocPage): string {
	return snippetParams(route).map(p => `${p.name}=${sampleValue(p)}`).join('&');
}

// Builds a curl snippet for a route
export function buildCurl(route: DocPage, baseURL = 'https://api.poi.trippier.dev'): string {
	const path = route.path ?? '';
	const query = buildQueryString(route);
	const url = baseURL + path + (query && route.method === 'GET' ? `?${query}` : '');
	if (route.method === 'GET') {
		return `curl "${url}" \\\n  -H "X-API-Key: YOUR_API_KEY"`;
	}
	return `curl -X ${route.method} "${url}" \\\n  -H "X-API-Key: YOUR_API_KEY" \\\n  -H "Content-Type: application/json" \\\n  -d '${(route.body ?? '{}').replace(/\n\s*/g, ' ')}'`;
}

// Builds a JavaScript fetch snippet for a route
export function buildJs(route: DocPage, baseURL = 'https://api.poi.trippier.dev'): string {
	const path = route.path ?? '';
	const query = buildQueryString(route);
	if (route.method === 'GET') {
		return `const res = await fetch(\n  '${baseURL}${path}${query ? `?${query}` : ''}',\n  { headers: { 'X-API-Key': process.env.TRIPPIER_API_KEY } }\n);\nconst data = await res.json();`;
	}
	return `const res = await fetch('${baseURL}${path}', {\n  method: 'POST',\n  headers: {\n    'X-API-Key': process.env.TRIPPIER_API_KEY,\n    'Content-Type': 'application/json',\n  },\n  body: JSON.stringify(${route.body ?? '{}'}),\n});\nconst data = await res.json();`;
}

// Builds a Python requests snippet for a route
export function buildPy(route: DocPage, baseURL = 'https://api.poi.trippier.dev'): string {
	const path = route.path ?? '';
	if (route.method === 'GET') {
		const params = snippetParams(route).map(p => `"${p.name}": ${isNaN(Number(sampleValue(p))) ? `"${sampleValue(p)}"` : sampleValue(p)}`).join(', ');
		return `import requests, os\n\nres = requests.get(\n    "${baseURL}${path}",\n    params={${params}},\n    headers={"X-API-Key": os.environ["TRIPPIER_API_KEY"]},\n)\ndata = res.json()`;
	}
	return `import requests, os\n\nres = requests.post(\n    "${baseURL}${path}",\n    json=${route.body ?? '{}'},\n    headers={\n        "X-API-Key": os.environ["TRIPPIER_API_KEY"],\n        "Content-Type": "application/json",\n    },\n)\ndata = res.json()`;
}
