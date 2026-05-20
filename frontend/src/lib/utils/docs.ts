import type { DocPage, Param } from '$lib/data/docs';

// Maps HTTP method to its accent color
export function methodColor(m: string): string {
	return ({ GET: 'var(--accent)', POST: '#e4b07a', DELETE: '#e78f8f', PUT: '#c89be0' } as Record<string, string>)[m] ?? 'var(--text)';
}

// Returns a representative sample value for a given parameter
export function sampleValue(p: Param): string {
	const samples: Record<string, string> = {
		lat: '45.76', lng: '4.83', radius: '1500',
		mode: 'radius', types: 'see,eat', lang: 'fr',
		limit: '20', offset: '0', min_score: '0.3',
		district: 'Vieux-Lyon', polygon: '45.76,4.83;45.77,4.84;45.76,4.85',
		providers: 'osm,wikidata',
	};
	return samples[p.name] ?? 'value';
}

// Builds a curl snippet for a route
export function buildCurl(route: DocPage, baseURL = 'https://api.trippier.dev'): string {
	const query = (route.params ?? [])
		.filter(p => p.in !== 'path' && ['lat','lng','radius'].includes(p.name))
		.map(p => `${p.name}=${sampleValue(p)}`)
		.join('&');
	const path = route.path ?? '';
	const url  = baseURL + path + (query && route.method === 'GET' ? `?${query}` : '');
	if (route.method === 'GET') {
		return `curl "${url}" \\\n  -H "X-API-Key: YOUR_API_KEY"`;
	}
	return `curl -X ${route.method} "${url}" \\\n  -H "X-API-Key: YOUR_API_KEY" \\\n  -H "Content-Type: application/json" \\\n  -d '${(route.body ?? '{}').replace(/\n\s*/g, ' ')}'`;
}

// Builds a JavaScript fetch snippet for a route
export function buildJs(route: DocPage): string {
	const path = route.path ?? '';
	if (route.method === 'GET') {
		return `const res = await fetch(\n  'https://api.trippier.dev${path}?lat=45.76&lng=4.83&radius=1500',\n  { headers: { 'X-API-Key': process.env.TRIPPIER_API_KEY } }\n);\nconst data = await res.json();`;
	}
	return `const res = await fetch('https://api.trippier.dev${path}', {\n  method: 'POST',\n  headers: {\n    'X-API-Key': process.env.TRIPPIER_API_KEY,\n    'Content-Type': 'application/json',\n  },\n  body: JSON.stringify(${route.body ?? '{}'}),\n});\nconst data = await res.json();`;
}

// Builds a Python requests snippet for a route
export function buildPy(route: DocPage): string {
	const path = route.path ?? '';
	if (route.method === 'GET') {
		return `import requests, os\n\nres = requests.get(\n    "https://api.trippier.dev${path}",\n    params={"lat": 45.76, "lng": 4.83, "radius": 1500},\n    headers={"X-API-Key": os.environ["TRIPPIER_API_KEY"]},\n)\ndata = res.json()`;
	}
	return `import requests, os\n\nres = requests.post(\n    "https://api.trippier.dev${path}",\n    json=${route.body ?? '{}'},\n    headers={\n        "X-API-Key": os.environ["TRIPPIER_API_KEY"],\n        "Content-Type": "application/json",\n    },\n)\ndata = res.json()`;
}
