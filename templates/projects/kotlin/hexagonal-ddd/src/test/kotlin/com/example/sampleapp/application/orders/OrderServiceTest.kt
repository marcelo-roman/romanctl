package com.example.sampleapp.application.orders

import com.example.sampleapp.application.shared.EventPublisher
import com.example.sampleapp.application.shared.NotFoundException
import com.example.sampleapp.domain.order.Order
import com.example.sampleapp.domain.order.OrderPlaced
import com.example.sampleapp.domain.shared.BusinessRuleException
import com.example.sampleapp.domain.shared.DomainEvent
import java.math.BigDecimal
import java.time.Clock
import java.time.Instant
import java.time.ZoneOffset
import java.util.UUID
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlin.test.assertIs

class OrderServiceTest {
    private class FakeOrderRepository : OrderRepository {
        val saved = mutableMapOf<UUID, Order>()

        override fun save(order: Order) {
            saved[order.id] = order
        }

        override fun findById(id: UUID): Order? = saved[id]

        override fun list(limit: Int): List<Order> = saved.values.take(limit)
    }

    private class RecordingPublisher : EventPublisher {
        val published = mutableListOf<DomainEvent>()

        override fun publish(events: List<DomainEvent>) {
            published.addAll(events)
        }
    }

    private val repository = FakeOrderRepository()
    private val events = RecordingPublisher()
    private val clock = Clock.fixed(Instant.parse("2026-01-01T10:00:00Z"), ZoneOffset.UTC)
    private val service = OrderService(repository, events, clock)

    @Test
    fun `places an order and publishes the event`() {
        val id = service.create("Ana")
        service.addItem(id.toString(), AddItem("Caneta", 2, BigDecimal("3.5"), null))

        service.place(id.toString())

        val view = service.get(id.toString())
        assertEquals("Placed", view.status)
        assertEquals(BigDecimal("7.00"), view.total)
        assertEquals("BRL", view.currency)
        assertEquals(clock.instant(), view.createdAt)
        assertIs<OrderPlaced>(events.published.single())
    }

    @Test
    fun `unknown order is not found`() {
        assertFailsWith<NotFoundException> { service.get(UUID.randomUUID().toString()) }
    }

    @Test
    fun `malformed id breaks a rule`() {
        val error = assertFailsWith<BusinessRuleException> { service.place("abc") }

        assertEquals("Identificador inválido.", error.message)
    }

    @Test
    fun `list honours the limit`() {
        repeat(3) { service.create("Cliente $it") }

        assertEquals(2, service.list(2).size)
        assertEquals(3, service.list(null).size)
    }
}
