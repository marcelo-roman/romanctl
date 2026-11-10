package com.example.sampleapp.adapters.postgres

import com.example.sampleapp.domain.order.Money
import com.example.sampleapp.domain.order.Order
import com.example.sampleapp.domain.order.OrderStatus
import java.math.BigDecimal
import java.time.Instant
import java.time.temporal.ChronoUnit
import java.util.UUID
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNotNull
import kotlin.test.assertNull
import kotlin.test.assertTrue

class PostgresOrderRepositoryTest {
    @Test
    fun `round trip`() {
        val repository = PostgresOrderRepository(TestDatabase.connectOrSkip())
        val now = Instant.now().truncatedTo(ChronoUnit.MICROS)
        val order = Order.create("Ana", now)
        order.addItem("Caneta", 2, Money.of(BigDecimal("3.5"), null))
        order.addItem("Caderno", 1, Money.of(BigDecimal("12.25"), null))
        repository.save(order)
        order.place(now)
        repository.save(order)

        val loaded = assertNotNull(repository.findById(order.id))

        assertEquals("Ana", loaded.customerName)
        assertEquals(OrderStatus.PLACED, loaded.status)
        assertEquals(now, loaded.createdAt)
        assertEquals(listOf("Caneta", "Caderno"), loaded.items.map { it.productName })
        assertEquals(Money.of(BigDecimal("19.25"), "BRL"), loaded.total)
        assertTrue(repository.list(200).any { it.id == order.id })
        assertNull(repository.findById(UUID.randomUUID()))
    }
}
