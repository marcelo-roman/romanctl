package com.example.sampleapp

import com.example.sampleapp.adapters.http.Services
import com.example.sampleapp.adapters.http.api
import com.example.sampleapp.adapters.logging.LoggingEventPublisher
import com.example.sampleapp.adapters.memory.MemoryOrderRepository
import com.example.sampleapp.application.orders.OrderService
// romanctl:imports
import io.ktor.server.engine.embeddedServer
import io.ktor.server.netty.Netty
import java.time.Clock

fun main() {
    val environment = System.getenv()
    val events = LoggingEventPublisher()
    val clock = Clock.systemUTC()
    val services = Services(
        orders = OrderService(MemoryOrderRepository(), events, clock),
        // romanctl:services
    )
    val port = environment["PORT"]?.toIntOrNull() ?: 8081
    embeddedServer(Netty, port = port) { api(services) }.start(wait = true)
}
