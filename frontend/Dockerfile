# Production runner image using pre-built Nuxt .output
FROM node:20-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production

# Copy pre-built application from local build
COPY .output ./.output

EXPOSE 3000

USER node

CMD ["node", ".output/server/index.mjs"]
