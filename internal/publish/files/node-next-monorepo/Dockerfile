FROM node:24-alpine AS build
WORKDIR /app
ENV NEXT_TELEMETRY_DISABLED=1
COPY . .
RUN npm ci
RUN npm run build

FROM node:24-alpine
WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV HOSTNAME=0.0.0.0
ENV PORT=3000
COPY --from=build --chown=node:node /app/[[or (index . "Web") "apps/web"]]/.next/standalone ./
COPY --from=build --chown=node:node /app/[[or (index . "Web") "apps/web"]]/.next/static ./[[or (index . "Web") "apps/web"]]/.next/static
USER node
EXPOSE 3000
CMD ["node", "[[or (index . "Web") "apps/web"]]/server.js"]
