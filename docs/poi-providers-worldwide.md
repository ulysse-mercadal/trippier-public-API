# POI Providers Worldwide — Research & Roadmap

> **Contexte :** Trippier est une API POI mondiale. Ce document liste toutes les pistes de providers identifiées, leur pricing, leurs limites, et les régions qu'ils couvrent. Déjà intégrés : OpenStreetMap/Overpass, Wikipedia, Wikivoyage, GeoNames.

---

## Déjà intégrés (référence)

| Provider | Type | Couverture | Modèle |
|---|---|---|---|
| **OpenStreetMap / Overpass** | POI (tout type) | Mondial | Gratuit, open data |
| **Wikipedia / Wikidata** | POI culturel, géolocalisé | Mondial | Gratuit, open data |
| **Wikivoyage** | Attractions touristiques | Mondial | Gratuit, open data |
| **GeoNames** | Lieux géographiques | Mondial | Gratuit (compte requis) |
| **Ticketmaster** | Événements | US/EU/AU fort | BYOK |
| **Eventbrite** | Événements | Mondial (32 pays) | BYOK |
| **Meetup** | Événements communautaires | Mondial | BYOK |
| **OpenAgenda** | Événements | France + EU | BYOK |

---

## TIER 1 — Providers globaux (forte couverture mondiale)

### Google Places API (New)
- **Type :** POI — restaurants, landmarks, hotels, retail, toutes catégories
- **Couverture :** 200M+ places ; fort en Amérique du Nord, Europe, Japon, Australie ; inégal en Afrique sub-saharienne, Asie centrale, Amérique du Sud rurale ; **Chine exclue**
- **API :** REST, clé API ou OAuth 2.0
- **Pricing :**
  - Essentials Text/Nearby Search : **32 $/1 000 req** → 2,40 $ à 5M+
  - Place Details Essentials : 5 $/1 000 ; Pro : 17 $/1 000
  - Abonnements : Starter 100 $/mois (50K calls), Essentials 275 $/mois (100K), Pro 1 200 $/mois (250K)
  - Depuis mars 2025 : suppression du crédit 200 $/mois
- **Limites :** Le plus cher à l'échelle ; Chine absente ; pricing revu en hausse en 2025
- **Verdict :** Richesse de données maximale, mais coût prohibitif pour un service public. À réserver aux routes premium ou comme fallback.

---

### Foursquare FSQ Places API
- **Type :** POI — 100M+ POIs vérifiés ; restaurants, nightlife, shopping, landmarks, arts, outdoor
- **Couverture :** 247 pays/territoires ; fort en US, Europe, Japon, Australie ; présent en Chine (partenariat Ctrip), Corée, Inde, Asie du Sud-Est, Amérique du Sud ; plus faible en Afrique rurale et Asie centrale
- **API :** REST, Bearer token
- **Pricing :**
  - Gratuit : **10 000 calls/mois** (→ 500/mois à partir de juin 2026 !)
  - Pro endpoints (Search, Details, Autocomplete) : **15 $/1 000** (10K–100K) → 1,25 $ à 5M+
  - Premium endpoints (photos, tips, horaires) : 18,75 $/1 000 → 1,75 $ à 5M+
- **Limites :** Free tier en forte réduction juin 2026 ; couverture plus faible dans les marchés émergents ; pas de free tier pour les endpoints Premium
- **Verdict :** Meilleur rapport couverture/coût pour un provider global. **Priorité haute pour intégration BYOK.**

---

### HERE Geocoding & Search (Places)
- **Type :** POI — 120M+ places, 400+ catégories ; inclut des données TripAdvisor
- **Couverture :** 200+ pays ; excellent Europe Ouest et Amérique du Nord ; Japon et Corée du Sud (couverture de base) ; Moyen-Orient fort (UAE, Arabie, Israël) ; Afrique : villes majeures
- **API :** REST, clé API ou Bearer token
- **Pricing :**
  - Gratuit : **1 000 req/jour** (5 req/sec)
  - Pay-as-you-go : **1 $/1 000 req**
  - Pro : 449 $/mois avec 1M transactions incluses
- **Limites :** Données tierces en marchés émergents (qualité variable) ; billing par transaction (cross-services)
- **Verdict :** Solide option globale, pricing parmi les plus abordables. À envisager en complément d'OSM.

---

### TomTom Places Search API
- **Type :** POI + geocoding ; Fuzzy Search, POI par type/catégorie
- **Couverture :** 200+ pays ; fort en Amériques (US/Canada/Brésil/Mexique), Europe, Asie-Pacifique (40+ marchés, fort à Singapour, HK, Malaysia, Taïwan, Indonésie) ; Chine et Japon : couverture limitée/conditionnelle ; Moyen-Orient et Afrique : 60+ territoires, UAE/Arabie/Israël/Égypte/Maroc forts, Afrique sub-saharienne : niveau ville uniquement
- **API :** REST, clé API en paramètre URL
- **Pricing :**
  - Gratuit : **2 500 req non-tile/jour** (shared toutes APIs)
  - Payant : **2,50 $/1 000 req**
  - Entreprise : contrat custom
- **Limites :** Chine ville uniquement ; Japon et Corée du Sud conditionnels
- **Verdict :** Bon complément pour Moyen-Orient et Asie-Pacifique. Free tier utilisable pour dev.

---

### Yelp Fusion / Yelp Places API
- **Type :** POI — restaurants, nightlife, shopping, services ; 5M+ businesses ; notes, avis, photos, horaires
- **Couverture :** **32 pays** — US, Canada, UK, France, Allemagne, Australie, Japon, Singapour, HK, Philippines, Malaisie, Taïwan, Mexique, Brésil, Argentine, Chili + Europe occidentale. **Absent : Chine, Corée, Inde, quasi toute l'Afrique, Moyen-Orient.**
- **API :** REST, OAuth 2.0 Bearer token
- **Pricing :**
  - Essai gratuit 30 jours (5 000 calls)
  - Starter : 7,99 $/1 000 — 300 calls/jour max
  - Plus : 9,99 $/1 000 — 500 calls/jour max (photos + avis)
  - Enterprise : 14,99 $/1 000 — 500 calls/jour max
- **Limites :** Devenu payant en 2024 ; quotas journaliers stricts ; couverture géographique limitée
- **Verdict :** Pertinent uniquement pour restauration en US/EU/JP. Couverture trop faible pour un service mondial.

---

### TripAdvisor Content API
- **Type :** POI — hôtels, restaurants, attractions ; 8M+ lieux, 1B+ avis, 29 langues
- **Couverture :** Mondiale sur les destinations touristiques ; présent en Asie, Moyen-Orient, parties de l'Afrique ; faible hors zones touristiques
- **API :** REST, clé API (carte bancaire requise à l'inscription)
- **Pricing :**
  - Gratuit : **5 000 calls/mois**
  - Pay-as-you-go : tarification par volume (visible au checkout)
  - Contact sales au-delà de 500K calls/mois
- **Limites :** Données exclusivement touristiques (hôtels/restaurants/attractions) ; reflet des destinations touristiques plutôt que du quotidien local ; photos et avis soumis à licence
- **Verdict :** Bon complément pour les attractions et hôtels sur routes touristiques mondiales.

---

### Amadeus Points of Interest API
- **Type :** POI — attractions et landmarks touristiques (liste curated Amadeus)
- **Couverture :** Mondiale, focus hubs touristiques ; fort en Europe et Amériques ; Asie/Afrique/Moyen-Orient concentrés sur les corridors touristiques
- **API :** REST, OAuth 2.0 (client credentials) ; inscription self-service sur developers.amadeus.com
- **Pricing :**
  - Sandbox : quota mensuel gratuit (quelques centaines à quelques milliers de calls)
  - Production : pay-as-you-go (tarifs non publics — voir console développeur)
  - Entreprise : contrat mensuel pré-négocié
- **Limites :** Dataset touristique curated, pas exhaustif ; accès production soumis à validation Amadeus
- **Verdict :** Idéal pour enrichir les routes `/pois/events` avec des attractions curated. Gratuit en sandbox.

---

### Overture Maps Foundation (Places Dataset)
- **Type :** POI open data — 64M+ places ; businesses, écoles, hôpitaux, landmarks ; données Meta + Microsoft + Esri + OSM
- **Couverture :** Mondiale ; qualité suit OSM + contributeurs ; fort US/Europe ; en croissance globale
- **API :** Pas d'API REST — données open en GeoParquet sur AWS S3 et Azure Blob ; queryable via DuckDB, Athena, BigQuery ; une API REST tierce existe (overturemapsapi.com)
- **Pricing :**
  - **Gratuit et open** (licence CDLA Permissive 2.0)
  - BigQuery : 1 To/mois gratuit via Google
- **Limites :** 64M places (moins que Foursquare 100M, dataplor 370M) ; pas de API temps réel ; données d'enrichissement (horaires, photos) incomplètes vs. commercial ; schéma encore en évolution
- **Verdict :** Excellent complément open data pour enrichir Overpass/OSM. Pipeline batch plutôt qu'API temps réel.

---

### Geoapify Places API
- **Type :** POI — basé OSM ; 800+ catégories
- **Couverture :** Mondiale (qualité OSM) ; excellent Europe/US ; bon en Asie dans les villes majeures ; plus faible en Afrique rurale
- **API :** REST, clé API
- **Pricing :**
  - Gratuit : **3 000 crédits/jour** (~60 000 résultats POI/jour)
  - Payant : à partir de 59 $/mois
- **Limites :** Qualité bornée par la densité OSM ; Chine/Corée/Japon : limitations pour la géocodage (POI OSM accessible quand même)
- **Verdict :** Alternative budget à Overpass avec une API plus simple. Utile si Overpass est surchargé.

---

### OpenTripMap API
- **Type :** POI — attractions touristiques, patrimoine culturel, sites naturels ; 10M+ objets ; sources : OSM, Wikidata, Wikipedia, ministères du patrimoine
- **Couverture :** Mondiale, focus tourisme/culture ; fort en Europe et régions touristiques populaires
- **API :** REST, clé API (disponible via RapidAPI)
- **Pricing :**
  - À partir de **19 $/mois** (RapidAPI) ; trial gratuit limité
- **Limites :** Focus attraction/patrimoine uniquement, pas de businesses locaux ; qualité OSM
- **Verdict :** Bon pour enrichir les données culturelles et patrimoine. Complémentaire à Wikivoyage.

---

### Radar.io Places API
- **Type :** POI — détection de chaînes et catégories de lieux (McDonald's, Walmart, etc.) ; pas une base POI générale
- **Couverture :** Mondiale pour les grandes chaînes et catégories ; non exhaustif pour les commerces indépendants
- **API :** REST, clé API
- **Pricing :**
  - **2 $/1 000 req** ; free tier disponible
  - Entreprise : custom
- **Limites :** Spécialisé chaînes/marques ; peu d'attributs détaillés vs. Foursquare/Google
- **Verdict :** Utile pour les cas d'usage "détecter si l'utilisateur est dans un McDonald's". Pas pour discovery générale.

---

## TIER 2 — Providers régionaux Asie

### Baidu Maps API (百度地图) — Chine
- **Type :** POI — recherche de lieux, géocodage, détails POI ; couverture Chine exhaustive
- **Couverture :** Chine (exhaustif) ; international très limité
- **API :** REST, clé API (developer.map.baidu.com) ; portail principalement en chinois
- **Pricing :**
  - Gratuit : **100 appels/jour** pour Place Search (très restrictif) ; Géocodage 5 000/jour ; Reverse géocodage 300/jour
  - Payant : ¥40/QPS/mois de concurrence ; abonnements de ¥20 000/mois (service unique) à ¥60 000/mois (multi-services)
- **Limites :** Free tier Place Search extrêmement bas (100/jour) ; portail en chinois ; enregistrement difficile sans numéro de téléphone chinois ; couverture hors Chine minimale
- **Verdict :** Seule option viable pour la Chine avec API publique. Barrière à l'entrée forte. **BYOK recommandé** (utilisateur apporte sa propre clé).

---

### Amap / AutoNavi / Gaode (高德地图) — Chine
- **Type :** POI — POI Chine le plus complet, géocodage, navigation ; filiale Alibaba ; app n°1 de navigation en Chine
- **Couverture :** Chine (le plus complet de tous les providers pour la Chine continentale) ; international limité
- **API :** REST, lbs.amap.com ; portail principalement en chinois
- **Pricing :**
  - Free tier développeur disponible (rapporté comme généreux pour dev en Chine)
  - Production/commercial : quotas par niveau de compte entreprise ; pay-as-you-go pour les dépassements
- **Limites :** Portail en chinois ; enregistrement étranger difficile (nécessite entité commerciale chinoise) ; non conçu pour usage international hors Chine
- **Verdict :** Meilleur provider Chine, mais inaccessible sans entité locale. Complémentaire à Baidu en **BYOK**.

---

### Meituan / Dianping (美团/大众点评) — Chine
- **Type :** POI — 20M+ commerces locaux dans 2 800 villes chinoises ; restaurants (principal), divertissement, beauté, hôtels ; le "Yelp chinois"
- **Couverture :** Chine uniquement
- **API :** **Pas d'API publique** — accès uniquement via partenariat officiel (Merchant API, Logistics API, Analytics API)
- **Pricing :** Partenariat entreprise uniquement ; contacter Meituan business development
- **Limites :** Aucune API publique ; anti-scraping rotatif toutes 72h ; barrières significatives pour les entreprises étrangères
- **Verdict :** Inaccessible sans partenariat formel. À écarter pour l'instant.

---

### Kakao Maps Local API (카카오맵) — Corée du Sud
- **Type :** POI — Corée uniquement ; keyword search, category search ; restaurants, cafés, pharmacies, écoles, etc.
- **Couverture :** Corée du Sud (exhaustif) ; aucune couverture internationale
- **API :** REST, `GET /v2/local/search/keyword` ; `Authorization: KakaoAK {key}`
- **Pricing :**
  - Quota gratuit par app (montant exact dans la console Kakao Developers)
  - À partir de fév. 2026 : ₩10 KRW/appel avec remise 80% ; augmentation de quota via demande de partenariat
- **Limites :** Corée uniquement ; augmentation de quota soumise à validation Kakao ; données en coréen principalement
- **Verdict :** Meilleur provider pour la Corée du Sud. **BYOK recommandé.**

---

### Naver Maps API (네이버 지도) — Corée du Sud
- **Type :** POI + géocodage + itinéraires ; Corée du Sud exhaustif ; POI classés par comportement utilisateur
- **Couverture :** Corée du Sud (principal, exhaustif) ; international très limité
- **API :** REST, NAVER Cloud Platform (ncloud.com) ; headers `X-NCP-APIGW-API-KEY-ID` + `X-NCP-APIGW-API-KEY`
- **Pricing :**
  - **Aucun free tier** sur les APIs Maps
  - Pay-as-you-go par requête (tarifs dans la console NAVER Cloud, en coréen)
- **Limites :** Pas de free tier ; écosystème principalement coréen ; documentation anglaise limitée
- **Verdict :** Alternative à Kakao pour la Corée. Sans free tier, plus difficile à adopter sans engagement.

---

### NAVITIME API — Japon
- **Type :** POI + routage + transit — focus Japon ; spots touristiques, parkings, restaurants ; transit 100% des réseaux de subway mondiaux (déc. 2024)
- **Couverture :** Japon (exhaustif pour routage et POI) ; transit global (subway)
- **API :** REST ; via RapidAPI ou contrat direct NAVITIME Japan ; trial gratuit disponible
- **Pricing :**
  - Via RapidAPI : plans visibles sur rapidapi.com/navitimejapan ; trial gratuit offert
  - Entreprise : contrat direct NAVITIME Japan
- **Limites :** Japon-centrique ; documentation anglaise limitée ; usage commercial = contrat formel
- **Verdict :** Meilleur pour intégrer des données de transit et touristiques japonaises. **BYOK via RapidAPI envisageable.**

---

### ZENRIN Maps API — Japon
- **Type :** POI — Japon uniquement ; données commerciales les plus autoritatives du Japon (standard pour le GPS auto et le gouvernement) ; adresse, téléphone, horaires, parking, catégorie
- **Couverture :** Japon uniquement
- **API :** REST, developers.zmaps-api.com ; licence commerciale requise
- **Pricing :** Tarification entreprise custom uniquement (pas de self-service)
- **Limites :** Japon uniquement ; overhead de licence élevé ; non adapté aux petits développeurs
- **Verdict :** La référence qualité au Japon, mais accessible uniquement avec un budget entreprise.

---

### Korea Tourism Organization — TourAPI 4.0 — Corée du Sud
- **Type :** POI — attractions touristiques coréennes, hébergements, restaurants, festivals, événements culturels ; 260 000+ records ; 15 types de données
- **Couverture :** Corée du Sud (focus tourisme ; couverture nationale des destinations)
- **API :** REST, inscription gratuite sur data.go.kr ; données anglaises disponibles séparément
- **Pricing :** **Gratuit** (open data gouvernemental coréen)
- **Limites :** Données tourisme/attraction uniquement ; principalement coréen (version anglaise disponible pour le contenu touristique) ; fraîcheur variable
- **Verdict :** Complément gratuit idéal pour enrichir la couverture Corée. Intégration simple. **Priorité haute.**

---

### GrabMaps — Asie du Sud-Est
- **Type :** POI — 33M+ POIs et adresses en Asie du Sud-Est ; données hyperlocales y compris les villes de tier 3 ; alimenté par la communauté chauffeurs/livreurs Grab
- **Couverture :** Singapour, Cambodge, Vietnam, Philippines, Indonésie, Malaisie, Myanmar, Thaïlande (**8 pays SEA uniquement**)
- **API :** Via **Amazon Location Service** (provider GrabMaps) ; Snowflake Marketplace ; partenariat B2B direct
- **Pricing :**
  - Via Amazon Location Service : tarification ALS standard
  - Direct : tarification entreprise custom
- **Limites :** Asie du Sud-Est uniquement ; pas de free tier self-service pour accès direct ; partenariat B2B requis hors AWS
- **Verdict :** La référence hyperlocale pour l'Asie du Sud-Est. Accessible via ALS. **Priorité haute pour la région SEA.**

---

## TIER 3 — Providers régionaux (Inde, Amérique Latine, Afrique, Moyen-Orient)

### Mappls / MapmyIndia — Inde
- **Type :** POI — Inde exhaustive ; couverture jusqu'au pas-de-porte ; 30M places dans 7 000 villes et 700 000 villages ; Nearby API, géocodage, détails de lieu
- **Couverture :** Inde (le plus complet disponible — partenariat gouvernemental pour la cartographie nationale) ; couverture internationale limitée via "Global API"
- **API :** REST, about.mappls.com/api ; clé API
- **Pricing :**
  - Free tier disponible (quota non documenté publiquement)
  - Commercial : tarification sur devis
- **Limites :** Détails du free tier opaques ; couverture hors Inde limitée
- **Verdict :** Incontournable pour l'Inde. **BYOK recommandé.** Niveau village = unique.

---

### Zomato API — Inde / International limité
- **Type :** POI — restaurants principalement ; 1,5M+ restaurants dans 10 000+ villes ; notes, menus, photos, avis
- **Couverture :** Inde (exhaustif, 20 grandes villes détaillées) ; international limité et non accessible via API standard
- **API :** REST, developers.zomato.com ; **Free : 1 000 req/jour, max 20 records/requête** ; accès complet via partenariat entreprise
- **Pricing :**
  - Basique : 1 000 calls/jour gratuit
  - Commercial : partenariat Zomato requis
- **Limites :** Vertical restauration uniquement ; accès API enterprise restreint ; free API officiellement limitée
- **Verdict :** Complément pour la restauration indienne, mais limité. Foursquare couvre mieux à l'international.

---

### dataplor — Marchés émergents (Afrique, Amérique Latine, Moyen-Orient)
- **Type :** POI — 370M+ places dans 250+ pays ; données vérifiées sur le terrain ; mises à jour hebdomadaires
- **Couverture :**
  - **Afrique :** 58 pays, 11M+ businesses ; Afrique du Sud : 1,9M+ POIs ; Nigeria, Égypte, Kenya, Ghana forts
  - **Amérique Latine :** 12 pays (voir aussi InfobelPRO)
  - **Moyen-Orient :** présent
  - Conçu spécifiquement pour les marchés sous-représentés
- **API :** REST + livraison bulk (S3, SFTP, Snowflake, BigQuery)
- **Pricing :** Tarification entreprise custom uniquement (pas de self-service)
- **Limites :** Aucun free tier ; focus entreprise ; prix sur devis
- **Verdict :** Le meilleur provider pour l'Afrique et les marchés émergents. Budget entreprise requis.

---

### InfobelPRO POI API — Amérique Latine & Global
- **Type :** POI — 202M+ POIs globaux ; annuaire commercial ; fort en Europe et Amérique Latine
- **Couverture :** 220+ pays ; **Amérique Latine : 38M+ POIs dans 12 pays** (Brésil, Mexique, Argentine, Colombie, Chili, Pérou, Équateur, Bolivie, Paraguay, Uruguay, Venezuela, République Dominicaine)
- **API :** REST + bulk delivery (S3, SFTP, REST, SOAP, Streaming)
- **Pricing :** Custom uniquement ; contacter les ventes
- **Limites :** Pas de self-service ; focus entreprise
- **Verdict :** Le plus fort en Amérique Latine pour les commerces locaux. À envisager si budget entreprise.

---

### Tokyo Tourism Data Catalog — Japon
- **Type :** POI — attractions touristiques multilingues ; spots touristiques de Tokyo avec photos et descriptions ; données du gouvernement métropolitain de Tokyo
- **Couverture :** Tokyo principalement ; quelques données nationales Japon
- **API :** REST JSON/GeoJSON, data.tourism.metro.tokyo.lg.jp/en ; accès libre
- **Pricing :** **Gratuit** (open data gouvernemental)
- **Limites :** Contenu touristique uniquement ; limité à Tokyo pour le catalogue le plus riche ; fraîcheur variable
- **Verdict :** Complément gratuit pour enrichir Tokyo. Intégration facile.

---

### Korea Open Government Data (data.go.kr)
Couvert par TourAPI 4.0 ci-dessus.

---

### India Open Government Data (data.gov.in)
- **Type :** Datasets tourisme et POI variés (sites du patrimoine, spots touristiques, infrastructures)
- **Couverture :** Inde (datasets gouvernementaux fragmentés)
- **API :** REST via data.gov.in/apis ; clé API gratuite
- **Pricing :** Gratuit
- **Limites :** Datasets fragmentés ; pas d'API POI unifiée ; qualité et fraîcheur inconsistantes ; focus gouvernement/patrimoine
- **Verdict :** Complément pour le patrimoine indien, mais MapmyIndia est bien supérieur pour les POI généraux.

---

## Tableau de synthèse

| Provider | Couverture | Free Tier | Payant (ordre de grandeur) | Priorité |
|---|---|---|---|---|
| **Foursquare FSQ** | 247 pays, 100M+ POIs | 10K/mois (→500 juin 2026) | 15 $/1K | 🔴 Haute |
| **HERE Search** | 200+ pays, 120M places | 1 000 req/jour | 1 $/1K | 🔴 Haute |
| **KTO TourAPI** | Corée du Sud (tourisme) | Gratuit | Gratuit | 🔴 Haute |
| **GrabMaps (via ALS)** | SEA — 8 pays | Via AWS free tier | Tarification ALS | 🔴 Haute |
| **TomTom Search** | 200+ pays | 2 500 req/jour | 2,50 $/1K | 🟡 Moyenne |
| **TripAdvisor Content** | Mondial (tourisme) | 5 000/mois | Volume-tiered | 🟡 Moyenne |
| **Amadeus POI** | Mondial (tourisme curated) | Sandbox gratuit | Pay-per-use | 🟡 Moyenne |
| **Geoapify Places** | Mondial (OSM) | 3 000 crédits/jour | 59 $/mois | 🟡 Moyenne |
| **OpenTripMap** | Mondial (tourisme/patrimoine) | Trial limité | 19 $/mois | 🟡 Moyenne |
| **Mappls/MapmyIndia** | Inde (30M places) | Free tier | Sur devis | 🟡 Moyenne (Inde) |
| **Kakao Maps** | Corée du Sud | Quota app gratuit | ₩10/appel | 🟡 Moyenne (Corée) |
| **NAVITIME** | Japon (POI + transit) | Trial gratuit | RapidAPI tiers | 🟡 Moyenne (Japon) |
| **Baidu Maps** | Chine (100/jour !) | 100 place searches/jour | ¥40/QPS/mois | 🟠 BYOK uniquement |
| **Amap/AutoNavi** | Chine (le meilleur) | Dev tier généreux | Enterprise | 🟠 BYOK uniquement |
| **Overture Maps** | Mondial, 64M POIs | Gratuit (open data) | Gratuit | 🟠 Pipeline batch |
| **Google Places** | 200M+ places, global | 1K–10K/mois | 5–32 $/1K | ⚫ Coût prohibitif |
| **Yelp Fusion** | 32 pays (US/EU/JP) | 30j trial seulement | 8–15 $/1K | ⚫ Couverture trop étroite |
| **dataplor** | 250+ pays, 370M POIs | Aucun | Enterprise custom | ⚫ Budget entreprise |
| **InfobelPRO** | 220+ pays, 202M POIs | Aucun | Enterprise custom | ⚫ Budget entreprise |
| **SafeGraph** | 300M+ places | Aucun | Enterprise custom | ⚫ Budget entreprise |
| **Meituan/Dianping** | Chine | **Aucune API publique** | Partenariat | ⚫ Inaccessible |
| **Naver Maps** | Corée du Sud | Aucun | Pay-per-use | ⚫ Pas de free tier |
| **ZENRIN** | Japon uniquement | Aucun (licence) | Enterprise custom | ⚫ Budget entreprise |

---

## Analyse stratégique par région

### 🇨🇳 Chine
Aucune API publique viable pour les entreprises étrangères. Baidu (100/jour) et Amap sont derrière des barrières d'enregistrement significatives (entité commerciale chinoise souvent requise). Meituan/Dianping : pas d'API publique. Foursquare a un partenariat Ctrip qui améliore sa couverture Chine. **Stratégie recommandée :** intégrer Baidu + Amap en BYOK (l'utilisateur apporte sa propre clé), compléter avec Foursquare.

### 🇯🇵 Japon
NAVITIME est le plus accessible pour POI touristique + transit. ZENRIN est le plus autoritatif mais réservé aux grandes entreprises. Tokyo Tourism Data Catalog est gratuit et couvre Tokyo. Foursquare et HERE ont une couverture Japon mais moins dense. **Stratégie recommandée :** NAVITIME BYOK + Tokyo Tourism Data Catalog (gratuit).

### 🇰🇷 Corée du Sud
KTO TourAPI est gratuit et couvre les attractions touristiques. Kakao Maps est le meilleur pour les POI généraux (BYOK). Naver Maps sans free tier. **Stratégie recommandée :** KTO TourAPI (intégration directe) + Kakao Maps BYOK.

### 🌏 Asie du Sud-Est
GrabMaps via Amazon Location Service est la référence (33M+ POIs hyperlocaux, 8 pays). **Stratégie recommandée :** intégrer GrabMaps via ALS en BYOK.

### 🇮🇳 Inde
MapmyIndia est incontestable (30M places, niveau village). Zomato pour la restauration. Google a une bonne couverture mais coûteuse. **Stratégie recommandée :** MapmyIndia BYOK.

### 🌎 Amérique Latine
Aucun provider régional avec API publique self-service. dataplor et InfobelPRO sont les meilleurs mais uniquement en enterprise. Foursquare, HERE, TomTom ont une couverture décente dans les grandes villes. **Stratégie recommandée :** Foursquare BYOK comme provider principal, dataplor/InfobelPRO si budget enterprise.

### 🌍 Afrique
dataplor est le seul provider avec une couverture sérieuse (58 pays, 11M+ businesses) mais enterprise uniquement. HERE/Google/Foursquare couvrent les grandes villes. Aucun provider africain régional avec API publique. **Stratégie recommandée :** OSM/Overpass (déjà intégré) + Foursquare BYOK pour les zones urbaines.

### 🌍 Moyen-Orient
HERE (UAE, Arabie Saoudite, Israël, Égypte fort), TomTom (similaire), Google offrent une couverture raisonnable. Aucun provider régional dédié identifié. **Stratégie recommandée :** HERE BYOK ou TomTom BYOK.

---

## Roadmap d'intégration suggérée

### Phase 1 — BYOK (utilisateur apporte sa clé, zéro coût serveur)
1. **Foursquare FSQ** — global, 100M POIs, meilleur rapport couverture/coût
2. **HERE Search** — global, 1K req/jour gratuit, 1$/1K en payant
3. **Kakao Maps** — Corée du Sud
4. **MapmyIndia** — Inde
5. **Baidu Maps** — Chine (avec disclaimer sur les limites d'enregistrement)
6. **Amap/AutoNavi** — Chine (idem)
7. **NAVITIME** — Japon

### Phase 2 — Intégration directe (gratuit ou low-cost)
1. **KTO TourAPI** — Corée du Sud, attractions touristiques, **gratuit**
2. **Tokyo Tourism Data Catalog** — Japon, **gratuit**
3. **TomTom Search** — 2 500 req/jour gratuit, complément Moyen-Orient/SEA
4. **Geoapify** — fallback OSM-based, 3 000 crédits/jour

### Phase 3 — Partenariats / Budget entreprise
1. **GrabMaps (via ALS)** — SEA hyperlocal
2. **dataplor** — Afrique et marchés émergents
3. **TripAdvisor Content** — enrichissement tourisme
4. **Amadeus POI** — attractions curated mondiale

---

*Document généré le 2026-05-21. Les pricing et free tiers sont susceptibles de changer — vérifier les sources officielles avant intégration.*
