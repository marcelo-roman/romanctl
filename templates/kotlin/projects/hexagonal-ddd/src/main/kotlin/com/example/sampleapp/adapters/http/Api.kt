package com.example.sampleapp.adapters.http

import io.ktor.serialization.kotlinx.json.json
import io.ktor.server.application.Application
import io.ktor.server.application.install
import io.ktor.server.plugins.contentnegotiation.ContentNegotiation
import io.ktor.server.plugins.statuspages.StatusPages
import io.ktor.server.routing.routing

fun Application.api(services: Services) {
    install(ContentNegotiation) { json(ApiJson.format) }
    install(StatusPages) { problems() }
    routing {
        health()
        orders(services.orders)
        // romanctl:routes
    }
}
