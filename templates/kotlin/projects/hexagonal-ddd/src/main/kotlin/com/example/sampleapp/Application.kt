package com.example.sampleapp

import com.example.sampleapp.adapters.http.Services
import com.example.sampleapp.adapters.http.api
import com.example.sampleapp.adapters.logging.LoggingEventPublisher
import com.example.sampleapp.adapters.postgres.Database
import com.example.sampleapp.adapters.postgres.DatabaseSettings
import com.example.sampleapp.adapters.postgres.PostgresOrderRepository
import com.example.sampleapp.application.orders.OrderService
// romanctl:imports
import io.ktor.server.application.ApplicationStopped
import io.ktor.server.engine.embeddedServer
import io.ktor.server.netty.Netty
import java.time.Clock

fun main() {
    val environment = System.getenv()
    val database = Database.connect(DatabaseSettings.from(environment))
    val events = LoggingEventPublisher()
    val clock = Clock.systemUTC()
    val services = Services(
        orders = OrderService(PostgresOrderRepository(database), events, clock),
        // romanctl:services
    )
    val port = environment["PORT"]?.toIntOrNull() ?: 8081
    embeddedServer(Netty, port = port) {
        monitor.subscribe(ApplicationStopped) { database.close() }
        api(services)
    }.start(wait = true)
}
