REGISTRY  ?= ghcr.io
OWNER     ?= trippier
TAG       ?= latest

POI_IMAGE        = $(REGISTRY)/$(OWNER)/poi-api:$(TAG)
AUTH_IMAGE       = $(REGISTRY)/$(OWNER)/auth-api:$(TAG)
ITINERARY_IMAGE  = $(REGISTRY)/$(OWNER)/itinerary-api:$(TAG)
FRONTEND_IMAGE   = $(REGISTRY)/$(OWNER)/frontend:$(TAG)

.PHONY: build push \
        dev dev-full stop \
        test-go-poi test-go-auth test-python test \
        lint-go-poi lint-go-auth lint-python lint \
        tidy

# ── Build ─────────────────────────────────────────────────────────────────────

build:
	docker build -t $(POI_IMAGE) ./poi-api
	docker build -t $(AUTH_IMAGE) ./auth-api
	docker build -t $(ITINERARY_IMAGE) ./itinerary-api
	docker build -t $(FRONTEND_IMAGE) ./frontend

push: build
	docker push $(POI_IMAGE)
	docker push $(AUTH_IMAGE)
	docker push $(ITINERARY_IMAGE)
	docker push $(FRONTEND_IMAGE)

# ── Dev ───────────────────────────────────────────────────────────────────────

dev:
	docker compose -f docker-compose.simple.yml up --build

dev-full:
	docker compose -f docker-compose.full.yml up --build

stop:
	docker compose -f docker-compose.simple.yml down -v

stop-full:
	docker compose -f docker-compose.full.yml down -v

# ── Tests ─────────────────────────────────────────────────────────────────────

test-go-poi:
	cd poi-api && go test -race ./...

test-go-auth:
	cd auth-api && go test -race ./...

test-python:
	cd itinerary-api && pytest --tb=short

test: test-go-poi test-go-auth test-python

# ── Lint ──────────────────────────────────────────────────────────────────────

lint-go-poi:
	cd poi-api && golangci-lint run

lint-go-auth:
	cd auth-api && golangci-lint run

lint-python:
	cd itinerary-api && ruff check . && mypy app

lint: lint-go-poi lint-go-auth lint-python

# ── Misc ──────────────────────────────────────────────────────────────────────

tidy:
	cd poi-api && go mod tidy
	cd auth-api && go mod tidy
