package com.example.sampleapp.application.orders;

import com.example.sampleapp.domain.order.Money;
import com.example.sampleapp.domain.order.Order;
import java.math.BigDecimal;
import java.time.Instant;
import java.util.List;
import java.util.UUID;

public record OrderView(
        UUID id,
        String customerName,
        String status,
        Instant createdAt,
        BigDecimal total,
        String currency,
        List<OrderItemView> items) {

    public record OrderItemView(String productName, int quantity, BigDecimal unitPrice, BigDecimal subtotal) {
    }

    static OrderView of(Order order) {
        Money total = order.total();
        List<OrderItemView> items = order.items().stream()
                .map(item -> new OrderItemView(item.productName(), item.quantity(), item.unitPrice().amount(), item.subtotal().amount()))
                .toList();
        return new OrderView(order.id(), order.customerName(), order.status().label(), order.createdAt(),
                total.amount(), total.currency(), items);
    }
}
