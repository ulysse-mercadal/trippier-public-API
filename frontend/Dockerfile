# syntax=docker/dockerfile:1

FROM oven/bun:1-alpine AS builder
WORKDIR /app
COPY package.json bun.lock* ./
RUN --mount=type=cache,target=/root/.bun/install/cache \
    bun install --frozen-lockfile
COPY . .
RUN bunx svelte-kit sync && bun run build


FROM oven/bun:1-alpine AS release
WORKDIR /app
ENV NODE_ENV=production

# curl powers the container HEALTHCHECK.
RUN apk add --no-cache curl

COPY --from=builder /app/build ./build
COPY --from=builder /app/data ./data
COPY --from=builder /app/package.json /app/bun.lock* ./
RUN --mount=type=cache,target=/root/.bun/install/cache \
    bun install --production --frozen-lockfile

USER bun
EXPOSE 3000

HEALTHCHECK --interval=10s --timeout=5s --start-period=20s --retries=5 \
    CMD curl -fsS http://localhost:3000/ || exit 1

CMD ["bun", "build/index.js"]
