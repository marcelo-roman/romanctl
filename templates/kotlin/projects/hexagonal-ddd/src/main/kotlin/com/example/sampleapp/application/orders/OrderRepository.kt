package com.example.sampleapp.application.orders

import com.example.sampleapp.domain.order.Order
import java.util.UUID

interface OrderRepository {
    fun save(order: Order)

    fun findById(id: UUID): Order?

    fun list(limit: Int): List<Order>
}
