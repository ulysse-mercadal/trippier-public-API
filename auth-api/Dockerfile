# syntax=docker/dockerfile:1

FROM golang:1.24-alpine AS builder
WORKDIR /app

ENV GOFLAGS=-mod=mod
ENV GONOSUMDB=*

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o /auth-api ./cmd/server
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o /admin ./cmd/admin


FROM alpine:3.24

# curl powers the container HEALTHCHECK; ca-certificates for outbound HTTPS/SMTP.
RUN apk add --no-cache ca-certificates curl \
    && adduser -D -u 1001 appuser

COPY --from=builder /auth-api /auth-api
COPY --from=builder /admin    /admin

USER appuser
EXPOSE 8081

HEALTHCHECK --interval=10s --timeout=5s --start-period=10s --retries=5 \
    CMD curl -fsS http://localhost:8081/health || exit 1

ENTRYPOINT ["/auth-api"]
