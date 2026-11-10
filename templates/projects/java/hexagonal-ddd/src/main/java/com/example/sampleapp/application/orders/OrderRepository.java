package com.example.sampleapp.application.orders;

import com.example.sampleapp.domain.order.Order;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

public interface OrderRepository {

    void save(Order order);

    Optional<Order> findById(UUID id);

    List<Order> list(int limit);
}
