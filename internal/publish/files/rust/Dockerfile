FROM rust:1 AS build
WORKDIR /source
COPY . .
RUN cargo build --release --locked --bin [[.Bin]]

FROM debian:stable-slim
COPY --from=build /source/target/release/[[.Bin]] /usr/local/bin/app
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/app"]
