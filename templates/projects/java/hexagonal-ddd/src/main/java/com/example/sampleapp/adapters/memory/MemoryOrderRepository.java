package com.example.sampleapp.adapters.memory;

import com.example.sampleapp.application.orders.OrderRepository;
import com.example.sampleapp.domain.order.Order;
import java.util.Comparator;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.stereotype.Repository;

@Repository
@ConditionalOnProperty(name = "app.persistence", havingValue = "memory", matchIfMissing = true)
public class MemoryOrderRepository implements OrderRepository {

    private final Map<UUID, Order> orders = new ConcurrentHashMap<>();

    @Override
    public void save(Order order) {
        orders.put(order.id(), copy(order));
    }

    @Override
    public Optional<Order> findById(UUID id) {
        return Optional.ofNullable(orders.get(id)).map(MemoryOrderRepository::copy);
    }

    @Override
    public List<Order> list(int limit) {
        return orders.values().stream()
                .sorted(Comparator.comparing(Order::createdAt).reversed())
                .limit(limit)
                .map(MemoryOrderRepository::copy)
                .toList();
    }

    private static Order copy(Order order) {
        return Order.restore(order.id(), order.customerName(), order.status(), order.createdAt(), order.items());
    }
}
