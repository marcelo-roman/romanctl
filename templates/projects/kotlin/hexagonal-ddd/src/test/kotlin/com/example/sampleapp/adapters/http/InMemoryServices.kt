package com.example.sampleapp.adapters.http

import com.example.sampleapp.adapters.logging.LoggingEventPublisher
import com.example.sampleapp.adapters.memory.MemoryOrderRepository
import com.example.sampleapp.application.orders.OrderService
// romanctl:imports
import java.time.Clock

fun inMemoryServices(): Services {
    val events = LoggingEventPublisher()
    val clock = Clock.systemUTC()
    return Services(
        orders = OrderService(MemoryOrderRepository(), events, clock),
        // romanctl:services
    )
}
