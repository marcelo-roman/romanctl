FROM node:24-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
ARG VITE_API_PROXY=true
ARG VITE_API_URL=
ENV VITE_API_PROXY=$VITE_API_PROXY VITE_API_URL=$VITE_API_URL
RUN npm run build

FROM nginx:alpine
COPY nginx/default.conf.template /etc/nginx/templates/default.conf.template
COPY --from=build /app/dist /usr/share/nginx/html
ENV API_URL=http://api:8090
EXPOSE 80
