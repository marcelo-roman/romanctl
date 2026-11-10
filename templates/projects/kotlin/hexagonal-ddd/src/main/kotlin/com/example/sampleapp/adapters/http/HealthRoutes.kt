package com.example.sampleapp.adapters.http

import io.ktor.server.response.respond
import io.ktor.server.routing.Route
import io.ktor.server.routing.get

fun Route.health() {
    get("/health") {
        call.respond(mapOf("status" to "ok"))
    }
}
