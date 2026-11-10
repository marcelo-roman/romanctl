package com.example.sampleapp.adapters.postgres

import com.example.sampleapp.application.orders.OrderRepository
import com.example.sampleapp.domain.order.Money
import com.example.sampleapp.domain.order.Order
import com.example.sampleapp.domain.order.OrderItem
import com.example.sampleapp.domain.order.OrderStatus
import java.sql.Connection
import java.sql.PreparedStatement
import java.sql.ResultSet
import java.time.OffsetDateTime
import java.time.ZoneOffset
import java.util.UUID

class PostgresOrderRepository(private val database: Database) : OrderRepository {
    private data class OrderRow(val id: UUID, val customerName: String, val status: OrderStatus, val createdAt: OffsetDateTime)

    override fun save(order: Order) {
        database.transaction { connection ->
            connection.prepareStatement(UPSERT_ORDER).use {
                it.setObject(1, order.id)
                it.setString(2, order.customerName)
                it.setString(3, order.status.name)
                it.setObject(4, OffsetDateTime.ofInstant(order.createdAt, ZoneOffset.UTC))
                it.executeUpdate()
            }
            connection.prepareStatement(DELETE_ITEMS).use {
                it.setObject(1, order.id)
                it.executeUpdate()
            }
            insertItems(connection, order)
        }
    }

    override fun findById(id: UUID): Order? =
        database.read { load(it, "$SELECT_ORDERS WHERE id = ?") { setObject(1, id) } }.firstOrNull()

    override fun list(limit: Int): List<Order> =
        database.read { load(it, "$SELECT_ORDERS ORDER BY created_at DESC LIMIT ?") { setInt(1, limit) } }

    private fun insertItems(connection: Connection, order: Order) {
        connection.prepareStatement(INSERT_ITEM).use { statement ->
            order.items.forEachIndexed { position, item ->
                statement.setObject(1, item.id)
                statement.setObject(2, order.id)
                statement.setInt(3, position)
                statement.setString(4, item.productName)
                statement.setInt(5, item.quantity)
                statement.setBigDecimal(6, item.unitPrice.amount)
                statement.setString(7, item.unitPrice.currency)
                statement.addBatch()
            }
            statement.executeBatch()
        }
    }

    private fun load(connection: Connection, sql: String, bind: PreparedStatement.() -> Unit): List<Order> {
        val rows = connection.prepareStatement(sql).use { statement ->
            statement.bind()
            statement.executeQuery().use { it.collect(::orderRow) }
        }
        if (rows.isEmpty()) {
            return emptyList()
        }
        val items = itemsOf(connection, rows.map { it.id })
        return rows.map {
            Order.restore(it.id, it.customerName, it.status, it.createdAt.toInstant(), items[it.id].orEmpty())
        }
    }

    private fun itemsOf(connection: Connection, orderIds: List<UUID>): Map<UUID, List<OrderItem>> =
        connection.prepareStatement(SELECT_ITEMS).use { statement ->
            statement.setArray(1, connection.createArrayOf("uuid", orderIds.toTypedArray()))
            statement.executeQuery().use { it.collect(::itemRow) }
        }.groupBy({ it.first }, { it.second })

    private fun orderRow(rs: ResultSet): OrderRow = OrderRow(
        rs.getObject("id", UUID::class.java),
        rs.getString("customer_name"),
        OrderStatus.valueOf(rs.getString("status")),
        rs.getObject("created_at", OffsetDateTime::class.java),
    )

    private fun itemRow(rs: ResultSet): Pair<UUID, OrderItem> {
        val unitPrice = Money(rs.getBigDecimal("unit_price"), rs.getString("currency"))
        val item = OrderItem(rs.getObject("id", UUID::class.java), rs.getString("product_name"), rs.getInt("quantity"), unitPrice)
        return rs.getObject("order_id", UUID::class.java) to item
    }

    private companion object {
        const val SELECT_ORDERS = "SELECT id, customer_name, status, created_at FROM orders"
        const val UPSERT_ORDER = """
            INSERT INTO orders (id, customer_name, status, created_at) VALUES (?, ?, ?, ?)
            ON CONFLICT (id) DO UPDATE SET customer_name = EXCLUDED.customer_name, status = EXCLUDED.status
        """
        const val DELETE_ITEMS = "DELETE FROM order_items WHERE order_id = ?"
        const val INSERT_ITEM = """
            INSERT INTO order_items (id, order_id, position, product_name, quantity, unit_price, currency)
            VALUES (?, ?, ?, ?, ?, ?, ?)
        """
        const val SELECT_ITEMS = """
            SELECT order_id, id, product_name, quantity, unit_price, currency
            FROM order_items WHERE order_id = ANY(?) ORDER BY position
        """
    }
}
