FROM golang:1.22-alpine AS builder
WORKDIR /app

ENV GOFLAGS=-mod=mod
ENV GONOSUMDB=*

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /auth-api ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /admin     ./cmd/admin

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /auth-api /auth-api
COPY --from=builder /admin    /admin

EXPOSE 8081
ENTRYPOINT ["/auth-api"]
