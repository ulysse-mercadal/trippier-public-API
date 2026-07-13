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
		kind: 'route', method: 'GET', path: '/v1/pois/search', cost: 1,
		summary: "Recherche multi-critères de points d'intérêt (restaurants, monuments, parcs…) par rayon, polygone ou quartier. Les résultats sont agrégés depuis plusieurs providers, dédoublonnés et scorés.",
		params: [
			{ name: 'mode',      type: 'enum',    required: false, desc: 'Stratégie géographique : radius (défaut), polygon, district.' },
			{ name: 'lat',       type: 'number',  required: false, desc: 'Latitude du centre (requis pour mode=radius). Domaine [-90, 90].' },
			{ name: 'lng',       type: 'number',  required: false, desc: 'Longitude du centre (requis pour mode=radius). Domaine [-180, 180].' },
			{ name: 'radius',    type: 'integer', required: false, desc: 'Rayon en mètres (max 50 000). Défaut : 5000.' },
			{ name: 'polygon',   type: 'string',  required: false, desc: 'Polygone encodé "lat1 lng1 lat2 lng2 …" (3 à 100 points, requis pour mode=polygon).' },
			{ name: 'district',  type: 'string',  required: false, desc: 'Nom du quartier/ville géocodé via Nominatim (requis pour mode=district).' },
			{ name: 'types',     type: 'string',  required: false, desc: 'Filtre par type CSV : see, eat, drink, do, buy, sleep, generic, event.' },
			{ name: 'weights',   type: 'string',  required: false, desc: 'Pondération par type JSON, valeurs ∈ [0,1] — mutuellement exclusif avec types. Ex : {"see":1,"eat":0.5}.' },
			{ name: 'providers', type: 'string',  required: false, desc: 'Filtre les providers sources (CSV). Voir GET /v1/pois/providers.' },
			{ name: 'lang',      type: 'string',  required: false, desc: 'Langue des libellés (ISO 639-1). Défaut : en.' },
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
      "id": "overpass:309832711",
      "name": "Cathédrale Saint-Jean-Baptiste",
      "kind": "poi",
      "type": "see",
      "score": 0.91,
      "coords": { "lat": 45.7606, "lng": 4.8278, "approximate": false },
      "distance": 142.5,
      "description": "Cathédrale gothique du XIIe siècle, siège de l'archevêché de Lyon.",
      "contact": { "website": "https://lyon.catholique.fr", "opening_hours": "Mo-Sa 08:00-18:00" },
      "thumbnail": "https://upload.wikimedia.org/wikipedia/commons/thumb/…/320px-…jpg",
      "sources": [
        { "provider": "overpass",  "url": "https://www.openstreetmap.org/node/309832711" },
        { "provider": "wikipedia", "url": "https://fr.wikipedia.org/wiki/Cathédrale_Saint-Jean_de_Lyon" }
      ]
    }
  ]
}`,
	},
	{
		id: 'poi-search-slim', group: 'POI', title: 'Recherche slim (carte)',
		kind: 'route', method: 'GET', path: '/v1/pois/search/slim', cost: 1,
		summary: "Mêmes paramètres que GET /v1/pois/search mais retourne uniquement name, type et coords. Conçu pour les rendus cartographiques où le poids de la réponse importe.",
		params: [
			{ name: 'mode',     type: 'enum',    required: false, desc: 'radius (défaut), polygon, district.' },
			{ name: 'lat',      type: 'number',  required: false, desc: 'Latitude du centre.' },
			{ name: 'lng',      type: 'number',  required: false, desc: 'Longitude du centre.' },
			{ name: 'radius',   type: 'integer', required: false, desc: 'Rayon en mètres (max 50 000). Défaut : 5000.' },
			{ name: 'polygon',  type: 'string',  required: false, desc: 'Polygone encodé.' },
			{ name: 'district', type: 'string',  required: false, desc: 'Nom du quartier/ville.' },
			{ name: 'types',    type: 'string',  required: false, desc: 'Filtre par type CSV.' },
			{ name: 'limit',    type: 'integer', required: false, desc: 'Nombre de résultats (1–100). Défaut : 20.' },
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
		kind: 'route', method: 'GET', path: '/v1/pois/events', cost: 10,
		summary: "Évènements (concerts, expos, marchés…) agrégés depuis Ticketmaster, Eventbrite et Wikipedia Events. Coût élevé car l'appel fan-out frappe des providers à quota limité ; certains providers ont un rayon minimum qui clampe automatiquement votre requête.",
		params: [
			{ name: 'mode',     type: 'enum',    required: false, desc: 'radius (défaut), polygon, district.' },
			{ name: 'lat',      type: 'number',  required: false, desc: 'Latitude du centre.' },
			{ name: 'lng',      type: 'number',  required: false, desc: 'Longitude du centre.' },
			{ name: 'radius',   type: 'integer', required: false, desc: 'Rayon en mètres (max 50 000). Défaut : 5000. Étiré automatiquement au minimum du provider (ex : Ticketmaster impose 50 000 m).' },
			{ name: 'polygon',  type: 'string',  required: false, desc: 'Polygone encodé.' },
			{ name: 'district', type: 'string',  required: false, desc: 'Nom du quartier/ville.' },
			{ name: 'date',     type: 'string',  required: false, desc: 'Date de début au format YYYY-MM-DD. Retourne les events commençant après 00h00 ce jour-là. Défaut : aujourd\'hui.' },
			{ name: 'limit',    type: 'integer', required: false, desc: 'Nombre de résultats (1–100). Défaut : 20.' },
			{ name: 'offset',   type: 'integer', required: false, desc: 'Décalage pour la pagination.' },
		],
		response: `{
  "query": { "mode": "radius", "lat": 45.76, "lng": 4.83, "radius": 50000 },
  "total": 8,
  "results": [
    {
      "id": "ticketmaster:G5v0Z9EALtXdm",
      "name": "Nuits de Fourvière — Tinariwen",
      "kind": "event",
      "type": "event",
      "score": 0.87,
      "coords": { "lat": 45.7602, "lng": 4.8222, "approximate": false },
      "distance": 680.2,
      "contact": { "website": "https://www.nuitsdefourviere.com" },
      "date_start": "2026-06-18T21:00:00Z",
      "date_end":   "2026-06-18T23:00:00Z",
      "recurring": false,
      "sources": [
        { "provider": "ticketmaster", "url": "https://www.ticketmaster.fr/event/G5v0Z9EALtXdm" }
      ]
    }
  ]
}`,
	},
	{
		id: 'poi-events-slim', group: 'POI', title: 'Évènements slim (carte)',
		kind: 'route', method: 'GET', path: '/v1/pois/events/slim', cost: 10,
		summary: "Version allégée de GET /v1/pois/events : retourne uniquement name, coords et dates. Même coût car le fan-out provider est identique.",
		params: [
			{ name: 'mode',     type: 'enum',    required: false, desc: 'radius (défaut), polygon, district.' },
			{ name: 'lat',      type: 'number',  required: false, desc: 'Latitude du centre.' },
			{ name: 'lng',      type: 'number',  required: false, desc: 'Longitude du centre.' },
			{ name: 'radius',   type: 'integer', required: false, desc: 'Rayon en mètres (max 50 000). Défaut : 5000.' },
			{ name: 'date',     type: 'string',  required: false, desc: 'Date de début au format YYYY-MM-DD. Défaut : aujourd\'hui.' },
			{ name: 'limit',    type: 'integer', required: false, desc: 'Nombre de résultats (1–100). Défaut : 20.' },
		],
		response: `{
  "total": 8,
  "results": [
    {
      "name": "Nuits de Fourvière — Tinariwen",
      "coords": { "lat": 45.7602, "lng": 4.8222, "approximate": false },
      "date_start": "2026-06-18T21:00:00Z",
      "date_end": "2026-06-18T23:00:00Z",
      "recurring": false
    }
  ]
}`,
	},
	{
		id: 'itinerary', group: 'Itinéraire', title: 'Générer un itinéraire',
		kind: 'route', method: 'POST', path: '/v1/itinerary/generate', cost: 50,
		summary: "Génère un itinéraire optimisé sur N jours. Fournissez soit une liste de POIs déjà sélectionnés (pois), soit une zone à explorer (poi_query) que le service délègue à GET /v1/pois/search. Note : ce service est exposé sur un sous-domaine distinct (api.ai.* en production).",
		params: [
			{ name: 'pois',           type: 'Poi[]',       required: false, desc: 'Liste de POIs déjà choisis. Schéma : { id, name, type, coords?, description?, distance? }. Mutuellement complémentaire avec poi_query — au moins l\'un des deux est obligatoire.' },
			{ name: 'poi_query',      type: 'PoiQuery',    required: false, desc: 'Requête forwardée à GET /v1/pois/search si pois est absent. Champs : mode, lat, lng, radius (≤50 000), polygon, district, types, lang, limit (1–400), offset, min_score.' },
			{ name: 'days',           type: 'integer',     required: false, desc: 'Nombre de jours (1–30). Défaut : 1.' },
			{ name: 'start_location', type: 'Coordinates', required: false, desc: 'Position de départ pour l\'optimisation : { lat, lng, approximate? }.' },
			{ name: 'preferences',    type: 'Preferences', required: false, desc: 'Contraintes : { pace: relaxed|moderate|intensive, priorities: PoiType[], avoid: PoiType[], start_time: "HH:MM", end_time: "HH:MM" }. Défaut : pace=moderate, start_time=09:00, end_time=21:00.' },
		],
		body: `{
  "poi_query": {
    "mode": "radius",
    "lat": 45.76, "lng": 4.83,
    "radius": 5000,
    "types": ["see", "eat"],
    "limit": 60
  },
  "days": 2,
  "start_location": { "lat": 45.7610, "lng": 4.8320 },
  "preferences": {
    "pace": "moderate",
    "priorities": ["see"],
    "avoid": ["sleep"],
    "start_time": "09:00",
    "end_time": "21:00"
  }
}`,
		response: `{
  "days": [
    {
      "day": 1,
      "pois": [
        { "id": "overpass:309832711", "name": "Cathédrale Saint-Jean-Baptiste", "type": "see",
          "coords": { "lat": 45.7606, "lng": 4.8278, "approximate": false } },
        { "id": "overpass:118820560", "name": "Brasserie Georges",              "type": "eat",
          "coords": { "lat": 45.7512, "lng": 4.8192, "approximate": false } }
      ],
      "estimated_duration_hours": 6.5,
      "description": null
    },
    {
      "day": 2,
      "pois": [
        { "id": "overpass:412034521", "name": "Parc de la Tête d'Or", "type": "do",
          "coords": { "lat": 45.7772, "lng": 4.8563, "approximate": false } }
      ],
      "estimated_duration_hours": 4.0,
      "description": null
    }
  ],
  "total_pois": 3,
  "summary": null
}`,
	},
	{
		id: 'poi-search-custom', group: 'POI · Custom', title: 'Recherche personnalisée',
		kind: 'route', method: 'GET', path: '/v1/pois/search/custom', cost: 1,
		summary: "Contrôle total sur la sélection des providers. Choisissez vos providers explicitement, pondérez-les, excluez-en certains, ou laissez la sélection géo-intelligente fonctionner avec un country_hint. Supporte les clés BYOK par header (Foursquare, HERE, Baidu, Amap, Kakao, Navitime, Mappls, GrabMaps).",
		params: [
			{ name: 'mode',             type: 'enum',    required: false, desc: 'radius (défaut), polygon, district.' },
			{ name: 'lat',              type: 'number',  required: false, desc: 'Latitude du centre.' },
			{ name: 'lng',              type: 'number',  required: false, desc: 'Longitude du centre.' },
			{ name: 'radius',           type: 'integer', required: false, desc: 'Rayon en mètres (max 50 000). Défaut : 5000.' },
			{ name: 'providers',        type: 'string',  required: false, desc: 'Liste de providers à activer (CSV). Ex : providers=overpass,foursquare.' },
			{ name: 'exclude_providers',type: 'string',  required: false, desc: 'Providers à exclure du résultat (CSV).' },
			{ name: 'provider_weights', type: 'JSON',    required: false, desc: 'Pondération de confiance par provider, valeurs ∈ [0,1]. Ex : {"overpass":0.8,"foursquare":0.5}. Un score effectif < 0,1 exclut le provider de la sélection auto.' },
			{ name: 'country_hint',     type: 'string',  required: false, desc: 'Code pays ISO 3166-1 α-2 (ex : CN, JP, KR). Force la sélection géo-intelligente sur ce pays sans appel Nominatim.' },
			{ name: 'types',            type: 'string',  required: false, desc: 'Filtre par type CSV.' },
			{ name: 'limit',            type: 'integer', required: false, desc: 'Nombre de résultats (1–100). Défaut : 20.' },
		],
		response: `{
  "query": { "mode": "radius", "lat": 35.68, "lng": 139.69, "radius": 2000, "country_hint": "JP" },
  "total": 12,
  "results": [
    {
      "id": "overpass:123456",
      "name": "Senso-ji Temple",
      "kind": "poi",
      "type": "see",
      "score": 0.94,
      "coords": { "lat": 35.7148, "lng": 139.7967, "approximate": false },
      "sources": [
        { "provider": "overpass", "url": "https://www.openstreetmap.org/node/123456" }
      ]
    }
  ]
}`,
	},
	{
		id: 'poi-search-custom-slim', group: 'POI · Custom', title: 'Recherche personnalisée slim',
		kind: 'route', method: 'GET', path: '/v1/pois/search/custom/slim', cost: 1,
		summary: "Version allégée de GET /v1/pois/search/custom — même contrôle provider, réponse réduite à name, type et coords. Idéal pour les cartes.",
		params: [
			{ name: 'mode',             type: 'enum',    required: false, desc: 'radius (défaut), polygon, district.' },
			{ name: 'lat',              type: 'number',  required: false, desc: 'Latitude du centre.' },
			{ name: 'lng',              type: 'number',  required: false, desc: 'Longitude du centre.' },
			{ name: 'radius',           type: 'integer', required: false, desc: 'Rayon en mètres (max 50 000). Défaut : 5000.' },
			{ name: 'providers',        type: 'string',  required: false, desc: 'Providers à activer (CSV).' },
			{ name: 'exclude_providers',type: 'string',  required: false, desc: 'Providers à exclure (CSV).' },
			{ name: 'country_hint',     type: 'string',  required: false, desc: 'Code pays ISO 3166-1 α-2.' },
			{ name: 'limit',            type: 'integer', required: false, desc: 'Nombre de résultats (1–100). Défaut : 20.' },
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
		kind: 'route', method: 'GET', path: '/v1/pois/events/custom', cost: 10,
		summary: "Contrôle total sur les providers d'évènements. Choisissez quels providers interroger (Ticketmaster, Eventbrite, Meetup, OpenAgenda, Wikipedia Events…), passez vos clés BYOK via X-Ticketmaster-Key, X-Eventbrite-Token, X-Meetup-Token, X-OpenAgenda-Key, et pondérez les résultats par provider.",
		params: [
			{ name: 'mode',              type: 'enum',    required: false, desc: 'radius (défaut), polygon, district.' },
			{ name: 'lat',               type: 'number',  required: false, desc: 'Latitude du centre.' },
			{ name: 'lng',               type: 'number',  required: false, desc: 'Longitude du centre.' },
			{ name: 'radius',            type: 'integer', required: false, desc: 'Rayon en mètres (max 50 000). Étiré au minimum du provider (Ticketmaster impose 50 000 m).' },
			{ name: 'providers',         type: 'string',  required: false, desc: 'Providers à activer (CSV). Ex : providers=ticketmaster,meetup.' },
			{ name: 'exclude_providers', type: 'string',  required: false, desc: 'Providers à exclure (CSV).' },
			{ name: 'provider_weights',  type: 'JSON',    required: false, desc: 'Pondération par provider, valeurs ∈ [0,1].' },
			{ name: 'country_hint',      type: 'string',  required: false, desc: 'Code pays ISO 3166-1 α-2 pour la sélection auto.' },
			{ name: 'date',              type: 'string',  required: false, desc: 'Date de début YYYY-MM-DD.' },
			{ name: 'limit',             type: 'integer', required: false, desc: 'Nombre de résultats (1–100). Défaut : 20.' },
		],
		response: `{
  "query": { "mode": "radius", "lat": 48.85, "lng": 2.35, "radius": 50000 },
  "total": 5,
  "results": [
    {
      "id": "ticketmaster:G5v0Z9EA",
      "name": "Paris Jazz Festival",
      "kind": "event",
      "type": "event",
      "score": 0.82,
      "coords": { "lat": 48.86, "lng": 2.33, "approximate": false },
      "date_start": "2026-06-20T19:00:00Z",
      "date_end": "2026-06-20T23:30:00Z",
      "recurring": false,
      "sources": [
        { "provider": "ticketmaster", "url": "https://www.ticketmaster.fr/event/G5v0Z9EA" }
      ]
    }
  ]
}`,
	},
	{
		id: 'poi-events-custom-slim', group: 'POI · Custom', title: 'Évènements personnalisés slim',
		kind: 'route', method: 'GET', path: '/v1/pois/events/custom/slim', cost: 10,
		summary: "Version allégée de GET /v1/pois/events/custom — réponse réduite à name, coords et dates.",
		params: [
			{ name: 'mode',      type: 'enum',    required: false, desc: 'radius (défaut), polygon, district.' },
			{ name: 'lat',       type: 'number',  required: false, desc: 'Latitude du centre.' },
			{ name: 'lng',       type: 'number',  required: false, desc: 'Longitude du centre.' },
			{ name: 'radius',    type: 'integer', required: false, desc: 'Rayon en mètres (max 50 000).' },
			{ name: 'providers', type: 'string',  required: false, desc: 'Providers à activer (CSV).' },
			{ name: 'date',      type: 'string',  required: false, desc: 'Date de début YYYY-MM-DD.' },
			{ name: 'limit',     type: 'integer', required: false, desc: 'Nombre de résultats (1–100). Défaut : 20.' },
		],
		response: `{
  "total": 5,
  "results": [
    {
      "name": "Paris Jazz Festival",
      "coords": { "lat": 48.86, "lng": 2.33, "approximate": false },
      "date_start": "2026-06-20T19:00:00Z",
      "date_end": "2026-06-20T23:30:00Z",
      "recurring": false
    }
  ]
}`,
	},
	{
		id: 'poi-providers-catalog', group: 'POI · Custom', title: 'Catalogue des providers',
		kind: 'route', method: 'GET', path: '/v1/pois/providers/catalog', cost: 1,
		summary: "Liste tous les providers connus du registre (implémentés ou non), avec leurs métadonnées : catégories supportées, BYOK, header HTTP attendu, scores de confiance par pays et par catégorie. Utile pour construire un sélecteur de providers dans votre UI.",
		response: `[
  {
    "id": "overpass",
    "label": "OpenStreetMap / Overpass",
    "byok": false,
    "kinds": ["poi"],
    "categories": ["see", "eat", "drink", "do", "buy", "sleep"],
    "country_scores": { "*": 0.75, "DE": 0.97, "US": 0.82, "CN": 0.60 },
    "category_scores": { "see": 0.88, "do": 0.82, "eat": 0.72 },
    "implemented": true
  },
  {
    "id": "ticketmaster",
    "label": "Ticketmaster",
    "byok": true,
    "byok_header": "X-Ticketmaster-Key",
    "kinds": ["event"],
    "categories": ["event"],
    "country_scores": { "*": 0.38, "US": 0.97, "GB": 0.91, "CA": 0.93 },
    "implemented": true
  },
  {
    "id": "foursquare",
    "label": "Foursquare / FSQ",
    "byok": true,
    "byok_header": "X-Foursquare-Key",
    "kinds": ["poi"],
    "categories": ["eat", "drink", "buy"],
    "country_scores": { "*": 0.70, "US": 0.96 },
    "implemented": false
  }
]`,
	},
	{
		id: 'poi-providers-recommend', group: 'POI · Custom', title: 'Recommandation de providers',
		kind: 'route', method: 'GET', path: '/v1/pois/providers/recommend', cost: 1,
		summary: "Retourne les providers les plus adaptés à une position géographique et à des catégories de POIs, triés par score de confiance. La sélection est géo-intelligente : Baidu/Amap pour la Chine, Kakao pour la Corée, Navitime pour le Japon, Mappls pour l'Inde, GrabMaps pour l'Asie du Sud-Est.",
		params: [
			{ name: 'lat',        type: 'number',  required: true,  desc: 'Latitude.' },
			{ name: 'lng',        type: 'number',  required: true,  desc: 'Longitude.' },
			{ name: 'for_events', type: 'boolean', required: false, desc: 'Accepte true / 1 pour n\'afficher que les providers d\'évènements. Défaut : false.' },
			{ name: 'types',      type: 'string',  required: false, desc: 'Types CSV pour affiner le scoring par catégorie.' },
			{ name: 'limit',      type: 'integer', required: false, desc: 'Nombre max de providers retournés. Défaut : 10.' },
		],
		response: `{
  "country_code": "JP",
  "providers": [
    { "id": "navitime", "label": "Navitime",               "score": 0.95,
      "byok": true,  "byok_header": "X-Navitime-Key", "kinds": ["poi"], "implemented": true },
    { "id": "overpass", "label": "OpenStreetMap / Overpass","score": 0.78,
      "byok": false, "kinds": ["poi"], "implemented": true },
    { "id": "geonames", "label": "GeoNames",               "score": 0.65,
      "byok": false, "kinds": ["poi"], "implemented": true }
  ]
}`,
	},
	{
		id: 'poi-providers', group: 'POI', title: 'Lister les providers actifs',
		kind: 'route', method: 'GET', path: '/v1/pois/providers', cost: 1,
		summary: "Liste les providers actuellement actifs (instances enregistrées au démarrage) avec leur flag BYOK. Pour le catalogue complet (providers non implémentés, scores par pays/catégorie, header BYOK), voir GET /v1/pois/providers/catalog.",
		response: `[
  { "name": "overpass"         },
  { "name": "wikivoyage"       },
  { "name": "wikipedia_events" },
  { "name": "geonames"         },
  { "name": "ticketmaster", "byok": true },
  { "name": "eventbrite",   "byok": true }
]`,
	},
	{
		id: 'health', group: 'Utilitaires', title: 'Health check',
		kind: 'route', method: 'GET', path: '/health', cost: 0,
		summary: "Endpoint de santé. Aucun token consommé, aucune authentification requise. Retourne 200 si l'API est opérationnelle.",
		response: `{ "status": "ok" }`,
	},
];
