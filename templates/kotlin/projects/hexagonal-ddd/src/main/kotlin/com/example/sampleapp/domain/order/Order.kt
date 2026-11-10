package com.example.sampleapp.domain.order

import com.example.sampleapp.domain.shared.BusinessRuleException
import com.example.sampleapp.domain.shared.DomainEvent
import java.time.Instant
import java.util.UUID

class Order private constructor(
    val id: UUID,
    val customerName: String,
    status: OrderStatus,
    val createdAt: Instant,
    items: List<OrderItem>,
) {
    private val lines = items.toMutableList()
    private val pending = mutableListOf<DomainEvent>()

    var status: OrderStatus = status
        private set

    val items: List<OrderItem>
        get() = lines.toList()

    val total: Money
        get() {
            val currency = lines.firstOrNull()?.unitPrice?.currency ?: Money.DEFAULT_CURRENCY
            return lines.fold(Money.zero(currency)) { sum, item -> sum + item.subtotal }
        }

    fun addItem(productName: String, quantity: Int, unitPrice: Money) {
        if (status != OrderStatus.DRAFT) {
            throw BusinessRuleException("Um pedido já realizado não pode ser alterado.")
        }
        val currency = lines.firstOrNull()?.unitPrice?.currency
        if (currency != null && currency != unitPrice.currency) {
            throw BusinessRuleException("Todos os itens do pedido devem usar a mesma moeda.")
        }
        lines.add(OrderItem.create(productName, quantity, unitPrice))
    }

    fun place(now: Instant) {
        if (status != OrderStatus.DRAFT) {
            throw BusinessRuleException("O pedido já foi realizado.")
        }
        if (lines.isEmpty()) {
            throw BusinessRuleException("Um pedido sem itens não pode ser realizado.")
        }
        status = OrderStatus.PLACED
        val total = total
        pending.add(OrderPlaced(id, total.amount, total.currency, now))
    }

    fun pullEvents(): List<DomainEvent> {
        val pulled = pending.toList()
        pending.clear()
        return pulled
    }

    companion object {
        fun create(customerName: String, now: Instant): Order {
            if (customerName.isBlank()) {
                throw BusinessRuleException("O nome do cliente é obrigatório.")
            }
            return Order(UUID.randomUUID(), customerName.trim(), OrderStatus.DRAFT, now, emptyList())
        }

        fun restore(id: UUID, customerName: String, status: OrderStatus, createdAt: Instant, items: List<OrderItem>): Order =
            Order(id, customerName, status, createdAt, items)
    }
}
