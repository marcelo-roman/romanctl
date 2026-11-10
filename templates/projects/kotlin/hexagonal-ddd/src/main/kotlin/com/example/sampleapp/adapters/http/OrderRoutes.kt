package com.example.sampleapp.adapters.http

import com.example.sampleapp.application.orders.OrderService
import io.ktor.http.HttpHeaders
import io.ktor.http.HttpStatusCode
import io.ktor.server.request.receive
import io.ktor.server.response.header
import io.ktor.server.response.respond
import io.ktor.server.routing.Route
import io.ktor.server.routing.get
import io.ktor.server.routing.post
import io.ktor.server.routing.route

fun Route.orders(orders: OrderService) {
    route("/orders") {
        post {
            val request = call.receive<CreateOrderRequest>()
            val id = orders.create(request.customerName)
            call.response.header(HttpHeaders.Location, "/orders/$id")
            call.respond(HttpStatusCode.Created, CreatedResponse(id.toString()))
        }
        post("/{id}/items") {
            val request = call.receive<AddItemRequest>()
            orders.addItem(call.parameters["id"].orEmpty(), request.toCommand())
            call.respond(HttpStatusCode.NoContent)
        }
        post("/{id}/place") {
            orders.place(call.parameters["id"].orEmpty())
            call.respond(HttpStatusCode.NoContent)
        }
        get("/{id}") {
            call.respond(OrderResponse.of(orders.get(call.parameters["id"].orEmpty())))
        }
        get {
            val limit = call.request.queryParameters["limit"]?.toIntOrNull()
            call.respond(orders.list(limit).map(OrderResponse::of))
        }
    }
}
