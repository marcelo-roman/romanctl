package com.example.sampleapp.application.orders;

import com.example.sampleapp.application.shared.EventPublisher;
import com.example.sampleapp.application.shared.ListLimit;
import com.example.sampleapp.application.shared.NotFoundException;
import com.example.sampleapp.domain.order.Money;
import com.example.sampleapp.domain.order.Order;
import java.time.Clock;
import java.util.List;
import java.util.UUID;
import java.util.function.Consumer;

public class OrderService {

    private final OrderRepository repository;
    private final EventPublisher events;
    private final Clock clock;

    public OrderService(OrderRepository repository, EventPublisher events, Clock clock) {
        this.repository = repository;
        this.events = events;
        this.clock = clock;
    }

    public UUID create(String customerName) {
        Order order = Order.create(customerName, clock.instant());
        repository.save(order);
        return order.id();
    }

    public void addItem(UUID id, AddItem command) {
        Money unitPrice = Money.of(command.unitPrice(), command.currency());
        change(id, order -> order.addItem(command.productName(), command.quantity(), unitPrice));
    }

    public void place(UUID id) {
        change(id, order -> order.place(clock.instant()));
    }

    public OrderView get(UUID id) {
        return OrderView.of(load(id));
    }

    public List<OrderView> list(Integer limit) {
        return repository.list(ListLimit.of(limit)).stream().map(OrderView::of).toList();
    }

    private void change(UUID id, Consumer<Order> apply) {
        Order order = load(id);
        apply.accept(order);
        repository.save(order);
        events.publish(order.pullEvents());
    }

    private Order load(UUID id) {
        return repository.findById(id).orElseThrow(() -> new NotFoundException("Pedido não encontrado."));
    }
}
