package com.example.sampleapp.adapters.postgres;

import com.example.sampleapp.application.orders.OrderRepository;
import com.example.sampleapp.domain.order.Money;
import com.example.sampleapp.domain.order.Order;
import com.example.sampleapp.domain.order.OrderItem;
import com.example.sampleapp.domain.order.OrderStatus;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.time.OffsetDateTime;
import java.time.ZoneOffset;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.UUID;
import java.util.stream.Collectors;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.jdbc.core.simple.JdbcClient;
import org.springframework.stereotype.Repository;
import org.springframework.transaction.PlatformTransactionManager;
import org.springframework.transaction.support.TransactionTemplate;

@Repository
@ConditionalOnProperty(name = "app.persistence", havingValue = "postgres")
public class PostgresOrderRepository implements OrderRepository {

    private static final String SELECT_ORDERS = "SELECT id, customer_name, status, created_at FROM orders";

    private record OrderRow(UUID id, String customerName, OrderStatus status, OffsetDateTime createdAt) {
    }

    private record ItemRow(UUID orderId, OrderItem item) {
    }

    private final JdbcClient jdbc;
    private final TransactionTemplate transactions;

    public PostgresOrderRepository(JdbcClient jdbc, PlatformTransactionManager transactionManager) {
        this.jdbc = jdbc;
        this.transactions = new TransactionTemplate(transactionManager);
    }

    @Override
    public void save(Order order) {
        transactions.executeWithoutResult(status -> {
            jdbc.sql("""
                    INSERT INTO orders (id, customer_name, status, created_at) VALUES (:id, :customerName, :status, :createdAt)
                    ON CONFLICT (id) DO UPDATE SET customer_name = EXCLUDED.customer_name, status = EXCLUDED.status
                    """)
                    .param("id", order.id())
                    .param("customerName", order.customerName())
                    .param("status", order.status().name())
                    .param("createdAt", OffsetDateTime.ofInstant(order.createdAt(), ZoneOffset.UTC))
                    .update();
            jdbc.sql("DELETE FROM order_items WHERE order_id = :id").param("id", order.id()).update();
            List<OrderItem> items = order.items();
            for (int position = 0; position < items.size(); position++) {
                insertItem(order.id(), position, items.get(position));
            }
        });
    }

    @Override
    public Optional<Order> findById(UUID id) {
        List<OrderRow> rows = jdbc.sql(SELECT_ORDERS + " WHERE id = :id").param("id", id).query(PostgresOrderRepository::orderRow).list();
        return load(rows).stream().findFirst();
    }

    @Override
    public List<Order> list(int limit) {
        List<OrderRow> rows = jdbc.sql(SELECT_ORDERS + " ORDER BY created_at DESC LIMIT :limit")
                .param("limit", limit)
                .query(PostgresOrderRepository::orderRow)
                .list();
        return load(rows);
    }

    private void insertItem(UUID orderId, int position, OrderItem item) {
        jdbc.sql("""
                INSERT INTO order_items (id, order_id, position, product_name, quantity, unit_price, currency)
                VALUES (:id, :orderId, :position, :productName, :quantity, :unitPrice, :currency)
                """)
                .param("id", item.id())
                .param("orderId", orderId)
                .param("position", position)
                .param("productName", item.productName())
                .param("quantity", item.quantity())
                .param("unitPrice", item.unitPrice().amount())
                .param("currency", item.unitPrice().currency())
                .update();
    }

    private List<Order> load(List<OrderRow> rows) {
        if (rows.isEmpty()) {
            return List.of();
        }
        Map<UUID, List<OrderItem>> items = itemsOf(rows.stream().map(OrderRow::id).toList());
        return rows.stream()
                .map(row -> Order.restore(row.id(), row.customerName(), row.status(), row.createdAt().toInstant(),
                        items.getOrDefault(row.id(), List.of())))
                .toList();
    }

    private Map<UUID, List<OrderItem>> itemsOf(List<UUID> orderIds) {
        return jdbc.sql("""
                        SELECT order_id, id, product_name, quantity, unit_price, currency
                        FROM order_items WHERE order_id IN (:ids) ORDER BY position
                        """)
                .param("ids", orderIds)
                .query(PostgresOrderRepository::itemRow)
                .list()
                .stream()
                .collect(Collectors.groupingBy(ItemRow::orderId, Collectors.mapping(ItemRow::item, Collectors.toList())));
    }

    private static OrderRow orderRow(ResultSet rs, int rowNum) throws SQLException {
        return new OrderRow(rs.getObject("id", UUID.class), rs.getString("customer_name"),
                OrderStatus.valueOf(rs.getString("status")), rs.getObject("created_at", OffsetDateTime.class));
    }

    private static ItemRow itemRow(ResultSet rs, int rowNum) throws SQLException {
        Money unitPrice = new Money(rs.getBigDecimal("unit_price"), rs.getString("currency"));
        OrderItem item = new OrderItem(rs.getObject("id", UUID.class), rs.getString("product_name"), rs.getInt("quantity"), unitPrice);
        return new ItemRow(rs.getObject("order_id", UUID.class), item);
    }
}
