package com.example.sampleapp.domain.order

import com.example.sampleapp.domain.shared.BusinessRuleException
import com.example.sampleapp.domain.shared.Identifier
import java.math.BigDecimal
import java.time.Instant
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlin.test.assertIs
import kotlin.test.assertTrue

class OrderTest {
    private val now = Instant.parse("2026-01-01T10:00:00Z")

    private fun brl(amount: String) = Money.of(BigDecimal(amount), null)

    @Test
    fun `new order is a draft without items`() {
        val order = Order.create("  Ana  ", now)

        assertEquals("Ana", order.customerName)
        assertEquals(OrderStatus.DRAFT, order.status)
        assertEquals("Draft", order.status.label)
        assertEquals(now, order.createdAt)
        assertEquals(brl("0"), order.total)
    }

    @Test
    fun `customer name is required`() {
        val error = assertFailsWith<BusinessRuleException> { Order.create(" ", now) }

        assertEquals("O nome do cliente é obrigatório.", error.message)
    }

    @Test
    fun `total sums the subtotals`() {
        val order = Order.create("Ana", now)
        order.addItem("Caneta", 2, brl("3.50"))
        order.addItem("Caderno", 1, brl("12.25"))

        assertEquals(brl("19.25"), order.total)
        assertEquals(2, order.items.size)
    }

    @Test
    fun `item needs a product name and a positive quantity`() {
        val order = Order.create("Ana", now)

        assertFailsWith<BusinessRuleException> { order.addItem(" ", 1, brl("1")) }
        val error = assertFailsWith<BusinessRuleException> { order.addItem("Caneta", 0, brl("1")) }
        assertEquals("A quantidade deve ser maior que zero.", error.message)
    }

    @Test
    fun `items must share the currency`() {
        val order = Order.create("Ana", now)
        order.addItem("Caneta", 1, brl("1"))

        val error = assertFailsWith<BusinessRuleException> { order.addItem("Livro", 1, Money.of(BigDecimal.ONE, "USD")) }

        assertEquals("Todos os itens do pedido devem usar a mesma moeda.", error.message)
    }

    @Test
    fun `place raises order placed`() {
        val order = Order.create("Ana", now)
        order.addItem("Caneta", 2, brl("3.50"))

        order.place(now)

        assertEquals(OrderStatus.PLACED, order.status)
        val event = assertIs<OrderPlaced>(order.pullEvents().single())
        assertEquals(order.id, event.orderId)
        assertEquals(BigDecimal("7.00"), event.total)
        assertEquals("order.placed", event.name)
        assertTrue(order.pullEvents().isEmpty())
    }

    @Test
    fun `empty order cannot be placed`() {
        val error = assertFailsWith<BusinessRuleException> { Order.create("Ana", now).place(now) }

        assertEquals("Um pedido sem itens não pode ser realizado.", error.message)
    }

    @Test
    fun `placed order cannot change`() {
        val order = Order.create("Ana", now)
        order.addItem("Caneta", 1, brl("1"))
        order.place(now)

        assertFailsWith<BusinessRuleException> { order.place(now) }
        assertFailsWith<BusinessRuleException> { order.addItem("Livro", 1, brl("1")) }
    }

    @Test
    fun `money validates amount and currency`() {
        assertFailsWith<BusinessRuleException> { Money.of(BigDecimal("-1"), "BRL") }
        assertFailsWith<BusinessRuleException> { Money.of(BigDecimal.ONE, "REAL") }
        assertFailsWith<BusinessRuleException> { Money.of(null, "BRL") }
        assertFailsWith<BusinessRuleException> { brl("1") + Money.of(BigDecimal.ONE, "USD") }
        assertEquals("BRL", brl("1").currency)
        assertEquals(Money.of(BigDecimal("2.5"), "usd"), Money.of(BigDecimal("2.50"), "USD"))
    }

    @Test
    fun `identifier rejects malformed values`() {
        val error = assertFailsWith<BusinessRuleException> { Identifier.parse("1-2-3-4-5") }

        assertEquals("Identificador inválido.", error.message)
    }
}
