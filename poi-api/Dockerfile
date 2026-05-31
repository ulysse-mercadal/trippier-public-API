FROM golang:1.24-alpine AS builder

# build-base provides gcc + musl-dev, required by the H3 v4 cgo binding
# (github.com/uber/h3-go/v4 links against a statically-compiled H3 C library).
RUN apk add --no-cache build-base

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# CGO is mandatory now that we depend on h3-go (cgo binding). We force static
# linking against musl so the final binary can still run in a minimal base
# without dynamic libc surprises.
RUN CGO_ENABLED=1 GOOS=linux \
    go build -ldflags="-s -w -extldflags '-static'" -o /poi-api ./cmd/server


FROM alpine:3.20

RUN apk add --no-cache ca-certificates

COPY --from=builder /poi-api /poi-api

EXPOSE 8080

ENTRYPOINT ["/poi-api"]
