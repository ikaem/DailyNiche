# Build stage - installs all dependencies (SvelteKit/Vite/adapter-node are
# devDependencies here, but they're exactly what's needed to produce the
# build) and runs the production build via adapter-node.
FROM node:22-alpine AS builder

WORKDIR /build

COPY package.json package-lock.json ./
RUN npm ci

COPY . .

RUN npm run build

# Runtime stage - small final image running the built adapter-node server.
FROM node:22-alpine

WORKDIR /app

COPY --from=builder /build/build ./build
COPY --from=builder /build/package.json /build/package-lock.json ./

# This project has no runtime "dependencies" section (everything
# SvelteKit/Vite-related is a devDependency, only needed to produce the
# build) - --omit=dev installs essentially nothing here. Confirmed via a
# live run (not just a successful build) that the adapter-node output still
# starts and serves real requests correctly without node_modules present -
# Vite's server bundle inlines what it needs at build time.
RUN npm ci --omit=dev

EXPOSE 3000

ENV PORT=3000
ENV NODE_ENV=production

CMD ["node", "build/index.js"]
