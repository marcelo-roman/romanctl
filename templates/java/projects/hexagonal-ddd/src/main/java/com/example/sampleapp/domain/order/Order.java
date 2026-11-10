package com.example.sampleapp.domain.order;

import com.example.sampleapp.domain.shared.BusinessRuleException;
import com.example.sampleapp.domain.shared.DomainEvent;
import java.time.Instant;
import java.util.ArrayList;
import java.util.List;
import java.util.UUID;

public final class Order {

    private final UUID id;
    private final String customerName;
    private final Instant createdAt;
    private final List<OrderItem> items;
    private final List<DomainEvent> events = new ArrayList<>();
    private OrderStatus status;

    private Order(UUID id, String customerName, OrderStatus status, Instant createdAt, List<OrderItem> items) {
        this.id = id;
        this.customerName = customerName;
        this.status = status;
        this.createdAt = createdAt;
        this.items = new ArrayList<>(items);
    }

    public static Order create(String customerName, Instant now) {
        if (customerName == null || customerName.isBlank()) {
            throw new BusinessRuleException("O nome do cliente é obrigatório.");
        }
        return new Order(UUID.randomUUID(), customerName.strip(), OrderStatus.DRAFT, now, List.of());
    }

    public static Order restore(UUID id, String customerName, OrderStatus status, Instant createdAt, List<OrderItem> items) {
        return new Order(id, customerName, status, createdAt, items);
    }

    public UUID id() {
        return id;
    }

    public String customerName() {
        return customerName;
    }

    public OrderStatus status() {
        return status;
    }

    public Instant createdAt() {
        return createdAt;
    }

    public List<OrderItem> items() {
        return List.copyOf(items);
    }

    public Money total() {
        String currency = items.isEmpty() ? Money.DEFAULT_CURRENCY : items.getFirst().unitPrice().currency();
        return items.stream().map(OrderItem::subtotal).reduce(Money.zero(currency), Money::add);
    }

    public void addItem(String productName, int quantity, Money unitPrice) {
        if (status != OrderStatus.DRAFT) {
            throw new BusinessRuleException("Um pedido já realizado não pode ser alterado.");
        }
        if (!items.isEmpty() && !items.getFirst().unitPrice().currency().equals(unitPrice.currency())) {
            throw new BusinessRuleException("Todos os itens do pedido devem usar a mesma moeda.");
        }
        items.add(OrderItem.create(productName, quantity, unitPrice));
    }

    public void place(Instant now) {
        if (status != OrderStatus.DRAFT) {
            throw new BusinessRuleException("O pedido já foi realizado.");
        }
        if (items.isEmpty()) {
            throw new BusinessRuleException("Um pedido sem itens não pode ser realizado.");
        }
        status = OrderStatus.PLACED;
        Money total = total();
        events.add(new OrderPlaced(id, total.amount(), total.currency(), now));
    }

    public List<DomainEvent> pullEvents() {
        List<DomainEvent> pulled = List.copyOf(events);
        events.clear();
        return pulled;
    }
}
