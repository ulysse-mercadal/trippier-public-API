export type Param = {
	name: string;
	type: string;
	required: boolean;
	desc: string;
	in?: 'path' | 'query';
};

export type DocPage = {
	id: string;
	group: string;
	title: string;
	kind: 'guide' | 'route';
	method?: string;
	path?: string;
	cost?: number;
	summary?: string;
	params?: Param[];
	body?: string;
	response?: string;
};

export const PAGES: DocPage[] = [
	{ id: 'quickstart', group: 'Démarrage', title: 'Quickstart',           kind: 'guide' },
	{ id: 'auth',       group: 'Démarrage', title: 'Authentification',     kind: 'guide' },
	{ id: 'errors',     group: 'Démarrage', title: 'Erreurs',              kind: 'guide' },
	{ id: 'ratelimit',  group: 'Démarrage', title: 'Rate-limit',           kind: 'guide' },
	{ id: 'tokens',     group: 'Démarrage', title: 'Tokens & facturation', kind: 'guide' },
	{
		id: 'poi-search', group: 'POI', title: 'Recherche de POIs',
		kind: 'route', method: 'GET', path: '/pois/search', cost: 1,
		summary: "Recherche multi-critères de points d'intérêt (restaurants, monuments, parcs…) par rayon, polygone ou quartier. Les résultats sont agrégés depuis plusieurs providers, dédoublonnés et scorés.",
		params: [
			{ name: 'mode',      type: 'enum',    required: false, desc: 'Stratégie géographique : radius (défaut), polygon, district.' },
			{ name: 'lat',       type: 'number',  required: false, desc: 'Latitude du centre (requis pour mode=radius).' },
			{ name: 'lng',       type: 'number',  required: false, desc: 'Longitude du centre (requis pour mode=radius).' },
			{ name: 'radius',    type: 'integer', required: false, desc: 'Rayon en mètres. Défaut : 1000.' },
			{ name: 'polygon',   type: 'string',  required: false, desc: 'Polygone encodé "lat1,lng1;lat2,lng2;…" (requis pour mode=polygon).' },
			{ name: 'district',  type: 'string',  required: false, desc: 'Nom du quartier/ville géocodé via Nominatim (requis pour mode=district).' },
			{ name: 'types',     type: 'string',  required: false, desc: 'Filtre par type CSV : see, eat, drink, do, buy, sleep, generic, event.' },
			{ name: 'weights',   type: 'string',  required: false, desc: 'Pondération par type JSON — mutuellement exclusif avec types. Ex : {"see":2,"eat":1}.' },
			{ name: 'providers', type: 'string',  required: false, desc: 'Filtre les providers sources (CSV). Voir GET /pois/providers.' },
			{ name: 'lang',      type: 'string',  required: false, desc: 'Langue des libellés (ISO 639-1). Défaut : fr.' },
			{ name: 'limit',     type: 'integer', required: false, desc: 'Nombre de résultats (1–100). Défaut : 20.' },
			{ name: 'offset',    type: 'integer', required: false, desc: 'Décalage pour la pagination. Défaut : 0.' },
			{ name: 'min_score', type: 'number',  required: false, desc: 'Score minimum (0–1). Filtre les POIs peu pertinents.' },
		],
		response: `{
  "query": {
    "mode": "radius",
    "lat": 45.76, "lng": 4.83,
    "radius": 1500,
    "lang": "fr", "limit": 20, "offset": 0
  },
  "total": 34,
  "results": [
    {
      "id": "osm:309832711",
      "name": "Cathédrale Saint-Jean-Baptiste",
      "type": "see",
      "score": 0.91,
      "coords": { "lat": 45.7606, "lng": 4.8278, "approximate": false },
      "distance": 142.5,
      "description": "Cathédrale gothique du XIIe siècle, siège de l'archevêché de Lyon.",
      "contact": { "website": "https://lyon.catholique.fr", "opening_hours": "Mo-Sa 08:00-18:00" },
      "thumbnail": "https://upload.wikimedia.org/wikipedia/commons/thumb/…/320px-…jpg",
      "sources": ["osm", "wikidata"]
    }
  ]
}`,
	},
	{
		id: 'poi-search-slim', group: 'POI', title: 'Recherche slim (carte)',
		kind: 'route', method: 'GET', path: '/pois/search/slim', cost: 1,
		summary: "Même paramètres que GET /pois/search mais retourne uniquement name, type et coords. Conçu pour les rendus cartographiques où le poids de la réponse importe.",
		params: [
			{ name: 'mode',     type: 'enum',    required: false, desc: 'radius (défaut), polygon, district.' },
			{ name: 'lat',      type: 'number',  required: false, desc: 'Latitude du centre.' },
			{ name: 'lng',      type: 'number',  required: false, desc: 'Longitude du centre.' },
			{ name: 'radius',   type: 'integer', required: false, desc: 'Rayon en mètres. Défaut : 1000.' },
			{ name: 'polygon',  type: 'string',  required: false, desc: 'Polygone encodé.' },
			{ name: 'district', type: 'string',  required: false, desc: 'Nom du quartier/ville.' },
			{ name: 'types',    type: 'string',  required: false, desc: 'Filtre par type CSV.' },
			{ name: 'limit',    type: 'integer', required: false, desc: 'Nombre de résultats. Défaut : 20.' },
			{ name: 'offset',   type: 'integer', required: false, desc: 'Décalage pour la pagination.' },
		],
		response: `{
  "total": 34,
  "results": [
    { "name": "Cathédrale Saint-Jean-Baptiste", "type": "see",   "coords": { "lat": 45.7606, "lng": 4.8278, "approximate": false } },
    { "name": "Brasserie des Confluences",       "type": "eat",   "coords": { "lat": 45.7512, "lng": 4.8192, "approximate": false } },
    { "name": "Parc de la Tête d'Or",            "type": "do",    "coords": { "lat": 45.7772, "lng": 4.8563, "approximate": false } }
  ]
}`,
	},
	{
		id: 'poi-events', group: 'POI', title: 'Recherche d\'évènements',
		kind: 'route', method: 'GET', path: '/pois/events', cost: 10,
		summary: "Évènements en temps réel (concerts, expos, marchés…) depuis Ticketmaster et Eventbrite. Coût plus élevé car l'appel fan-out vers des providers à quota limité.",
		params: [
			{ name: 'mode',     type: 'enum',    required: false, desc: 'radius (défaut), polygon, district.' },
			{ name: 'lat',      type: 'number',  required: false, desc: 'Latitude du centre.' },
			{ name: 'lng',      type: 'number',  required: false, desc: 'Longitude du centre.' },
			{ name: 'radius',   type: 'integer', required: false, desc: 'Rayon en mètres. Défaut : 1000.' },
			{ name: 'polygon',  type: 'string',  required: false, desc: 'Polygone encodé.' },
			{ name: 'district', type: 'string',  required: false, desc: 'Nom du quartier/ville.' },
			{ name: 'limit',    type: 'integer', required: false, desc: 'Nombre de résultats. Défaut : 20.' },
			{ name: 'offset',   type: 'integer', required: false, desc: 'Décalage pour la pagination.' },
		],
		response: `{
  "query": { "mode": "radius", "lat": 45.76, "lng": 4.83, "radius": 5000 },
  "total": 8,
  "results": [
    {
      "id": "ticketmaster:G5v0Z9EALtXdm",
      "name": "Nuits de Fourvière — Tinariwen",
      "type": "event",
      "score": 0.87,
      "coords": { "lat": 45.7602, "lng": 4.8222, "approximate": false },
      "distance": 680.2,
      "contact": { "website": "https://www.nuitsdefourviere.com" },
      "date_start": "2026-06-18T21:00:00Z",
      "date_end":   "2026-06-18T23:00:00Z",
      "sources": ["ticketmaster"]
    }
  ]
}`,
	},
	{
		id: 'poi-events-slim', group: 'POI', title: 'Évènements slim (carte)',
		kind: 'route', method: 'GET', path: '/pois/events/slim', cost: 10,
		summary: "Version allégée de GET /pois/events : retourne uniquement name, coords et dates. Même coût car le fan-out provider est identique.",
		params: [
			{ name: 'mode',     type: 'enum',    required: false, desc: 'radius (défaut), polygon, district.' },
			{ name: 'lat',      type: 'number',  required: false, desc: 'Latitude du centre.' },
			{ name: 'lng',      type: 'number',  required: false, desc: 'Longitude du centre.' },
			{ name: 'radius',   type: 'integer', required: false, desc: 'Rayon en mètres. Défaut : 1000.' },
			{ name: 'limit',    type: 'integer', required: false, desc: 'Nombre de résultats. Défaut : 20.' },
		],
		response: `{
  "total": 8,
  "results": [
    {
      "name": "Nuits de Fourvière — Tinariwen",
      "coords": { "lat": 45.7602, "lng": 4.8222, "approximate": false },
      "date_start": "2026-06-18T21:00:00Z",
      "date_end": "2026-06-18T23:00:00Z"
    }
  ]
}`,
	},
	{
		id: 'poi-providers', group: 'POI', title: 'Lister les providers',
		kind: 'route', method: 'GET', path: '/pois/providers', cost: 0,
		summary: "Liste tous les providers de données actifs. Aucun token consommé. Utile pour construire un filtre providers dans vos requêtes de recherche.",
		response: `[
  "overpass",
  "wikivoyage",
  "wikipedia",
  "geonames",
  "ticketmaster",
  "eventbrite"
]`,
	},
	{
		id: 'health', group: 'Utilitaires', title: 'Health check',
		kind: 'route', method: 'GET', path: '/health', cost: 0,
		summary: "Endpoint de santé. Aucun token consommé, aucune authentification requise. Retourne 200 si l'API est opérationnelle.",
		response: `{ "status": "ok" }`,
	},
];
