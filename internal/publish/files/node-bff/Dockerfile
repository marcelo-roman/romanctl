FROM node:24-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
COPY apps/api/package.json apps/api/
COPY apps/bff/package.json apps/bff/
COPY apps/web/package.json apps/web/
RUN npm ci
COPY . .
RUN npm run build

FROM node:24-alpine
ARG APP=bff
WORKDIR /app
ENV NODE_ENV=production APP=${APP} WEB_DIST=/app/apps/web/dist
COPY package.json package-lock.json ./
COPY apps/api/package.json apps/api/
COPY apps/bff/package.json apps/bff/
COPY apps/web/package.json apps/web/
RUN npm ci --omit=dev --workspace apps/${APP}
COPY --from=build /app/apps/${APP}/dist apps/${APP}/dist
COPY --from=build /app/apps/web/dist apps/web/dist
EXPOSE 3301 3302
CMD ["sh", "-c", "exec node apps/$APP/dist/server.js"]
