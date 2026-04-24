REGISTRY  ?= ghcr.io
OWNER     ?= trippier
TAG       ?= latest

POI_IMAGE        = $(REGISTRY)/$(OWNER)/poi-api:$(TAG)
AUTH_IMAGE       = $(REGISTRY)/$(OWNER)/auth-api:$(TAG)
ITINERARY_IMAGE  = $(REGISTRY)/$(OWNER)/itinerary-api:$(TAG)
FRONTEND_IMAGE   = $(REGISTRY)/$(OWNER)/frontend:$(TAG)

UID_GID  := $(shell id -u):$(shell id -g)
DRUN     := docker run --rm -u $(UID_GID)
DRUN_GO  := $(DRUN) -e GOCACHE=/tmp/go-build -e GOPATH=/tmp/go -e GOLANGCI_LINT_CACHE=/tmp/golangci-cache
DRUN_PY  := docker run --rm

.PHONY: dev dev-stop \
        prod-build prod-up prod-stop \
        test-go-poi test-go-auth test-python test \
        lint-go-poi lint-go-auth lint-python lint \
        push tidy

# ── Dev (hot reload) ──────────────────────────────────────────────────────────
# All services reload on file changes:
#   Go (auth-api, poi-api) → air
#   Python (itinerary-api) → uvicorn --reload
#   Frontend               → Vite HMR on :3000  (websocket :24678)

dev:
	docker compose -f docker-compose.dev.yml up --build

dev-stop:
	docker compose -f docker-compose.dev.yml down -v

# ── Production ────────────────────────────────────────────────────────────────

prod-build:
	docker build -t $(POI_IMAGE)       ./poi-api
	docker build -t $(AUTH_IMAGE)      ./auth-api
	docker build -t $(ITINERARY_IMAGE) ./itinerary-api
	docker build -t $(FRONTEND_IMAGE)  ./frontend

prod-up:
	docker compose -f docker-compose.full.yml up

prod-stop:
	docker compose -f docker-compose.full.yml down -v

push: prod-build
	docker push $(POI_IMAGE)
	docker push $(AUTH_IMAGE)
	docker push $(ITINERARY_IMAGE)
	docker push $(FRONTEND_IMAGE)

# ── Tests ─────────────────────────────────────────────────────────────────────

test-go-poi:
	$(DRUN_GO) -v $(CURDIR)/poi-api:/app:z -w /app golang:1.23 \
		go test -race ./...

test-go-auth:
	$(DRUN_GO) -v $(CURDIR)/auth-api:/app:z -w /app golang:1.23 \
		go test -race ./...

test-python:
	$(DRUN_PY) -v $(CURDIR)/itinerary-api:/app:z -w /app python:3.12-slim \
		sh -c "pip install -q -r requirements-dev.txt && pytest --tb=short"

test: test-go-poi test-go-auth test-python

# ── Lint ──────────────────────────────────────────────────────────────────────

lint-go-poi:
	$(DRUN_GO) -v $(CURDIR)/poi-api:/app:z -w /app golangci/golangci-lint:v1.64 \
		golangci-lint run

lint-go-auth:
	$(DRUN_GO) -v $(CURDIR)/auth-api:/app:z -w /app golangci/golangci-lint:v1.64 \
		golangci-lint run

lint-python:
	$(DRUN_PY) -v $(CURDIR)/itinerary-api:/app:z -w /app python:3.12-slim \
		sh -c "pip install -q -r requirements-dev.txt && ruff check . && mypy app"

lint: lint-go-poi lint-go-auth lint-python

# ── Misc ──────────────────────────────────────────────────────────────────────

tidy:
	$(DRUN_GO) -v $(CURDIR)/poi-api:/app:z -w /app golang:1.23-alpine go mod tidy
	$(DRUN_GO) -v $(CURDIR)/auth-api:/app:z -w /app golang:1.23-alpine go mod tidy
