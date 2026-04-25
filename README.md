# trippier — travel API platform

Self-hosted backend that exposes a POI search API and an itinerary generator behind API-key auth. Sign up, verify your email, grab your key, start querying.

---

## Architecture

Four services behind a single Docker Compose stack:

| Service | Language | Role |
|---|---|---|
| `auth-api` | Go 1.23 · Gin | Registration, email verification, JWT login, API key management |
| `poi-api` | Go 1.23 · Gin | Point-of-interest search — OpenStreetMap, Wikipedia, Wikivoyage, GeoNames |
| `itinerary-api` | Python 3.12 · FastAPI | Day-by-day itinerary generation from a POI list |
| `frontend` | SvelteKit · Bun | Landing page + auth + dashboard |

Supporting infra: PostgreSQL 16, Redis 7, MailHog (SMTP in dev).

---

## Quick start

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

All API calls require `Authorization: Bearer <api-key>` except the health endpoints.

### POI search

```
GET /pois/search?lat=45.83&lng=6.87&radius=5000
```

Aggregates geo-enriched POIs from multiple sources, deduplicates them, scores by relevance [0–100]. Supports radius, polygon, and district search modes. Results are Redis-cached.

**Cost:** 1 token per request.

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

Every account starts with **1 000 tokens per month**, shared across all API keys. The bucket refills every 30 days (configurable via `AUTH_DEFAULT_TOKENS_LIMIT` and `AUTH_DEFAULT_RESET_INTERVAL_S`).

| Endpoint | Cost |
|---|---|
| `GET /pois/search` | 1 token |
| `POST /itinerary/generate` | 50 tokens |

---

## Configuration

Copy `.env.example` to `.env` and fill in the values you actually need to change. The dev compose file uses `.env.example` directly so you can run `make dev` without touching anything.

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
```

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
itinerary-api/      Python — itinerary logic (FastAPI)
frontend/           SvelteKit app

docker-compose.dev.yml    hot-reload dev stack
docker-compose.full.yml   production-like stack
Makefile                  dev / test / lint / push targets
```

---

## Docker Compose files

There are two Compose files depending on what you need:

- **`docker-compose.dev.yml`** — for local development. All services mount source code as volumes and reload on file changes (air, uvicorn --reload, Vite HMR). MailHog catches all outbound email.
- **`docker-compose.full.yml`** — production-like. No hot reload, no MailHog, Postgres and Redis exposed only on the internal Docker network. Use this for staging or to smoke-test the final images before deploying.

`make dev` and `make dev-stop` both target `docker-compose.dev.yml`. For the full stack, run Compose directly:

```bash
docker compose -f docker-compose.full.yml up -d
docker compose -f docker-compose.full.yml down -v
```

---

## Database migrations

Migrations live in `auth-api/internal/db/migrations/` as numbered SQL files (`001_init.sql`, `002_…`, etc.). They are embedded into the binary at compile time via `//go:embed` and run sequentially at startup. Each file is idempotent (`CREATE TABLE IF NOT EXISTS`, `ALTER TABLE … ADD COLUMN IF NOT EXISTS`, etc.), so re-running against an existing schema is safe.

---

## License

MIT
