# Stage 1: Builder
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package configuration
COPY package*.json ./

# Clean lockfile inside container to avoid npm edgesOut issue and install dependencies
RUN rm -f package-lock.json && npm install --legacy-peer-deps

# Copy full frontend source code and build Nuxt 4 application
COPY . .
RUN npm run build

# Stage 2: Runner
FROM node:20-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production
ENV HOST=0.0.0.0
ENV PORT=3000

# Copy pre-built application bundle from builder stage
COPY --from=builder /app/.output ./.output

EXPOSE 3000

USER node

CMD ["node", ".output/server/index.mjs"]
