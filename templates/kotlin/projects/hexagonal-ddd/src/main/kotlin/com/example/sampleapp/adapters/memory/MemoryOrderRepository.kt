package com.example.sampleapp.adapters.memory

import com.example.sampleapp.application.orders.OrderRepository
import com.example.sampleapp.domain.order.Order
import java.util.UUID
import java.util.concurrent.ConcurrentHashMap

class MemoryOrderRepository : OrderRepository {
    private val orders = ConcurrentHashMap<UUID, Order>()

    override fun save(order: Order) {
        orders[order.id] = copy(order)
    }

    override fun findById(id: UUID): Order? = orders[id]?.let(::copy)

    override fun list(limit: Int): List<Order> =
        orders.values.sortedByDescending { it.createdAt }.take(limit).map(::copy)

    private fun copy(order: Order): Order =
        Order.restore(order.id, order.customerName, order.status, order.createdAt, order.items)
}
