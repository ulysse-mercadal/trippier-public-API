FROM oven/bun:1-alpine AS builder
WORKDIR /app
COPY package.json bun.lockb* ./
RUN bun install --frozen-lockfile
COPY . .
# svelte-kit sync generates .svelte-kit/tsconfig.json and route types.
# Vite build runs under Node to avoid Bun's ESM resolution incompatibility
# with @sveltejs/vite-plugin-svelte (Bun resolves src/ instead of dist/).
RUN bunx svelte-kit sync && node node_modules/.bin/vite build

FROM oven/bun:1-alpine
WORKDIR /app
COPY --from=builder /app/build ./build
COPY --from=builder /app/package.json ./
RUN bun install --production --frozen-lockfile
ENV NODE_ENV=production
EXPOSE 3000
CMD ["bun", "build/index.js"]
