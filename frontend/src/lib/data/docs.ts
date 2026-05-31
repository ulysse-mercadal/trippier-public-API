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
			{ name: 'date',     type: 'string',  required: false, desc: 'Date de début au format YYYY-MM-DD. Retourne les events commençant après 00h00 ce jour-là. Défaut : aujourd\'hui.' },
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
			{ name: 'date',     type: 'string',  required: false, desc: 'Date de début au format YYYY-MM-DD. Défaut : aujourd\'hui.' },
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
		id: 'itinerary', group: 'Itinéraire', title: 'Générer un itinéraire',
		kind: 'route', method: 'POST', path: '/itinerary/generate', cost: 50,
		summary: "Génère un itinéraire optimisé sur N jours à partir d'une zone géographique et de préférences. L'IA sélectionne les POIs, ordonne les étapes selon les horaires d'ouverture, le transport et le rythme souhaité.",
		params: [
			{ name: 'days',         type: 'integer', required: true,  desc: 'Nombre de jours (1–7).' },
			{ name: 'poi_query',    type: 'object',  required: true,  desc: 'Zone de recherche : { lat, lng, radius }.' },
			{ name: 'preferences',  type: 'object',  required: false, desc: 'Préférences : { pace: slow|moderate|fast, start_time: "HH:MM" }.' },
		],
		body: `{
  "days": 1,
  "poi_query": { "lat": 45.76, "lng": 4.83, "radius": 5000 },
  "preferences": { "pace": "moderate", "start_time": "09:00" }
}`,
		response: `{
  "days": [
    {
      "day": 1,
      "steps": [
        {
          "poi": { "id": "osm:309832711", "name": "Cathédrale Saint-Jean-Baptiste", "type": "see" },
          "arrival": "09:15", "duration_min": 60
        },
        {
          "poi": { "id": "osm:118820560", "name": "Brasserie Georges", "type": "eat" },
          "arrival": "12:30", "duration_min": 75
        }
      ]
    }
  ]
}`,
	},
	{
		id: 'poi-search-custom', group: 'POI · Custom', title: 'Recherche personnalisée',
		kind: 'route', method: 'GET', path: '/pois/search/custom', cost: 1,
		summary: "Contrôle total sur la sélection des providers. Choisissez vos providers explicitement, pondérez-les, excluez-en certains, ou laissez la sélection géo-intelligente fonctionner avec un country_hint. Supporte les clés BYOK par header.",
		params: [
			{ name: 'mode',             type: 'enum',    required: false, desc: 'radius (défaut), polygon, district.' },
			{ name: 'lat',              type: 'number',  required: false, desc: 'Latitude du centre.' },
			{ name: 'lng',              type: 'number',  required: false, desc: 'Longitude du centre.' },
			{ name: 'radius',           type: 'integer', required: false, desc: 'Rayon en mètres.' },
			{ name: 'providers',        type: 'string[]',required: false, desc: 'Liste de providers à activer (répéter le param). Ex : providers=overpass&providers=foursquare.' },
			{ name: 'exclude_providers',type: 'string[]',required: false, desc: 'Providers à exclure du résultat (répéter le param).' },
			{ name: 'provider_weights', type: 'JSON',    required: false, desc: 'Pondération de confiance par provider [0,1]. Ex : {"overpass":0.8,"foursquare":0.5}.' },
			{ name: 'country_hint',     type: 'string',  required: false, desc: 'Code pays ISO 3166-1 α-2 (ex : CN, JP, KR). Force la sélection géo-intelligente sur ce pays sans appel Nominatim.' },
			{ name: 'types',            type: 'string',  required: false, desc: 'Filtre par type CSV.' },
			{ name: 'limit',            type: 'integer', required: false, desc: 'Nombre de résultats. Défaut : 20.' },
		],
		response: `{
  "query": { "mode": "radius", "lat": 35.68, "lng": 139.69, "radius": 2000, "country_hint": "JP" },
  "total": 12,
  "results": [
    {
      "id": "osm:123456",
      "name": "Senso-ji Temple",
      "type": "see",
      "score": 0.94,
      "coords": { "lat": 35.7148, "lng": 139.7967, "approximate": false },
      "sources": ["overpass"]
    }
  ]
}`,
	},
	{
		id: 'poi-search-custom-slim', group: 'POI · Custom', title: 'Recherche personnalisée slim',
		kind: 'route', method: 'GET', path: '/pois/search/custom/slim', cost: 1,
		summary: "Version allégée de GET /pois/search/custom — même contrôle provider, réponse réduite à name, type et coords. Idéal pour les cartes.",
		params: [
			{ name: 'mode',             type: 'enum',    required: false, desc: 'radius (défaut), polygon, district.' },
			{ name: 'lat',              type: 'number',  required: false, desc: 'Latitude du centre.' },
			{ name: 'lng',              type: 'number',  required: false, desc: 'Longitude du centre.' },
			{ name: 'radius',           type: 'integer', required: false, desc: 'Rayon en mètres.' },
			{ name: 'providers',        type: 'string[]',required: false, desc: 'Providers à activer (répéter le param).' },
			{ name: 'exclude_providers',type: 'string[]',required: false, desc: 'Providers à exclure.' },
			{ name: 'country_hint',     type: 'string',  required: false, desc: 'Code pays ISO 3166-1 α-2.' },
			{ name: 'limit',            type: 'integer', required: false, desc: 'Nombre de résultats. Défaut : 20.' },
		],
		response: `{
  "total": 12,
  "results": [
    { "name": "Senso-ji Temple",   "type": "see", "coords": { "lat": 35.7148, "lng": 139.7967 } },
    { "name": "Nakamise-dori",     "type": "buy", "coords": { "lat": 35.7144, "lng": 139.7965 } }
  ]
}`,
	},
	{
		id: 'poi-events-custom', group: 'POI · Custom', title: 'Évènements personnalisés',
		kind: 'route', method: 'GET', path: '/pois/events/custom', cost: 10,
		summary: "Contrôle total sur les providers d'évènements. Choisissez quels providers interroger (Ticketmaster, Eventbrite, Meetup, OpenAgenda…), passez vos clés BYOK, et pondérez les résultats par provider.",
		params: [
			{ name: 'mode',              type: 'enum',    required: false, desc: 'radius (défaut), polygon, district.' },
			{ name: 'lat',               type: 'number',  required: false, desc: 'Latitude du centre.' },
			{ name: 'lng',               type: 'number',  required: false, desc: 'Longitude du centre.' },
			{ name: 'radius',            type: 'integer', required: false, desc: 'Rayon en mètres (min 50 000 pour les évènements).' },
			{ name: 'providers',         type: 'string[]',required: false, desc: 'Providers à activer (répéter le param). Ex : providers=ticketmaster&providers=meetup.' },
			{ name: 'exclude_providers', type: 'string[]',required: false, desc: 'Providers à exclure.' },
			{ name: 'provider_weights',  type: 'JSON',    required: false, desc: 'Pondération par provider [0,1].' },
			{ name: 'date',              type: 'string',  required: false, desc: 'Date de début YYYY-MM-DD.' },
			{ name: 'limit',             type: 'integer', required: false, desc: 'Nombre de résultats. Défaut : 20.' },
		],
		response: `{
  "query": { "mode": "radius", "lat": 48.85, "lng": 2.35, "radius": 50000 },
  "total": 5,
  "results": [
    {
      "id": "ticketmaster:G5v0Z9EA",
      "name": "Paris Jazz Festival",
      "type": "event",
      "date_start": "2026-06-20T19:00:00Z",
      "sources": ["ticketmaster"]
    }
  ]
}`,
	},
	{
		id: 'poi-events-custom-slim', group: 'POI · Custom', title: 'Évènements personnalisés slim',
		kind: 'route', method: 'GET', path: '/pois/events/custom/slim', cost: 10,
		summary: "Version allégée de GET /pois/events/custom — réponse réduite à name, coords et dates.",
		params: [
			{ name: 'mode',      type: 'enum',    required: false, desc: 'radius (défaut), polygon, district.' },
			{ name: 'lat',       type: 'number',  required: false, desc: 'Latitude du centre.' },
			{ name: 'lng',       type: 'number',  required: false, desc: 'Longitude du centre.' },
			{ name: 'radius',    type: 'integer', required: false, desc: 'Rayon en mètres.' },
			{ name: 'providers', type: 'string[]',required: false, desc: 'Providers à activer (répéter le param).' },
			{ name: 'date',      type: 'string',  required: false, desc: 'Date de début YYYY-MM-DD.' },
			{ name: 'limit',     type: 'integer', required: false, desc: 'Nombre de résultats. Défaut : 20.' },
		],
		response: `{
  "total": 5,
  "results": [
    { "name": "Paris Jazz Festival", "coords": { "lat": 48.86, "lng": 2.33 }, "date_start": "2026-06-20T19:00:00Z" }
  ]
}`,
	},
	{
		id: 'poi-providers-catalog', group: 'POI · Custom', title: 'Catalogue des providers',
		kind: 'route', method: 'GET', path: '/pois/providers/catalog', cost: 0,
		summary: "Liste tous les providers connus du registre (implémentés ou non), avec leurs métadonnées : catégorie, BYOK, header HTTP attendu, spécialisation géographique. Utile pour construire un sélecteur de providers dans votre UI.",
		response: `[
  { "id": "overpass",      "label": "OpenStreetMap / Overpass",  "byok": false, "for_events": false, "implemented": true  },
  { "id": "ticketmaster",  "label": "Ticketmaster",              "byok": true,  "for_events": true,  "implemented": true,
    "byok_header": "X-Ticketmaster-Key"  },
  { "id": "foursquare",    "label": "Foursquare / FSQ",          "byok": true,  "for_events": false, "implemented": false,
    "byok_header": "X-Foursquare-Key"   },
  { "id": "baidu",         "label": "Baidu Maps",                "byok": true,  "for_events": false, "implemented": false,
    "byok_header": "X-Baidu-Key"        }
]`,
	},
	{
		id: 'poi-providers-recommend', group: 'POI · Custom', title: 'Recommandation de providers',
		kind: 'route', method: 'GET', path: '/pois/providers/recommend', cost: 0,
		summary: "Retourne les providers les plus adaptés à une position géographique et à des catégories de POIs, triés par score de confiance. La sélection est géo-intelligente : Baidu pour la Chine, Kakao pour la Corée, Navitime pour le Japon, etc.",
		params: [
			{ name: 'lat',        type: 'number',  required: true,  desc: 'Latitude.' },
			{ name: 'lng',        type: 'number',  required: true,  desc: 'Longitude.' },
			{ name: 'for_events', type: 'boolean', required: false, desc: 'true pour n\'afficher que les providers d\'évènements.' },
			{ name: 'types',      type: 'string',  required: false, desc: 'Types CSV pour affiner le scoring par catégorie.' },
			{ name: 'limit',      type: 'integer', required: false, desc: 'Nombre max de providers retournés. Défaut : 10.' },
		],
		response: `{
  "country_code": "JP",
  "providers": [
    { "id": "navitime",  "label": "Navitime",              "score": 0.95, "byok": true, "for_events": false },
    { "id": "overpass",  "label": "OpenStreetMap / Overpass","score": 0.78, "byok": false,"for_events": false },
    { "id": "geonames",  "label": "GeoNames",              "score": 0.65, "byok": false,"for_events": false }
  ]
}`,
	},
	{
		id: 'poi-providers', group: 'POI', title: 'Lister les providers actifs',
		kind: 'route', method: 'GET', path: '/pois/providers', cost: 0,
		summary: "Liste les providers actuellement actifs (instances enregistrées au démarrage). Pour le catalogue complet incluant providers non implémentés et métadonnées BYOK, voir GET /pois/providers/catalog.",
		response: `[
  "overpass",
  "wikivoyage",
  "wikipedia_events",
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
