# ─── Stage 1: install dependencies ───────────────────────────────────────────
FROM node:18-alpine AS deps
WORKDIR /app
COPY package*.json ./
RUN npm ci

# ─── Stage 2: build ───────────────────────────────────────────────────────────
FROM node:18-alpine AS builder
WORKDIR /app

COPY --from=deps /app/node_modules ./node_modules
COPY . .

ENV NEXT_TELEMETRY_DISABLED=1

# BACKEND_URL — server-side URL used by Next.js rewrites (container → container).
# In Docker Compose this is overridden to http://backend:8080.
ARG BACKEND_URL=http://localhost:8080
ENV BACKEND_URL=${BACKEND_URL}

# NEXT_PUBLIC_API_URL — baked into the client bundle; browser uses this to call
# the backend. Browser always reaches the HOST's published port, not the service name.
ARG NEXT_PUBLIC_API_URL=http://localhost:8080
ENV NEXT_PUBLIC_API_URL=${NEXT_PUBLIC_API_URL}

RUN npm run build

# ─── Stage 3: runtime (minimal standalone image) ─────────────────────────────
FROM node:18-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

RUN addgroup --system --gid 1001 nodejs && \
    adduser  --system --uid 1001 nextjs

# Standalone output contains a self-contained server.js + bundled node_modules.
# Static assets and public files must be copied separately.
COPY --from=builder /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 3000

ENV PORT=3000
ENV HOSTNAME=0.0.0.0

CMD ["node", "server.js"]
