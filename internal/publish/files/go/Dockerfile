FROM golang:alpine AS build
WORKDIR /source
COPY . .
RUN CGO_ENABLED=0 go build -o /app [[.Main]]

FROM alpine
COPY --from=build /app /app
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/app"]
