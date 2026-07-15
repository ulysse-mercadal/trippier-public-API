ENGINE   ?= docker
COMPOSE  := $(ENGINE) compose

REGISTRY ?= ghcr.io
OWNER    ?= trippier
TAG      ?= latest

POI_IMAGE       = $(REGISTRY)/$(OWNER)/poi-api:$(TAG)
AUTH_IMAGE      = $(REGISTRY)/$(OWNER)/auth-api:$(TAG)
ITINERARY_IMAGE = $(REGISTRY)/$(OWNER)/itinerary-api:$(TAG)
FRONTEND_IMAGE  = $(REGISTRY)/$(OWNER)/frontend:$(TAG)

UID_GID := $(shell id -u):$(shell id -g)
CACHE   := $(HOME)/.cache/trippier

DRUN    := $(ENGINE) run --rm -u $(UID_GID)
DRUN_GO := $(DRUN) \
	-e GOCACHE=/cache/go-build \
	-e GOPATH=/cache/go \
	-e GOLANGCI_LINT_CACHE=/cache/golangci \
	-v $(CACHE)/go-build:/cache/go-build:z \
	-v $(CACHE)/go:/cache/go:z \
	-v $(CACHE)/golangci:/cache/golangci:z
DRUN_PY := $(ENGINE) run --rm

SERVICE ?=

ifndef NO_COLOR
BLUE := \033[1;34m
CYAN := \033[1;36m
BOLD := \033[1m
DIM  := \033[2m
RST  := \033[0m
endif

.PHONY: help setup init doctor \
	dev dev-stop logs up stop \
	build push \
	prod-pull prod-up prod-stop \
	standalone standalone-stop \
	test test-go-poi test-go-auth test-python \
	lint lint-go-poi lint-go-auth lint-python \
	tidy clean

.DEFAULT_GOAL := help

#################################### Setup #####################################

setup:
	@if [ -f .env ]; then echo ".env already exists, nothing to do."; else \
		cp .env.example .env; \
		secret=$$(openssl rand -hex 32 2>/dev/null || head -c 32 /dev/urandom | od -An -tx1 | tr -d ' \n'); \
		jwt=$$(openssl rand -hex 32 2>/dev/null || head -c 32 /dev/urandom | od -An -tx1 | tr -d ' \n'); \
		sed "s/change-me-internal-secret/$$secret/g; s/change-me-in-production-32-chars!!/$$jwt/g" .env > .env.tmp && mv .env.tmp .env; \
		echo "Created .env"; \
	fi

init: setup

doctor:
	@printf "$(BLUE)Trippier doctor$(RST)\n"
	@command -v $(ENGINE) >/dev/null 2>&1 \
		&& printf "  [ok] $(ENGINE): %s\n" "$$($(ENGINE) --version | head -1)" \
		|| printf "  [!!] $(ENGINE) not found\n"
	@$(COMPOSE) version >/dev/null 2>&1 \
		&& printf "  [ok] '$(ENGINE) compose' available\n" \
		|| printf "  [!!] '$(ENGINE) compose' not available\n"
	@[ -f .env ] && printf "  [ok] .env present\n" || printf "  [!!] .env missing — run 'make setup'\n"
	@$(COMPOSE) config -q >/dev/null 2>&1 \
		&& printf "  [ok] compose files are valid\n" \
		|| printf "  [!!] compose files have errors\n"

########## Development (hot reload + Traefik on *.trippier.localhost) ##########

dev:
	$(COMPOSE) up --build

dev-stop:
	$(COMPOSE) down -v

logs:
	$(COMPOSE) logs -f $(SERVICE)

up:
	$(COMPOSE) -f docker-compose.yml up -d --build

stop:
	$(COMPOSE) -f docker-compose.yml down

############################ Build & publish images ############################

build:
	$(ENGINE) build -t $(AUTH_IMAGE)      ./auth-api
	$(ENGINE) build -t $(POI_IMAGE)       ./poi-api
	$(ENGINE) build -t $(ITINERARY_IMAGE) ./itinerary-api
	$(ENGINE) build -t $(FRONTEND_IMAGE)  ./frontend

push: build
	$(ENGINE) push $(AUTH_IMAGE)
	$(ENGINE) push $(POI_IMAGE)
	$(ENGINE) push $(ITINERARY_IMAGE)
	$(ENGINE) push $(FRONTEND_IMAGE)

############ Production (Traefik + Let's Encrypt, pulls from GHCR) #############

prod-pull:
	$(COMPOSE) -f docker-compose.prod.yml pull

prod-up:
	$(COMPOSE) -f docker-compose.prod.yml up -d

prod-stop:
	$(COMPOSE) -f docker-compose.prod.yml down

############## Standalone (poi-api + itinerary-api only, no auth) ##############

standalone:
	$(COMPOSE) -f docker-compose.standalone.yml up -d

standalone-stop:
	$(COMPOSE) -f docker-compose.standalone.yml down

########### Tests (throwaway containers — no local toolchain needed) ###########

test-go-poi:
	$(DRUN_GO) -v $(CURDIR)/poi-api:/app:z -w /app golang:1.25 go test -race ./...

test-go-auth:
	$(DRUN_GO) -v $(CURDIR)/auth-api:/app:z -w /app golang:1.25 go test -race ./...

test-python:
	$(DRUN_PY) -v $(CURDIR)/itinerary-api:/app:z -w /app python:3.12-slim \
		sh -c "pip install -q -r requirements-dev.txt && pytest --tb=short"

test: test-go-poi test-go-auth test-python

##################################### Lint #####################################

lint-go-poi:
	$(DRUN_GO) -v $(CURDIR)/poi-api:/app:z -w /app golangci/golangci-lint:v2.12.2 golangci-lint run --timeout 5m

lint-go-auth:
	$(DRUN_GO) -v $(CURDIR)/auth-api:/app:z -w /app golangci/golangci-lint:v2.12.2 golangci-lint run --timeout 5m

lint-python:
	$(DRUN_PY) -v $(CURDIR)/itinerary-api:/app:z -w /app python:3.12-slim \
		sh -c "pip install -q -r requirements-dev.txt && ruff check . && mypy app"

lint: lint-go-poi lint-go-auth lint-python

##################################### Misc #####################################

tidy:
	$(DRUN_GO) -v $(CURDIR)/poi-api:/app:z  -w /app golang:1.25-alpine go mod tidy
	$(DRUN_GO) -v $(CURDIR)/auth-api:/app:z -w /app golang:1.25-alpine go mod tidy

clean:
	-$(COMPOSE) down -v --remove-orphans
	-$(COMPOSE) -f docker-compose.prod.yml down -v --remove-orphans
	-$(COMPOSE) -f docker-compose.standalone.yml down -v --remove-orphans

help:
	@printf "$(BLUE)Trippier$(RST) — travel API platform\n\n"
	@printf "$(BOLD)Usage:$(RST) make $(CYAN)<target>$(RST)  [ENGINE=podman] [OWNER=… TAG=…]\n"
	@awk 'BEGIN {FS = ":.*## "} \
		/^#+ .+ #+$$/ { line = $$0; sub(/^#+ +/, "", line); sub(/ +#+$$/, "", line); printf "\n$(BOLD)%s$(RST)\n", line; next } \
		/^[a-zA-Z0-9_-]+:.*## / { printf "  $(CYAN)%-16s$(RST) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
