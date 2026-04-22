REGISTRY  ?= ghcr.io
OWNER     ?= trippier
TAG       ?= latest

POI_IMAGE        = $(REGISTRY)/$(OWNER)/poi-api:$(TAG)
ITINERARY_IMAGE  = $(REGISTRY)/$(OWNER)/itinerary-api:$(TAG)

.PHONY: build-simple build-full push-simple push-full \
        dev dev-full stop \
        test-go test-python test \
        lint-go lint-python lint \
        tidy

build-simple:
	docker build -t $(POI_IMAGE) ./poi-api

build-full: build-simple
	docker build -t $(ITINERARY_IMAGE) ./itinerary-api

push-simple: build-simple
	docker push $(POI_IMAGE)

push-full: build-full
	docker push $(ITINERARY_IMAGE)

dev:
	docker compose -f docker-compose.simple.yml up --build

dev-full:
	docker compose -f docker-compose.full.yml up --build

stop:
	docker compose -f docker-compose.full.yml down -v

test-go:
	cd poi-api && go test -race ./...

test-python:
	cd itinerary-api && pytest --tb=short

test: test-go test-python

lint-go:
	cd poi-api && golangci-lint run

lint-python:
	cd itinerary-api && ruff check . && mypy app

lint: lint-go lint-python

tidy:
	cd poi-api && go mod tidy
