FROM golang:alpine AS build
WORKDIR /source
COPY . .
RUN CGO_ENABLED=0 go build -o /app .

FROM alpine
COPY --from=build /app /app
EXPOSE 8080
ENTRYPOINT ["/app"]
