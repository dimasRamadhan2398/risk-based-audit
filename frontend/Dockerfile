FROM node:20-alpine

WORKDIR /app

ENV NODE_ENV=production
ENV HOST=0.0.0.0
ENV PORT=3000

COPY .output ./.output

EXPOSE 3000

USER node

CMD ["node", ".output/server/index.mjs"]

