package com.example.sampleapp.application.orders

import com.example.sampleapp.application.shared.EventPublisher
import com.example.sampleapp.application.shared.ListLimit
import com.example.sampleapp.application.shared.NotFoundException
import com.example.sampleapp.domain.order.Money
import com.example.sampleapp.domain.order.Order
import com.example.sampleapp.domain.shared.Identifier
import java.time.Clock
import java.util.UUID

class OrderService(
    private val repository: OrderRepository,
    private val events: EventPublisher,
    private val clock: Clock,
) {
    fun create(customerName: String): UUID {
        val order = Order.create(customerName, clock.instant())
        repository.save(order)
        return order.id
    }

    fun addItem(id: String, command: AddItem) {
        val unitPrice = Money.of(command.unitPrice, command.currency)
        change(id) { it.addItem(command.productName, command.quantity, unitPrice) }
    }

    fun place(id: String) {
        change(id) { it.place(clock.instant()) }
    }

    fun get(id: String): OrderView = OrderView.of(load(id))

    fun list(limit: Int?): List<OrderView> = repository.list(ListLimit.of(limit)).map(OrderView::of)

    private fun change(id: String, apply: (Order) -> Unit) {
        val order = load(id)
        apply(order)
        repository.save(order)
        events.publish(order.pullEvents())
    }

    private fun load(id: String): Order =
        repository.findById(Identifier.parse(id)) ?: throw NotFoundException("Pedido não encontrado.")
}
