# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder

# build-base provides gcc + musl-dev, required by the H3 v4 cgo binding
# (github.com/uber/h3-go/v4 links against a statically-compiled H3 C library).
RUN apk add --no-cache build-base

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY . .
# CGO is mandatory now that we depend on h3-go (cgo binding). We force static
# linking against musl so the final binary can still run in a minimal base
# without dynamic libc surprises.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 GOOS=linux \
    go build -ldflags="-s -w -extldflags '-static'" -o /poi-api ./cmd/server


FROM alpine:3.24

# curl powers the container HEALTHCHECK; ca-certificates for outbound HTTPS.
RUN apk add --no-cache ca-certificates curl \
    && adduser -D -u 1001 appuser

COPY --from=builder /poi-api /poi-api

USER appuser
EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=5s --start-period=10s --retries=5 \
    CMD curl -fsS http://localhost:8080/health || exit 1

ENTRYPOINT ["/poi-api"]
