package com.example.sampleapp.adapters.http

import com.example.sampleapp.application.sampleitems.SampleItemService
import io.ktor.http.HttpHeaders
import io.ktor.http.HttpStatusCode
import io.ktor.server.request.receive
import io.ktor.server.response.header
import io.ktor.server.response.respond
import io.ktor.server.routing.Route
import io.ktor.server.routing.delete
import io.ktor.server.routing.get
import io.ktor.server.routing.post
import io.ktor.server.routing.put
import io.ktor.server.routing.route

fun Route.sampleItems(sampleItems: SampleItemService) {
    route("/sample-items") {
        post {
            val request = call.receive<SampleItemRequest>()
            val id = sampleItems.create(request.name)
            call.response.header(HttpHeaders.Location, "/sample-items/$id")
            call.respond(HttpStatusCode.Created, CreatedResponse(id.toString()))
        }
        get {
            val limit = call.request.queryParameters["limit"]?.toIntOrNull()
            call.respond(sampleItems.list(limit).map(SampleItemResponse::of))
        }
        get("/{id}") {
            call.respond(SampleItemResponse.of(sampleItems.get(call.parameters["id"].orEmpty())))
        }
        put("/{id}") {
            val request = call.receive<SampleItemRequest>()
            sampleItems.rename(call.parameters["id"].orEmpty(), request.name)
            call.respond(HttpStatusCode.NoContent)
        }
        delete("/{id}") {
            sampleItems.delete(call.parameters["id"].orEmpty())
            call.respond(HttpStatusCode.NoContent)
        }
    }
}
