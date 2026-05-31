# trippier — travel API platform

Self-hosted backend that exposes a POI search API and an itinerary generator. Run the full stack with auth and rate limiting, or drop into standalone mode and query without any account.

---

## Architecture

Four services behind a single Docker Compose stack:

| Service | Language | Role |
|---|---|---|
| `auth-api` | Go 1.23 · Gin | Registration, email verification, JWT login, API key management |
| `poi-api` | Go 1.23 · Gin | Point-of-interest search — OpenStreetMap, Wikipedia, Wikivoyage, GeoNames, Ticketmaster, Eventbrite |
| `itinerary-api` | Python 3.12 · FastAPI | Day-by-day itinerary generation from a POI list |
| `frontend` | SvelteKit · Bun | Landing page + auth + dashboard |

Supporting infra: PostgreSQL 16, Redis 7, MailHog (SMTP in dev).

---

## Quick start

### Standalone (no auth, no rate limiting)

The fastest way to run the API. No account, no keys — just the Docker images straight from GHCR:

```bash
docker compose -f docker-compose.standalone.yml up
```

- `http://localhost:8080` — poi-api
- `http://localhost:8000` — itinerary-api

The standalone stack sets `POI_AUTH_DISABLED=true`, so all endpoints are open. Ticketmaster and Eventbrite use BYOK (see below) — no server-side keys needed.

### Full dev stack

```bash
cp .env.example .env
make dev          # builds + starts everything with hot reload
```

The stack comes up on:
- `http://localhost:3000` — frontend
- `http://localhost:8081` — auth-api
- `http://localhost:8080` — poi-api
- `http://localhost:8000` — itinerary-api
- `http://localhost:8025` — MailHog (catch-all SMTP UI)

Hot reload is on everywhere: Go services use [air](https://github.com/air-verse/air), Python uses `uvicorn --reload`, the frontend uses Vite HMR.

```bash
make dev-stop     # tears down the stack and removes volumes
```

---

## API

All API calls require `Authorization: Bearer <api-key>` except health endpoints and standalone mode.

### POI search

```
GET /pois/search?lat=45.83&lng=6.87&radius=5000
```

Aggregates geo-enriched POIs from multiple sources, deduplicates, scores by relevance [0–100]. Supports radius, polygon, and district modes. Results are Redis-cached.

**Cost:** 1 token per request.

### Events

```
GET /pois/events?lat=48.85&lng=2.35&radius=50000
```

Returns cultural events and festivals. Always includes Wikipedia/Wikidata (recurring festivals). Optionally activates Ticketmaster and Eventbrite via BYOK headers — the server never stores those keys:

```bash
curl http://localhost:8080/pois/events \
  -H "X-Ticketmaster-Key: YOUR_TM_KEY" \
  -H "X-Eventbrite-Token: YOUR_EB_TOKEN" \
  -G -d lat=40.71 -d lng=-74.00 -d radius=50000
```

**Minimum radius:** 50 km (enforced server-side — paid event APIs return few results at smaller radii).  
**Cost:** 10 tokens per request.

### Itinerary generation

```
POST /itinerary/generate
Content-Type: application/json

{
  "location": "Chamonix",
  "days": 3,
  "pace": "moderate",
  "transport": "walking"
}
```

Calls the POI API internally, then produces a day-by-day schedule with opening hours and proximity awareness.

**Cost:** 50 tokens per request.

### Auth

```
POST /auth/register          { email, password }
GET  /auth/verify-email?token=…
POST /auth/login             { email, password } → { token }
GET  /auth/me                Bearer <jwt>
POST /auth/keys              Bearer <jwt>    → create API key
GET  /auth/keys              Bearer <jwt>    → list API keys
DELETE /auth/keys/:id        Bearer <jwt>    → revoke API key
```

---

## Token model

Every account starts with **1 000 tokens per month**, shared across all API keys. The bucket refills every 30 days. Quota is stored per-user in `users.tokens_limit` / `users.tokens_reset_interval_secs`.

Two ways to raise a user's quota:

```bash
# 1) From the host, against the running container (operator-only):
sudo docker exec <auth-api-container> /admin set-quota --email=user@example.com --limit=100000

# 2) Service-to-service (e.g. Stripe webhook from the frontend), HMAC-signed:
POST /internal/admin/user-quota
X-Internal-Auth: <ts>.<hmac-sha256(secret, ts)>
{ "email": "user@example.com", "tokens_limit": 100000 }
```

| Endpoint | Cost |
|---|---|
| `GET /pois/search` | 1 token |
| `GET /pois/events` | 10 tokens |
| `GET /pois/events/slim` | 10 tokens |
| `POST /itinerary/generate` | 50 tokens |

---

## BYOK — Ticketmaster & Eventbrite

The POI API stores **zero** third-party API keys. Ticketmaster and Eventbrite are always registered as providers but only activate when the caller passes their own key in the request headers:

| Header | Provider |
|---|---|
| `X-Ticketmaster-Key` | Ticketmaster Discovery API |
| `X-Eventbrite-Token` | Eventbrite private token |

The cache key encodes *which* BYOK providers are active (not their values), so two users with different keys for the same location share the same cache slot — all paid event APIs return the same public event listings regardless of which dev key is used.

---

## Configuration

Copy `.env.example` to `.env` and fill in the values you actually need. The dev compose file uses `.env.example` directly so you can run `make dev` without touching anything.

Required for a production deployment:

```
AUTH_JWT_SECRET=<32+ random chars>
AUTH_DATABASE_URL=postgres://…
AUTH_SMTP_HOST=…
AUTH_SMTP_FROM=noreply@yourdomain.com
AUTH_APP_URL=https://yourdomain.com
INTERNAL_SECRET=<shared secret between services>
```

Optional:
```
POI_GEONAMES_USERNAME=   # free account at geonames.org, enables GeoNames source
POI_AUTH_DISABLED=true   # disables auth-api dependency — open access, no rate limiting
```

Ticketmaster and Eventbrite require **no server configuration** — keys come from callers via request headers.

---

## Tests & lint

Tests run in Docker — no local Go or Python required.

```bash
make test           # all services
make test-go-poi
make test-go-auth
make test-python

make lint           # golangci-lint + ruff + mypy
```

CI runs both test and lint jobs on every push and PR (GitHub Actions, see `.github/workflows/ci.yml`).

---

## Directory layout

```
auth-api/           Go — auth, API key management, email
  cmd/server/       entrypoint
  internal/
    auth/           register / login / verify
    apikey/         create / list / revoke
    email/          SMTP transactional mail
    middleware/     JWT, rate-limit, security headers
    ratelimit/      Redis token-bucket

poi-api/            Go — POI aggregation + caching
  internal/
    byok/           context keys for per-request BYOK credentials
    dedup/          duplicate merging (Wikidata ID + proximity + Jaro-Winkler)
    providers/      overpass · wikivoyage · wikipedia · geonames · ticketmaster · eventbrite
    scoring/        multi-signal relevance scoring
    search/         fan-out pipeline, HTTP handlers
  pkg/types/        shared domain types

itinerary-api/      Python — itinerary logic (FastAPI)
frontend/           SvelteKit app

docker-compose.dev.yml        hot-reload dev stack
docker-compose.full.yml       production-like stack
docker-compose.standalone.yml poi-api + itinerary-api, no auth, pulls from GHCR
Makefile                      dev / test / lint / push targets
```

---

## Docker Compose files

| File | Purpose |
|---|---|
| `docker-compose.dev.yml` | Local development — source-mounted volumes, hot reload, MailHog |
| `docker-compose.full.yml` | Production-like — no hot reload, Postgres/Redis on internal network only |
| `docker-compose.standalone.yml` | Zero-config standalone — pulls GHCR images, no auth, no Postgres |

`make dev` and `make dev-stop` both target `docker-compose.dev.yml`. For other stacks:

```bash
docker compose -f docker-compose.full.yml up -d
docker compose -f docker-compose.standalone.yml up
```

---

## Database migrations

Migrations live in `auth-api/internal/db/migrations/` as numbered SQL files (`001_init.sql`, `002_…`, etc.). They are embedded into the binary at compile time via `//go:embed` and run sequentially at startup. Each file is idempotent (`CREATE TABLE IF NOT EXISTS`, `ALTER TABLE … ADD COLUMN IF NOT EXISTS`, etc.), so re-running against an existing schema is safe.

---

## License

MIT
