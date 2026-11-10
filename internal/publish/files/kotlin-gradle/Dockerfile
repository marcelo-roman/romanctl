FROM eclipse-temurin:21 AS build
WORKDIR /source
COPY gradlew settings.gradle.kts build.gradle.kts gradle.properties ./
COPY gradle gradle
RUN ./gradlew --no-daemon --quiet dependencies
COPY . .
RUN ./gradlew --no-daemon --quiet installDist && mv build/install/[[.Dist]] /app

FROM eclipse-temurin:21-jre
WORKDIR /app
COPY --from=build /app .
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/app/bin/[[.Dist]]"]
