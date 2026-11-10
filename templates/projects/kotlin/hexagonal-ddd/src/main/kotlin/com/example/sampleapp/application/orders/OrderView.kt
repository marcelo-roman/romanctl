package com.example.sampleapp.application.orders

import com.example.sampleapp.domain.order.Order
import java.math.BigDecimal
import java.time.Instant
import java.util.UUID

data class OrderView(
    val id: UUID,
    val customerName: String,
    val status: String,
    val createdAt: Instant,
    val total: BigDecimal,
    val currency: String,
    val items: List<OrderItemView>,
) {
    companion object {
        internal fun of(order: Order): OrderView {
            val total = order.total
            val items = order.items.map {
                OrderItemView(it.productName, it.quantity, it.unitPrice.amount, it.subtotal.amount)
            }
            return OrderView(order.id, order.customerName, order.status.label, order.createdAt, total.amount, total.currency, items)
        }
    }
}

data class OrderItemView(val productName: String, val quantity: Int, val unitPrice: BigDecimal, val subtotal: BigDecimal)
