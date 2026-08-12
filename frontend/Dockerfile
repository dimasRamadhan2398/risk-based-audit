# Stage 1: Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# Install dependencies
COPY package.json package-lock.json ./
RUN npm ci

# Copy source code
COPY . .

# Build Nuxt application
RUN npm run build

# Stage 2: Production runner stage
FROM node:20-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production

# Copy pre-built application from builder stage
COPY --from=builder /app/.output ./.output

EXPOSE 3000

USER node

CMD ["node", ".output/server/index.mjs"]

