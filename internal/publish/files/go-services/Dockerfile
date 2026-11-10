FROM golang:alpine AS build
ARG SERVICE=orders
WORKDIR /source
COPY . .
RUN CGO_ENABLED=0 go build -o /app ./cmd/${SERVICE}

FROM alpine
COPY --from=build /app /app
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/app"]
